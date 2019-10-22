import { isEqual } from 'lodash'
import { useEffect, useMemo, useState } from 'react'
import { combineLatest, merge, Observable, of, Subject } from 'rxjs'
import {
    catchError,
    debounceTime,
    distinctUntilChanged,
    map,
    mapTo,
    share,
    switchMap,
    tap,
    throttleTime,
} from 'rxjs/operators'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import { dataOrThrowErrors, gql } from '../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { asError, ErrorLike, isErrorLike } from '../../../../../shared/src/util/errors'
import { ActorFragment, ActorQuery } from '../../../actor/graphql'
import { queryGraphQL } from '../../../backend/graphql'
import {
    diffStatFieldsFragment,
    fileDiffHunkRangeFieldsFragment,
} from '../../../repo/compare/RepositoryCompareDiffPage'
import { ThreadConnectionFiltersFragment } from '../../threads/list/useThreads'
import { ExtensionDataStatus, getCampaignExtensionData } from '../extensionData'
import { parseJSON } from '../../../settings/configuration'
import { Workflow } from '../../../schema/workflow.schema'

export const RepositoryComparisonQuery = gql`
baseRepository {
    id
    name
    url
}
headRepository {
    id
    name
    url
}
range {
    expr
    baseRevSpec {
        object {
            oid
        }
        expr
    }
    headRevSpec {
        expr
    }
}
fileDiffs {
    nodes {
        oldPath
        newPath
        hunks {
            oldRange {
                ...FileDiffHunkRangeFields
            }
            oldNoNewlineAt
            newRange {
                ...FileDiffHunkRangeFields
            }
            section
            body
        }
        stat {
            ...DiffStatFields
        }
        internalID
    }
    totalCount
    pageInfo {
        hasNextPage
    }
    diffStat {
        ...DiffStatFields
    }
}`

export const ThreadPreviewFragment = gql`
fragment ThreadPreviewFragment on ThreadPreview {
    author {
        ${ActorQuery}
    }
    repository {
        id
        name
        url
    }
    title
    bodyHTML
    isDraft
    isPendingExternalCreation
    kind
    assignees {
        nodes {
            ${ActorQuery}
        }
    }
    internalID
}`

export const CampaignPreviewFragment = gql`
    fragment CampaignPreviewFragment on ExpCampaignPreview {
        name
        threads {
            nodes {
                __typename
                ... on ThreadPreview {
                    ...ThreadPreviewFragment
                    repositoryComparison {
                        ${RepositoryComparisonQuery}
                    }
                }
            }
            totalCount
            filters {
                ...ThreadConnectionFiltersFragment
            }
        }
        participants {
            edges {
                actor {
                    ${ActorQuery}
                }
                reasons
            }
            totalCount
        }
        diagnostics {
            nodes {
                type
                data
            }
            totalCount
        }
        repositories {
            id
        }
        repositoryComparisons {
            ${RepositoryComparisonQuery}
        }
        sideEffects {
            nodes {
                title
                detail
            }
            totalCount
        }
        logMessages {
            nodes {
                body
            }
            totalCount
        }
    }
    ${ThreadPreviewFragment}
    ${fileDiffHunkRangeFieldsFragment}
    ${diffStatFieldsFragment}
`

const LOADING: 'loading' = 'loading'

export type CampaignPreviewResult = typeof LOADING | GQL.IExpCampaignPreview | ErrorLike

type CreateCampaignInputWithoutExtensionData = Pick<
    GQL.IExpCreateCampaignInput,
    Exclude<keyof GQL.IExpCreateCampaignInput, 'extensionData'>
>

const queryCampaignPreview = ({
    extensionsController,
    input,
}: ExtensionsControllerProps & {
    input: Pick<GQL.IExpCreateCampaignInput, Exclude<keyof GQL.IExpCreateCampaignInput, 'extensionData'>>
}): Observable<readonly [GQL.IExpCampaignPreview | ErrorLike, GQL.IExpCampaignExtensionData, ExtensionDataStatus]> => {
    const workflow: Workflow = parseJSON(input.workflowAsJSONCString)
    const extensionDataAndStatus = getCampaignExtensionData(extensionsController, workflow, input).pipe(share())
    const campaignPreview = extensionDataAndStatus.pipe(
        map(([extensionData]) => extensionData),
        switchMap(extensionData =>
            queryGraphQL(
                gql`
                    query CampaignPreview($input: ExpCampaignPreviewInput!) {
                        expCampaignPreview(input: $input) {
                            ...CampaignPreviewFragment
                        }
                    }
                    ${CampaignPreviewFragment}
                    ${ThreadConnectionFiltersFragment}
                    ${ActorFragment}
                `,
                // tslint:disable-next-line: no-object-literal-type-assertion
                {
                    input: {
                        campaign: {
                            ...input,
                            extensionData,
                        },
                    },
                } as GQL.IExpCampaignPreviewOnQueryArguments
            ).pipe(
                map(dataOrThrowErrors),
                map(data => data.expCampaignPreview),
                tap(data => {
                    // TODO!(sqs) hack, compensate for the RepositoryComparison head not existing
                    const fixup = (c: GQL.IRepositoryComparison): void => {
                        if (c.range) {
                            c.range.headRevSpec.object = { oid: '' } as any
                        }
                        for (const d of c.fileDiffs.nodes) {
                            d.mostRelevantFile = { path: d.newPath, url: '' } as any
                        }
                    }
                    for (const c of data.repositoryComparisons) {
                        fixup(c)
                    }
                    for (const t of data.threads.nodes) {
                        if (t.repositoryComparison) {
                            fixup(t.repositoryComparison)
                        }
                    }
                }),
                catchError(err => of(asError(err)))
            )
        )
    )
    return combineLatest([campaignPreview, extensionDataAndStatus]).pipe(
        map(([campaignPreview, [data, status]]) => [campaignPreview, data, status] as const)
    )
}

const EMPTY_EXT_DATA: GQL.IExpCampaignExtensionData = {
    rawChangesets: [],
    rawDiagnostics: [],
    rawLogMessages: [],
    rawSideEffects: [],
}

/**
 * A React hook that observes a campaign preview queried from the GraphQL API.
 */
export const useCampaignPreview = (
    { extensionsController }: ExtensionsControllerProps,
    input: CreateCampaignInputWithoutExtensionData
): [CampaignPreviewResult, GQL.IExpCampaignExtensionData, ExtensionDataStatus, boolean] => {
    const inputSubject = useMemo(() => new Subject<CreateCampaignInputWithoutExtensionData>(), [])
    const [isLoading, setIsLoading] = useState(true)
    const [result, setResult] = useState<CampaignPreviewResult>(LOADING)
    const [data, setData] = useState<GQL.IExpCampaignExtensionData>(EMPTY_EXT_DATA)
    const [status, setStatus] = useState<ExtensionDataStatus>({ isLoading: false })
    useEffect(() => {
        // Refresh more slowly on changes to the name or description.
        const inputSubjectChanges = merge(
            inputSubject.pipe(distinctUntilChanged((a, b) => a.namespace === b.namespace && a.draft === b.draft)),
            inputSubject.pipe(
                debounceTime(250),
                distinctUntilChanged((a, b) => a.workflowAsJSONCString === b.workflowAsJSONCString)
            ),
            inputSubject.pipe(
                distinctUntilChanged(
                    (a, b) =>
                        a.name === b.name && a.body === b.body && a.startDate === b.startDate && a.dueDate === b.dueDate
                ),
                debounceTime(2000)
            )
        )
        const subscription = merge(
            inputSubjectChanges.pipe(
                distinctUntilChanged((a, b) => isEqual(a, b)),
                mapTo([LOADING, EMPTY_EXT_DATA, { isLoading: true }] as [
                    typeof LOADING,
                    GQL.IExpCampaignExtensionData,
                    ExtensionDataStatus
                ])
            ),
            inputSubjectChanges.pipe(
                throttleTime(1000, undefined, { leading: true, trailing: true }),
                distinctUntilChanged((a, b) => isEqual(a, b)),
                switchMap(input => queryCampaignPreview({ extensionsController, input }))
            )
        ).subscribe(([result, data, status]) => {
            setStatus(status)
            setData(data)
            setResult(prevResult => {
                setIsLoading(result === LOADING || status.isLoading)
                // Reuse last result while loading, to reduce UI jitter.
                return result === LOADING && prevResult !== LOADING
                    ? isErrorLike(prevResult)
                        ? LOADING
                        : prevResult
                    : result
            })
        })
        return () => subscription.unsubscribe()
    }, [extensionsController, inputSubject])
    useEffect(() => inputSubject.next(input), [input, inputSubject])
    return [result, data, status, isLoading]
}
