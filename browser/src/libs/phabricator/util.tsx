import { PlatformContext } from '../../../../shared/src/platform/context'
import { ChangeState, DifferentialState, DiffusionState, PhabricatorMode, RevisionState } from '.'
import { getRepoDetailsFromCallsign, getRepoDetailsFromRevisionID, QueryConduitHelper } from './backend'
import { map } from 'rxjs/operators'
import { Observable, throwError } from 'rxjs'

const TAG_PATTERN = /r([0-9A-z]+)([0-9a-f]{40})/
function matchPageTag(): RegExpExecArray | null {
    const el = document.getElementsByClassName('phui-tag-core').item(0)
    if (!el) {
        throw new Error('Could not find Phabricator page tag')
    }
    return TAG_PATTERN.exec(el.children[0].getAttribute('href') as string)
}

function getCallsignFromPageTag(): string {
    const match = matchPageTag()
    if (!match) {
        throw new Error('Could not determine callsign from page tag')
    }
    return match[1]
}

function getCommitIDFromPageTag(): string {
    const match = matchPageTag()
    if (!match) {
        throw new Error('Could not determine commitID from page tag')
    }
    return match[2]
}

function isDifferentialLanded(): boolean {
    const closedElement = document.getElementsByClassName('visual-only phui-icon-view phui-font-fa fa-check-square-o')
    return closedElement.length > 0
}

const DIFF_LINK = /D[0-9]+\?id=([0-9]+)/i
function getMaxDiffFromTabView(): { diffID: number; revDescription: string } | null {
    // first, find Revision contents table box
    const headerShells = document.getElementsByClassName('phui-header-header')
    let revisionContents: Element | null = null
    for (const headerShell of Array.from(headerShells)) {
        if (headerShell.textContent === 'Revision Contents') {
            revisionContents = headerShell
        }
    }
    if (!revisionContents) {
        return null
    }
    const parentContainer = revisionContents.parentElement!.parentElement!.parentElement!.parentElement!.parentElement!
    const tables = parentContainer.getElementsByClassName('aphront-table-view')
    for (const table of Array.from(tables)) {
        const tableRows = (table as HTMLTableElement).rows
        const row = tableRows[0]
        // looking for the history tab of the revision contents table
        if (row.children[0].textContent !== 'Diff') {
            continue
        }
        const links = table.getElementsByTagName('a')
        let max: { diffID: number; revDescription: string } | null = null
        for (const link of Array.from(links)) {
            const linkHref = link.getAttribute('href')
            if (!linkHref) {
                continue
            }
            const matches = DIFF_LINK.exec(linkHref)
            if (!matches) {
                continue
            }
            if (!link.parentNode!.parentNode!.childNodes[2].childNodes[0]) {
                continue
            }
            const revDescription = (link.parentNode!.parentNode!.childNodes[2].childNodes[0] as any).href
            const shaMatch = TAG_PATTERN.exec(revDescription)
            if (!shaMatch) {
                continue
            }
            max =
                max && max.diffID > parseInt(matches[1], 10)
                    ? max
                    : { diffID: parseInt(matches[1], 10), revDescription: shaMatch[2] }
        }
        return max
    }
    return null
}

const DIFF_PATTERN = /Diff ([0-9]+)/
function getDiffIdFromDifferentialPage(): number {
    const diffsContainer = document.getElementById('differential-review-stage')
    if (!diffsContainer) {
        throw new Error('no element with id differential-review-stage found on page.')
    }
    const wrappingDiffBox = diffsContainer.parentElement
    if (!wrappingDiffBox) {
        throw new Error('parent container of diff container not found.')
    }
    const diffTitle = wrappingDiffBox.children[0].getElementsByClassName('phui-header-header').item(0)
    if (!diffTitle || !diffTitle.textContent) {
        throw new Error('Could not find diffTitle element, or it had no text content')
    }
    const matches = DIFF_PATTERN.exec(diffTitle.textContent)
    if (!matches) {
        throw new Error(`diffTitle element does not match pattern. Content: '${diffTitle.textContent}'`)
    }
    return parseInt(matches[1], 10)
}

// https://phabricator.sgdev.org/source/gorilla/browse/master/mux.go
const PHAB_DIFFUSION_REGEX = /^\/?(source|diffusion)\/([A-Za-z0-9\-_]+)\/browse\/([\w-]+\/)?([^;$]+)(;[0-9a-f]{40})?(?:\$[0-9]+)?/i
// https://phabricator.sgdev.org/D2
const PHAB_DIFFERENTIAL_REGEX = /^\/?(D[0-9]+)(?:\?(?:(?:id=([0-9]+))|(vs=(?:[0-9]+|on)&id=[0-9]+)))?/i
// https://phabricator.sgdev.org/rMUXfb619131e25d82897c9de11789aa479941cfd415
const PHAB_REVISION_REGEX = /^\/?r([0-9A-z]+)([0-9a-f]{40})/i
// https://phabricator.sgdev.org/source/gorilla/change/master/mux.go
const PHAB_CHANGE_REGEX = /^\/?(source|diffusion)\/([A-Za-z0-9]+)\/change\/([\w-]+)\/([^;]+)(;[0-9a-f]{40})?/i
const PHAB_CHANGESET_REGEX = /^\/?\/differential\/changeset.*/i
const COMPARISON_REGEX = /^vs=((?:[0-9]+|on))&id=([0-9]+)/i

function getBaseCommitIDFromRevisionPage(): string {
    const keyElements = document.getElementsByClassName('phui-property-list-key')
    for (const keyElement of Array.from(keyElements)) {
        if (keyElement.textContent === 'Parents ') {
            const parentUrl = ((keyElement.nextSibling as HTMLElement).children[0].children[0] as HTMLLinkElement).href
            const url = new URL(parentUrl)
            const revisionMatch = PHAB_REVISION_REGEX.exec(url.pathname)
            if (revisionMatch) {
                return revisionMatch[2]
            }
        }
    }
    throw new Error('Could not determine base commit ID from revision page')
}

export function getPhabricatorState(
    loc: Location,
    requestGraphQL: PlatformContext['requestGraphQL'],
    queryConduit: QueryConduitHelper<any>
): Observable<DiffusionState | DifferentialState | RevisionState | ChangeState> {
    try {
        const stateUrl = loc.href.replace(loc.origin, '')
        const diffusionMatch = PHAB_DIFFUSION_REGEX.exec(stateUrl)
        if (diffusionMatch) {
            const filePath = diffusionMatch[4]
            if (!filePath) {
                throw new Error(`Could not determine file path from diffusionMatch, stateUrl: ${stateUrl}`)
            }
            const callsign = getCallsignFromPageTag()
            return getRepoDetailsFromCallsign(callsign, requestGraphQL, queryConduit).pipe(
                map(
                    ({ rawRepoName }): DiffusionState => ({
                        mode: PhabricatorMode.Diffusion,
                        rawRepoName,
                        filePath,
                        commitID: getCommitIDFromPageTag(),
                    })
                )
            )
        }
        const differentialMatch = PHAB_DIFFERENTIAL_REGEX.exec(stateUrl)
        if (differentialMatch) {
            const differentialID = differentialMatch[1]
            const comparison = differentialMatch[7]
            const revisionID = parseInt(differentialID.split('D')[1], 10)
            let diffID = differentialMatch[6] ? parseInt(differentialMatch[6], 10) : undefined
            if (!diffID) {
                diffID = getDiffIdFromDifferentialPage()
            }
            let baseRev = `phabricator/base/${diffID}`
            let headRev = `phabricator/diff/${diffID}`

            let baseDiffID: number | undefined

            const maxDiff = getMaxDiffFromTabView()
            const diffLanded = isDifferentialLanded()
            if (diffLanded && !maxDiff) {
                throw new Error(
                    'looking for the final diff id in the revision contents table failed. expected final row to have the commit in the description field.'
                )
            }
            if (comparison) {
                // urls that looks like this: http://phabricator.aws.sgdev.org/D3?vs=on&id=8&whitespace=ignore-most#toc
                // if the first parameter (vs=) is not 'on', not sure how to handle
                const comparisonMatch = COMPARISON_REGEX.exec(comparison)!
                const leftID = comparisonMatch[1]
                if (leftID !== 'on') {
                    baseDiffID = parseInt(leftID, 10)
                    baseRev = `phabricator/diff/${baseDiffID}`
                } else {
                    baseRev = `phabricator/base/${comparisonMatch[2]}`
                }
                headRev = `phabricator/diff/${comparisonMatch[2]}`
                if (diffLanded && maxDiff && comparisonMatch[2] === `${maxDiff.diffID}`) {
                    headRev = maxDiff.revDescription
                    baseRev = headRev.concat('~1')
                }
            }
            // check if the diff we are viewing is the max diff. if so,
            // right is the merged rev into master, and left is master~1
            else if (diffLanded && maxDiff && diffID === maxDiff.diffID) {
                headRev = maxDiff.revDescription
                baseRev = maxDiff.revDescription.concat('~1')
            }

            return getRepoDetailsFromRevisionID(revisionID, requestGraphQL, queryConduit).pipe(
                map(
                    ({ rawRepoName }): DifferentialState => ({
                        baseRawRepoName: rawRepoName,
                        baseRev,
                        headRawRepoName: rawRepoName,
                        headRev, // This will be blank on GitHub, but on a manually staged instance should exist
                        revisionID,
                        diffID: diffID!,
                        baseDiffID,
                        mode: PhabricatorMode.Differential,
                    })
                )
            )
        }

        const revisionMatch = PHAB_REVISION_REGEX.exec(stateUrl)
        if (revisionMatch) {
            const callsign = revisionMatch[1]
            const headCommitID = revisionMatch[2]
            const baseCommitID = getBaseCommitIDFromRevisionPage()
            return getRepoDetailsFromCallsign(callsign, requestGraphQL, queryConduit).pipe(
                map(
                    ({ rawRepoName }): RevisionState => ({
                        mode: PhabricatorMode.Revision,
                        rawRepoName,
                        baseCommitID,
                        headCommitID,
                    })
                )
            )
        }

        const changeMatch = PHAB_CHANGE_REGEX.exec(stateUrl)
        if (changeMatch) {
            const filePath = changeMatch[8]
            const callsign = getCallsignFromPageTag()
            const commitID = getCommitIDFromPageTag()
            return getRepoDetailsFromCallsign(callsign, requestGraphQL, queryConduit).pipe(
                map(
                    ({ rawRepoName }): ChangeState => ({
                        mode: PhabricatorMode.Change,
                        filePath,
                        rawRepoName,
                        commitID,
                    })
                )
            )
        }

        const changesetMatch = PHAB_CHANGESET_REGEX.exec(stateUrl)
        if (changesetMatch) {
            const crumbs = document.querySelector('.phui-crumbs-view')
            if (!crumbs) {
                throw new Error('failed parsing changeset dom')
            }

            const [, differentialHref, diffHref] = crumbs.querySelectorAll('a')

            const differentialMatch = differentialHref.getAttribute('href')!.match(/D(\d+)/)
            if (!differentialMatch) {
                throw new Error('failed parsing differentialID')
            }
            const revisionID = parseInt(differentialMatch[1], 10)

            const diffMatch = diffHref.getAttribute('href')!.match(/\/differential\/diff\/(\d+)/)
            if (!diffMatch) {
                throw new Error('failed parsing diffID')
            }
            const diffID = parseInt(diffMatch[1], 10)
            return getRepoDetailsFromRevisionID(revisionID, requestGraphQL, queryConduit).pipe(
                map(
                    ({ rawRepoName }): DifferentialState => {
                        let baseRev = `phabricator/base/${diffID}`
                        let headRev = `phabricator/diff/${diffID}`

                        const maxDiff = getMaxDiffFromTabView()
                        const diffLanded = isDifferentialLanded()
                        if (diffLanded && !maxDiff) {
                            throw new Error(
                                'Could not find the final diffID in the revision contents table. expected final row to have the commit in the description field.'
                            )
                        }

                        // check if the diff we are viewing is the max diff. if so,
                        // right is the merged rev into master, and left is master~1
                        if (diffLanded && maxDiff && diffID === maxDiff.diffID) {
                            headRev = maxDiff.revDescription
                            baseRev = maxDiff.revDescription.concat('~1')
                        }

                        return {
                            baseRawRepoName: rawRepoName,
                            baseRev,
                            headRawRepoName: rawRepoName,
                            headRev, // This will be blank on GitHub, but on a manually staged instance should exist
                            revisionID,
                            diffID,
                            mode: PhabricatorMode.Differential,
                        }
                    }
                )
            )
        }

        throw new Error(`Could not determine Phabricator state from stateUrl ${stateUrl}`)
    } catch (err) {
        return throwError(err)
    }
}

/**
 * This hacks javelin Stratcom to ignore command + click actions on sg-clickable tokens.
 * Without this, two windows open when a user command + clicks on a token.
 *
 * TODO could this be eliminated with shadow DOM?
 */
export function metaClickOverride(): void {
    const JX = (window as any).JX
    if (JX.Stratcom._dispatchProxyPreMeta) {
        return
    }
    JX.Stratcom._dispatchProxyPreMeta = JX.Stratcom._dispatchProxy
    JX.Stratcom._dispatchProxy = (proxyEvent: {
        __auto__type: string
        __auto__rawEvent: KeyboardEvent
        __auto__target: HTMLElement
    }) => {
        if (
            proxyEvent.__auto__type === 'click' &&
            proxyEvent.__auto__rawEvent.metaKey &&
            proxyEvent.__auto__target.classList.contains('sg-clickable')
        ) {
            return
        }
        return JX.Stratcom._dispatchProxyPreMeta(proxyEvent)
    }
}

export function normalizeRepoName(origin: string): string {
    let repoName = origin
    repoName = repoName.replace('\\', '')
    if (origin.startsWith('git@')) {
        repoName = origin.substr('git@'.length)
        repoName = repoName.replace(':', '/')
    } else if (origin.startsWith('git://')) {
        repoName = origin.substr('git://'.length)
    } else if (origin.startsWith('https://')) {
        repoName = origin.substr('https://'.length)
    } else if (origin.includes('@')) {
        // Assume the origin looks like `username@host:repo/path`
        const split = origin.split('@')
        repoName = split[1]
        repoName = repoName.replace(':', '/')
    }
    return repoName.replace(/.git$/, '')
}
