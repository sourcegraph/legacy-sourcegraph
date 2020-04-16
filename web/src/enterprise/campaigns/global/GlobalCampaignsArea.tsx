import React from 'react'
import { RouteComponentProps, Switch, Route } from 'react-router'
import { GlobalCampaignListPage } from './list/GlobalCampaignListPage'
import { CampaignDetails } from '../detail/CampaignDetails'
import { IUser } from '../../../../../shared/src/graphql/schema'
import { withAuthenticatedUser } from '../../../auth/withAuthenticatedUser'
import { ThemeProps } from '../../../../../shared/src/theme'
import { CreateCampaign } from './create/CreateCampaign'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import { PlatformContextProps } from '../../../../../shared/src/platform/context'
import { TelemetryProps } from '../../../../../shared/src/telemetry/telemetryService'
import { CampaignUpdateSelection } from '../detail/CampaignUpdateSelection'
import { CampaignCLIHelp } from './create/CampaignCLIHelp'
import { CampaignsDotComPage } from './marketing/CampaignsDotComPage'
import { CampaignsSiteAdminMarketingPage } from './marketing/CampaignsSiteAdminMarketingPage'
import { CampaignsUserMarketingPage } from './marketing/CampaignsUserMarketingPage'

interface Props
    extends RouteComponentProps<{}>,
        ThemeProps,
        ExtensionsControllerProps,
        TelemetryProps,
        PlatformContextProps {
    authenticatedUser: IUser
    isSourcegraphDotCom: boolean
}

/**
 * The global campaigns area.
 */
export const GlobalCampaignsArea = withAuthenticatedUser<Props>(({ match, ...outerProps }) => {
    let content: React.ReactFragment
    if (outerProps.isSourcegraphDotCom) {
        content = <CampaignsDotComPage {...outerProps} />
    } else if (window.context.experimentalFeatures?.automation === 'enabled') {
        if (!outerProps.authenticatedUser.siteAdmin && window.context.site['campaigns.readAccess.enabled'] !== true) {
            content = <CampaignsUserMarketingPage {...outerProps} enableReadAccess={true} />
        } else {
            content = (
                <>
                    {/* eslint-disable react/jsx-no-bind */}
                    <Switch>
                        <Route
                            render={props => <GlobalCampaignListPage {...outerProps} {...props} />}
                            path={match.url}
                            exact={true}
                        />
                        <Route
                            path={`${match.url}/create`}
                            render={props => <CreateCampaign {...outerProps} {...props} />}
                            exact={true}
                        />
                        <Route
                            path={`${match.url}/cli`}
                            render={props => <CampaignCLIHelp {...outerProps} {...props} />}
                            exact={true}
                        />
                        <Route
                            path={`${match.url}/new`}
                            render={props => <CampaignDetails {...outerProps} {...props} />}
                            exact={true}
                        />
                        <Route
                            path={`${match.url}/update`}
                            render={props => <CampaignUpdateSelection {...outerProps} {...props} />}
                            exact={true}
                        />
                        <Route
                            path={`${match.url}/:campaignID`}
                            render={({ match, ...props }: RouteComponentProps<{ campaignID: string }>) => (
                                <CampaignDetails {...outerProps} {...props} campaignID={match.params.campaignID} />
                            )}
                        />
                    </Switch>
                    <div className="fixed-bottom text center ml-4 mr-4 mt-4">
                        <p className="font-italic">
                            Campaigns are currently in <span className="badge badge-info badge-outline">Beta</span>.
                            During the beta period, Campaigns are free to use. After the beta period, Campaigns will be
                            available as a paid add-on. We're looking forward to your feedback! Get in touch on Twitter{' '}
                            <a href="https://twitter.com/srcgraph">@srcgraph</a>, file an issue in our{' '}
                            <a href="https://github.com/sourcegraph/sourcegraph/issues">public issue tracker</a>, or
                            email{' '}
                            <a href="mailto:feedback@sourcegraph.com?subject=Feedback on Campaigns">
                                feedback@sourcegraph.com
                            </a>
                            . We look forward to hearing from you!
                        </p>
                    </div>
                    {/* eslint-enable react/jsx-no-bind */}
                </>
            )
        }
    } else if (outerProps.authenticatedUser.siteAdmin) {
        content = <CampaignsSiteAdminMarketingPage {...outerProps} />
    } else {
        content = <CampaignsUserMarketingPage {...outerProps} enableReadAccess={false} />
    }
    return <div className="container mt-4">{content}</div>
})
