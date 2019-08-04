import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import * as React from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'
import { Subscription } from 'rxjs'
import { map } from 'rxjs/operators'
import * as GQL from '../../../../shared/src/graphql/schema'
import { withAuthenticatedUser } from '../../auth/withAuthenticatedUser'
import { ErrorBoundary } from '../../components/ErrorBoundary'
import { HeroPage } from '../../components/HeroPage'
import { siteFlags } from '../../site/backend'
import { ThemeProps } from '../../theme'
import { RouteDescriptor } from '../../util/contributions'
import { UserAreaRouteContext } from '../area/UserArea'
import { UserSettingsSidebar, UserSettingsSidebarItems } from './UserSettingsSidebar'

const NotFoundPage: React.FunctionComponent = () => <HeroPage icon={MapSearchIcon} title="404: Not Found" />

export interface UserSettingsAreaRoute extends RouteDescriptor<UserSettingsAreaRouteContext> {}

export interface UserSettingsAreaProps extends UserAreaRouteContext, RouteComponentProps<{}>, ThemeProps {
    authenticatedUser: GQL.IUser
    sideBarItems: UserSettingsSidebarItems
    routes: readonly UserSettingsAreaRoute[]
}

export interface UserSettingsAreaRouteContext extends UserSettingsAreaProps {
    /**
     * The user who is the subject of the page. This can differ from the authenticatedUser (e.g., when a site admin
     * is viewing another user's account page).
     */
    user: GQL.IUser
    authProviders: GQL.IAuthProvider[]
    newToken?: GQL.ICreateAccessTokenResult
    onDidCreateAccessToken: (value?: GQL.ICreateAccessTokenResult) => void
    onDidPresentNewToken: (value?: GQL.ICreateAccessTokenResult) => void
}

interface UserSettingsAreaState {
    authProviders: GQL.IAuthProvider[]

    /**
     * Holds the newly created access token (from UserSettingsCreateAccessTokenPage), if any. After
     * it is displayed to the user, this subject's value is set back to undefined.
     */
    newlyCreatedAccessToken?: GQL.ICreateAccessTokenResult
}

/**
 * Renders a layout of a sidebar and a content area to display user settings.
 */
export const UserSettingsArea = withAuthenticatedUser(
    class UserSettingsArea extends React.Component<UserSettingsAreaProps, UserSettingsAreaState> {
        public state: UserSettingsAreaState = { authProviders: [] }
        private subscriptions = new Subscription()

        public componentDidMount(): void {
            this.subscriptions.add(
                siteFlags.pipe(map(({ authProviders }) => authProviders)).subscribe(({ nodes }) => {
                    this.setState({ authProviders: nodes })
                })
            )
        }

        public componentWillUnmount(): void {
            this.subscriptions.unsubscribe()
        }

        public render(): JSX.Element | null {
            if (!this.props.user) {
                return null
            }

            if (this.props.authenticatedUser.id !== this.props.user.id && !this.props.user.viewerCanAdminister) {
                return (
                    <HeroPage
                        icon={MapSearchIcon}
                        title="403: Forbidden"
                        subtitle="You are not authorized to view or edit this user's settings."
                    />
                )
            }

            const { children, ...props } = this.props
            const context: UserSettingsAreaRouteContext = {
                ...props,
                newToken: this.state.newlyCreatedAccessToken,
                user: this.props.user,
                onDidCreateAccessToken: this.setNewToken,
                onDidPresentNewToken: this.setNewToken,
                authProviders: this.state.authProviders,
            }

            return (
                <div className="d-flex">
                    <UserSettingsSidebar
                        items={this.props.sideBarItems}
                        authProviders={this.state.authProviders}
                        {...this.props}
                        className="flex-0 mr-3"
                    />
                    <div className="flex-1">
                        <ErrorBoundary location={this.props.location}>
                            <React.Suspense fallback={<LoadingSpinner className="icon-inline m-2" />}>
                                <Switch>
                                    {this.props.routes.map(
                                        ({ path, exact, render, condition = () => true }) =>
                                            condition(context) && (
                                                <Route
                                                    path={this.props.match.url + path}
                                                    key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                                                    exact={exact}
                                                    // tslint:disable-next-line:jsx-no-lambda
                                                    render={routeComponentProps =>
                                                        render({ ...context, ...routeComponentProps })
                                                    }
                                                />
                                            )
                                    )}
                                    <Route component={NotFoundPage} key="hardcoded-key" />
                                </Switch>
                            </React.Suspense>
                        </ErrorBoundary>
                    </div>
                </div>
            )
        }

        private setNewToken = (value?: GQL.ICreateAccessTokenResult): void => {
            this.setState({ newlyCreatedAccessToken: value })
        }
    }
)
