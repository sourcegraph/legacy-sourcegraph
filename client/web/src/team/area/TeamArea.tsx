import * as React from 'react'

import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import { Route, Routes, useLocation, useParams } from 'react-router-dom-v5-compat'

import { LoadingSpinner, ErrorMessage } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../auth'
import { ErrorBoundary } from '../../components/ErrorBoundary'
import { HeroPage } from '../../components/HeroPage'
import { TeamAreaTeamFields } from '../../graphql-operations'
import { RouteV6Descriptor } from '../../util/contributions'

import { useTeam } from './backend'
import type { TeamProfilePageProps } from './TeamProfilePage'
import type { TeamMembersPageProps } from './TeamMembersPage'
import type { TeamChildTeamsPageProps } from './TeamChildTeamsPage'
import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

const TeamProfilePage = lazyComponent<TeamProfilePageProps, 'TeamProfilePage'>(
    () => import('./TeamProfilePage'),
    'TeamProfilePage'
)
const TeamMembersPage = lazyComponent<TeamMembersPageProps, 'TeamMembersPage'>(
    () => import('./TeamMembersPage'),
    'TeamMembersPage'
)
const TeamChildTeamsPage = lazyComponent<TeamChildTeamsPageProps, 'TeamChildTeamsPage'>(
    () => import('./TeamChildTeamsPage'),
    'TeamChildTeamsPage'
)

const NotFoundPage: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => (
    <HeroPage icon={MapSearchIcon} title="404: Not Found" subtitle="Sorry, the requested team was not found." />
)

export interface TeamAreaRoute extends RouteV6Descriptor<TeamAreaRouteContext> {}

export interface TeamAreaProps {
    /**
     * The currently authenticated user.
     */
    authenticatedUser: AuthenticatedUser
}

/**
 * Properties passed to all page components in the team area.
 */
export interface TeamAreaRouteContext {
    /** The team that is the subject of the page. */
    team: TeamAreaTeamFields

    /** Called when the team is updated and must be reloaded. */
    onTeamUpdate: () => void

    /** The currently authenticated user. */
    authenticatedUser: AuthenticatedUser
}

export const TeamArea: React.FunctionComponent<TeamAreaProps> = ({ authenticatedUser }) => {
    const { teamName } = useParams<{ teamName: string }>()

    const location = useLocation()

    const { data, loading, error, refetch } = useTeam(teamName!)

    if (loading) {
        return null
    }
    if (error) {
        return <HeroPage icon={AlertCircleIcon} title="Error" subtitle={<ErrorMessage error={error} />} />
    }

    if (!data?.team) {
        return (
            <HeroPage
                icon={AlertCircleIcon}
                title="Error"
                subtitle={<ErrorMessage error={new Error(`Team not found: ${JSON.stringify(teamName)}`)} />}
            />
        )
    }

    const context: TeamAreaRouteContext = {
        authenticatedUser: authenticatedUser,
        team: data.team,
        onTeamUpdate: refetch,
    }

    return (
        <ErrorBoundary location={location}>
            <React.Suspense fallback={<LoadingSpinner className="m-2" />}>
                <Routes>
                    <Route
                        path=""
                        key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                        element={<TeamProfilePage {...context} />}
                    />
                    <Route
                        path="members"
                        key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                        element={<TeamMembersPage {...context} />}
                    />
                    <Route
                        path="child-teams"
                        key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                        element={<TeamChildTeamsPage {...context} />}
                    />
                    <Route element={<NotFoundPage />} />
                </Routes>
            </React.Suspense>
        </ErrorBoundary>
    )
}
