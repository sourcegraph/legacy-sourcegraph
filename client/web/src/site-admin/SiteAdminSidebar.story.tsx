import { DecoratorFn, Meta, Story } from '@storybook/react'

import { Grid } from '@sourcegraph/wildcard'

import { WebStory } from '../components/WebStory'
import { enterpriseSiteAdminSidebarGroups } from '../enterprise/site-admin/sidebaritems'

import { SiteAdminSidebar } from './SiteAdminSidebar'

const decorator: DecoratorFn = story => <div style={{ width: '192px' }}>{story()}</div>

const config: Meta = {
    title: 'web/site-admin/AdminSidebar',
    decorators: [decorator],
    parameters: {
        chromatic: {
            disableSnapshot: false,
        },
    },
}

export default config

export const AdminSidebarItems: Story = () => (
    <WebStory>
        {webProps => (
            <Grid columnCount={5}>
                <code>isSourcegraphApp=true</code>
                <code>default</code>
                <code>isSourcegraphDotCom=true</code>
                <code>batchChangesEnabled=false</code>
                <code>codeInsightsEnabled=false</code>
                <SiteAdminSidebar
                    {...webProps}
                    groups={enterpriseSiteAdminSidebarGroups}
                    isSourcegraphDotCom={false}
                    isSourcegraphApp={true}
                    batchChangesEnabled={true}
                    batchChangesExecutionEnabled={true}
                    batchChangesWebhookLogsEnabled={true}
                    codeInsightsEnabled={true}
                />
                <SiteAdminSidebar
                    {...webProps}
                    groups={enterpriseSiteAdminSidebarGroups}
                    isSourcegraphDotCom={false}
                    isSourcegraphApp={false}
                    batchChangesEnabled={true}
                    batchChangesExecutionEnabled={true}
                    batchChangesWebhookLogsEnabled={true}
                    codeInsightsEnabled={true}
                />
                <SiteAdminSidebar
                    {...webProps}
                    groups={enterpriseSiteAdminSidebarGroups}
                    isSourcegraphDotCom={true}
                    isSourcegraphApp={false}
                    batchChangesEnabled={true}
                    batchChangesExecutionEnabled={true}
                    batchChangesWebhookLogsEnabled={true}
                    codeInsightsEnabled={true}
                />
                <SiteAdminSidebar
                    {...webProps}
                    groups={enterpriseSiteAdminSidebarGroups}
                    isSourcegraphDotCom={false}
                    isSourcegraphApp={false}
                    batchChangesEnabled={false}
                    batchChangesExecutionEnabled={false}
                    batchChangesWebhookLogsEnabled={false}
                    codeInsightsEnabled={true}
                />
                <SiteAdminSidebar
                    {...webProps}
                    groups={enterpriseSiteAdminSidebarGroups}
                    isSourcegraphDotCom={false}
                    isSourcegraphApp={false}
                    batchChangesEnabled={true}
                    batchChangesExecutionEnabled={true}
                    batchChangesWebhookLogsEnabled={true}
                    codeInsightsEnabled={false}
                />
            </Grid>
        )}
    </WebStory>
)

AdminSidebarItems.storyName = 'Admin Sidebar Items'
AdminSidebarItems.parameters = {
    design: {
        type: 'figma',
        url: 'https://www.figma.com/file/cT23UcGqbdjdV4H2yxweIu/%2311775-Map-the-current-information-architecture-%5BApproved%5D?node-id=68%3A1',
    },
}
