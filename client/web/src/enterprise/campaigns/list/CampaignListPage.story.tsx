import { storiesOf } from '@storybook/react'
import React from 'react'
import { CampaignListPage } from './CampaignListPage'
import { nodes } from './CampaignNode.story'
import { of } from 'rxjs'
import { EnterpriseWebStory } from '../../components/EnterpriseWebStory'

const { add } = storiesOf('web/campaigns/CampaignListPage', module)
    .addDecorator(story => <div className="p-3 container web-content">{story()}</div>)
    .addParameters({
        chromatic: {
            viewports: [320, 576, 978, 1440],
        },
    })

const queryCampaigns = () =>
    of({
        campaigns: {
            totalCount: Object.values(nodes).length,
            nodes: Object.values(nodes),
            pageInfo: { endCursor: null, hasNextPage: false },
        },
        totalCount: Object.values(nodes).length,
    })

add('List of campaigns', () => (
    <EnterpriseWebStory>{props => <CampaignListPage {...props} queryCampaigns={queryCampaigns} />}</EnterpriseWebStory>
))

add('No campaigns', () => (
    <EnterpriseWebStory>
        {props => (
            <CampaignListPage
                {...props}
                queryCampaigns={() =>
                    of({
                        campaigns: {
                            totalCount: 0,
                            nodes: [],
                            pageInfo: {
                                endCursor: null,
                                hasNextPage: false,
                            },
                        },
                        totalCount: 0,
                    })
                }
            />
        )}
    </EnterpriseWebStory>
))
