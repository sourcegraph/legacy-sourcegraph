/**
 * Provides convenience functions for interacting with the Sourcegraph API from tests.
 */

import {
    gql,
    dataOrThrowErrors,
    createInvalidGraphQLMutationResponseError,
} from '../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../shared/src/graphql/schema'
import { GraphQLClient } from './GraphQLClient'
import { map, tap, retryWhen, delayWhen, take, mergeMap } from 'rxjs/operators'
import { zip, timer, concat, throwError, defer, Observable } from 'rxjs'
import { CloneInProgressError, ECLONEINPROGESS } from '../../../../shared/src/backend/errors'
import { isErrorLike, createAggregateError } from '../../../../shared/src/util/errors'
import { ResourceDestructor } from './TestResourceManager'
import { Config } from '../../../../shared/src/e2e/config'
import { PlatformContext } from '../../../../shared/src/platform/context'

/**
 * Wait until all repositories in the list exist.
 */
export async function waitForRepos(
    gqlClient: GraphQLClient,
    ensureRepos: string[],
    config?: Partial<Pick<Config, 'logStatusMessages'>>
): Promise<void> {
    await zip(
        // List of Observables that complete after each repository is successfully fetched.
        ...ensureRepos.map(repoName =>
            gqlClient
                .queryGraphQL(
                    gql`
                        query ResolveRev($repoName: String!) {
                            repository(name: $repoName) {
                                mirrorInfo {
                                    cloned
                                }
                            }
                        }
                    `,
                    { repoName }
                )
                .pipe(
                    map(dataOrThrowErrors),
                    // Wait until the repository is cloned even if it doesn't yet exist.
                    // waitForRepos might be called immediately after adding a new external service,
                    // and we have no guarantee that all the repositories from that external service
                    // will exist when the add-external-service endpoint returns.
                    tap(({ repository }) => {
                        if (!repository || !repository.mirrorInfo || !repository.mirrorInfo.cloned) {
                            throw new CloneInProgressError(repoName)
                        }
                    }),
                    retryWhen(errors =>
                        concat(
                            errors.pipe(
                                delayWhen(error => {
                                    if (isErrorLike(error) && error.code === ECLONEINPROGESS) {
                                        // Delay retry by 2s.
                                        if (config && config.logStatusMessages) {
                                            console.log(`Waiting for ${repoName} to finish cloning...`)
                                        }
                                        return timer(2 * 1000)
                                    }
                                    // Throw all errors other than ECLONEINPROGRESS
                                    throw error
                                }),
                                take(60) // Up to 60 retries (an effective timeout of 2 minutes)
                            ),
                            defer(() => throwError(new Error(`Could not resolve repo ${repoName}: too many retries`)))
                        )
                    )
                )
        )
    ).toPromise()
}

export async function ensureNoTestExternalServices(
    gqlClient: GraphQLClient,
    options: {
        kind: GQL.ExternalServiceKind
        uniqueDisplayName: string
        deleteIfExist?: boolean
    }
): Promise<void> {
    if (!options.uniqueDisplayName.startsWith('[TEST]')) {
        throw new Error(
            `Test external service name ${JSON.stringify(options.uniqueDisplayName)} must start with "[TEST]".`
        )
    }

    const externalServices = await getExternalServices(gqlClient, options)
    if (externalServices.length === 0) {
        return
    }
    if (!options.deleteIfExist) {
        throw new Error('external services already exist, not deleting')
    }

    for (const externalService of externalServices) {
        await gqlClient
            .mutateGraphQL(
                gql`
                    mutation DeleteExternalService($externalService: ID!) {
                        deleteExternalService(externalService: $externalService) {
                            alwaysNil
                        }
                    }
                `,
                { externalService: externalService.id }
            )
            .toPromise()
    }
}

/**
 * TODO(beyang): remove this after the corresponding API in the main code has been updated to use a
 * dependency-injected `requestGraphQL`.
 */
export function getExternalServices(
    gqlClient: GraphQLClient,
    options: {
        kind?: GQL.ExternalServiceKind
        uniqueDisplayName?: string
    } = {}
): Promise<GQL.IExternalService[]> {
    return gqlClient
        .queryGraphQL(
            gql`
                query ExternalServices($first: Int) {
                    externalServices(first: $first) {
                        nodes {
                            id
                            kind
                            displayName
                            config
                            createdAt
                            updatedAt
                            warning
                        }
                    }
                }
            `,
            { first: 100 }
        )
        .pipe(
            map(dataOrThrowErrors),
            map(({ externalServices }) =>
                externalServices.nodes.filter(
                    ({ displayName, kind }) =>
                        (options.uniqueDisplayName === undefined || options.uniqueDisplayName === displayName) &&
                        (options.kind === undefined || options.kind === kind)
                )
            )
        )
        .toPromise()
}

export async function ensureTestExternalService(
    gqlClient: GraphQLClient,
    options: {
        kind: GQL.ExternalServiceKind
        uniqueDisplayName: string
        config: Record<string, any>
        waitForRepos?: string[]
    },
    e2eConfig?: Partial<Pick<Config, 'logStatusMessages'>>
): Promise<ResourceDestructor> {
    if (!options.uniqueDisplayName.startsWith('[TEST]')) {
        throw new Error(
            `Test external service name ${JSON.stringify(options.uniqueDisplayName)} must start with "[TEST]".`
        )
    }

    const destroy = (): Promise<void> => ensureNoTestExternalServices(gqlClient, { ...options, deleteIfExist: true })

    const externalServices = await getExternalServices(gqlClient, options)
    if (externalServices.length > 0) {
        return destroy
    }

    // Add a new external service if one doesn't already exist.
    const input: GQL.IAddExternalServiceInput = {
        kind: options.kind,
        displayName: options.uniqueDisplayName,
        config: JSON.stringify(options.config),
    }
    dataOrThrowErrors(
        await gqlClient
            .mutateGraphQL(
                gql`
                    mutation addExternalService($input: AddExternalServiceInput!) {
                        addExternalService(input: $input) {
                            kind
                            displayName
                            config
                        }
                    }
                `,
                { input }
            )
            .toPromise()
    )

    if (options.waitForRepos && options.waitForRepos.length > 0) {
        await waitForRepos(gqlClient, options.waitForRepos, e2eConfig)
    }

    return destroy
}

/**
 * TODO(beyang): remove this after the corresponding API in the main code has been updated to use a
 * dependency-injected `requestGraphQL`.
 */
export async function deleteUser(
    { requestGraphQL }: Pick<PlatformContext, 'requestGraphQL'>,
    username: string,
    mustAlreadyExist: boolean = true
): Promise<void> {
    let user: GQL.IUser | null
    try {
        user = await getUser({ requestGraphQL }, username)
    } catch (err) {
        if (mustAlreadyExist) {
            throw err
        } else {
            return
        }
    }

    if (!user) {
        if (mustAlreadyExist) {
            throw new Error(`Fetched user ${username} was null`)
        } else {
            return
        }
    }

    await requestGraphQL<GQL.IMutation>({
        request: gql`
            mutation DeleteUser($user: ID!, $hard: Boolean) {
                deleteUser(user: $user, hard: $hard) {
                    alwaysNil
                }
            }
        `,
        variables: { hard: false, user: user.id },
        mightContainPrivateInfo: false,
    }).toPromise()
}

/**
 * TODO(beyang): remove this after the corresponding API in the main code has been updated to use a
 * dependency-injected `requestGraphQL`.
 */
export async function setUserSiteAdmin(gqlClient: GraphQLClient, userID: GQL.ID, siteAdmin: boolean): Promise<void> {
    await gqlClient
        .mutateGraphQL(
            gql`
                mutation SetUserIsSiteAdmin($userID: ID!, $siteAdmin: Boolean!) {
                    setUserIsSiteAdmin(userID: $userID, siteAdmin: $siteAdmin) {
                        alwaysNil
                    }
                }
            `,
            { userID, siteAdmin }
        )
        .toPromise()
}

/**
 * TODO(beyang): remove this after the corresponding API in the main code has been updated to use a
 * dependency-injected `requestGraphQL`.
 */
export function currentProductVersion(gqlClient: GraphQLClient): Promise<string> {
    return gqlClient
        .queryGraphQL(
            gql`
                query SiteFlags {
                    site {
                        productVersion
                    }
                }
            `,
            {}
        )
        .pipe(
            map(dataOrThrowErrors),
            map(({ site }) => site.productVersion)
        )
        .toPromise()
}

/**
 * TODO(beyang): remove this after the corresponding API in the main code has been updated to use a
 * dependency-injected `requestGraphQL`.
 */
export function getManagementConsoleState(gqlClient: GraphQLClient): Promise<GQL.IManagementConsoleState> {
    return gqlClient
        .queryGraphQL(
            gql`
                query ManagementConsoleState {
                    site {
                        managementConsoleState {
                            plaintextPassword
                        }
                    }
                }
            `
        )
        .pipe(
            map(dataOrThrowErrors),
            map(({ site }) => site.managementConsoleState)
        )
        .toPromise()
}

/**
 * TODO(beyang): remove this after the corresponding API in the main code has been updated to use a
 * dependency-injected `requestGraphQL`.
 */
export async function setUserEmailVerified(
    gqlClient: GraphQLClient,
    username: string,
    email: string,
    verified: boolean
): Promise<void> {
    const user = await getUser(gqlClient, username)
    if (!user) {
        throw new Error(`User ${username} does not exist`)
    }
    await gqlClient
        .mutateGraphQL(
            gql`
                mutation SetUserEmailVerified($user: ID!, $email: String!, $verified: Boolean!) {
                    setUserEmailVerified(user: $user, email: $email, verified: $verified) {
                        alwaysNil
                    }
                }
            `,
            { user: user.id, email, verified }
        )
        .pipe(map(dataOrThrowErrors))
        .toPromise()
}

/**
 * TODO(beyang): remove this after the corresponding API in the main code has been updated to use a
 * dependency-injected `requestGraphQL`.
 */
export function getViewerSettings({
    requestGraphQL,
}: Pick<PlatformContext, 'requestGraphQL'>): Promise<GQL.ISettingsCascade> {
    return requestGraphQL<GQL.IQuery>({
        request: gql`
            query ViewerSettings {
                viewerSettings {
                    ...SettingsCascadeFields
                }
            }

            fragment SettingsCascadeFields on SettingsCascade {
                subjects {
                    __typename
                    ... on Org {
                        id
                        name
                        displayName
                    }
                    ... on User {
                        id
                        username
                        displayName
                    }
                    ... on Site {
                        id
                        siteID
                    }
                    latestSettings {
                        id
                        contents
                    }
                    settingsURL
                    viewerCanAdminister
                }
                final
            }
        `,
        variables: {},
        mightContainPrivateInfo: true,
    })
        .pipe(
            map(dataOrThrowErrors),
            map(data => data.viewerSettings)
        )
        .toPromise()
}

/**
 * TODO(beyang): remove this after the corresponding API in the main code has been updated to use a
 * dependency-injected `requestGraphQL`.
 */
export function deleteOrganization(
    { requestGraphQL }: Pick<PlatformContext, 'requestGraphQL'>,
    organization: GQL.ID
): Observable<void> {
    return requestGraphQL<GQL.IMutation>({
        request: gql`
            mutation DeleteOrganization($organization: ID!) {
                deleteOrganization(organization: $organization) {
                    alwaysNil
                }
            }
        `,
        variables: { organization },
        mightContainPrivateInfo: true,
    }).pipe(
        map(dataOrThrowErrors),
        map(data => {
            if (!data.deleteOrganization) {
                throw createInvalidGraphQLMutationResponseError('DeleteOrganization')
            }
        })
    )
}

/**
 * TODO(beyang): remove this after the corresponding API in the main code has been updated to use a
 * dependency-injected `requestGraphQL`.
 */
export function fetchAllOrganizations(
    { requestGraphQL }: Pick<PlatformContext, 'requestGraphQL'>,
    args: { first?: number; query?: string }
): Observable<GQL.IOrgConnection> {
    return requestGraphQL<GQL.IQuery>({
        request: gql`
            query Organizations($first: Int, $query: String) {
                organizations(first: $first, query: $query) {
                    nodes {
                        id
                        name
                        displayName
                        createdAt
                        latestSettings {
                            createdAt
                            contents
                        }
                        members {
                            totalCount
                        }
                    }
                    totalCount
                }
            }
        `,
        variables: args,
        mightContainPrivateInfo: true,
    }).pipe(
        map(dataOrThrowErrors),
        map(data => data.organizations)
    )
}

/**
 * TODO(beyang): remove this after the corresponding API in the main code has been updated to use a
 * dependency-injected `requestGraphQL`.
 */
export function createOrganization(
    {
        requestGraphQL,
        eventLogger = { log: () => undefined },
    }: Pick<PlatformContext, 'requestGraphQL'> & {
        eventLogger?: { log: (eventLabel: string, eventProperties?: any) => void }
    },
    variables: {
        /** The name of the organization. */
        name: string
        /** The new organization's display name (e.g. full name) in the organization profile. */
        displayName?: string
    }
): Observable<GQL.IOrg> {
    return requestGraphQL<GQL.IMutation>({
        request: gql`
            mutation createOrganization($name: String!, $displayName: String) {
                createOrganization(name: $name, displayName: $displayName) {
                    id
                    name
                }
            }
        `,
        variables,
        mightContainPrivateInfo: false,
    }).pipe(
        mergeMap(({ data, errors }) => {
            if (!data || !data.createOrganization) {
                eventLogger.log('NewOrgFailed')
                throw createAggregateError(errors)
            }
            eventLogger.log('NewOrgCreated', {
                organization: {
                    org_id: data.createOrganization.id,
                    org_name: data.createOrganization.name,
                },
            })
            return concat([data.createOrganization])
        })
    )
}

/**
 * TODO(beyang): remove this after the corresponding API in the main code has been updated to use a
 * dependency-injected `requestGraphQL`.
 */
export function createUser(
    { requestGraphQL }: Pick<PlatformContext, 'requestGraphQL'>,
    username: string,
    email: string | undefined
): Observable<GQL.ICreateUserResult> {
    return requestGraphQL<GQL.IMutation>({
        request: gql`
            mutation CreateUser($username: String!, $email: String) {
                createUser(username: $username, email: $email) {
                    resetPasswordURL
                }
            }
        `,
        variables: { username, email },
        mightContainPrivateInfo: true,
    }).pipe(
        map(dataOrThrowErrors),
        map(data => data.createUser)
    )
}

/**
 * TODO(beyang): remove this after the corresponding API in the main code has been updated to use a
 * dependency-injected `requestGraphQL`.
 */
export async function getUser(
    { requestGraphQL }: Pick<PlatformContext, 'requestGraphQL'>,
    username: string
): Promise<GQL.IUser | null> {
    const user = await requestGraphQL<GQL.IQuery>({
        request: gql`
            query User($username: String!) {
                user(username: $username) {
                    __typename
                    id
                    username
                    displayName
                    url
                    settingsURL
                    avatarURL
                    viewerCanAdminister
                    siteAdmin
                    createdAt
                    emails {
                        email
                        verified
                    }
                    organizations {
                        nodes {
                            id
                            displayName
                            name
                        }
                    }
                    settingsCascade {
                        subjects {
                            latestSettings {
                                id
                                contents
                            }
                        }
                    }
                }
            }
        `,
        variables: { username },
        mightContainPrivateInfo: true,
    })
        .pipe(
            map(dataOrThrowErrors),
            map(({ user }) => user)
        )
        .toPromise()
    return user
}
