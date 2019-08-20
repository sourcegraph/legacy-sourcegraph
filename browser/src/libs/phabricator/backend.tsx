import { from, Observable, of, throwError } from 'rxjs'
import { map, mapTo, switchMap, catchError, tap } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../shared/src/graphql/schema'
import { PlatformContext } from '../../../../shared/src/platform/context'
import { memoizeObservable } from '../../../../shared/src/util/memoizeObservable'
import { storage } from '../../browser/storage'
import { isExtension } from '../../context'
import { resolveRepo } from '../../shared/repo/backend'
import { normalizeRepoName } from './util'
import { ajax } from 'rxjs/ajax'
import { EREPONOTFOUND } from '../../../../shared/src/backend/errors'
import { RepoSpec, FileSpec, ResolvedRevSpec } from '../../../../shared/src/util/url'
import { RevisionSpec, DiffSpec, BaseDiffSpec } from '.'

interface PhabEntity {
    id: string // e.g. "48"
    type: string // e.g. "RHURI"
    phid: string // e.g. "PHID-RHURI-..."
}

interface ConduitURI extends PhabEntity {
    fields: {
        uri: {
            raw: string // e.g. https://secure.phabricator.com/source/phabricator.git",
            display: string // e.g. https://secure.phabricator.com/source/phabricator.git",
            effective: string // e.g. https://secure.phabricator.com/source/phabricator.git",
            normalized: string // e.g. secure.phabricator.com/source/phabricator",
        }
    }
}

interface ConduitRepo extends PhabEntity {
    fields: {
        name: string
        vcs: string // e.g. 'git'
        callsign: string
        shortName: string
        status: 'active' | 'inactive'
    }
    attachments: {
        uris: {
            uris: ConduitURI[]
        }
    }
}

export interface ConduitReposResponse {
    data: ConduitRepo[]
}

interface ConduitRef {
    ref: string
    type: 'base' | 'diff'
    commit: string // a SHA
    remote: {
        uri: string
    }
}

interface ConduitDiffChange {
    oldPath: string
    currentPath: string
}

interface ConduitDiffDetails {
    branch: string
    sourceControlBaseRevision: string // the merge base commit
    description: string // e.g. 'rNZAP9bee3bc2cd3068dd97dfa87068c4431c5d6093ef'
    changes: ConduitDiffChange[]
    dateCreated: string
    authorName: string
    authorEmail: string
    properties: {
        'arc.staging': {
            status: string
            refs: ConduitRef[]
        }
        'local:commits': string[]
    }
}

interface ConduitDiffDetailsResponse {
    [id: string]: ConduitDiffDetails
}

function createConduitRequestForm(): FormData {
    const searchForm = document.querySelector('.phabricator-search-menu form')
    if (!searchForm) {
        throw new Error('cannot create conduit request form')
    }
    const form = new FormData()
    form.set('__csrf__', searchForm.querySelector<HTMLInputElement>('input[name=__csrf__]')!.value)
    form.set('__form__', searchForm.querySelector<HTMLInputElement>('input[name=__form__]')!.value)
    return form
}

/**
 * Native installation of the Phabricator extension does not allow for us to fetch the style.bundle from a script element.
 * To get around this we fetch the bundled CSS contents and append it to the DOM.
 */
export async function getPhabricatorCSS(sourcegraphURL: string): Promise<string> {
    const bundleUID = process.env.BUNDLE_UID
    const resp = await fetch(sourcegraphURL + `/.assets/extension/css/style.bundle.css?v=${bundleUID}`, {
        method: 'GET',
        credentials: 'include',
        headers: new Headers({ Accept: 'text/html' }),
    })
    return resp.text()
}

type ConduitResponse<T> =
    | { error_code: null; error_info: null; result: T }
    | { error_code: string; error_info: string; result: null }

export type QueryConduitHelper<T> = (endpoint: string, params: {}) => Observable<T>

export function queryConduit<T>(endpoint: string, params: {}): Observable<T> {
    const form = createConduitRequestForm()
    for (const [key, value] of Object.entries(params)) {
        form.set(`params[${key}]`, JSON.stringify(value))
    }
    return ajax({
        url: window.location.origin + endpoint,
        method: 'POST',
        body: form,
        withCredentials: true,
        headers: {
            Accept: 'application/json',
        },
    }).pipe(
        map(({ response }: { response: ConduitResponse<T> }) => {
            if (response.error_code !== null) {
                throw new Error(`error ${response.error_code}: ${response.error_info}`)
            }
            return response.result
        })
    )
}

function getDiffDetailsFromConduit(
    diffID: number,
    differentialID: number,
    queryConduit: QueryConduitHelper<ConduitDiffDetailsResponse>
): Observable<ConduitDiffDetails> {
    return queryConduit('/api/differential.querydiffs', {
        ids: [diffID],
        revisionIDs: [differentialID],
    }).pipe(map(diffDetails => diffDetails[`${diffID}`]))
}

function getRawDiffFromConduit(diffID: number, queryConduit: QueryConduitHelper<string>): Observable<string> {
    return queryConduit('/api/differential.getrawdiff', { diffID })
}

interface ConduitDifferentialQueryResponse {
    [index: string]: {
        repositoryPHID: string | null
    }
}

function getRepoPHIDForRevisionID(
    revisionID: number,
    queryConduit: QueryConduitHelper<ConduitDifferentialQueryResponse>
): Observable<string> {
    return queryConduit('/api/differential.query', { ids: [revisionID] }).pipe(
        map(result => {
            const phid = result['0'].repositoryPHID
            if (!phid) {
                // This happens for revisions that were created without an associated repository
                throw new Error(`no repositoryPHID for revision ${revisionID}`)
            }
            return phid
        })
    )
}

interface CreatePhabricatorRepoOptions extends Pick<PlatformContext, 'requestGraphQL'> {
    callsign: string
    repoName: string
    phabricatorURL: string
}

const createPhabricatorRepo = memoizeObservable(
    ({ requestGraphQL, ...variables }: CreatePhabricatorRepoOptions): Observable<void> =>
        requestGraphQL<GQL.IMutation>({
            request: gql`
                mutation addPhabricatorRepo($callsign: String!, $repoName: String!, $phabricatorURL: String!) {
                    addPhabricatorRepo(callsign: $callsign, uri: $repoName, url: $phabricatorURL) {
                        alwaysNil
                    }
                }
            `,
            variables,
            mightContainPrivateInfo: true,
        }).pipe(mapTo(undefined)),
    ({ callsign }) => callsign
)

interface PhabricatorRepoDetails {
    callsign: string
    rawRepoName: string
}

export function getRepoDetailsFromCallsign(
    callsign: string,
    requestGraphQL: PlatformContext['requestGraphQL'],
    queryConduit: QueryConduitHelper<ConduitReposResponse>
): Observable<PhabricatorRepoDetails> {
    return queryConduit('/api/diffusion.repository.search', {
        constraints: { callsigns: [callsign] },
        attachments: { uris: true },
    }).pipe(
        switchMap(({ data }) => {
            const repo = data[0]
            if (!repo) {
                throw new Error(`could not locate repo with callsign ${callsign}`)
            }
            if (!repo.attachments || !repo.attachments.uris) {
                throw new Error(`could not locate git uri for repo with callsign ${callsign}`)
            }
            return convertConduitRepoToRepoDetails(repo)
        }),
        switchMap((details: PhabricatorRepoDetails | null) => {
            if (!details) {
                return throwError(new Error('could not parse repo details'))
            }
            return createPhabricatorRepo({
                callsign,
                repoName: details.rawRepoName,
                phabricatorURL: window.location.origin,
                requestGraphQL,
            }).pipe(mapTo(details)) as Observable<PhabricatorRepoDetails>
        })
    )
}

/**
 * Queries the sourcegraph.configuration conduit API endpoint.
 *
 * The Phabricator extension updates the window object automatically, but in the
 * case it fails we query the conduit API.
 */
export function getSourcegraphURLFromConduit(): Observable<string> {
    return queryConduit<{ url: string }>('/api/sourcegraph.configuration', {}).pipe(map(({ url }) => url))
}

function getRepoDetailsFromRepoPHID(
    phid: string,
    requestGraphQL: PlatformContext['requestGraphQL'],
    queryConduit: QueryConduitHelper<ConduitReposResponse>
): Observable<PhabricatorRepoDetails> {
    const form = createConduitRequestForm()
    form.set('params[constraints]', JSON.stringify({ phids: [phid] }))
    form.set('params[attachments]', '{ "uris": true }')

    return queryConduit('/api/diffusion.repository.search', {
        constraints: {
            phids: [phid],
        },
        attachments: {
            uris: true,
        },
    }).pipe(
        switchMap(({ data }) => {
            const repo = data[0]
            if (!repo) {
                throw new Error(`could not locate repo with phid ${phid}`)
            }
            if (!repo.attachments || !repo.attachments.uris) {
                throw new Error(`could not locate git uri for repo with phid ${phid}`)
            }
            return from(convertConduitRepoToRepoDetails(repo)).pipe(
                switchMap((details: PhabricatorRepoDetails | null) => {
                    if (!details) {
                        return throwError(new Error('could not parse repo details'))
                    }
                    if (!repo.fields || !repo.fields.callsign) {
                        return throwError(new Error('callsign not found'))
                    }
                    return createPhabricatorRepo({
                        callsign: repo.fields.callsign,
                        repoName: details.rawRepoName,
                        phabricatorURL: window.location.origin,
                        requestGraphQL,
                    }).pipe(mapTo(details))
                })
            ) as Observable<PhabricatorRepoDetails>
        })
    )
}

export function getRepoDetailsFromRevisionID(
    revisionID: number,
    requestGraphQL: PlatformContext['requestGraphQL'],
    queryConduit: QueryConduitHelper<any>
): Observable<PhabricatorRepoDetails> {
    return getRepoPHIDForRevisionID(revisionID, queryConduit).pipe(
        switchMap(repositoryPHID => getRepoDetailsFromRepoPHID(repositoryPHID, requestGraphQL, queryConduit))
    )
}

async function convertConduitRepoToRepoDetails(repo: ConduitRepo): Promise<PhabricatorRepoDetails | null> {
    if (isExtension) {
        const items = await storage.sync.get()
        if (items.phabricatorMappings) {
            for (const mapping of items.phabricatorMappings) {
                if (mapping.callsign === repo.fields.callsign) {
                    return {
                        callsign: repo.fields.callsign,
                        rawRepoName: mapping.path,
                    }
                }
            }
        }
        return convertToDetails(repo)
    }
    // The path to a phabricator repository on a Sourcegraph instance may differ than it's URI / name from the
    // phabricator conduit API. Since we do not currently send the PHID with the Phabricator repository this a
    // backwards work around configuration setting to ensure mappings are correct. This logic currently exists
    // in the browser extension options menu.
    type Mappings = { callsign: string; path: string }[]
    const mappingsString = window.localStorage.getItem('PHABRICATOR_CALLSIGN_MAPPINGS')
    const callsignMappings = mappingsString
        ? (JSON.parse(mappingsString) as Mappings)
        : window.PHABRICATOR_CALLSIGN_MAPPINGS || []
    const details = convertToDetails(repo)
    if (callsignMappings) {
        for (const mapping of callsignMappings) {
            if (mapping.callsign === repo.fields.callsign) {
                return {
                    callsign: repo.fields.callsign,
                    rawRepoName: mapping.path,
                }
            }
        }
    }
    return details
}

function convertToDetails(repo: ConduitRepo): PhabricatorRepoDetails | null {
    let uri: ConduitURI | undefined
    for (const u of repo.attachments.uris.uris) {
        const normalPath = u.fields.uri.normalized.replace('\\', '')
        if (normalPath.startsWith(window.location.host + '/')) {
            continue
        }
        uri = u
        break
    }
    if (!uri) {
        return null
    }
    const rawURI = uri.fields.uri.raw
    const rawRepoName = normalizeRepoName(rawURI)
    return { callsign: repo.fields.callsign, rawRepoName }
}

interface ResolveStagingOptions extends Pick<PlatformContext, 'requestGraphQL'>, RepoSpec, DiffSpec {
    baseRev: string
    patch?: string
    date?: string
    authorName?: string
    authorEmail?: string
    description?: string
}

const resolveStagingRev = memoizeObservable(
    ({ requestGraphQL, ...variables }: ResolveStagingOptions): Observable<ResolvedRevSpec> =>
        requestGraphQL<GQL.IMutation>({
            request: gql`
                mutation ResolveStagingRev(
                    $repoName: String!
                    $diffID: ID!
                    $baseRev: String!
                    $patch: String
                    $date: String
                    $authorName: String
                    $authorEmail: String
                    $description: String
                ) {
                    resolvePhabricatorDiff(
                        repoName: $repoName
                        diffID: $diffID
                        baseRev: $baseRev
                        patch: $patch
                        date: $date
                        authorName: $authorName
                        authorEmail: $authorEmail
                        description: $description
                    ) {
                        oid
                    }
                }
            `,
            variables,
            mightContainPrivateInfo: true,
        }).pipe(
            map(dataOrThrowErrors),
            map(({ resolvePhabricatorDiff }) => {
                if (!resolvePhabricatorDiff) {
                    throw new Error('Empty resolvePhabricatorDiff')
                }
                const { oid } = resolvePhabricatorDiff
                if (!oid) {
                    throw new Error('Could not resolve staging rev: empty oid')
                }
                return { commitID: oid }
            })
        ),
    ({ diffID }: ResolveStagingOptions) => diffID.toString()
)

function hasThisFileChanged(filePath: string, changes: ConduitDiffChange[]): boolean {
    for (const change of changes) {
        if (change.currentPath === filePath) {
            return true
        }
    }
    return false
}

interface ResolveDiffOpt extends RepoSpec, FileSpec, RevisionSpec, DiffSpec, BaseDiffSpec {
    isBase: boolean
    useDiffForBase: boolean // indicates whether the base should use the diff commit
    useBaseForDiff: boolean // indicates whether the diff should use the base commit
}

interface PropsWithDiffDetails extends ResolveDiffOpt {
    diffDetails: ConduitDiffDetails
}

function getPropsWithDiffDetails(
    props: ResolveDiffOpt,
    queryConduit: QueryConduitHelper<any>
): Observable<PropsWithDiffDetails> {
    return getDiffDetailsFromConduit(props.diffID, props.revisionID, queryConduit).pipe(
        switchMap(diffDetails => {
            if (props.isBase || !props.baseDiffID || hasThisFileChanged(props.filePath, diffDetails.changes)) {
                // no need to update props
                return of({
                    ...props,
                    diffDetails,
                })
            }
            return getDiffDetailsFromConduit(props.baseDiffID, props.revisionID, queryConduit).pipe(
                map(
                    (diffDetails): PropsWithDiffDetails => ({
                        ...props,
                        diffDetails,
                        diffID: props.baseDiffID!,
                        useBaseForDiff: true,
                    })
                )
            )
        })
    )
}

function getStagingDetails(
    propsWithInfo: PropsWithDiffDetails
): { repoName: string; ref: ConduitRef; unconfigured: boolean } | undefined {
    const stagingInfo = propsWithInfo.diffDetails.properties['arc.staging']
    if (!stagingInfo) {
        return undefined
    }
    let key: string
    if (propsWithInfo.isBase) {
        const type = propsWithInfo.useDiffForBase ? 'diff' : 'base'
        key = `refs/tags/phabricator/${type}/${propsWithInfo.diffID}`
    } else {
        const type = propsWithInfo.useBaseForDiff ? 'base' : 'diff'
        key = `refs/tags/phabricator/${type}/${propsWithInfo.diffID}`
    }
    for (const ref of propsWithInfo.diffDetails.properties['arc.staging'].refs) {
        if (ref.ref === key) {
            const remote = ref.remote.uri
            if (remote) {
                return {
                    repoName: normalizeRepoName(remote),
                    ref,
                    unconfigured: stagingInfo.status === 'repository.unconfigured',
                }
            }
        }
    }
    return undefined
}

interface ResolvedDiff extends ResolvedRevSpec {
    /**
     * Whether this commit lives on a staging repository.
     * See https://secure.phabricator.com/book/phabricator/article/harbormaster/#change-handoff
     */
    isStagingCommit?: boolean
    /**
     * The name of the staging repository, if it is synced to the Sourcegraph instance.
     */
    stagingRepoName?: string
}

export function resolveDiffRev(
    props: ResolveDiffOpt,
    requestGraphQL: PlatformContext['requestGraphQL'],
    queryConduit: QueryConduitHelper<any>
): Observable<ResolvedDiff> {
    return getPropsWithDiffDetails(props, queryConduit).pipe(
        switchMap(({ diffDetails, ...props }) => {
            const stagingDetails = getStagingDetails({ diffDetails, ...props })
            const conduitProps = {
                repoName: props.repoName,
                diffID: props.diffID,
                baseRev: diffDetails.sourceControlBaseRevision,
                date: diffDetails.dateCreated,
                authorName: diffDetails.authorName,
                authorEmail: diffDetails.authorEmail,
                description: diffDetails.description,
            }

            // When resolving the base, use the commit ID from the diff details.
            if (props.isBase && !props.useDiffForBase) {
                return of({ commitID: diffDetails.sourceControlBaseRevision })
            }
            if (!stagingDetails || stagingDetails.unconfigured) {
                // If there are no staging details, get the patch from the conduit API,
                // create a one-off commit on the Sourcegraph instance from the patch,
                // and resolve to the commit ID returned by the Sourcegraph instance.
                return getRawDiffFromConduit(props.diffID, queryConduit).pipe(
                    switchMap(patch => resolveStagingRev({ ...conduitProps, patch, requestGraphQL }))
                )
            }

            // If staging details are configured, first check if the repo is present on the Sourcegraph instance.
            return resolveRepo({ rawRepoName: stagingDetails.repoName, requestGraphQL }).pipe(
                // If the repo is present on the Sourcegraph instance,
                // use the commitID and repo name from the staging details.
                mapTo({
                    commitID: stagingDetails.ref.commit,
                    stagingRepoName: stagingDetails.repoName,
                }),
                // Otherwise, create a one-off commit containing the patch on the Sourcegraph instance,
                // and resolve to the commit ID returned by the Sourcegraph instance.
                catchError(error => {
                    if (error.code !== EREPONOTFOUND) {
                        throw error
                    }
                    return getRawDiffFromConduit(props.diffID, queryConduit).pipe(
                        switchMap(patch => resolveStagingRev({ ...conduitProps, patch, requestGraphQL }))
                    )
                })
            )
        })
    )
}
