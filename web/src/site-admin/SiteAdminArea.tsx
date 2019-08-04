import MapSearchIcon from 'mdi-react/MapSearchIcon'
import * as React from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'
import { ActivationProps } from '../../../shared/src/components/activation/Activation'
import * as GQL from '../../../shared/src/graphql/schema'
import { PlatformContextProps } from '../../../shared/src/platform/context'
import { SettingsCascadeProps } from '../../../shared/src/settings/settings'
import { withAuthenticatedUser } from '../auth/withAuthenticatedUser'
import { ErrorBoundary } from '../components/ErrorBoundary'
import { HeroPage } from '../components/HeroPage'
import { RouteDescriptor } from '../util/contributions'
import { SiteAdminSidebar, SiteAdminSideBarGroups } from './SiteAdminSidebar'

const NotFoundPage: React.ComponentType<{}> = () => (
    <HeroPage
        icon={MapSearchIcon}
        title="404: Not Found"
        subtitle="Sorry, the requested site admin page was not found."
    />
)

const NotSiteAdminPage: React.ComponentType<{}> = () => (
    <HeroPage icon={MapSearchIcon} title="403: Forbidden" subtitle="Only site admins are allowed here." />
)

export interface SiteAdminAreaRouteContext extends PlatformContextProps, SettingsCascadeProps, ActivationProps {
    site: Pick<GQL.ISite, '__typename' | 'id'>
    authenticatedUser: GQL.IUser
    isLightTheme: boolean

    /** This property is only used by {@link SiteAdminOverviewPage}. */
    overviewComponents: readonly React.ComponentType[]
}

export interface SiteAdminAreaRoute extends RouteDescriptor<SiteAdminAreaRouteContext> {}

interface SiteAdminAreaProps
    extends RouteComponentProps<{}>,
        PlatformContextProps,
        SettingsCascadeProps,
        ActivationProps {
    routes: readonly SiteAdminAreaRoute[]
    sideBarGroups: SiteAdminSideBarGroups
    overviewComponents: readonly React.ComponentType[]
    authenticatedUser: GQL.IUser
    isLightTheme: boolean
}

const AuthenticatedSiteAdminArea: React.FunctionComponent<SiteAdminAreaProps> = props => {
    // If not site admin, redirect to sign in.
    if (!props.authenticatedUser.siteAdmin) {
        return <NotSiteAdminPage />
    }

    const context: SiteAdminAreaRouteContext = {
        authenticatedUser: props.authenticatedUser,
        platformContext: props.platformContext,
        settingsCascade: props.settingsCascade,
        isLightTheme: props.isLightTheme,
        activation: props.activation,
        site: { __typename: 'Site' as const, id: window.context.siteGQLID },
        overviewComponents: props.overviewComponents,
    }

    return (
        <div className="site-admin-area d-flex container">
            <SiteAdminSidebar className="flex-0 mr-3" groups={props.sideBarGroups} />
            <div className="flex-1">
                <ErrorBoundary location={props.location}>
                    <Switch>
                        {props.routes.map(
                            ({ render, path, exact, condition = () => true }) =>
                                condition(context) && (
                                    <Route
                                        // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                                        key="hardcoded-key"
                                        path={props.match.url + path}
                                        exact={exact}
                                        // tslint:disable-next-line:jsx-no-lambda RouteProps.render is an exception
                                        render={routeComponentProps => render({ ...context, ...routeComponentProps })}
                                    />
                                )
                        )}
                        <Route component={NotFoundPage} />
                    </Switch>
                </ErrorBoundary>
            </div>
        </div>
    )
}

/**
 * Renders a layout of a sidebar and a content area to display site admin information.
 */
export const SiteAdminArea = withAuthenticatedUser(AuthenticatedSiteAdminArea)
