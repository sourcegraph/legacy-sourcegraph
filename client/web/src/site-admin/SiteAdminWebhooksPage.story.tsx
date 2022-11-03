import { DecoratorFn, Meta, Story } from '@storybook/react'

import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { WebStory } from '../components/WebStory'

import { SiteAdminWebhooksPage } from './SiteAdminWebhooksPage'
import { getDocumentNode } from '@sourcegraph/http-client'

import { ExternalServiceKind } from '../../../shared/src/graphql-operations'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import * as H from 'history'
import { WEBHOOKS } from './backend'
import { MATCH_ANY_PARAMETERS, WildcardMockLink } from 'wildcard-mock-link'
import { WebhookFields } from '../graphql-operations'

const decorator: DecoratorFn = Story => <Story />

const config: Meta = {
    title: 'web/src/site-admin/SiteAdminWebhooksPage',
    decorators: [decorator],
}

export default config

export const NoWebhooksFound: Story = () => (
    <WebStory>
        {() => (
            <MockedTestProvider
                link={
                    new WildcardMockLink([
                        {
                            request: {
                                query: getDocumentNode(WEBHOOKS),
                                variables: MATCH_ANY_PARAMETERS,
                            },
                            result: {
                                data: {
                                    webhooks: {
                                        nodes: [],
                                        totalCount: 0,
                                        pageInfo: {
                                            hasNextPage: false,
                                        },
                                    },
                                },
                            },
                        },
                    ])
                }
            >
                <SiteAdminWebhooksPage
                    match={{} as any}
                    history={H.createMemoryHistory()}
                    location={{} as any}
                    telemetryService={NOOP_TELEMETRY_SERVICE}
                />
            </MockedTestProvider>
        )}
    </WebStory>
)

NoWebhooksFound.storyName = 'No webhooks found'

export const FiveWebhooksFound: Story = () => (
    <WebStory>
        {() => (
            <MockedTestProvider
                link={
                    new WildcardMockLink([
                        {
                            request: {
                                query: getDocumentNode(WEBHOOKS),
                                variables: MATCH_ANY_PARAMETERS,
                            },
                            result: {
                                data: {
                                    webhooks: {
                                        nodes: [
                                            createWebhookMock(
                                                ExternalServiceKind.BITBUCKETCLOUD,
                                                'bitbucket.com/repo1'
                                            ),
                                            createWebhookMock(ExternalServiceKind.GITHUB, 'github.com/repo1'),
                                            createWebhookMock(ExternalServiceKind.GITHUB, 'github.com/repo2'),
                                            createWebhookMock(ExternalServiceKind.GITHUB, 'github.com/repo3'),
                                            createWebhookMock(ExternalServiceKind.BITBUCKETCLOUD, 'github.com/repo2'),
                                        ],
                                        totalCount: 5,
                                        pageInfo: {
                                            hasNextPage: false,
                                        },
                                    },
                                },
                            },
                        },
                    ])
                }
            >
                <SiteAdminWebhooksPage
                    match={{} as any}
                    history={H.createMemoryHistory()}
                    location={{} as any}
                    telemetryService={NOOP_TELEMETRY_SERVICE}
                />
            </MockedTestProvider>
        )}
    </WebStory>
)

FiveWebhooksFound.storyName = '5 webhooks found'

function createWebhookMock(kind: ExternalServiceKind, urn: string): WebhookFields {
    return {
        __typename: 'Webhook',
        createdAt: '',
        id: '',
        secret: null,
        updatedAt: '',
        url: '',
        uuid: '',
        codeHostKind: kind,
        codeHostURN: urn,
    }
}
