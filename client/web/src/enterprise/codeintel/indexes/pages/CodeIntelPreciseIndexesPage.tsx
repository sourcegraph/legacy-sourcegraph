import { FunctionComponent, useCallback, useEffect, useMemo, useState } from 'react'

import { useApolloClient } from '@apollo/client'
import classNames from 'classnames'
import * as H from 'history'
import { RouteComponentProps, useLocation } from 'react-router'
import { of, Subject } from 'rxjs'
import { tap } from 'rxjs/operators'

import { Timestamp } from '@sourcegraph/branded/src/components/Timestamp'
import { isErrorLike } from '@sourcegraph/common'
import { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import {
    Alert,
    Button,
    Checkbox,
    Container,
    ErrorAlert,
    H3,
    Icon,
    Link,
    PageHeader,
    Text,
    Tooltip,
    useObservable,
} from '@sourcegraph/wildcard'

import {
    FilteredConnection,
    FilteredConnectionFilter,
    FilteredConnectionQueryArguments,
} from '../../../../components/FilteredConnection'
import { PageTitle } from '../../../../components/PageTitle'
import { PreciseIndexesVariables, PreciseIndexFields, PreciseIndexState } from '../../../../graphql-operations'
import { FlashMessage } from '../../configuration/components/FlashMessage'
import { PreciseIndexLastUpdated } from '../components/CodeIntelLastUpdated'
import { CodeIntelStateIcon } from '../components/CodeIntelStateIcon'
import { CodeIntelStateLabel } from '../components/CodeIntelStateLabel'
import { ProjectDescription } from '../components/ProjectDescription'
import { queryCommitGraph as defaultQueryCommitGraph } from '../hooks/queryCommitGraph'
import { queryPreciseIndexes as defaultQueryPreciseIndexes, statesFromString } from '../hooks/queryPreciseIndexes'
import { useDeletePreciseIndex as defaultUseDeletePreciseIndex } from '../hooks/useDeletePreciseIndex'
import { useDeletePreciseIndexes as defaultUseDeletePreciseIndexes } from '../hooks/useDeletePreciseIndexes'
import { useReindexPreciseIndex as defaultUseReindexPreciseIndex } from '../hooks/useReindexPreciseIndex'
import { useReindexPreciseIndexes as defaultUseReindexPreciseIndexes } from '../hooks/useReindexPreciseIndexes'

import { EnqueueForm } from '../components/EnqueueForm'
import styles from './CodeIntelPreciseIndexesPage.module.scss'
import { mdiMapSearch } from '@mdi/js'

export interface CodeIntelPreciseIndexesPageProps extends RouteComponentProps<{}>, ThemeProps, TelemetryProps {
    authenticatedUser: AuthenticatedUser | null
    repo?: { id: string }
    now?: () => Date
    queryCommitGraph?: typeof defaultQueryCommitGraph
    queryPreciseIndexes?: typeof defaultQueryPreciseIndexes
    useDeletePreciseIndex?: typeof defaultUseDeletePreciseIndex
    useDeletePreciseIndexes?: typeof defaultUseDeletePreciseIndexes
    useReindexPreciseIndex?: typeof defaultUseReindexPreciseIndex
    useReindexPreciseIndexes?: typeof defaultUseReindexPreciseIndexes
}

const filters: FilteredConnectionFilter[] = [
    {
        id: 'filters',
        label: 'State',
        type: 'select',
        values: [
            {
                label: 'All',
                value: 'all',
                tooltip: 'Show all indexes',
                args: {},
            },
            {
                label: 'Completed',
                value: 'completed',
                tooltip: 'Show completed indexes only',
                args: { states: PreciseIndexState.COMPLETED },
            },

            {
                label: 'Queued',
                value: 'queued',
                tooltip: 'Show queued indexes only',
                args: {
                    states: [
                        PreciseIndexState.UPLOADING_INDEX,
                        PreciseIndexState.QUEUED_FOR_INDEXING,
                        PreciseIndexState.QUEUED_FOR_PROCESSING,
                    ].join(','),
                },
            },
            {
                label: 'In progress',
                value: 'in-progress',
                tooltip: 'Show in-progress indexes only',
                args: { states: [PreciseIndexState.INDEXING, PreciseIndexState.PROCESSING].join(',') },
            },
            {
                label: 'Errored',
                value: 'errored',
                tooltip: 'Show errored indexes only',
                args: { states: [PreciseIndexState.INDEXING_ERRORED, PreciseIndexState.PROCESSING_ERRORED].join(',') },
            },
        ],
    },
]

export const CodeIntelPreciseIndexesPage: FunctionComponent<CodeIntelPreciseIndexesPageProps> = ({
    authenticatedUser,
    repo,
    now,
    queryCommitGraph = defaultQueryCommitGraph,
    queryPreciseIndexes = defaultQueryPreciseIndexes,
    useDeletePreciseIndex = defaultUseDeletePreciseIndex,
    useDeletePreciseIndexes = defaultUseDeletePreciseIndexes,
    useReindexPreciseIndex = defaultUseReindexPreciseIndex,
    useReindexPreciseIndexes = defaultUseReindexPreciseIndexes,
    telemetryService,
    history,
}) => {
    useEffect(() => telemetryService.logViewEvent('CodeIntelPreciseIndexesPage'), [telemetryService])
    const location = useLocation<{ message: string; modal: string }>()

    const apolloClient = useApolloClient()
    const { handleDeletePreciseIndex, deleteError } = useDeletePreciseIndex()
    const { handleDeletePreciseIndexes, deletesError } = useDeletePreciseIndexes()
    const { handleReindexPreciseIndex, reindexError } = useReindexPreciseIndex()
    const { handleReindexPreciseIndexes, reindexesError } = useReindexPreciseIndexes()
    const commitGraphMetadata = useObservable(
        useMemo(
            () => (repo ? queryCommitGraph(repo?.id, apolloClient) : of(undefined)),
            [repo, queryCommitGraph, apolloClient]
        )
    )

    // Poke filtered connection to refresh
    const refresh = useMemo(() => new Subject<undefined>(), [])
    const querySubject = useMemo(() => new Subject<string>(), [])

    // State used to control bulk index selection
    const [selection, setSelection] = useState<Set<string> | 'all'>(new Set())

    // Updates state of bulk index selection
    const onCheckboxToggle = useCallback(
        (id: string, checked: boolean) =>
            setSelection(selection =>
                selection === 'all'
                    ? selection
                    : checked
                    ? new Set([...selection, id])
                    : new Set([...selection].filter(selected => selected !== id))
            ),
        [setSelection]
    )

    // State used to spy on the query connection callback defined below
    const [args, setArgs] = useState<
        Partial<Omit<PreciseIndexesVariables, 'states'>> & { states?: string; isLatestForRepo?: boolean }
    >()
    const [totalCount, setTotalCount] = useState<number | undefined>(undefined)

    // Query indexes matching filter criteria
    const queryConnection = useCallback(
        (args: FilteredConnectionQueryArguments) => {
            const stashArgs = {
                repo: repo?.id,
                query: args.query,
                states: (args as any).states,
                isLatestForRepo: (args as any).isLatestForRepo,
            }

            setArgs(stashArgs)
            setSelection(new Set())

            return queryPreciseIndexes(stashArgs, apolloClient).pipe(
                tap(connection => {
                    setTotalCount(connection.totalCount ?? undefined)
                })
            )
        },
        [queryPreciseIndexes, apolloClient]
    )

    const onRawDelete = () => {
        if (selection === 'all') {
            if (args !== undefined && confirm(`Delete ${totalCount} indexes?`)) {
                const typedStates = statesFromString(args?.states)
                return handleDeletePreciseIndexes({
                    variables: {
                        repo: args.repo ?? null,
                        query: args.query ?? null,
                        states: typedStates.length > 0 ? typedStates : null,
                        isLatestForRepo: args.isLatestForRepo ?? null,
                    },
                    update: cache => cache.modify({ fields: { node: () => {} } }),
                })
            }

            return Promise.resolve()
        }

        return Promise.all(
            [...selection].map(id =>
                handleDeletePreciseIndex({
                    variables: { id },
                    update: cache => cache.modify({ fields: { node: () => {} } }),
                })
            )
        )
    }

    const onRawReindex = () => {
        if (selection === 'all') {
            if (args !== undefined && confirm(`Mark ${totalCount} indexes as replaceable by auto-indexing?`)) {
                const typedStates = statesFromString(args?.states)
                return handleReindexPreciseIndexes({
                    variables: {
                        repo: args.repo ?? null,
                        query: args.query ?? null,
                        states: typedStates.length > 0 ? typedStates : null,
                        isLatestForRepo: args.isLatestForRepo ?? null,
                    },
                    update: cache => cache.modify({ fields: { node: () => {} } }),
                })
            }

            return Promise.resolve()
        }

        return Promise.all(
            [...selection].map(id =>
                handleReindexPreciseIndex({
                    variables: { id },
                    update: cache => cache.modify({ fields: { node: () => {} } }),
                })
            )
        )
    }

    const onDelete = () => onRawDelete().then(() => refresh.next())
    const onReindex = () => onRawReindex().then(() => refresh.next())

    return (
        <div>
            <PageTitle title="Precise indexes" />
            <PageHeader
                headingElement="h2"
                path={[{ text: 'Precise indexes' }]}
                description="Precise code intelligence index data and auto-indexing jobs."
                className="mb-3"
            />

            {!!location.state && <FlashMessage state={location.state.modal} message={location.state.message} />}

            {repo && commitGraphMetadata && (
                <Alert variant={commitGraphMetadata.stale ? 'primary' : 'success'} aria-live="off">
                    {commitGraphMetadata.stale ? (
                        <>
                            Repository commit graph is currently stale and is queued to be refreshed. Refreshing the
                            commit graph updates which uploads are visible from which commits.
                        </>
                    ) : (
                        <>Repository commit graph is currently up to date.</>
                    )}{' '}
                    {commitGraphMetadata.updatedAt && (
                        <>
                            Last refreshed <Timestamp date={commitGraphMetadata.updatedAt} now={now} />.
                        </>
                    )}
                </Alert>
            )}

            {repo && authenticatedUser?.siteAdmin && (
                <Container className="mb-2">
                    <EnqueueForm repoId={repo.id} querySubject={querySubject} />
                </Container>
            )}

            {isErrorLike(deleteError) && <ErrorAlert prefix="Error deleting precise index" error={deleteError} />}
            {isErrorLike(deletesError) && <ErrorAlert prefix="Error deleting precise indexes" error={deletesError} />}
            {isErrorLike(reindexError) && (
                <ErrorAlert prefix="Error marking precise index as replaceable by auto-indexing" error={reindexError} />
            )}
            {isErrorLike(reindexesError) && (
                <ErrorAlert
                    prefix="Error marking precise indexes as replaceable by auto-indexing"
                    error={reindexesError}
                />
            )}

            <Container>
                <div className="list-group position-relative">
                    <FilteredConnection<PreciseIndexFields, Omit<IndexNodeProps, 'node'>>
                        listComponent="div"
                        inputClassName="ml-2 flex-1"
                        listClassName="mb-3"
                        noun="precise index"
                        pluralNoun="precise indexes"
                        querySubject={querySubject}
                        nodeComponent={IndexNode}
                        nodeComponentProps={{ repo, selection, onCheckboxToggle, history }}
                        headComponent={() => (
                            <div className={classNames(styles.header, 'px-4 py-3')}>
                                <div className="px-3">
                                    <Checkbox
                                        label=""
                                        id="checkAll"
                                        checked={selection === 'all'}
                                        onChange={() =>
                                            setSelection(selection => (selection === 'all' ? new Set() : 'all'))
                                        }
                                    />
                                </div>

                                {authenticatedUser?.siteAdmin && (
                                    <div className="text-right">
                                        <Button
                                            className="mr-2"
                                            variant="danger"
                                            disabled={selection !== 'all' && selection.size === 0}
                                            // eslint-disable-next-line @typescript-eslint/no-misused-promises
                                            onClick={onDelete}
                                        >
                                            Delete{' '}
                                            {(selection === 'all' ? totalCount : selection.size) === 0 ? (
                                                ''
                                            ) : (
                                                <>
                                                    {selection === 'all' ? totalCount : selection.size}{' '}
                                                    {(selection === 'all' ? totalCount : selection.size) === 1
                                                        ? 'index'
                                                        : 'indexes'}
                                                </>
                                            )}
                                        </Button>

                                        <Button
                                            className="mr-2"
                                            variant="secondary"
                                            disabled={selection !== 'all' && selection.size === 0}
                                            // eslint-disable-next-line @typescript-eslint/no-misused-promises
                                            onClick={onReindex}
                                        >
                                            Reindex{' '}
                                            {(selection === 'all' ? totalCount : selection.size) === 0 ? (
                                                ''
                                            ) : (
                                                <>
                                                    {selection === 'all' ? totalCount : selection.size}{' '}
                                                    {(selection === 'all' ? totalCount : selection.size) === 1
                                                        ? 'index'
                                                        : 'indexes'}
                                                </>
                                            )}
                                        </Button>
                                    </div>
                                )}
                            </div>
                        )}
                        queryConnection={queryConnection}
                        history={history}
                        location={location}
                        cursorPaging={true}
                        filters={filters}
                        emptyElement={<EmptyIndex />}
                        updates={refresh}
                    />
                </div>
            </Container>
        </div>
    )
}

interface IndexNodeProps {
    node: PreciseIndexFields
    repo?: { id: string }
    selection: Set<string> | 'all'
    onCheckboxToggle: (id: string, checked: boolean) => void
    history: H.History
}

const IndexNode: FunctionComponent<IndexNodeProps> = ({ node, repo, selection, onCheckboxToggle, history }) => (
    <>
        <div className={classNames(styles.grid, 'px-4')} onClick={() => history.push(`./indexes/${node.id}`)}>
            <div className="px-3 py-4" onClick={event => event.stopPropagation()}>
                <Checkbox
                    label=""
                    id="disabledFieldsetCheck"
                    disabled={selection === 'all'}
                    checked={selection === 'all' ? true : selection.has(node.id)}
                    onClick={event => event.stopPropagation()}
                    onChange={input => onCheckboxToggle(node.id, input.target.checked)}
                />
            </div>

            <div className={classNames(styles.information, 'd-flex flex-column')}>
                {!repo && (
                    <div>
                        <H3 className="m-0 mb-1">
                            {node.projectRoot ? (
                                <Link to={node.projectRoot.repository.url} onClick={event => event.stopPropagation()}>
                                    {node.projectRoot.repository.name}
                                </Link>
                            ) : (
                                <span>Unknown repository</span>
                            )}
                        </H3>
                    </div>
                )}

                <div>
                    <span className="mr-2 d-block d-mdinline-block">
                        <ProjectDescription index={node} onLinkClick={event => event.stopPropagation()} />
                    </span>

                    <small className="text-mute">
                        <PreciseIndexLastUpdated index={node} />{' '}
                        {node.shouldReindex && (
                            <Tooltip content="This index has been marked as replaceable by auto-indexing.">
                                <span className={classNames(styles.tag, 'ml-1 rounded')}>
                                    (replaceable by auto-indexing)
                                </span>
                            </Tooltip>
                        )}
                    </small>
                </div>
            </div>

            <span className={classNames(styles.state, 'd-none d-md-inline')}>
                <div className="d-flex flex-column align-items-center">
                    <CodeIntelStateIcon state={node.state} autoIndexed={!!node.indexingFinishedAt} />
                    <CodeIntelStateLabel
                        state={node.state}
                        autoIndexed={!!node.indexingFinishedAt}
                        placeInQueue={node.placeInQueue}
                        className="mt-2"
                    />
                </div>
            </span>
        </div>
    </>
)

const EmptyIndex: React.FunctionComponent<{}> = () => (
    <Text alignment="center" className="text-muted w-100 mb-0 mt-1">
        <Icon className="mb-2" svgPath={mdiMapSearch} inline={false} aria-hidden={true} />
        <br />
        No indexes.
    </Text>
)
