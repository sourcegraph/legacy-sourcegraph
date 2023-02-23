import { DecoratorFn, Meta, Story } from '@storybook/react'
import { addMinutes, formatRFC3339, subMinutes } from 'date-fns'
import { MATCH_ANY_PARAMETERS, WildcardMockLink } from 'wildcard-mock-link'

import { getDocumentNode } from '@sourcegraph/http-client'
import { PermissionSyncJobReasonGroup, PermissionSyncJobState } from '@sourcegraph/shared/src/graphql-operations'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { WebStory } from '../../components/WebStory'
import { PermissionsSyncJob } from '../../graphql-operations'

import { PERMISSIONS_SYNC_JOBS_QUERY } from './backend'
import { PermissionsSyncJobsTable } from './PermissionsSyncJobsTable'

const decorator: DecoratorFn = Story => <Story />

const config: Meta = {
    title: 'web/src/site-admin/permissions-center/PermissionsSyncJobsTable',
    decorators: [decorator],
}

export default config

const TIMESTAMP_MOCK = subMinutes(Date.now(), 5)

export const FiveSyncJobsFound: Story = () => (
    <WebStory>
        {() => (
            <MockedTestProvider
                link={
                    new WildcardMockLink([
                        {
                            request: {
                                query: getDocumentNode(PERMISSIONS_SYNC_JOBS_QUERY),
                                variables: MATCH_ANY_PARAMETERS,
                            },
                            result: {
                                data: {
                                    permissionSyncJobs: {
                                        nodes: [
                                            createSyncJobMock(
                                                '1',
                                                PermissionSyncJobState.COMPLETED,
                                                {
                                                    __typename: 'Repository',
                                                    name: 'sourcegraph/sourcegraph',
                                                },
                                                {
                                                    group: PermissionSyncJobReasonGroup.WEBHOOK,
                                                    message: 'REASON_GITHUB_REPO_EVENT',
                                                }
                                            ),
                                            createSyncJobMock(
                                                '2',
                                                PermissionSyncJobState.ERRORED,
                                                {
                                                    __typename: 'User',
                                                    username: 'abdul',
                                                },
                                                {
                                                    group: PermissionSyncJobReasonGroup.SOURCEGRAPH,
                                                    message: 'REASON_USER_EMAIL_VERIFIED',
                                                }
                                            ),
                                            createSyncJobMock(
                                                '3',
                                                PermissionSyncJobState.FAILED,
                                                {
                                                    __typename: 'Repository',
                                                    name: 'sourcegraph/hoursegraph',
                                                },
                                                {
                                                    group: PermissionSyncJobReasonGroup.SCHEDULE,
                                                    message: 'REASON_REPO_OUTDATED_PERMS',
                                                }
                                            ),
                                            createSyncJobMock(
                                                '4',
                                                PermissionSyncJobState.PROCESSING,
                                                {
                                                    __typename: 'User',
                                                    username: 'omar',
                                                },
                                                {
                                                    group: PermissionSyncJobReasonGroup.MANUAL,
                                                    message: 'REASON_MANUAL_USER_SYNC',
                                                }
                                            ),
                                            createSyncJobMock(
                                                '5',
                                                PermissionSyncJobState.QUEUED,
                                                {
                                                    __typename: 'Repository',
                                                    name: 'sourcegraph/stillfunny',
                                                },
                                                {
                                                    group: PermissionSyncJobReasonGroup.MANUAL,
                                                    message: 'REASON_MANUAL_REPO_SYNC',
                                                }
                                            ),
                                        ],
                                        totalCount: 5,
                                        pageInfo: {
                                            hasNextPage: true,
                                            hasPreviousPage: false,
                                            startCursor: null,
                                            endCursor: null,
                                        },
                                    },
                                },
                            },
                            nMatches: Number.POSITIVE_INFINITY,
                        },
                    ])
                }
            >
                <PermissionsSyncJobsTable telemetryService={NOOP_TELEMETRY_SERVICE} />
            </MockedTestProvider>
        )}
    </WebStory>
)

FiveSyncJobsFound.storyName = 'Five sync jobs'

interface repo {
    __typename: 'Repository'
    name: string
}

interface user {
    __typename: 'User'
    username: string
}

type subject = repo | user

interface reason {
    __typename?: 'PermissionSyncJobReason'
    group: PermissionSyncJobReasonGroup
    message: string
}

function createSyncJobMock(
    id: string,
    state: PermissionSyncJobState,
    subject: subject,
    reason: reason
): PermissionsSyncJob {
    return {
        __typename: 'PermissionSyncJob',
        id,
        state,
        subject,
        reason,
        triggeredByUser: {
            username: 'super-site-admin',
        },
        queuedAt: formatRFC3339(TIMESTAMP_MOCK),
        startedAt: state !== PermissionSyncJobState.QUEUED ? formatRFC3339(addMinutes(TIMESTAMP_MOCK, 1)) : null,
        finishedAt:
            state !== PermissionSyncJobState.QUEUED && state !== PermissionSyncJobState.PROCESSING
                ? formatRFC3339(addMinutes(TIMESTAMP_MOCK, 2))
                : null,
        processAfter: null,
        permissionsAdded: 1337,
        permissionsRemoved: 42,
        permissionsFound: 1337 + 42,
    }
}
