import { wrap } from 'async-middleware'
import bodyParser from 'body-parser'
import express from 'express'
import LRU from 'lru-cache'
import moveFile from 'move-file'
import { fs, child_process } from 'mz'
import * as path from 'path'
import { withFile } from 'tmp-promise'
import { Database, noopTransformer } from './database'
import { JsonDatabase } from './json'
import { BlobStore } from './blobStore';
import { GraphStore } from './graphStore';
import { MultiDatabase, NamedDatabase, instrumentPromise } from './multidb';

/**
 * Reads an integer from an environment variable or defaults to the given value.
 */
function readEnvInt({ key, defaultValue }: { key: string; defaultValue: number }): number {
    const value = process.env[key]
    if (!value) {
        return defaultValue
    }
    const n = parseInt(value, 10)
    if (isNaN(n)) {
        return defaultValue
    }
    return n
}

/**
 * Reads an boolean from an environment variable or defaults to the given value.
 */
function readEnvBoolean({ key, defaultValue }: { key: string; defaultValue: boolean }): boolean {
    const value = process.env[key]

    if (!value) {
        return defaultValue
    }

    try {
        return Boolean(JSON.parse(value))
    } catch (e) {
        return defaultValue
    }
}

/**
 * Where on the file system to store LSIF files.
 */
const STORAGE_ROOT = process.env.LSIF_STORAGE_ROOT || 'lsif-storage'

/**
 * Soft limit on the amount of storage used by LSIF files. Storage can exceed
 * this limit if a single LSIF file is larger than this, otherwise storage will
 * be kept under this limit. Defaults to 100GB.
 */
const SOFT_MAX_STORAGE = readEnvInt({ key: 'LSIF_SOFT_MAX_STORAGE', defaultValue: 100 * 1024 * 1024 * 1024 })

/**
 * Limit on the file size accepted by the /upload endpoint. Defaults to 100MB.
 */
const MAX_FILE_SIZE = readEnvInt({ key: 'LSIF_MAX_FILE_SIZE', defaultValue: 100 * 1024 * 1024 })

/**
 * Soft limit on the total amount of storage occupied by LSIF data loaded in
 * memory. The actual amount can exceed this if a single LSIF file is larger
 * than this limit, otherwise memory will be kept under this limit. Defaults to
 * 100MB.
 *
 * Empirically based on github.com/sourcegraph/codeintellify, each byte of
 * storage (uncompressed newline-delimited JSON) expands to 3 bytes in memory.
 */
const SOFT_MAX_STORAGE_IN_MEMORY = readEnvInt({
    key: 'LSIF_SOFT_MAX_STORAGE_IN_MEMORY',
    defaultValue: 100 * 1024 * 1024,
})

/**
 * Which port to run the LSIF server on. Defaults to 3186.
 */
const PORT = readEnvInt({ key: 'LSIF_HTTP_PORT', defaultValue: 3186 })

/**
 * Feature flag used to disable json-encoded file dumps.
 */
const ENABLE_JSON_DUMPS = readEnvBoolean({ key: 'LSIF_ENABLE_JSON', defaultValue: true })

/**
 * Feature flag used to disable graph-encoded sqlite dumps.
 */
const ENABLE_GRAPH_DUMPS = readEnvBoolean({ key: 'LSIF_ENABLE_GRAPH', defaultValue: true })

/**
 * Feature flag used to disable blob-encoded sqlite dumps.
 */
const ENABLE_BLOB_DUMPS = readEnvBoolean({ key: 'LSIF_ENABLE_BLOB', defaultValue: true })

/**
 * The extension used for json-encoded LSIF dumps.
 */
const JSON_EXTENSION = '.lsif'

/**
 * The file extension used for graph-encoded sqlite dumps.
 */
const GRAPH_EXTENSION = '.graph.db'

/**
 * The file extension used for blob-encoded sqlite dumps.
 */
const BLOB_EXTENSION = '.blob.db'

/**
 * A list of extensions supposed extensions for the on-disk LSIF dump.
 */
const EXTENSIONS = [JSON_EXTENSION, GRAPH_EXTENSION, BLOB_EXTENSION]

/**
 * The path of the binary used to convert json dumps to sqlite dumps.
 * See https://github.com/microsoft/lsif-node/tree/master/sqlite.
 */
const SQLITE_CONVERTER_BINARY = './node_modules/lsif-sqlite/bin/lsif-sqlite'

/**
 * An opaque repository ID.
 */
interface Repository {
    repository: string
}

/**
 * A 40-character commit hash.
 */
interface Commit {
    commit: string
}

/**
 * Combines `Repository` and `Commit`.
 */
interface RepositoryCommit extends Repository, Commit {}

/**
 * Deletes old files (sorted by last modified time) to keep the disk usage below
 * the given `max`.
 */
async function enforceMaxDiskUsage({
    flatDirectory,
    max,
    onBeforeDelete,
}: {
    flatDirectory: string
    max: number
    onBeforeDelete: (filePath: string) => void
}): Promise<void> {
    if (!(await fs.exists(flatDirectory))) {
        return
    }
    const files = await Promise.all(
        (await fs.readdir(flatDirectory)).map(async f => ({
            path: path.join(flatDirectory, f),
            stat: await fs.stat(path.join(flatDirectory, f)),
        }))
    )
    let totalSize = files.reduce((subtotal, f) => subtotal + f.stat.size, 0)
    for (const f of files.sort((a, b) => a.stat.atimeMs - b.stat.atimeMs)) {
        if (totalSize <= max) {
            break
        }
        onBeforeDelete(f.path)
        await fs.unlink(f.path)
        totalSize = totalSize - f.stat.size
    }
}

/**
 * Computes the bare filename (without extension) that contains LSIF data
 * encoded as either a json database or a sqlite blob store for the given
 * repository@commit. This is used everywhere as a hash key, as well as a
 * few places as an actual filename (when FS interaction is necessary).
 */
function hashKey({ repository, commit }: RepositoryCommit): string {
    const urlEncodedRepository = encodeURIComponent(repository)
    return path.join(STORAGE_ROOT, `${urlEncodedRepository}@${commit}`)
}

/**
 * Loads LSIF data from disk and returns a promise to the resulting `Database`.
 * Throws ENOENT when there is no LSIF data for the given repository@commit.
 */
async function createDB(key: string): Promise<Database> {
    const multidb = new MultiDatabase([
        new NamedDatabase('json', ENABLE_JSON_DUMPS, JSON_EXTENSION, new JsonDatabase()),
        new NamedDatabase('graph', ENABLE_GRAPH_DUMPS, GRAPH_EXTENSION, new GraphStore()),
        new NamedDatabase('blob', ENABLE_BLOB_DUMPS, BLOB_EXTENSION, new BlobStore())
    ])

    await multidb.load(key, (projectRootURI: string) => ({
        toDatabase: (pathRelativeToProjectRoot: string) => projectRootURI + '/' + pathRelativeToProjectRoot,
        fromDatabase: (uri: string) => (uri.startsWith(projectRootURI) ? uri.slice(`${projectRootURI}/`.length) : uri),
    }))

    return multidb
}

/**
 * List of supported `Database` methods.
 */
type SupportedMethods = 'hover' | 'definitions' | 'references'

const SUPPORTED_METHODS: Set<SupportedMethods> = new Set(['hover', 'definitions', 'references'])

/**
 * Type guard for SupportedMethods.
 */
function isSupportedMethod(method: string): method is SupportedMethods {
    return (SUPPORTED_METHODS as Set<string>).has(method)
}

/**
 * Throws an error with status 400 if the repository is invalid.
 */
function checkRepository(repository: any): void {
    if (typeof repository !== 'string') {
        throw Object.assign(new Error('Must specify the repository (usually of the form github.com/user/repo)'), {
            status: 400,
        })
    }
}

/**
 * Throws an error with status 400 if the commit is invalid.
 */
function checkCommit(commit: any): void {
    if (typeof commit !== 'string' || commit.length !== 40 || !/^[0-9a-f]+$/.test(commit)) {
        throw Object.assign(new Error('Must specify the commit as a 40 character hash ' + commit), { status: 400 })
    }
}

/**
 * A `Database`, the size of the LSIF file it was loaded from, and a callback to
 * dispose of it when evicted from the cache.
 */
interface LRUDBEntry {
    dbPromise: Promise<Database>
    /**
     * The size of the underlying LSIF file. This directly contributes to the
     * size of the cache. Ideally, this would be set to the amount of memory
     * that the `Database` uses, but calculating the memory usage is difficult
     * so this uses the file size as a rough heuristic.
     */
    length: number
    dispose: () => void
}

/**
 * An LRU cache mapping `repository@commit`s to in-memory `Database`s. Old
 * `Database`s are evicted from the cache to prevent OOM errors.
 */
const dbLRU = new LRU<string, LRUDBEntry>({
    max: SOFT_MAX_STORAGE_IN_MEMORY,
    // `length` contributes to the total size of the cache, with a `max` specified
    // above, after which old items get evicted.
    length: (entry, key) => entry.length,
    dispose: (key, entry) => entry.dispose(),
})

/**
 * Get the size of the given filename if stat succeeds. If stat fails,
 * return zero as the file is likely just not on disk.
 */
async function getSize(filename: string): Promise<number> {
    try {
        return (await fs.stat(filename)).size
    } catch {
        return 0
    }
}

/**
 * Runs the given `action` with the `Database` associated with the given
 * repository@commit. Internally, it either gets the `Database` from the LRU
 * cache or loads it from storage.
 */
async function withDB(repositoryCommit: RepositoryCommit, action: (db: Database) => Promise<void>): Promise<void> {
    const key = hashKey(repositoryCommit)
    const entry = dbLRU.get(key)
    if (entry) {
        await action(await entry.dbPromise)
    } else {
        // Get the combined length of all files on disk related to this key.
        const lengths = (await Promise.all(EXTENSIONS.map(ext => getSize(key + ext))))
        const length = lengths.reduce((a, b) => a + b, 0)

        const dbPromise = createDB(key)
        dbLRU.set(key, {
            dbPromise,
            length,
            dispose: () => dbPromise.then(db => db.close()),
        })
        await action(await dbPromise)
    }
}

/**
 * Runs the HTTP server which accepts LSIF file uploads and responds to
 * hover/defs/refs requests.
 */
function main(): void {
    const app = express()

    app.use((err: any, req: express.Request, res: express.Response, next: express.NextFunction) => {
        if (err && err.status) {
            res.status(err.status).send({ message: err.message })
            return
        }
        res.status(500).send({ message: 'Unknown error' })
        console.error(err)
    })

    app.get('/ping', (req, res) => {
        res.send({ pong: 'pong' })
    })

    app.post(
        '/request',
        bodyParser.json({ limit: '1mb' }),
        wrap(async (req, res) => {
            const { repository, commit } = req.query
            const { path, position, method } = req.body

            checkRepository(repository)
            checkCommit(commit)
            if (!isSupportedMethod(method)) {
                throw Object.assign(new Error('Method must be one of ' + SUPPORTED_METHODS), { status: 422 })
            }

            try {
                await withDB({ repository, commit }, async db => {
                    let result: any
                    switch (method) {
                        case 'hover':
                            result = db.hover(path, position)
                            break
                        case 'definitions':
                            result = db.definitions(path, position)
                            break
                        case 'references':
                            result = db.references(path, position, { includeDeclaration: false })
                            break
                        default:
                            throw new Error(`Unknown method ${method}`)
                    }
                    res.json(result || null)
                })
            } catch (e) {
                if ('code' in e && e.code === 'ENOENT') {
                    throw Object.assign(new Error(`No LSIF data available for ${repository}@${commit}.`), {
                        status: 404,
                    })
                }
                throw e
            }
        })
    )

    app.post(
        '/exists',
        wrap(async (req, res) => {
            const { repository, commit, file } = req.query

            checkRepository(repository)
            checkCommit(commit)
            const key = hashKey({ repository, commit })

            // The database does not exist if there is no file to create the database
            if (!(await Promise.all(EXTENSIONS.map(ext => fs.exists(key + ext)))).some(v => v)) {
                res.send(false)
                return
            }

            if (!file) {
                res.send(true)
                return
            }

            if (typeof file !== 'string') {
                throw Object.assign(new Error('File must be a string'), { status: 400 })
            }

            try {
                res.send(Boolean((await createDB(key)).stat(file)))
            } catch (e) {
                if ('code' in e && e.code === 'ENOENT') {
                    res.send(false)
                    return
                }
                throw e
            }
        })
    )

    app.post(
        '/upload',
        wrap(async (req, res) => {
            const { repository, commit } = req.query

            checkRepository(repository)
            checkCommit(commit)

            if (req.header('Content-Length') && parseInt(req.header('Content-Length') || '', 10) > MAX_FILE_SIZE) {
                throw Object.assign(
                    new Error(
                        `The size of the given LSIF file (${req.header(
                            'Content-Length'
                        )} bytes) exceeds the max of ${MAX_FILE_SIZE}`
                    ),
                    { status: 413 }
                )
            }

            let contentLength = 0

            await withFile(async tempFile => {
                // Pipe the given LSIF data to a temp file.
                await new Promise((resolve, reject) => {
                    const tempFileWriteStream = fs.createWriteStream(tempFile.path)
                    req.on('data', chunk => {
                        contentLength += chunk.length
                        if (contentLength > MAX_FILE_SIZE) {
                            tempFileWriteStream.destroy()
                            reject(
                                Object.assign(
                                    new Error(
                                        `The size of the given LSIF file (${contentLength} bytes so far) exceeds the max of ${MAX_FILE_SIZE}`
                                    ),
                                    { status: 413 }
                                )
                            )
                        }
                    }).pipe(tempFileWriteStream)

                    tempFileWriteStream.on('close', resolve)
                    tempFileWriteStream.on('error', reject)
                })

                console.log(`Base file is ${contentLength} bytes`)

                // Load a `Database` from the file to check that it's valid.
                await new JsonDatabase().load(tempFile.path, () => noopTransformer)

                try {
                    await fs.mkdir(STORAGE_ROOT)
                } catch (e) {
                    if (e.code !== 'EEXIST') {
                        throw e
                    }
                }

                const key = hashKey({ repository, commit })

                if (ENABLE_GRAPH_DUMPS) {
                    console.log('Creating graph-encoded sqlite database')
                    const command = `${SQLITE_CONVERTER_BINARY} --in "${tempFile.path}" --out "${key + GRAPH_EXTENSION}"`
                    await instrumentPromise('graph: convert', child_process.exec(command))
                    const size = await getSize(key + GRAPH_EXTENSION)
                    console.log(`Sqlite file is ${size} bytes`)
                }

                if (ENABLE_BLOB_DUMPS) {
                    console.log('Creating blob-encoded sqlite database')
                    // TODO - set version somehow
                    const command = `${SQLITE_CONVERTER_BINARY} --in "${tempFile.path}" --out "${key + BLOB_EXTENSION}" --format blob --delete --projectVersion 0.1.0`
                    await instrumentPromise('blob: convert', child_process.exec(command))
                    const size = await getSize(key + BLOB_EXTENSION)
                    console.log(`Sqlite file is ${size} bytes`)
                }

                if (ENABLE_JSON_DUMPS) {
                    console.log('Moving json-encoded dump to cache path')
                    await moveFile(tempFile.path, key + JSON_EXTENSION)
                } else {
                    console.log('Removing json-encoded dump')

                    // We only needed the file for input to the sqlite converter
                    // We can remove this now to save on space in the temp dir
                    await fs.unlink(tempFile.path)
                }

                // Evict the old `Database` from the LRU cache to cause it to pick up the new LSIF data from disk.
                dbLRU.del(key)

                res.send('Upload successful.')
            })

            // TODO enforce max disk usage per-repository. Currently, a
            // misbehaving client could upload a bunch of LSIF files for one
            // repository and take up all of the disk space, causing all other
            // LSIF files to get deleted to make room for the new files.
            await enforceMaxDiskUsage({
                flatDirectory: STORAGE_ROOT,
                max: Math.max(0, SOFT_MAX_STORAGE - contentLength),
                onBeforeDelete: filePath =>
                    console.log(`Deleting ${filePath} to help keep disk usage under ${SOFT_MAX_STORAGE}.`),
            })
        })
    )

    app.listen(PORT, () => {
        console.log(`Listening for HTTP requests on port ${PORT}`)
    })
}

main()
