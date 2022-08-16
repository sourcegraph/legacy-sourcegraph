import React, { useEffect, useCallback, useState } from 'react'

import { mdiCloudDownload, mdiCog } from '@mdi/js'
import { RouteComponentProps } from 'react-router'
import { Observable } from 'rxjs'

import { useQuery } from '@sourcegraph/http-client'
import { RepoLink } from '@sourcegraph/shared/src/components/RepoLink'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import {
    Button,
    Link,
    Alert,
    Icon,
    H2,
    Text,
    Tooltip,
    Container,
    Card,
    CardBody,
    LoadingSpinner,
} from '@sourcegraph/wildcard'

import { TerminalLine } from '../auth/Terminal'
import {
    FilteredConnection,
    FilteredConnectionFilter,
    FilteredConnectionQueryArguments,
} from '../components/FilteredConnection'
import { PageTitle } from '../components/PageTitle'
import {
    RepositoriesResult,
    RepositoryStatsResult,
    RepositoryStatsVariables,
    SiteAdminRepositoryFields,
} from '../graphql-operations'
import { refreshSiteFlags } from '../site/backend'

import { fetchAllRepositoriesAndPollIfEmptyOrAnyCloning, REPOSITORY_STATS, REPO_PAGE_POLL_INTERVAL } from './backend'
import { ExternalRepositoryIcon } from './components/ExternalRepositoryIcon'
import { RepoMirrorInfo as RepoMirrorInfo } from './components/RepoMirrorInfo'

import styles from './SiteAdminRepositoriesPage.module.scss'

interface RepositoryNodeProps {
    node: SiteAdminRepositoryFields
}

const RepositoryNode: React.FunctionComponent<React.PropsWithChildren<RepositoryNodeProps>> = ({ node }) => (
    <li
        className="repository-node list-group-item py-2"
        data-test-repository={node.name}
        data-test-cloned={node.mirrorInfo.cloned}
    >
        <div className="d-flex align-items-center justify-content-between">
            <div>
                <ExternalRepositoryIcon externalRepo={node.externalRepository} />
                <RepoLink repoName={node.name} to={node.url} />
                <RepoMirrorInfo mirrorInfo={node.mirrorInfo} />
            </div>

            <div className="repository-node__actions">
                {!node.mirrorInfo.cloneInProgress && !node.mirrorInfo.cloned && (
                    <Button to={node.url} variant="secondary" size="sm" as={Link}>
                        <Icon aria-hidden={true} svgPath={mdiCloudDownload} /> Clone now
                    </Button>
                )}{' '}
                <Tooltip content="Repository settings">
                    <Button to={`/${node.name}/-/settings`} variant="secondary" size="sm" as={Link}>
                        <Icon aria-hidden={true} svgPath={mdiCog} /> Settings
                    </Button>
                </Tooltip>
            </div>
        </div>

        {node.mirrorInfo.lastError && (
            <div className={styles.alertWrapper}>
                <Alert variant="warning">
                    <Text className="font-weight-bold">Error syncing repository:</Text>
                    <TerminalLine className={styles.alertContent}>
                        {node.mirrorInfo.lastError.replaceAll('\r', '\n')}
                    </TerminalLine>
                </Alert>
            </div>
        )}
    </li>
)

interface Props extends RouteComponentProps<{}>, TelemetryProps {}

const FILTERS: FilteredConnectionFilter[] = [
    {
        id: 'status',
        label: 'Status',
        type: 'select',
        values: [
            {
                label: 'All',
                value: 'all',
                tooltip: 'Show all repositories',
                args: {},
            },
            {
                label: 'Cloned',
                value: 'cloned',
                tooltip: 'Show cloned repositories only',
                args: { cloned: true, notCloned: false },
            },
            {
                label: 'Not cloned',
                value: 'not-cloned',
                tooltip: 'Show only repositories that have not been cloned yet',
                args: { cloned: false, notCloned: true },
            },
            {
                label: 'Needs index',
                value: 'needs-index',
                tooltip: 'Show only repositories that need to be indexed',
                args: { indexed: false },
            },
            {
                label: 'Failed fetch/clone',
                value: 'failed-fetch',
                tooltip: 'Show only repositories that have failed to fetch or clone',
                args: { failedFetch: true },
            },
        ],
    },
]

/**
 * A page displaying the repositories on this site.
 */
export const SiteAdminRepositoriesPage: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    history,
    location,
    telemetryService,
}) => {
    useEffect(() => {
        telemetryService.logViewEvent('SiteAdminRepos')
    }, [telemetryService])

    // Refresh global alert about enabling repositories when the user visits & navigates away from this page.
    useEffect(() => {
        refreshSiteFlags()
            .toPromise()
            .then(null, error => console.error(error))
        return () => {
            refreshSiteFlags()
                .toPromise()
                .then(null, error => console.error(error))
        }
    }, [])

    const [pollRepositoryStats, setPollRepositoryStats] = useState(true)
    const { data, loading, error } = useQuery<RepositoryStatsResult, RepositoryStatsVariables>(REPOSITORY_STATS, {
        pollInterval: pollRepositoryStats ? REPO_PAGE_POLL_INTERVAL : 0,
    })

    const queryRepositories = useCallback(
        (args: FilteredConnectionQueryArguments): Observable<RepositoriesResult['repositories']> =>
            fetchAllRepositoriesAndPollIfEmptyOrAnyCloning(args),
        []
    )
    const showRepositoriesAddedBanner = new URLSearchParams(location.search).has('repositoriesUpdated')

    if (!error && !loading && data && data.repositoryStats.total !== 0 && data.repositoryStats.cloning === 0) {
        setPollRepositoryStats(false)
    }

    return (
        <div className="site-admin-repositories-page">
            <PageTitle title="Repositories - Admin" />
            {showRepositoriesAddedBanner && (
                <Alert variant="success" as="p">
                    Syncing repositories. It may take a few moments to clone and index each repository. Repository
                    statuses are displayed below.
                </Alert>
            )}
            <H2>Repositories</H2>
            <Text>
                Repositories are synced from connected{' '}
                <Link to="/site-admin/external-services" data-testid="test-repositories-code-host-connections-link">
                    code hosts
                </Link>
                .
            </Text>
            {error && !loading && (
                <Alert variant="warning" as="p">
                    {error.message}
                </Alert>
            )}
            {loading && !error && <LoadingSpinner />}
            {!loading && !error && data && (
                <div className="d-flex justify-content-between text-center">
                    <Card className="flex-grow-1">
                        <CardBody>
                            <span className={styles.repoStatsNumber}>{data.repositoryStats.total}</span>
                            <Text className="mb-0">Repositories</Text>
                        </CardBody>
                    </Card>
                    <Card className="flex-grow-1">
                        <CardBody>
                            <span className={styles.repoStatsNumber}>{data.repositoryStats.notCloned}</span>
                            <Text className="mb-0">Not cloned</Text>
                        </CardBody>
                    </Card>
                    <Card className="flex-grow-1">
                        <CardBody>
                            <span
                                className={classNames(
                                    styles.repoStatsNumber,
                                    data.repositoryStats.cloning > 0 && 'text-success'
                                )}
                            >
                                {data.repositoryStats.cloning}
                            </span>
                            <Text className="mb-0">Cloning</Text>
                        </CardBody>
                    </Card>
                    <Card className="flex-grow-1">
                        <CardBody>
                            <span className={styles.repoStatsNumber}>{data.repositoryStats.cloned}</span>
                            <Text className="mb-0">Cloned</Text>
                        </CardBody>
                    </Card>
                    <Card className="flex-grow-1">
                        <CardBody>
                            <span
                                className={classNames(
                                    styles.repoStatsNumber,
                                    data.repositoryStats.failedFetch > 0 && 'text-warning'
                                )}
                            >
                                {data.repositoryStats.failedFetch}
                            </span>
                            <Text className="mb-0">Failed clone/fetch</Text>
                        </CardBody>
                    </Card>
                </div>
            )}
            <Container className="mb-3">
                <FilteredConnection<SiteAdminRepositoryFields, Omit<RepositoryNodeProps, 'node'>>
                    className="mb-0"
                    listClassName="list-group list-group-flush mt-3"
                    noun="repository"
                    pluralNoun="repositories"
                    queryConnection={queryRepositories}
                    nodeComponent={RepositoryNode}
                    inputClassName="flex-1"
                    filters={FILTERS}
                    history={history}
                    location={location}
                />
            </Container>
        </div>
    )
}
