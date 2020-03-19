import { readEnvInt } from '../shared/settings'

/** Which port to run the LSIF server on. Defaults to 3186. */
export const HTTP_PORT = readEnvInt('HTTP_PORT', 3186)

/** HTTP address for internal LSIF dump manager server. */
export const LSIF_DUMP_MANAGER_URL = process.env.LSIF_DUMP_MANAGER_URL || 'http://lsif-dump-manager'

/** Where on the file system to store LSIF files. */
export const STORAGE_ROOT = process.env.LSIF_STORAGE_ROOT || 'lsif-storage'

/** The interval (in seconds) to invoke the cleanOldUploads task. */
export const CLEAN_OLD_UPLOADS_INTERVAL = readEnvInt('CLEAN_OLD_UPLOADS_INTERVAL', 60 * 60 * 8)

/** The interval (in seconds) to clean the dbs directory. */
export const PURGE_OLD_DUMPS_INTERVAL = readEnvInt('PURGE_OLD_DUMPS_INTERVAL', 60 * 30)

/** How many uploads to query at once when determining if a db file is unreferenced. */
export const DEAD_DUMP_CHUNK_SIZE = readEnvInt('DEAD_DUMP_CHUNK_SIZE', 100)

/** The default number of location results to return when performing a find-references operation. */
export const DEFAULT_REFERENCES_PAGE_SIZE = readEnvInt('DEFAULT_REFERENCES_PAGE_SIZE', 100)

/** The interval (in seconds) to invoke the cleanFailedUploads task. */
export const CLEAN_FAILED_UPLOADS_INTERVAL = readEnvInt('CLEAN_FAILED_UPLOADS_INTERVAL', 60 * 60 * 8)

/** The interval (in seconds) to invoke the updateQueueSizeGaugeInterval task. */
export const UPDATE_QUEUE_SIZE_GAUGE_INTERVAL = readEnvInt('UPDATE_QUEUE_SIZE_GAUGE_INTERVAL', 5)

/** The interval (in seconds) to run the resetStalledUploads task. */
export const RESET_STALLED_UPLOADS_INTERVAL = readEnvInt('RESET_STALLED_UPLOADS_INTERVAL', 60)

/** The default page size for the upload endpoints. */
export const DEFAULT_UPLOAD_PAGE_SIZE = readEnvInt('DEFAULT_UPLOAD_PAGE_SIZE', 50)

/** The default page size for the dumps endpoint. */
export const DEFAULT_DUMP_PAGE_SIZE = readEnvInt('DEFAULT_DUMP_PAGE_SIZE', 50)

/** The maximum age (in seconds) that an upload (completed or queued) will remain in Postgres. */
export const UPLOAD_MAX_AGE = readEnvInt('UPLOAD_UPLOAD_AGE', 60 * 60 * 24 * 7)

/** The maximum age (in seconds) that the files for an unprocessed upload can remain on disk. */
export const FAILED_UPLOAD_MAX_AGE = readEnvInt('FAILED_UPLOAD_MAX_AGE', 24 * 60 * 60)

/** The maximum age (in seconds) that the an upload can be unlocked and in the `processing` state. */
export const STALLED_UPLOAD_MAX_AGE = readEnvInt('STALLED_UPLOAD_MAX_AGE', 5)

/** The maximum space (in bytes) that the dbs directory can use. */
export const DBS_DIR_MAXIMUM_SIZE_BYTES = readEnvInt('DBS_DIR_MAXIMUM_SIZE_BYTES', 1024 * 1024 * 1024 * 10)
