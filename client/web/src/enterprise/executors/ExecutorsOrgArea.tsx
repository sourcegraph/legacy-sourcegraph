import React from 'react'

import MapSearchIcon from 'mdi-react/MapSearchIcon'
import { Route, RouteComponentProps, Switch } from 'react-router'

import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'
import { Link } from '@sourcegraph/wildcard'

import { HeroPage } from '../../components/HeroPage'
import { Scalars } from '../../graphql-operations'

import type { OrgExecutorSecretsListPageProps } from './secrets/ExecutorSecretsListPage'

const OrgExecutorSecretsListPage = lazyComponent<OrgExecutorSecretsListPageProps, 'OrgExecutorSecretsListPage'>(
    () => import('./secrets/ExecutorSecretsListPage'),
    'OrgExecutorSecretsListPage'
)

export interface ExecutorsOrgAreaProps<RouteProps extends {} = {}> extends RouteComponentProps<RouteProps> {
    namespaceID: Scalars['ID']
}

/** The page area for all executors settings in user settings. */
export const ExecutorsOrgArea: React.FunctionComponent<React.PropsWithChildren<ExecutorsOrgAreaProps>> = ({
    match,
    ...outerProps
}) => (
    <>
        <Switch>
            <Route
                path={`${match.url}/secrets`}
                render={props => (
                    <OrgExecutorSecretsListPage
                        orgID={outerProps.namespaceID}
                        headerLine={
                            <>
                                Configure executor secrets that will only be available to executions in this
                                organization.
                                <br />
                                Global secrets are available to executions in this namespace. Secrets in this namespace
                                with the same name as a global secret will overwrite the global secret. Site admins can
                                configure global secrets{' '}
                                <Link to="/admin/executors/secrets">in site admin settings</Link>.
                            </>
                        }
                        {...outerProps}
                        {...props}
                    />
                )}
                exact={true}
            />
            <Route component={NotFoundPage} key="hardcoded-key" />
        </Switch>
    </>
)

const NotFoundPage: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => (
    <HeroPage icon={MapSearchIcon} title="404: Not Found" />
)
