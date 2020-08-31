import { storiesOf } from '@storybook/react'
import React from 'react'
import webStyles from '../../../enterprise.scss'
import { ChangesetSpecList } from './ChangesetSpecList'
import { of, Observable } from 'rxjs'
import { CampaignSpecChangesetSpecsResult, ChangesetSpecFields } from '../../../graphql-operations'
import { visibleChangesetSpecStories } from './VisibleChangesetSpecNode.story'
import { hiddenChangesetSpecStories } from './HiddenChangesetSpecNode.story'
import { WebStory } from '../../../components/WebStory'

const { add } = storiesOf('web/campaigns/apply/ChangesetSpecList', module).addDecorator(story => (
    <div className="p-3 container web-content">{story()}</div>
))

const nodes: ChangesetSpecFields[] = [
    ...Object.values(visibleChangesetSpecStories),
    ...Object.values(hiddenChangesetSpecStories),
]

const queryChangesetSpecs = (): Observable<
    (CampaignSpecChangesetSpecsResult['node'] & { __typename: 'CampaignSpec' })['changesetSpecs']
> =>
    of({
        pageInfo: {
            endCursor: null,
            hasNextPage: false,
        },
        totalCount: nodes.length,
        nodes,
    })

const queryEmptyFileDiffs = () =>
    of({ fileDiffs: { totalCount: 0, pageInfo: { endCursor: null, hasNextPage: false }, nodes: [] } })

add('List view', () => (
    <WebStory webStyles={webStyles}>
        {props => (
            <ChangesetSpecList
                {...props}
                campaignSpecID="123123"
                queryChangesetSpecs={queryChangesetSpecs}
                queryChangesetSpecFileDiffs={queryEmptyFileDiffs}
            />
        )}
    </WebStory>
))
