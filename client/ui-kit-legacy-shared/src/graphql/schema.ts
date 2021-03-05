export type ID = string
export type GitObjectID = string
export type DateTime = string
export type JSONCString = string

export interface IGraphQLResponseRoot {
    data?: IQuery | IMutation
    errors?: IGraphQLResponseError[]
}

export interface IGraphQLResponseError {
    /** Required for all errors */
    message: string
    locations?: IGraphQLResponseErrorLocation[]
    /** 7.2.2 says 'GraphQL servers may provide additional entries to error' */
    [propName: string]: any
}

export interface IGraphQLResponseErrorLocation {
    line: number
    column: number
}

/**
 * This type is not returned by any resolver, but serves to document what an error
 * response will look like.
 */
export interface IError {
    __typename: 'Error'

    /**
     * A string giving more context about the error that ocurred.
     */
    message: string

    /**
     * The GraphQL path to where the error happened. For an error in the query
     * query {
     *     user {
     *         externalID # This is a nullable field that failed computing.
     *     }
     * }
     * the path would be ["user", "externalID"].
     */
    path: string[]

    /**
     * Optional additional context on the error.
     */
    extensions: IErrorExtensions | null
}

/**
 * Optional additional context on an error returned from a resolver.
 * It may also contain more properties, which aren't strictly typed here.
 */
export interface IErrorExtensions {
    __typename: 'ErrorExtensions'

    /**
     * An error code, which can be asserted on.
     * Possible error codes are communicated in the doc string of the field.
     */
    code: string | null
}

/**
 * Represents a null return value.
 */
export interface IEmptyResponse {
    __typename: 'EmptyResponse'

    /**
     * A dummy null value.
     */
    alwaysNil: string | null
}

/**
 * An object with an ID.
 */
export type Node =
    | IOutOfBandMigration
    | ISavedSearch
    | IMonitor
    | IMonitorQuery
    | IMonitorTriggerEvent
    | IMonitorEmail
    | IMonitorActionEvent
    | IExternalService
    | IRepository
    | ISearchContext
    | IGitRef
    | IGitCommit
    | IUser
    | IAccessToken
    | IExternalAccount
    | IOrg
    | IOrganizationInvitation
    | IRegistryExtension
    | ILSIFUpload
    | ILSIFIndex
    | IProductSubscription
    | IProductLicense
    | ICampaign
    | IHiddenExternalChangeset
    | IExternalChangeset
    | IChangesetEvent
    | ICampaignsCredential
    | IHiddenChangesetSpec
    | IVisibleChangesetSpec
    | ICampaignSpec
    | IBatchChange
    | IBatchSpec
    | IBatchChangesCredential

/**
 * An object with an ID.
 */
export interface INode {
    __typename: 'Node'

    /**
     * The ID of the node.
     */
    id: ID
}

/**
 * A mutation.
 */
export interface IMutation {
    __typename: 'Mutation'

    /**
     * Updates the user profile information for the user with the given ID.
     *
     * Only the user and site admins may perform this mutation.
     */
    updateUser: IUser

    /**
     * Creates an organization. The caller is added as a member of the newly created organization.
     *
     * Only authenticated users may perform this mutation.
     */
    createOrganization: IOrg

    /**
     * Updates an organization.
     *
     * Only site admins and any member of the organization may perform this mutation.
     */
    updateOrganization: IOrg

    /**
     * Deletes an organization. Only site admins may perform this mutation.
     */
    deleteOrganization: IEmptyResponse | null

    /**
     * Adds a external service. Only site admins may perform this mutation.
     */
    addExternalService: IExternalService

    /**
     * Updates a external service. Only site admins may perform this mutation.
     */
    updateExternalService: IExternalService

    /**
     * Delete an external service. Only site admins may perform this mutation.
     */
    deleteExternalService: IEmptyResponse

    /**
     * DEPRECATED: All repositories are accessible or deleted. To prevent a
     * repository from being accessed on Sourcegraph add it to the external
     * service exclude configuration. This mutation will be removed in 3.6.
     *
     * Enables or disables a repository. A disabled repository is only accessible
     * to site admins and never appears in search results.
     *
     * Only site admins may perform this mutation.
     * @deprecated "update external service exclude setting."
     */
    setRepositoryEnabled: IEmptyResponse | null

    /**
     * Tests the connection to a mirror repository's original source repository. This is an
     * expensive and slow operation, so it should only be used for interactive diagnostics.
     *
     * Only site admins may perform this mutation.
     */
    checkMirrorRepositoryConnection: ICheckMirrorRepositoryConnectionResult

    /**
     * Schedule the mirror repository to be updated from its original source repository. Updating
     * occurs automatically, so this should not normally be needed.
     *
     * Only site admins may perform this mutation.
     */
    updateMirrorRepository: IEmptyResponse

    /**
     * Creates a new user account.
     *
     * Only site admins may perform this mutation.
     */
    createUser: ICreateUserResult

    /**
     * Randomize a user's password so that they need to reset it before they can sign in again.
     *
     * Only site admins may perform this mutation.
     */
    randomizeUserPassword: IRandomizeUserPasswordResult

    /**
     * Adds an email address to the user's account. The email address will be marked as unverified until the user
     * has followed the email verification process.
     *
     * Only the user and site admins may perform this mutation.
     */
    addUserEmail: IEmptyResponse

    /**
     * Removes an email address from the user's account.
     *
     * Only the user and site admins may perform this mutation.
     */
    removeUserEmail: IEmptyResponse

    /**
     * Set an email address as the user's primary.
     *
     * Only the user and site admins may perform this mutation.
     */
    setUserEmailPrimary: IEmptyResponse

    /**
     * Manually set the verification status of a user's email, without going through the normal verification process
     * (of clicking on a link in the email with a verification code).
     *
     * Only site admins may perform this mutation.
     */
    setUserEmailVerified: IEmptyResponse

    /**
     * Resend a verification email, no op if the email is already verified.
     *
     * Only the user and site admins may perform this mutation.
     */
    resendVerificationEmail: IEmptyResponse

    /**
     * Deletes a user account. Only site admins may perform this mutation.
     *
     * If hard == true, a hard delete is performed. By default, deletes are
     * 'soft deletes' and could theoretically be undone with manual DB commands.
     * If a hard delete is performed, the data is truly removed from the
     * database and deletion can NEVER be undone.
     *
     * Data that is deleted as part of this operation:
     *
     * - All user data (access tokens, email addresses, external account info, survey responses, etc)
     * - Organization membership information (which organizations the user is a part of, any invitations created by or targeting the user).
     * - Sourcegraph extensions published by the user.
     * - User, Organization, or Global settings authored by the user.
     */
    deleteUser: IEmptyResponse | null

    /**
     * Updates the current user's password. The oldPassword arg must match the user's current password.
     */
    updatePassword: IEmptyResponse | null

    /**
     * Creates a password for the current user. It is only permitted if the user does not have a password and
     * they don't have any login connections.
     */
    createPassword: IEmptyResponse | null

    /**
     * Creates an access token that grants the privileges of the specified user (referred to as the access token's
     * "subject" user after token creation). The result is the access token value, which the caller is responsible
     * for storing (it is not accessible by Sourcegraph after creation).
     *
     * The supported scopes are:
     *
     * - "user:all": Full control of all resources accessible to the user account.
     * - "site-admin:sudo": Ability to perform any action as any other user. (Only site admins may create tokens
     *   with this scope.)
     *
     * Only the user or site admins may perform this mutation.
     */
    createAccessToken: ICreateAccessTokenResult

    /**
     * Deletes and immediately revokes the specified access token, specified by either its ID or by the token
     * itself.
     *
     * Only site admins or the user who owns the token may perform this mutation.
     */
    deleteAccessToken: IEmptyResponse

    /**
     * Deletes the association between an external account and its Sourcegraph user. It does NOT delete the external
     * account on the external service where it resides.
     *
     * Only site admins or the user who is associated with the external account may perform this mutation.
     */
    deleteExternalAccount: IEmptyResponse

    /**
     * Invite the user with the given username to join the organization. The invited user account must already
     * exist.
     *
     * Only site admins and any organization member may perform this mutation.
     */
    inviteUserToOrganization: IInviteUserToOrganizationResult

    /**
     * Accept or reject an existing organization invitation.
     *
     * Only the recipient of the invitation may perform this mutation.
     */
    respondToOrganizationInvitation: IEmptyResponse

    /**
     * Resend the notification about an organization invitation to the recipient.
     *
     * Only site admins and any member of the organization may perform this mutation.
     */
    resendOrganizationInvitationNotification: IEmptyResponse

    /**
     * Revoke an existing organization invitation.
     *
     * If the invitation has been accepted or rejected, it may no longer be revoked. After an
     * invitation is revoked, the recipient may not accept or reject it. Both cases yield an error.
     *
     * Only site admins and any member of the organization may perform this mutation.
     */
    revokeOrganizationInvitation: IEmptyResponse

    /**
     * Immediately add a user as a member to the organization, without sending an invitation email.
     *
     * Only site admins may perform this mutation. Organization members may use the inviteUserToOrganization
     * mutation to invite users.
     */
    addUserToOrganization: IEmptyResponse

    /**
     * Removes a user as a member from an organization.
     *
     * Only site admins and any member of the organization may perform this mutation.
     */
    removeUserFromOrganization: IEmptyResponse | null

    /**
     * Adds or removes a tag on a user.
     *
     * Tags are used internally by Sourcegraph as feature flags for experimental features.
     *
     * Only site admins may perform this mutation.
     */
    setTag: IEmptyResponse

    /**
     * Adds a Phabricator repository to Sourcegraph.
     */
    addPhabricatorRepo: IEmptyResponse | null

    /**
     * Resolves a revision for a given diff from Phabricator.
     */
    resolvePhabricatorDiff: IGitCommit | null

    /**
     * Logs a user event.
     * @deprecated "use logEvent instead"
     */
    logUserEvent: IEmptyResponse | null

    /**
     * Logs an event.
     */
    logEvent: IEmptyResponse | null

    /**
     * Sends a test notification for the saved search. Be careful: this will send a notifcation (email and other
     * types of notifications, if configured) to all subscribers of the saved search, which could be bothersome.
     *
     * Only subscribers to this saved search may perform this action.
     */
    sendSavedSearchTestNotification: IEmptyResponse | null

    /**
     * All mutations that update settings (global, organization, and user settings) are under this field.
     *
     * Only the settings subject whose settings are being mutated (and site admins) may perform this mutation.
     *
     * This mutation only affects global, organization, and user settings, not site configuration. For site
     * configuration (which is a separate set of configuration properties from global/organization/user settings),
     * use updateSiteConfiguration.
     */
    settingsMutation: ISettingsMutation | null

    /**
     * DEPRECATED: Use settingsMutation instead. This field is a deprecated alias for settingsMutation and will be
     * removed in a future release.
     * @deprecated "use settingsMutation instead"
     */
    configurationMutation: ISettingsMutation | null

    /**
     * Updates the site configuration. Returns whether or not a restart is required for the update to be applied.
     *
     * Only site admins may perform this mutation.
     */
    updateSiteConfiguration: boolean

    /**
     * Sets whether the user with the specified user ID is a site admin.
     *
     * Only site admins may perform this mutation.
     */
    setUserIsSiteAdmin: IEmptyResponse | null

    /**
     * Invalidates all sessions belonging to a user.
     *
     * Only site admins may perform this mutation.
     */
    invalidateSessionsByID: IEmptyResponse | null

    /**
     * Reloads the site by restarting the server. This is not supported for all deployment
     * types. This may cause downtime.
     *
     * Only site admins may perform this mutation.
     */
    reloadSite: IEmptyResponse | null

    /**
     * Submits a user satisfaction (NPS) survey.
     */
    submitSurvey: IEmptyResponse | null

    /**
     * Submits happiness feedback.
     */
    submitHappinessFeedback: IEmptyResponse | null

    /**
     * Submits a request for a Sourcegraph Enterprise trial license.
     */
    requestTrial: IEmptyResponse | null

    /**
     * Manages the extension registry.
     */
    extensionRegistry: IExtensionRegistryMutation

    /**
     * Mutations that are only used on Sourcegraph.com.
     *
     * FOR INTERNAL USE ONLY.
     */
    dotcom: IDotcomMutation

    /**
     * Creates a saved search.
     */
    createSavedSearch: ISavedSearch

    /**
     * Updates a saved search
     */
    updateSavedSearch: ISavedSearch

    /**
     * Deletes a saved search
     */
    deleteSavedSearch: IEmptyResponse | null

    /**
     * (experimental) The LSIF API may change substantially in the near future as we
     * continue to adjust it for our use cases. Changes will not be documented in the
     * CHANGELOG during this time.
     * Deletes an LSIF upload.
     */
    deleteLSIFUpload: IEmptyResponse | null

    /**
     * (experimental) The LSIF API may change substantially in the near future as we
     * continue to adjust it for our use cases. Changes will not be documented in the
     * CHANGELOG during this time.
     * Deletes an LSIF index.
     */
    deleteLSIFIndex: IEmptyResponse | null

    /**
     * Updates the indexing configuration associated with a repository.
     */
    updateRepositoryIndexConfiguration: IEmptyResponse | null

    /**
     * Queues the index jobs for a repository for execution.
     */
    queueAutoIndexJobForRepo: IEmptyResponse | null

    /**
     * Set the permissions of a repository (i.e., which users may view it on Sourcegraph). This
     * operation overwrites the previous permissions for the repository.
     */
    setRepositoryPermissionsForUsers: IEmptyResponse

    /**
     * Schedule a permissions sync for given repository. This queries the repository's code host for
     * all users' permissions associated with the repository, so that the current permissions apply
     * to all users' operations on that repository on Sourcegraph.
     */
    scheduleRepositoryPermissionsSync: IEmptyResponse

    /**
     * Schedule a permissions sync for given user. This queries all code hosts for the user's current
     * repository permissions and syncs them to Sourcegraph, so that the current permissions apply to
     * the user's operations on Sourcegraph.
     */
    scheduleUserPermissionsSync: IEmptyResponse

    /**
     * OBSERVABILITY
     *
     * Set the status of a test alert of the specified parameters - useful for validating
     * 'observability.alerts' configuration. Alerts may take up to a minute to fire.
     */
    triggerObservabilityTestAlert: IEmptyResponse

    /**
     * Create a code monitor.
     */
    createCodeMonitor: IMonitor

    /**
     * Set a code monitor to active/inactive.
     */
    toggleCodeMonitor: IMonitor

    /**
     * Delete a code monitor.
     */
    deleteCodeMonitor: IEmptyResponse

    /**
     * Update a code monitor. We assume that the request contains a complete code monitor,
     * including its trigger and all actions. Actions which are stored in the database,
     * but are missing from the request will be deleted from the database. Actions with id=null
     * will be created.
     */
    updateCodeMonitor: IMonitor

    /**
     * Set the repos synced by an external service
     */
    setExternalServiceRepos: IEmptyResponse

    /**
     * Reset the timestamps of a trigger query. The query will be queued immediately and return
     * all results without a limit on the timeframe. Only site admins may perform this mutation.
     */
    resetTriggerQueryTimestamps: IEmptyResponse

    /**
     * Triggers a test email for a code monitor action.
     */
    triggerTestEmailAction: IEmptyResponse

    /**
     * Updates an out-of-band migration to run in a particular direction.
     *
     * Applied in the forward direction, an out-of-band migration migrates data into a format that
     * is readable by newer Sourcegraph instances. This may be destructive or non-destructive process,
     * depending on the nature and implementation of the migration.
     *
     * Applied in the reverse direction, an out-of-band migration ensures that data is moved back into
     * a format that is readable by the previous Sourcegraph instance. Recently introduced migrations
     * should be applied in reverse prior to downgrading the instance.
     */
    SetMigrationDirection: IEmptyResponse

    /**
     * SetUserPublicRepos sets the list of public repos for a user's search context, ensuring those repos
     * exist and are cloned
     */
    SetUserPublicRepos: IEmptyResponse

    /**
     * Create a campaign from a campaign spec and locally computed changeset specs. The newly created
     * campaign is returned.
     * If a campaign in the same namespace with the same name already exists, an error with the error code
     * ErrMatchingCampaignExists is returned.
     * @deprecated "campaigns have been renamed to batch changes. Use createBatchChange instead."
     */
    createCampaign: ICampaign

    /**
     * Create or update a campaign from a campaign spec and locally computed changeset specs. If no
     * campaign exists in the namespace with the name given in the campaign spec, a campaign will be
     * created. Otherwise, the existing campaign will be updated. The campaign is returned.
     * Closed campaigns cannot be applied to. In that case, an error with the error code ErrApplyClosedCampaign
     * will be returned.
     * @deprecated "campaigns have been renamed to batch changes. Use applyBatchChange instead."
     */
    applyCampaign: ICampaign

    /**
     * Move a campaign to a different namespace, or rename it in the current namespace.
     * @deprecated "campaigns have been renamed to batch changes. Use moveBatchChange instead."
     */
    moveCampaign: ICampaign

    /**
     * Close a campaign.
     * @deprecated "campaigns have been renamed to batch changes. Use closeBatchChange instead."
     */
    closeCampaign: ICampaign

    /**
     * Delete a campaign. A deleted campaign is completely removed and can't be un-deleted. The
     * campaign's changesets are kept as-is; to close them, use the closeCampaign mutation first.
     * @deprecated "campaigns have been renamed to batch changes. Use deleteBatchChange instead."
     */
    deleteCampaign: IEmptyResponse | null

    /**
     * Upload a changeset spec that will be used in a future update to a campaign. The changeset spec
     * is stored and can be referenced by its ID in the applyCampaign mutation. Just uploading the
     * changeset spec does not result in changes to the campaign or any of its changesets; you need
     * to call applyCampaign to use it.
     *
     * You can use this mutation to upload large changeset specs (e.g., containing large diffs) in
     * individual HTTP requests. Then, in the eventual applyCampaign call, you just refer to the
     * changeset specs by their IDs. This lets you avoid problems when updating large campaigns where
     * a large HTTP request body (e.g., with many large diffs in the changeset specs) would be
     * rejected by the web server/proxy or would be very slow.
     *
     * The returned ChangesetSpec is immutable and expires after a certain period of time (if not
     * used in a call to applyCampaign), which can be queried on ChangesetSpec.expiresAt.
     */
    createChangesetSpec: ChangesetSpec

    /**
     * Create a campaign spec that will be used to create a campaign (with the createCampaign
     * mutation), or to update a campaign (with the applyCampaign mutation).
     *
     * The returned CampaignSpec is immutable and expires after a certain period of time (if not used
     * in a call to applyCampaign), which can be queried on CampaignSpec.expiresAt.
     *
     * If campaigns are unlicensed and the number of changesetSpecIDs is higher than what's allowed in
     * the free tier, an error with the error code ErrCampaignsUnlicensed is returned.
     * @deprecated "campaigns have been renamed to batch changes. Use createBatchSpec instead."
     */
    createCampaignSpec: ICampaignSpec

    /**
     * Enqueue the given changeset for high-priority syncing.
     */
    syncChangeset: IEmptyResponse

    /**
     * Re-enqueue the changeset for processing by the reconciler. The changeset must be in FAILED state.
     */
    reenqueueChangeset: Changeset

    /**
     * Create a new credential for the given user for the given code host.
     * If another token for that code host already exists, an error with the error code
     * ErrDuplicateCredential is returned.
     * @deprecated "campaigns have been renamed to batch changes. Use batchChanges instead."
     */
    createCampaignsCredential: ICampaignsCredential

    /**
     * Hard-deletes a given campaigns credential.
     * @deprecated "campaigns have been renamed to batch changes. Use batchChanges instead."
     */
    deleteCampaignsCredential: IEmptyResponse

    /**
     * Create a batch change from a batch spec and locally computed changeset specs. The newly created
     * batch change is returned.
     * If a batch change in the same namespace with the same name already exists,
     * an error with the error code ErrMatchingBatchChangeExists is returned.
     */
    createBatchChange: IBatchChange

    /**
     * Create a batch spec that will be used to create a campaign (with the createCampaign
     * mutation), or to update a campaign (with the applyCampaign mutation).
     *
     * The returned BatchSpec is immutable and expires after a certain period of time (if not used
     * in a call to applyCampaign), which can be queried on BatchSpec.expiresAt.
     *
     * If campaigns are unlicensed and the number of changesetSpecIDs is higher than what's allowed in
     * the free tier, an error with the error code ErrCampaignsUnlicensed is returned.
     */
    createBatchSpec: IBatchSpec

    /**
     * Create or update a batch change from a batch spec and locally computed changeset specs. If no
     * batch change exists in the namespace with the name given in the batch spec, a batch change will be
     * created. Otherwise, the existing batch change will be updated. The batch change is returned.
     * Closed batch changes cannot be applied to. In that case, an error with the error code ErrApplyClosedbatch change
     * will be returned.
     */
    applyBatchChange: IBatchChange

    /**
     * Close a batch change.
     */
    closeBatchChange: IBatchChange

    /**
     * Move a batch change to a different namespace, or rename it in the current namespace.
     */
    moveBatchChange: IBatchChange

    /**
     * Delete a batch change. A deleted batch change is completely removed and can't be un-deleted. The
     * batch change's changesets are kept as-is; to close them, use the closeBatchChange mutation first.
     */
    deleteBatchChange: IEmptyResponse | null

    /**
     * Create a new credential for the given user for the given code host.
     * If another token for that code host already exists, an error with the error code
     * ErrDuplicateCredential is returned.
     */
    createBatchChangesCredential: IBatchChangesCredential

    /**
     * Hard-deletes a given credential.
     */
    deleteBatchChangesCredential: IEmptyResponse
}

export interface IUpdateUserOnMutationArguments {
    user: ID
    username?: string | null
    displayName?: string | null
    avatarURL?: string | null
}

export interface ICreateOrganizationOnMutationArguments {
    name: string
    displayName?: string | null
}

export interface IUpdateOrganizationOnMutationArguments {
    id: ID
    displayName?: string | null
}

export interface IDeleteOrganizationOnMutationArguments {
    organization: ID
}

export interface IAddExternalServiceOnMutationArguments {
    input: IAddExternalServiceInput
}

export interface IUpdateExternalServiceOnMutationArguments {
    input: IUpdateExternalServiceInput
}

export interface IDeleteExternalServiceOnMutationArguments {
    externalService: ID
}

export interface ISetRepositoryEnabledOnMutationArguments {
    repository: ID
    enabled: boolean
}

export interface ICheckMirrorRepositoryConnectionOnMutationArguments {
    /**
     * The ID of the existing repository whose mirror to check.
     */
    repository?: ID | null

    /**
     * The name of a repository whose mirror to check. If the name is provided, the repository need not be added
     * to the site (but the site configuration must define a code host that knows how to handle the name).
     */
    name?: string | null
}

export interface IUpdateMirrorRepositoryOnMutationArguments {
    /**
     * The mirror repository to update.
     */
    repository: ID
}

export interface ICreateUserOnMutationArguments {
    /**
     * The new user's username.
     */
    username: string

    /**
     * The new user's optional email address. If given, it is marked as verified.
     */
    email?: string | null
}

export interface IRandomizeUserPasswordOnMutationArguments {
    user: ID
}

export interface IAddUserEmailOnMutationArguments {
    user: ID
    email: string
}

export interface IRemoveUserEmailOnMutationArguments {
    user: ID
    email: string
}

export interface ISetUserEmailPrimaryOnMutationArguments {
    user: ID
    email: string
}

export interface ISetUserEmailVerifiedOnMutationArguments {
    user: ID
    email: string
    verified: boolean
}

export interface IResendVerificationEmailOnMutationArguments {
    user: ID
    email: string
}

export interface IDeleteUserOnMutationArguments {
    user: ID
    hard?: boolean | null
}

export interface IUpdatePasswordOnMutationArguments {
    oldPassword: string
    newPassword: string
}

export interface ICreatePasswordOnMutationArguments {
    newPassword: string
}

export interface ICreateAccessTokenOnMutationArguments {
    user: ID
    scopes: string[]
    note: string
}

export interface IDeleteAccessTokenOnMutationArguments {
    byID?: ID | null
    byToken?: string | null
}

export interface IDeleteExternalAccountOnMutationArguments {
    externalAccount: ID
}

export interface IInviteUserToOrganizationOnMutationArguments {
    organization: ID
    username: string
}

export interface IRespondToOrganizationInvitationOnMutationArguments {
    /**
     * The organization invitation.
     */
    organizationInvitation: ID

    /**
     * The response to the invitation.
     */
    responseType: OrganizationInvitationResponseType
}

export interface IResendOrganizationInvitationNotificationOnMutationArguments {
    /**
     * The organization invitation.
     */
    organizationInvitation: ID
}

export interface IRevokeOrganizationInvitationOnMutationArguments {
    /**
     * The organization invitation.
     */
    organizationInvitation: ID
}

export interface IAddUserToOrganizationOnMutationArguments {
    organization: ID
    username: string
}

export interface IRemoveUserFromOrganizationOnMutationArguments {
    user: ID
    organization: ID
}

export interface ISetTagOnMutationArguments {
    /**
     * The ID of the user whose tags to set.
     *
     * (This parameter is named "node" to make it easy to support tagging other types of nodes
     * other than users in the future.)
     */
    node: ID

    /**
     * The tag to set.
     */
    tag: string

    /**
     * The desired state of the tag on the user (whether to add or remove): true to add, false to
     * remove.
     */
    present: boolean
}

export interface IAddPhabricatorRepoOnMutationArguments {
    /**
     * The callsign, for example "MUX".
     */
    callsign: string

    /**
     * The name, for example "github.com/gorilla/mux".
     */
    name?: string | null

    /**
     * An alias for name. DEPRECATED: use name instead.
     */
    uri?: string | null

    /**
     * The URL to the phabricator instance (e.g. http://phabricator.sgdev.org).
     */
    url: string
}

export interface IResolvePhabricatorDiffOnMutationArguments {
    /**
     * The name of the repository that the diff is based on.
     */
    repoName: string

    /**
     * The ID of the diff on Phabricator.
     */
    diffID: ID

    /**
     * The base revision this diff is based on.
     */
    baseRev: string

    /**
     * The raw contents of the diff from Phabricator.
     * Required if Sourcegraph doesn't have a Conduit API token.
     */
    patch?: string | null

    /**
     * The description of the diff. This will be used as the commit message.
     */
    description?: string | null

    /**
     * The name of author of the diff.
     */
    authorName?: string | null

    /**
     * The author's email.
     */
    authorEmail?: string | null

    /**
     * When the diff was created.
     */
    date?: string | null
}

export interface ILogUserEventOnMutationArguments {
    event: UserEvent
    userCookieID: string
}

export interface ILogEventOnMutationArguments {
    /**
     * The name of the event.
     */
    event: string

    /**
     * The randomly generated unique user ID stored in a browser cookie.
     */
    userCookieID: string

    /**
     * The URL when the event was logged.
     */
    url: string

    /**
     * The source of the event.
     */
    source: EventSource

    /**
     * The additional argument information.
     */
    argument?: string | null
}

export interface ISendSavedSearchTestNotificationOnMutationArguments {
    /**
     * ID of the saved search.
     */
    id: ID
}

export interface ISettingsMutationOnMutationArguments {
    input: ISettingsMutationGroupInput
}

export interface IConfigurationMutationOnMutationArguments {
    input: ISettingsMutationGroupInput
}

export interface IUpdateSiteConfigurationOnMutationArguments {
    /**
     * The last ID of the site configuration that is known by the client, to
     * prevent race conditions. An error will be returned if someone else
     * has already written a new update.
     */
    lastID: number

    /**
     * A JSON object containing the entire site configuration. The previous site configuration will be replaced
     * with this new value.
     */
    input: string
}

export interface ISetUserIsSiteAdminOnMutationArguments {
    userID: ID
    siteAdmin: boolean
}

export interface IInvalidateSessionsByIDOnMutationArguments {
    userID: ID
}

export interface ISubmitSurveyOnMutationArguments {
    input: ISurveySubmissionInput
}

export interface ISubmitHappinessFeedbackOnMutationArguments {
    input: IHappinessFeedbackSubmissionInput
}

export interface IRequestTrialOnMutationArguments {
    email: string
}

export interface ICreateSavedSearchOnMutationArguments {
    description: string
    query: string
    notifyOwner: boolean
    notifySlack: boolean
    orgID?: ID | null
    userID?: ID | null
}

export interface IUpdateSavedSearchOnMutationArguments {
    id: ID
    description: string
    query: string
    notifyOwner: boolean
    notifySlack: boolean
    orgID?: ID | null
    userID?: ID | null
}

export interface IDeleteSavedSearchOnMutationArguments {
    id: ID
}

export interface IDeleteLSIFUploadOnMutationArguments {
    id: ID
}

export interface IDeleteLSIFIndexOnMutationArguments {
    id: ID
}

export interface IUpdateRepositoryIndexConfigurationOnMutationArguments {
    repository: ID
    configuration: string
}

export interface IQueueAutoIndexJobForRepoOnMutationArguments {
    repository: ID
}

export interface ISetRepositoryPermissionsForUsersOnMutationArguments {
    /**
     * The repository whose permissions to set.
     */
    repository: ID

    /**
     * A list of user identifiers and their repository permissions, which defines the set of
     * users who may view the repository. All users not included in the list will not be
     * permitted to view the repository on Sourcegraph.
     */
    userPermissions: IUserPermission[]
}

export interface IScheduleRepositoryPermissionsSyncOnMutationArguments {
    repository: ID
}

export interface IScheduleUserPermissionsSyncOnMutationArguments {
    user: ID
}

export interface ITriggerObservabilityTestAlertOnMutationArguments {
    /**
     * Level of alert to test - either warning or critical.
     */
    level: string
}

export interface ICreateCodeMonitorOnMutationArguments {
    /**
     * A monitor.
     */
    monitor: IMonitorInput

    /**
     * A trigger.
     */
    trigger: IMonitorTriggerInput

    /**
     * A list of actions.
     */
    actions: IMonitorActionInput[]
}

export interface IToggleCodeMonitorOnMutationArguments {
    /**
     * The id of a code monitor.
     */
    id: ID

    /**
     * Whether the code monitor should be enabled or not.
     */
    enabled: boolean
}

export interface IDeleteCodeMonitorOnMutationArguments {
    /**
     * The id of a code monitor.
     */
    id: ID
}

export interface IUpdateCodeMonitorOnMutationArguments {
    /**
     * The input required to edit a monitor.
     */
    monitor: IMonitorEditInput

    /**
     * The input required to edit the trigger of a monitor. You can only edit triggers that are
     * associated with the monitor (value of field monitor).
     */
    trigger: IMonitorEditTriggerInput

    /**
     * The input required to edit the actions of a monitor. You can only edit actions that are
     * associated with the monitor (value of field monitor).
     */
    actions: IMonitorEditActionInput[]
}

export interface ISetExternalServiceReposOnMutationArguments {
    id: ID
    repos?: string[] | null
    allRepos: boolean
}

export interface IResetTriggerQueryTimestampsOnMutationArguments {
    /**
     * The id of the trigger query.
     */
    id: ID
}

export interface ITriggerTestEmailActionOnMutationArguments {
    namespace: ID
    description: string
    email: IMonitorEmailInput
}

export interface ISetMigrationDirectionOnMutationArguments {
    id: ID
    applyReverse: boolean
}

export interface ISetUserPublicReposOnMutationArguments {
    userID: ID
    repoURIs: string[]
}

export interface ICreateCampaignOnMutationArguments {
    /**
     * The campaign spec that describes the desired state of the campaign.
     */
    campaignSpec: ID
}

export interface IApplyCampaignOnMutationArguments {
    /**
     * The campaign spec that describes the new desired state of the campaign.
     */
    campaignSpec: ID

    /**
     * If set, return an error if the campaign identified using the namespace and campaignSpec
     * parameters does not match the campaign with this ID. This lets callers use a stable ID
     * that refers to a specific campaign during an edit session (and is not susceptible to
     * conflicts if the underlying campaign is moved to a different namespace, renamed, or
     * deleted). The returned error has the error code ErrEnsureCampaignFailed.
     */
    ensureCampaign?: ID | null
}

export interface IMoveCampaignOnMutationArguments {
    campaign: ID
    newName?: string | null
    newNamespace?: ID | null
}

export interface ICloseCampaignOnMutationArguments {
    campaign: ID

    /**
     * Whether to close the changesets associated with this campaign on their respective code
     * hosts. "Close" means the appropriate final state on the code host (e.g., "closed" on
     * GitHub and "declined" on Bitbucket Server).
     * @default false
     */
    closeChangesets?: boolean | null
}

export interface IDeleteCampaignOnMutationArguments {
    campaign: ID
}

export interface ICreateChangesetSpecOnMutationArguments {
    /**
     * The raw changeset spec (as JSON). See
     * https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/schema/changeset_spec.schema.json
     * for the JSON Schema that describes the structure of this input.
     */
    changesetSpec: string
}

export interface ICreateCampaignSpecOnMutationArguments {
    /**
     * The namespace (either a user or organization). A campaign spec can only be applied to (or
     * used to create) campaigns in this namespace.
     */
    namespace: ID

    /**
     * The campaign spec as YAML (or the equivalent JSON). See
     * https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/schema/campaign_spec.schema.json
     * for the JSON Schema that describes the structure of this input.
     */
    campaignSpec: string

    /**
     * Changeset specs that were locally computed and then uploaded using createChangesetSpec.
     */
    changesetSpecs: ID[]
}

export interface ISyncChangesetOnMutationArguments {
    changeset: ID
}

export interface IReenqueueChangesetOnMutationArguments {
    changeset: ID
}

export interface ICreateCampaignsCredentialOnMutationArguments {
    /**
     * The user for which to create the credential.
     */
    user: ID

    /**
     * The kind of external service being configured.
     */
    externalServiceKind: ExternalServiceKind

    /**
     * The URL of the external service being configured.
     */
    externalServiceURL: string

    /**
     * The credential to be stored. This can never be retrieved through the API and will be stored encrypted.
     */
    credential: string
}

export interface IDeleteCampaignsCredentialOnMutationArguments {
    campaignsCredential: ID
}

export interface ICreateBatchChangeOnMutationArguments {
    /**
     * The batch spec that describes the desired state of the batch change.
     */
    batchSpec: ID
}

export interface ICreateBatchSpecOnMutationArguments {
    /**
     * The namespace (either a user or organization). A batch spec can only be applied to (or
     * used to create) campaigns in this namespace.
     */
    namespace: ID

    /**
     * The batch spec as YAML (or the equivalent JSON). See
     * https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/schema/campaign_spec.schema.json
     * for the JSON Schema that describes the structure of this input.
     */
    batchSpec: string

    /**
     * Changeset specs that were locally computed and then uploaded using createChangesetSpec.
     */
    changesetSpecs: ID[]
}

export interface IApplyBatchChangeOnMutationArguments {
    /**
     * The batch spec that describes the new desired state of the batch change.
     */
    batchSpec: ID

    /**
     * If set, return an error if the batch change identified using the namespace and batch changeSpec
     * parameters does not match the batch change with this ID. This lets callers use a stable ID
     * that refers to a specific batch change during an edit session (and is not susceptible to
     * conflicts if the underlying batch change is moved to a different namespace, renamed, or
     * deleted). The returned error has the error code ErrEnsureBatchChangeFailed.
     */
    ensureBatchChange?: ID | null
}

export interface ICloseBatchChangeOnMutationArguments {
    batchChange: ID

    /**
     * Whether to close the changesets associated with this batch change on their respective code
     * hosts. "Close" means the appropriate final state on the code host (e.g., "closed" on
     * GitHub and "declined" on Bitbucket Server).
     * @default false
     */
    closeChangesets?: boolean | null
}

export interface IMoveBatchChangeOnMutationArguments {
    batchChange: ID
    newName?: string | null
    newNamespace?: ID | null
}

export interface IDeleteBatchChangeOnMutationArguments {
    batchChange: ID
}

export interface ICreateBatchChangesCredentialOnMutationArguments {
    /**
     * The user for which to create the credential.
     */
    user: ID

    /**
     * The kind of external service being configured.
     */
    externalServiceKind: ExternalServiceKind

    /**
     * The URL of the external service being configured.
     */
    externalServiceURL: string

    /**
     * The credential to be stored. This can never be retrieved through the API and will be stored encrypted.
     */
    credential: string
}

export interface IDeleteBatchChangesCredentialOnMutationArguments {
    batchChangesCredential: ID
}

/**
 * A user (identified either by username or email address) with its repository permission.
 */
export interface IUserPermission {
    /**
     * Depending on the bindID option in the permissions.userMapping site configuration property,
     * the elements of the list are either all usernames (bindID of "username") or all email
     * addresses (bindID of "email").
     */
    bindID: string

    /**
     * The highest level of repository permission.
     * @default "READ"
     */
    permission?: RepositoryPermission | null
}

/**
 * A list of insights.
 */
export interface IInsightConnection {
    __typename: 'InsightConnection'

    /**
     * A list of insights.
     */
    nodes: IInsight[]

    /**
     * The total number of insights in the connection.
     */
    totalCount: number

    /**
     * Pagination information.
     */
    pageInfo: IPageInfo
}

/**
 * An insight about code.
 */
export interface IInsight {
    __typename: 'Insight'

    /**
     * The short title of the insight.
     */
    title: string

    /**
     * The description of the insight.
     */
    description: string

    /**
     * Data points over a time range (inclusive)
     */
    series: IInsightsSeries[]
}

/**
 * A series of data about a code insight.
 */
export interface IInsightsSeries {
    __typename: 'InsightsSeries'

    /**
     * The label used to describe this series of data points.
     */
    label: string

    /**
     * Data points over a time range (inclusive)
     *
     * If no 'from' time range is specified, the last 90 days of data is assumed.
     *
     * If no 'to' time range is specified, the current point in time is assumed.
     */
    points: IInsightDataPoint[]
}

export interface IPointsOnInsightsSeriesArguments {
    from?: DateTime | null
    to?: DateTime | null
}

/**
 * A code insight data point.
 */
export interface IInsightDataPoint {
    __typename: 'InsightDataPoint'

    /**
     * The time of this data point.
     */
    dateTime: DateTime

    /**
     * The value of the insight at this point in time.
     */
    value: number
}

/**
 * A new external service.
 */
export interface IAddExternalServiceInput {
    /**
     * The kind of the external service.
     */
    kind: ExternalServiceKind

    /**
     * The display name of the external service.
     */
    displayName: string

    /**
     * The JSON configuration of the external service.
     */
    config: string

    /**
     * The namespace this external service belongs to.
     * Currently, this can only be used for a user.
     */
    namespace?: ID | null
}

/**
 * Fields to update for an existing external service.
 */
export interface IUpdateExternalServiceInput {
    /**
     * The id of the external service to update.
     */
    id: ID

    /**
     * The updated display name, if provided.
     */
    displayName?: string | null

    /**
     * The updated config, if provided.
     */
    config?: string | null
}

/**
 * Describes options for rendering Markdown.
 */
export interface IMarkdownOptions {
    /**
     * A dummy null value (empty input types are not allowed yet).
     */
    alwaysNil?: string | null
}

/**
 * The product sources where events can come from.
 */
export enum EventSource {
    WEB = 'WEB',
    CODEHOSTINTEGRATION = 'CODEHOSTINTEGRATION',
    BACKEND = 'BACKEND',
}

/**
 * Input for Mutation.settingsMutation, which contains fields that all settings (global, organization, and user
 * settings) mutations need.
 */
export interface ISettingsMutationGroupInput {
    /**
     * The subject whose settings to mutate (organization, user, etc.).
     */
    subject: ID

    /**
     * The ID of the last-known settings known to the client, or null if there is none. This field is used to
     * prevent race conditions when there are concurrent editors.
     */
    lastID?: number | null
}

/**
 * Mutations that update settings (global, organization, or user settings). These mutations are grouped together
 * because they:
 * - are all versioned to avoid race conditions with concurrent editors
 * - all apply to a specific settings subject (i.e., a user, an organization, or the whole site)
 *
 * Grouping them lets us extract those common parameters to the Mutation.settingsMutation field.
 */
export interface ISettingsMutation {
    __typename: 'SettingsMutation'

    /**
     * Edit a single property in the settings object.
     */
    editSettings: IUpdateSettingsPayload | null

    /**
     * DEPRECATED
     * @deprecated "Use editSettings instead. This field is a deprecated alias for it and will be removed in a future release."
     */
    editConfiguration: IUpdateSettingsPayload | null

    /**
     * Overwrite the existing settings with the new settings.
     */
    overwriteSettings: IUpdateSettingsPayload | null
}

export interface IEditSettingsOnSettingsMutationArguments {
    /**
     * The edit to apply to the settings.
     */
    edit: ISettingsEdit
}

export interface IEditConfigurationOnSettingsMutationArguments {
    edit: IConfigurationEdit
}

export interface IOverwriteSettingsOnSettingsMutationArguments {
    /**
     * A JSON object (stringified) of the settings. Trailing commas and "//"-style comments are supported. The
     * entire previous settings value will be overwritten by this new value.
     */
    contents: string
}

/**
 * An edit to a JSON property in a settings JSON object. The JSON property to edit can be nested.
 */
export interface ISettingsEdit {
    /**
     * The key path of the property to update.
     *
     * Inserting into an existing array is not yet supported.
     */
    keyPath: IKeyPathSegment[]

    /**
     * The new JSON-encoded value to insert. If the field's value is not set, the property is removed. (This is
     * different from the field's value being the JSON null value.)
     *
     * When the value is a non-primitive type, it must be specified using a GraphQL variable, not an inline literal,
     * or else the GraphQL parser will return an error.
     */
    value?: any | null

    /**
     * Whether to treat the value as a JSONC-encoded string, which makes it possible to perform an edit that
     * preserves (or adds/removes) comments.
     * @default false
     */
    valueIsJSONCEncodedString?: boolean | null
}

/**
 * DEPRECATED: This type was renamed to SettingsEdit.
 * NOTE: GraphQL does not support @deprecated directives on INPUT_FIELD_DEFINITION (input fields).
 */
export interface IConfigurationEdit {
    /**
     * DEPRECATED
     */
    keyPath: IKeyPathSegment[]

    /**
     * DEPRECATED
     */
    value?: any | null

    /**
     * DEPRECATED
     * @default false
     */
    valueIsJSONCEncodedString?: boolean | null
}

/**
 * A segment of a key path that locates a nested JSON value in a root JSON value. Exactly one field in each
 * KeyPathSegment must be non-null.
 * For example, in {"a": [0, {"b": 3}]}, the value 3 is located at the key path ["a", 1, "b"].
 */
export interface IKeyPathSegment {
    /**
     * The name of the property in the object at this location to descend into.
     */
    property?: string | null

    /**
     * The index of the array at this location to descend into.
     */
    index?: number | null
}

/**
 * The payload for SettingsMutation.updateConfiguration.
 */
export interface IUpdateSettingsPayload {
    __typename: 'UpdateSettingsPayload'

    /**
     * An empty response.
     */
    empty: IEmptyResponse | null
}

/**
 * The result for Mutation.createAccessToken.
 */
export interface ICreateAccessTokenResult {
    __typename: 'CreateAccessTokenResult'

    /**
     * The ID of the newly created access token.
     */
    id: ID

    /**
     * The secret token value that is used to authenticate API clients. The caller is responsible for storing this
     * value.
     */
    token: string
}

/**
 * The result for Mutation.checkMirrorRepositoryConnection.
 */
export interface ICheckMirrorRepositoryConnectionResult {
    __typename: 'CheckMirrorRepositoryConnectionResult'

    /**
     * The error message encountered during the update operation, if any. If null, then
     * the connection check succeeded.
     */
    error: string | null
}

/**
 * The result for Mutation.createUser.
 */
export interface ICreateUserResult {
    __typename: 'CreateUserResult'

    /**
     * The new user.
     */
    user: IUser

    /**
     * The reset password URL that the new user must visit to sign into their account. If the builtin
     * username-password authentication provider is not enabled, this field's value is null.
     */
    resetPasswordURL: string | null
}

/**
 * The result for Mutation.randomizeUserPassword.
 */
export interface IRandomizeUserPasswordResult {
    __typename: 'RandomizeUserPasswordResult'

    /**
     * The reset password URL that the user must visit to sign into their account again. If the builtin
     * username-password authentication provider is not enabled, this field's value is null.
     */
    resetPasswordURL: string | null
}

/**
 * Input for a user satisfaction (NPS) survey submission.
 */
export interface ISurveySubmissionInput {
    /**
     * User-provided email address, if there is no currently authenticated user. If there is, this value
     * will not be used.
     */
    email?: string | null

    /**
     * User's likelihood of recommending Sourcegraph to a friend, from 0-10.
     */
    score: number

    /**
     * The answer to "What is the most important reason for the score you gave".
     */
    reason?: string | null

    /**
     * The answer to "What can Sourcegraph do to provide a better product"
     */
    better?: string | null
}

/**
 * Input for a happiness feedback submission.
 */
export interface IHappinessFeedbackSubmissionInput {
    /**
     * User's happiness rating, from 1-4.
     */
    score: number

    /**
     * The answer to "What's going well? What could be better?".
     */
    feedback?: string | null

    /**
     * The path that the happiness feedback will be submitted from.
     */
    currentPath?: string | null
}

/**
 * A query.
 */
export interface IQuery {
    __typename: 'Query'

    /**
     * The root of the query.
     * @deprecated "this will be removed."
     */
    root: IQuery

    /**
     * Looks up a node by ID.
     */
    node: Node | null

    /**
     * EXPERIMENTAL: Queries code insights
     */
    insights: IInsightConnection | null

    /**
     * Looks up a repository by either name or cloneURL.
     */
    repository: IRepository | null

    /**
     * Looks up a repository by either name or cloneURL. When the repository does not exist on the server
     * and "disablePublicRepoRedirects" is "false" in the site configuration, it returns a Redirect to
     * an external Sourcegraph URL that may have this repository instead. Otherwise, this query returns
     * null.
     */
    repositoryRedirect: RepositoryRedirect | null

    /**
     * Lists external services under given namespace.
     * If no namespace is given, it returns all external services.
     */
    externalServices: IExternalServiceConnection

    /**
     * List all repositories.
     */
    repositories: IRepositoryConnection

    /**
     * Looks up a Phabricator repository by name.
     */
    phabricatorRepo: IPhabricatorRepo | null

    /**
     * The current user.
     */
    currentUser: IUser | null

    /**
     * Looks up a user by username or email address.
     */
    user: IUser | null

    /**
     * List all users.
     */
    users: IUserConnection

    /**
     * Looks up an organization by name.
     */
    organization: IOrg | null

    /**
     * List all organizations.
     */
    organizations: IOrgConnection

    /**
     * Renders Markdown to HTML. The returned HTML is already sanitized and
     * escaped and thus is always safe to render.
     */
    renderMarkdown: string

    /**
     * EXPERIMENTAL: Syntax highlights a code string.
     */
    highlightCode: string

    /**
     * Looks up an instance of a type that implements SettingsSubject (i.e., something that has settings). This can
     * be a site (which has global settings), an organization, or a user.
     */
    settingsSubject: SettingsSubject | null

    /**
     * The settings for the viewer. The viewer is either an anonymous visitor (in which case viewer settings is
     * global settings) or an authenticated user (in which case viewer settings are the user's settings).
     */
    viewerSettings: ISettingsCascade

    /**
     * DEPRECATED
     * @deprecated "use viewerSettings instead"
     */
    viewerConfiguration: IConfigurationCascade

    /**
     * The configuration for clients.
     */
    clientConfiguration: IClientConfigurationDetails

    /**
     * Fetch search filter suggestions for autocompletion.
     */
    searchFilterSuggestions: ISearchFilterSuggestions

    /**
     * Runs a search.
     */
    search: ISearch | null

    /**
     * All saved searches configured for the current user, merged from all configurations.
     */
    savedSearches: ISavedSearch[]

    /**
     * All repository groups for the current user, merged from all configurations.
     */
    repoGroups: IRepoGroup[]

    /**
     * (experimental) All version contexts.
     */
    versionContexts: IVersionContext[]

    /**
     * (experimental) All search contexts.
     */
    searchContexts: ISearchContext[]

    /**
     * (experimental) Return the parse tree of a search query.
     */
    parseSearchQuery: any | null

    /**
     * The current site.
     */
    site: ISite

    /**
     * Retrieve responses to surveys.
     */
    surveyResponses: ISurveyResponseConnection

    /**
     * The extension registry.
     */
    extensionRegistry: IExtensionRegistry

    /**
     * Queries that are only used on Sourcegraph.com.
     *
     * FOR INTERNAL USE ONLY.
     */
    dotcom: IDotcomQuery

    /**
     * FOR INTERNAL USE ONLY: Lists all status messages
     */
    statusMessages: StatusMessage[]

    /**
     * FOR INTERNAL USE ONLY: Query repository statistics for the site.
     */
    repositoryStats: IRepositoryStats

    /**
     * Look up a namespace by ID.
     */
    namespace: Namespace | null

    /**
     * Look up a namespace by name, which is a username or organization name.
     */
    namespaceByName: Namespace | null

    /**
     * The repositories a user is authorized to access with the given permission.
     * This isn’t defined in the User type because we store permissions for users
     * that don’t yet exist (i.e. late binding). Only one of "username" or "email"
     * is required to identify a user.
     */
    authorizedUserRepositories: IRepositoryConnection

    /**
     * Returns a list of usernames or emails that have associated pending permissions.
     * The returned list can be used to query authorizedUserRepositories for pending permissions.
     */
    usersWithPendingPermissions: string[]

    /**
     * (experimental) The LSIF API may change substantially in the near future as we
     * continue to adjust it for our use cases. Changes will not be documented in the
     * CHANGELOG during this time.
     * The repository's LSIF uploads.
     */
    lsifUploads: ILSIFUploadConnection

    /**
     * (experimental) The LSIF API may change substantially in the near future as we
     * continue to adjust it for our use cases. Changes will not be documented in the
     * CHANGELOG during this time.
     * The repository's LSIF uploads.
     */
    lsifIndexes: ILSIFIndexConnection

    /**
     * Repos affiliated with the user & code hosts, these repos are not necessarily synced, but ones that
     * the configured code hosts are able to see.
     */
    affiliatedRepositories: ICodeHostRepositoryConnection

    /**
     * Checks whether the given feature is enabled on Sourcegraph. Open source
     * installations will always return false for any feature.
     */
    enterpriseLicenseHasFeature: boolean

    /**
     * Retrieve all registered out-of-band migrations.
     */
    outOfBandMigrations: IOutOfBandMigration[]

    /**
     * A list of campaigns.
     * @deprecated "campaigns have been renamed to batch changes. Use batchChanges instead."
     */
    campaigns: ICampaignConnection

    /**
     * Looks up a campaign by namespace and campaign name.
     * @deprecated "campaigns have been renamed to batch changes. Use batchChange instead."
     */
    campaign: ICampaign | null

    /**
     * A list of batch changes.
     */
    batchChanges: IBatchChangeConnection

    /**
     * Looks up a batch change by namespace and campaign name.
     */
    batchChange: IBatchChange | null
}

export interface INodeOnQueryArguments {
    id: ID
}

export interface IRepositoryOnQueryArguments {
    /**
     * Query the repository by name, for example "github.com/gorilla/mux".
     */
    name?: string | null

    /**
     * Query the repository by a Git clone URL (format documented here: https://git-scm.com/docs/git-clone_git_urls_a_id_urls_a)
     * by checking for a code host configuration that matches the clone URL.
     * Will not actually check the code host to see if the repository actually exists.
     */
    cloneURL?: string | null

    /**
     * An alias for name. DEPRECATED: use name instead.
     */
    uri?: string | null
}

export interface IRepositoryRedirectOnQueryArguments {
    /**
     * Query the repository by name, for example "github.com/gorilla/mux".
     */
    name?: string | null

    /**
     * Query the repository by a Git clone URL (format documented here: https://git-scm.com/docs/git-clone_git_urls_a_id_urls_a)
     * by checking for a code host configuration that matches the clone URL.
     * Will not actually check the code host to see if the repository actually exists.
     */
    cloneURL?: string | null
}

export interface IExternalServicesOnQueryArguments {
    /**
     * The namespace to scope returned external services.
     * Currently, this can only be used for a user.
     */
    namespace?: ID | null

    /**
     * Returns the first n external services from the list.
     */
    first?: number | null

    /**
     * Opaque pagination cursor.
     */
    after?: string | null
}

export interface IRepositoriesOnQueryArguments {
    /**
     * Returns the first n repositories from the list.
     */
    first?: number | null

    /**
     * Return repositories whose names match the query.
     */
    query?: string | null

    /**
     * An opaque cursor that is used for pagination.
     */
    after?: string | null

    /**
     * Return repositories whose names are in the list.
     */
    names?: string[] | null

    /**
     * Include cloned repositories.
     * @default true
     */
    cloned?: boolean | null

    /**
     * Include repositories that are not yet cloned and for which cloning is not in progress.
     * @default true
     */
    notCloned?: boolean | null

    /**
     * Include repositories that have a text search index.
     * @default true
     */
    indexed?: boolean | null

    /**
     * Include repositories that do not have a text search index.
     * @default true
     */
    notIndexed?: boolean | null

    /**
     * Sort field.
     * @default "REPOSITORY_NAME"
     */
    orderBy?: RepositoryOrderBy | null

    /**
     * Sort direction.
     * @default false
     */
    descending?: boolean | null
}

export interface IPhabricatorRepoOnQueryArguments {
    /**
     * The name, for example "github.com/gorilla/mux".
     */
    name?: string | null

    /**
     * An alias for name. DEPRECATED: use name instead.
     */
    uri?: string | null
}

export interface IUserOnQueryArguments {
    /**
     * Query the user by username.
     */
    username?: string | null

    /**
     * Query the user by verified email address.
     */
    email?: string | null
}

export interface IUsersOnQueryArguments {
    /**
     * Returns the first n users from the list.
     */
    first?: number | null

    /**
     * Return users whose usernames or display names match the query.
     */
    query?: string | null

    /**
     * Return only users with the given tag.
     */
    tag?: string | null

    /**
     * Returns users who have been active in a given period of time.
     */
    activePeriod?: UserActivePeriod | null
}

export interface IOrganizationOnQueryArguments {
    name: string
}

export interface IOrganizationsOnQueryArguments {
    /**
     * Returns the first n organizations from the list.
     */
    first?: number | null

    /**
     * Return organizations whose names or display names match the query.
     */
    query?: string | null
}

export interface IRenderMarkdownOnQueryArguments {
    markdown: string
    options?: IMarkdownOptions | null
}

export interface IHighlightCodeOnQueryArguments {
    code: string
    fuzzyLanguage: string
    disableTimeout: boolean
    isLightTheme: boolean
}

export interface ISettingsSubjectOnQueryArguments {
    id: ID
}

export interface ISearchOnQueryArguments {
    /**
     * The version of the search syntax being used.
     * All new clients should use the latest version.
     * @default "V1"
     */
    version?: SearchVersion | null

    /**
     * PatternType controls the search pattern type, if and only if it is not specified in the query string using
     * the patternType: field.
     */
    patternType?: SearchPatternType | null

    /**
     * The search query (such as "foo" or "repo:myrepo foo").
     * @default ""
     */
    query?: string | null

    /**
     * (experimental) Optionally specify the versionContext. If not specified the
     * default version context is used (all repositories on the default branch).
     */
    versionContext?: string | null

    /**
     * (experimental) Sourcegraph 3.9 added support for cursor-based paginated
     * search requests when this field is specified. For details, see
     * https://docs.sourcegraph.com/api/graphql/search
     *
     * When specified, indicates that this request should be paginated and
     * to fetch results starting at this cursor.
     *
     * A future request can be made for more results by passing in the
     * 'SearchResults.pageInfo.endCursor' that is returned.
     */
    after?: string | null

    /**
     * (experimental) Sourcegraph 3.9 added support for cursor-based paginated
     * search requests when this field is specified. For details, see
     * https://docs.sourcegraph.com/api/graphql/search
     *
     * When specified, indicates that this request should be paginated and
     * the first N results (relative to the cursor) should be returned. i.e.
     * how many results to return per page. It must be in the range of 0-5000.
     */
    first?: number | null
}

export interface IParseSearchQueryOnQueryArguments {
    /**
     * The search query (such as "repo:myrepo foo").
     * @default ""
     */
    query?: string | null

    /**
     * The parser to use for this query.
     * @default "literal"
     */
    patternType?: SearchPatternType | null
}

export interface ISurveyResponsesOnQueryArguments {
    /**
     * Returns the first n survey responses from the list.
     */
    first?: number | null
}

export interface INamespaceOnQueryArguments {
    id: ID
}

export interface INamespaceByNameOnQueryArguments {
    /**
     * The name of the namespace.
     */
    name: string
}

export interface IAuthorizedUserRepositoriesOnQueryArguments {
    /**
     * The username.
     */
    username?: string | null

    /**
     * One of the email addresses.
     */
    email?: string | null

    /**
     * Permission that the user has on the repositories.
     * @default "READ"
     */
    perm?: RepositoryPermission | null

    /**
     * Number of repositories to return after the given cursor.
     */
    first: number

    /**
     * Opaque pagination cursor.
     */
    after?: string | null
}

export interface ILsifUploadsOnQueryArguments {
    /**
     * An (optional) search query that searches over the state, repository name,
     * commit, root, and indexer properties.
     */
    query?: string | null

    /**
     * The state of returned uploads.
     */
    state?: LSIFUploadState | null

    /**
     * When specified, shows only uploads that are latest for the given repository.
     */
    isLatestForRepo?: boolean | null

    /**
     * When specified, indicates that this request should be paginated and
     * the first N results (relative to the cursor) should be returned. i.e.
     * how many results to return per page. It must be in the range of 0-5000.
     */
    first?: number | null

    /**
     * When specified, indicates that this request should be paginated and
     * to fetch results starting at this cursor.
     *
     * A future request can be made for more results by passing in the
     * 'LSIFUploadConnection.pageInfo.endCursor' that is returned.
     */
    after?: string | null
}

export interface ILsifIndexesOnQueryArguments {
    /**
     * An (optional) search query that searches over the state, repository name,
     * and commit properties.
     */
    query?: string | null

    /**
     * The state of returned uploads.
     */
    state?: LSIFIndexState | null

    /**
     * When specified, indicates that this request should be paginated and
     * the first N results (relative to the cursor) should be returned. i.e.
     * how many results to return per page. It must be in the range of 0-5000.
     */
    first?: number | null

    /**
     * When specified, indicates that this request should be paginated and
     * to fetch results starting at this cursor.
     * A future request can be made for more results by passing in the
     * 'LSIFIndexConnection.pageInfo.endCursor' that is returned.
     */
    after?: string | null
}

export interface IAffiliatedRepositoriesOnQueryArguments {
    user: ID
    codeHost?: ID | null
    query?: string | null
}

export interface IEnterpriseLicenseHasFeatureOnQueryArguments {
    feature: string
}

export interface ICampaignsOnQueryArguments {
    /**
     * Returns the first n campaigns from the list.
     * @default 50
     */
    first?: number | null

    /**
     * Opaque pagination cursor.
     */
    after?: string | null

    /**
     * Only return campaigns in this state.
     */
    state?: CampaignState | null

    /**
     * Only include campaigns that the viewer can administer.
     */
    viewerCanAdminister?: boolean | null
}

export interface ICampaignOnQueryArguments {
    /**
     * The namespace where the campaign lives.
     */
    namespace: ID

    /**
     * The campaigns name.
     */
    name: string
}

export interface IBatchChangesOnQueryArguments {
    /**
     * Returns the first n batch changes from the list.
     * @default 50
     */
    first?: number | null

    /**
     * Opaque pagination cursor.
     */
    after?: string | null

    /**
     * Only return batch changes in this state.
     */
    state?: BatchChangeState | null

    /**
     * Only include batch changes that the viewer can administer.
     */
    viewerCanAdminister?: boolean | null
}

export interface IBatchChangeOnQueryArguments {
    /**
     * The namespace where the batch change lives.
     */
    namespace: ID

    /**
     * The batch changes name.
     */
    name: string
}

/**
 * An out-of-band migration is a process that runs in the background of the instance that moves
 * data from one format into another format. Out-of-band migrations
 */
export interface IOutOfBandMigration {
    __typename: 'OutOfBandMigration'

    /**
     * The unique identifier of this migration.
     */
    id: ID

    /**
     * The team that owns this migration (e.g., code-intelligence).
     */
    team: string

    /**
     * The component this migration affects (e.g., codeintel-db.lsif_data_documents).
     */
    component: string

    /**
     * A human-readable summary of the migration.
     */
    description: string

    /**
     * The Sourcegraph version in which this migration was introduced.
     *
     * It is necessary to completely this migration in reverse (if destructive) before downgrading
     * to or past this version. Otherwise, the previous instance version will not be aware of the
     * new data format.
     */
    introduced: string

    /**
     * The Sourcegraph version in which this migration is assumed to have completed.
     *
     * It is necessary to have completed this migration before upgrading to or past this version.
     * Otherwsie, the next instance version will no longer be aware of the old data format.
     */
    deprecated: string | null

    /**
     * The progress of the migration (in the forward direction). In the range [0, 1].
     */
    progress: number

    /**
     * The time the migration record was inserted.
     */
    created: DateTime

    /**
     * The last time the migration progress or error list was updated.
     */
    lastUpdated: DateTime | null

    /**
     * If false, the migration moves data destructively, and a previous version of Sourcegraph
     * will encounter errors when interfacing with the target data unless the migration is first
     * run in reverse prior to a downgrade.
     */
    nonDestructive: boolean

    /**
     * If true, the migration will run in reverse.
     */
    applyReverse: boolean

    /**
     * A list of errors that have occurred while performing this migration (in either direction).
     * This list is bounded by a maximum size, and older errors will replaced by newer errors as
     * the list capacity is reached.
     */
    errors: IOutOfBandMigrationError[]
}

/**
 * An error that occurred while performing an out-of-band migration.
 */
export interface IOutOfBandMigrationError {
    __typename: 'OutOfBandMigrationError'

    /**
     * The error message.
     */
    message: string

    /**
     * The time the error occurred.
     */
    created: DateTime
}

/**
 * The version of the search syntax.
 */
export enum SearchVersion {
    /**
     * Search syntax that defaults to regexp search.
     */
    V1 = 'V1',

    /**
     * Search syntax that defaults to literal search.
     */
    V2 = 'V2',
}

/**
 * The search pattern type.
 */
export enum SearchPatternType {
    literal = 'literal',
    regexp = 'regexp',
    structural = 'structural',
}

/**
 * Configuration details for the browser extension, editor extensions, etc.
 */
export interface IClientConfigurationDetails {
    __typename: 'ClientConfigurationDetails'

    /**
     * The list of phabricator/gitlab/bitbucket/etc instance URLs that specifies which pages the content script will be injected into.
     */
    contentScriptUrls: string[]

    /**
     * Returns details about the parent Sourcegraph instance.
     */
    parentSourcegraph: IParentSourcegraphDetails
}

/**
 * Parent Sourcegraph instance
 */
export interface IParentSourcegraphDetails {
    __typename: 'ParentSourcegraphDetails'

    /**
     * Sourcegraph instance URL.
     */
    url: string
}

/**
 * A search.
 */
export interface ISearch {
    __typename: 'Search'

    /**
     * The results.
     */
    results: ISearchResults

    /**
     * The suggestions.
     */
    suggestions: SearchSuggestion[]

    /**
     * A subset of results (excluding actual search results) which are heavily
     * cached and thus quicker to query. Useful for e.g. querying sparkline
     * data.
     */
    stats: ISearchResultsStats
}

export interface ISuggestionsOnSearchArguments {
    first?: number | null
}

/**
 * Predefined suggestions for search filters when backfill.
 */
export interface ISearchFilterSuggestions {
    __typename: 'SearchFilterSuggestions'

    /**
     * The suggestions for search filter "repogroup:".
     */
    repogroup: string[]

    /**
     * The suggestions for search filter "repo:".
     */
    repo: string[]
}

/**
 * A search result.
 */
export type SearchResult = IFileMatch | ICommitSearchResult | IRepository

/**
 * An object representing a markdown string.
 */
export interface IMarkdown {
    __typename: 'Markdown'

    /**
     * The raw markdown string.
     */
    text: string

    /**
     * HTML for the rendered markdown string, or null if there is no HTML representation provided.
     * If specified, clients should render this directly.
     */
    html: string
}

/**
 * A search result. Every type of search result, except FileMatch, must implement this interface.
 */
export type GenericSearchResultInterface = ICommitSearchResult | IRepository

/**
 * A search result. Every type of search result, except FileMatch, must implement this interface.
 */
export interface IGenericSearchResultInterface {
    __typename: 'GenericSearchResultInterface'

    /**
     * URL to an icon that is displayed with every search result.
     */
    icon: string

    /**
     * A markdown string that is rendered prominently.
     */
    label: IMarkdown

    /**
     * The URL of the result.
     */
    url: string

    /**
     * A markdown string that is rendered less prominently.
     */
    detail: IMarkdown

    /**
     * A list of matches in this search result.
     */
    matches: ISearchResultMatch[]
}

/**
 * A match in a search result. Matches make up the body content of a search result.
 */
export interface ISearchResultMatch {
    __typename: 'SearchResultMatch'

    /**
     * URL for the individual result match.
     */
    url: string

    /**
     * A markdown string containing the preview contents of the result match.
     */
    body: IMarkdown

    /**
     * A list of highlights that specify locations of matches of the query in the body. Each highlight is
     * a line number, character offset, and length. Currently, highlights are only displayed on match bodies
     * that are code blocks. If the result body is a code block, exclude the markdown code fence lines in
     * the line and character count. Leave as an empty list if no highlights are available.
     */
    highlights: IHighlight[]
}

/**
 * Search results.
 */
export interface ISearchResults {
    __typename: 'SearchResults'

    /**
     * The results. Inside each SearchResult there may be multiple matches, e.g.
     * a FileMatch may contain multiple line matches.
     */
    results: SearchResult[]

    /**
     * The total number of matches returned by this search. This is different
     * than the length of the results array in that e.g. a single results array
     * entry may contain multiple matches. For example, the results array may
     * contain two file matches and this field would report 6 ("3 line matches
     * per file") while the length of the results array would report 3
     * ("3 FileMatch results").
     * Typically, 'approximateResultCount', not this field, is shown to users.
     */
    matchCount: number

    /**
     * DEPRECATED: Renamed to 'matchCount' for less ambiguity.
     * @deprecated "renamed to matchCount for less ambiguity"
     */
    resultCount: number

    /**
     * The approximate number of results. This is like the length of the results
     * array, except it can indicate the number of results regardless of whether
     * or not the limit was hit. Currently, this is represented as e.g. "5+"
     * results.
     * This string is typically shown to users to indicate the true result count.
     */
    approximateResultCount: string

    /**
     * Whether or not the results limit was hit.
     * In paginated requests, this field is always false. Use 'pageInfo.hasNextPage' instead.
     */
    limitHit: boolean

    /**
     * Integers representing the sparkline for the search results.
     */
    sparkline: number[]

    /**
     * Repositories that were eligible to be searched.
     */
    repositories: IRepository[]

    /**
     * The number of repositories that were eligible to be searched (for clients
     * that just wish to know how many without querying the, sometimes extremely
     * large, list).
     */
    repositoriesCount: number

    /**
     * Repositories that are busy cloning onto gitserver.
     * In paginated search requests, some repositories may be cloning. These are reported here
     * and you may choose to retry the paginated request with the same cursor after they have
     * cloned OR you may simply continue making further paginated requests and choose to skip
     * the cloning repositories.
     */
    cloning: IRepository[]

    /**
     * Repositories or commits that do not exist.
     * In paginated search requests, some repositories may be missing (e.g. if Sourcegraph is
     * aware of them but is temporarily unable to serve them). These are reported here and you
     * may choose to retry the paginated request with the same cursor and they may no longer be
     * missing OR you may simply continue making further paginated requests and choose to skip
     * the missing repositories.
     */
    missing: IRepository[]

    /**
     * Repositories or commits which we did not manage to search in time. Trying
     * again usually will work.
     * In paginated search requests, this field is not relevant.
     */
    timedout: IRepository[]

    /**
     * True if indexed search is enabled but was not available during this search.
     */
    indexUnavailable: boolean

    /**
     * An alert message that should be displayed before any results.
     */
    alert: ISearchAlert | null

    /**
     * The time it took to generate these results.
     */
    elapsedMilliseconds: number

    /**
     * Dynamic filters generated by the search results
     */
    dynamicFilters: ISearchFilter[]

    /**
     * Pagination information.
     * This field is only applcable when the original request was a paginated one.
     */
    pageInfo: IPageInfo
}

/**
 * Statistics about search results.
 */
export interface ISearchResultsStats {
    __typename: 'SearchResultsStats'

    /**
     * The approximate number of results returned.
     */
    approximateResultCount: string

    /**
     * The sparkline.
     */
    sparkline: number[]

    /**
     * Statistics about the languages represented in the search results.
     * Known issue: The LanguageStatistics.totalBytes field values are incorrect in the result.
     */
    languages: ILanguageStatistics[]
}

/**
 * A search filter.
 */
export interface ISearchFilter {
    __typename: 'SearchFilter'

    /**
     * The value.
     */
    value: string

    /**
     * The string to be displayed in the UI.
     */
    label: string

    /**
     * Number of matches for a given filter.
     */
    count: number

    /**
     * Whether the results returned are incomplete.
     */
    limitHit: boolean

    /**
     * The kind of filter. Should be "file" or "repo".
     */
    kind: string
}

/**
 * A programming language.
 */
export interface ILanguage {
    __typename: 'Language'

    /**
     * Name of the programming language.
     */
    name: string
}

/**
 * A search suggestion.
 */
export type SearchSuggestion = IRepository | IFile | ISymbol | ILanguage

/**
 * A search-related alert message.
 */
export interface ISearchAlert {
    __typename: 'SearchAlert'

    /**
     * The title.
     */
    title: string

    /**
     * The description.
     */
    description: string | null

    /**
     * "Did you mean: ____" query proposals
     */
    proposedQueries: ISearchQueryDescription[] | null
}

/**
 * A saved search query, defined in settings.
 */
export interface ISavedSearch {
    __typename: 'SavedSearch'

    /**
     * The unique ID of this saved query.
     */
    id: ID

    /**
     * The description.
     */
    description: string

    /**
     * The query.
     */
    query: string

    /**
     * Whether or not to notify the owner of the saved search via email. This owner is either
     * a single user, or every member of an organization that owns the saved search.
     */
    notify: boolean

    /**
     * Whether or not to notify on Slack.
     */
    notifySlack: boolean

    /**
     * The user or org that owns this saved search.
     */
    namespace: Namespace

    /**
     * The Slack webhook URL associated with this saved search, if any.
     */
    slackWebhookURL: string | null
}

/**
 * A list of code monitors
 */
export interface IMonitorConnection {
    __typename: 'MonitorConnection'

    /**
     * A list of monitors.
     */
    nodes: IMonitor[]

    /**
     * The total number of monitors in the connection.
     */
    totalCount: number

    /**
     * Pagination information.
     */
    pageInfo: IPageInfo
}

/**
 * A code monitor with one trigger and possibly many actions.
 */
export interface IMonitor {
    __typename: 'Monitor'

    /**
     * The code monitor's unique ID.
     */
    id: ID

    /**
     * The user who created the code monitor.
     */
    createdBy: IUser

    /**
     * The time at which the code monitor was created.
     */
    createdAt: DateTime

    /**
     * A meaningful description of the code monitor.
     */
    description: string

    /**
     * Owners can edit the code monitor.
     */
    owner: Namespace

    /**
     * Whether the code monitor is currently enabled.
     */
    enabled: boolean

    /**
     * Triggers trigger actions. There can only be one trigger per monitor.
     */
    trigger: MonitorTrigger

    /**
     * One or more actions that are triggered by the trigger.
     */
    actions: IMonitorActionConnection
}

export interface IActionsOnMonitorArguments {
    /**
     * Returns the first n actions from the list.
     * @default 50
     */
    first?: number | null

    /**
     * Opaque pagination cursor.
     */
    after?: string | null
}

/**
 * A query that can serve as a trigger for code monitors.
 */
export interface IMonitorQuery {
    __typename: 'MonitorQuery'

    /**
     * The unique id of a trigger query.
     */
    id: ID

    /**
     * A query.
     */
    query: string

    /**
     * A list of events.
     */
    events: IMonitorTriggerEventConnection
}

export interface IEventsOnMonitorQueryArguments {
    /**
     * Returns the first n events from the list.
     * @default 50
     */
    first?: number | null

    /**
     * Opaque pagination cursor.
     */
    after?: string | null
}

/**
 * A list of trigger events.
 */
export interface IMonitorTriggerEventConnection {
    __typename: 'MonitorTriggerEventConnection'

    /**
     * A list of events.
     */
    nodes: IMonitorTriggerEvent[]

    /**
     * The total number of events in the connection.
     */
    totalCount: number

    /**
     * Pagination information.
     */
    pageInfo: IPageInfo
}

/**
 * A trigger event is an event together with a list of associated actions.
 */
export interface IMonitorTriggerEvent {
    __typename: 'MonitorTriggerEvent'

    /**
     * The unique id of an event.
     */
    id: ID

    /**
     * The status of an event.
     */
    status: EventStatus

    /**
     * A message with details regarding the status of the event.
     */
    message: string | null

    /**
     * The time and date of the event.
     */
    timestamp: DateTime

    /**
     * A list of actions.
     */
    actions: IMonitorActionConnection
}

export interface IActionsOnMonitorTriggerEventArguments {
    /**
     * Returns the first n events from the list.
     * @default 50
     */
    first?: number | null

    /**
     * Opaque pagination cursor.
     */
    after?: string | null
}

/**
 * Supported triggers for code monitors.
 */
export type MonitorTrigger = IMonitorQuery

/**
 * A list of actions.
 */
export interface IMonitorActionConnection {
    __typename: 'MonitorActionConnection'

    /**
     * A list of actions.
     */
    nodes: MonitorAction[]

    /**
     * The total number of actions in the connection.
     */
    totalCount: number

    /**
     * Pagination information.
     */
    pageInfo: IPageInfo
}

/**
 * Supported actions for code monitors.
 */
export type MonitorAction = IMonitorEmail

/**
 * Email is one of the supported actions of code monitors.
 */
export interface IMonitorEmail {
    __typename: 'MonitorEmail'

    /**
     * The unique id of an email action.
     */
    id: ID

    /**
     * Whether the email action is enabled or not.
     */
    enabled: boolean

    /**
     * The priority of the email action.
     */
    priority: MonitorEmailPriority

    /**
     * Use header to automatically approve the message in a read-only or moderated mailing list.
     */
    header: string

    /**
     * A list of recipients of the email.
     */
    recipients: IMonitorActionEmailRecipientsConnection

    /**
     * A list of events.
     */
    events: IMonitorActionEventConnection
}

export interface IRecipientsOnMonitorEmailArguments {
    /**
     * Returns the first n recipients from the list.
     * @default 50
     */
    first?: number | null

    /**
     * Opaque pagination cursor.
     */
    after?: string | null
}

export interface IEventsOnMonitorEmailArguments {
    /**
     * Returns the first n events from the list.
     * @default 50
     */
    first?: number | null

    /**
     * Opaque pagination cursor.
     */
    after?: string | null
}

/**
 * The priority of an email action.
 */
export enum MonitorEmailPriority {
    NORMAL = 'NORMAL',
    CRITICAL = 'CRITICAL',
}

/**
 * A list of events.
 */
export interface IMonitorActionEmailRecipientsConnection {
    __typename: 'MonitorActionEmailRecipientsConnection'

    /**
     * A list of recipients.
     */
    nodes: Namespace[]

    /**
     * The total number of recipients in the connection.
     */
    totalCount: number

    /**
     * Pagination information.
     */
    pageInfo: IPageInfo
}

/**
 * A list of events.
 */
export interface IMonitorActionEventConnection {
    __typename: 'MonitorActionEventConnection'

    /**
     * A list of events.
     */
    nodes: IMonitorActionEvent[]

    /**
     * The total number of events in the connection.
     */
    totalCount: number

    /**
     * Pagination information.
     */
    pageInfo: IPageInfo
}

/**
 * An event documents the result of a trigger or an execution of an action.
 */
export interface IMonitorActionEvent {
    __typename: 'MonitorActionEvent'

    /**
     * The unique id of an event.
     */
    id: ID

    /**
     * The status of an event.
     */
    status: EventStatus

    /**
     * A message with details regarding the status of the event.
     */
    message: string | null

    /**
     * The time and date of the event.
     */
    timestamp: DateTime
}

/**
 * Supported status of monitor events.
 */
export enum EventStatus {
    PENDING = 'PENDING',
    SUCCESS = 'SUCCESS',
    ERROR = 'ERROR',
}

/**
 * The input required to create a code monitor.
 */
export interface IMonitorInput {
    /**
     * The namespace represents the owner of the code monitor.
     * Owners can either be users or organizations.
     */
    namespace: ID

    /**
     * A meaningful description of the code monitor.
     */
    description: string

    /**
     * Whether the code monitor is enabled or not.
     */
    enabled: boolean
}

/**
 * The input required to edit a code monitor.
 */
export interface IMonitorEditInput {
    /**
     * The id of the monitor.
     */
    id: ID

    /**
     * The desired state after the udpate.
     */
    update: IMonitorInput
}

/**
 * The input required to create a trigger.
 */
export interface IMonitorTriggerInput {
    /**
     * The query string.
     */
    query: string
}

/**
 * The input required to edit a trigger.
 */
export interface IMonitorEditTriggerInput {
    /**
     * The id of the Trigger.
     */
    id: ID

    /**
     * The desired state after the udpate.
     */
    update: IMonitorTriggerInput
}

/**
 * The input required to create an action.
 */
export interface IMonitorActionInput {
    /**
     * An email action.
     */
    email?: IMonitorEmailInput | null
}

/**
 * The input required to create an email action.
 */
export interface IMonitorEmailInput {
    /**
     * Whether the email action is enabled or not.
     */
    enabled: boolean

    /**
     * The priority of the email.
     */
    priority: MonitorEmailPriority

    /**
     * A list of users or orgs which will receive the email.
     */
    recipients: ID[]

    /**
     * Use header to automatically approve the message in a read-only or moderated mailing list.
     */
    header: string
}

/**
 * The input required to edit an action.
 */
export interface IMonitorEditActionInput {
    /**
     * An email action.
     */
    email?: IMonitorEditEmailInput | null
}

/**
 * The input required to edit an email action.
 */
export interface IMonitorEditEmailInput {
    /**
     * The id of an email action.
     */
    id?: ID | null

    /**
     * The desired state after the update.
     */
    update: IMonitorEmailInput
}

/**
 * A search query description.
 */
export interface ISearchQueryDescription {
    __typename: 'SearchQueryDescription'

    /**
     * The description.
     */
    description: string | null

    /**
     * The query.
     */
    query: string
}

/**
 * A group of repositories.
 */
export interface IRepoGroup {
    __typename: 'RepoGroup'

    /**
     * The name.
     */
    name: string

    /**
     * The repositories.
     */
    repositories: string[]
}

/**
 * A diff between two diffable Git objects.
 */
export interface IDiff {
    __typename: 'Diff'

    /**
     * The diff's repository.
     */
    repository: IRepository

    /**
     * The revision range of the diff.
     */
    range: IGitRevisionRange
}

/**
 * A search result that is a Git commit.
 */
export interface ICommitSearchResult {
    __typename: 'CommitSearchResult'

    /**
     * Base64 data uri to an icon.
     */
    icon: string

    /**
     * A markdown string that is rendered prominently.
     */
    label: IMarkdown

    /**
     * The URL of the result.
     */
    url: string

    /**
     * A markdown string of that is rendered less prominently.
     */
    detail: IMarkdown

    /**
     * The result previews of the result.
     */
    matches: ISearchResultMatch[]

    /**
     * The commit that matched the search query.
     */
    commit: IGitCommit

    /**
     * The ref names of the commit.
     */
    refs: IGitRef[]

    /**
     * The refs by which this commit was reached.
     */
    sourceRefs: IGitRef[]

    /**
     * The matching portion of the commit message, if any.
     */
    messagePreview: IHighlightedString | null

    /**
     * The matching portion of the diff, if any.
     */
    diffPreview: IHighlightedString | null
}

/**
 * A string that has highlights (e.g, query matches).
 */
export interface IHighlightedString {
    __typename: 'HighlightedString'

    /**
     * The full contents of the string.
     */
    value: string

    /**
     * Highlighted matches of the query in the preview string.
     */
    highlights: IHighlight[]
}

/**
 * A highlighted region in a string (e.g., matched by a query).
 */
export interface IHighlight {
    __typename: 'Highlight'

    /**
     * The 1-indexed line number.
     */
    line: number

    /**
     * The 1-indexed character on the line.
     */
    character: number

    /**
     * The length of the highlight, in characters (on the same line).
     */
    length: number
}

/**
 * A list of external services.
 */
export interface IExternalServiceConnection {
    __typename: 'ExternalServiceConnection'

    /**
     * A list of external services.
     */
    nodes: IExternalService[]

    /**
     * The total number of external services in the connection.
     */
    totalCount: number

    /**
     * Pagination information.
     */
    pageInfo: IPageInfo
}

/**
 * A specific kind of external service.
 */
export enum ExternalServiceKind {
    AWSCODECOMMIT = 'AWSCODECOMMIT',
    BITBUCKETCLOUD = 'BITBUCKETCLOUD',
    BITBUCKETSERVER = 'BITBUCKETSERVER',
    GITHUB = 'GITHUB',
    GITLAB = 'GITLAB',
    GITOLITE = 'GITOLITE',
    PERFORCE = 'PERFORCE',
    PHABRICATOR = 'PHABRICATOR',
    OTHER = 'OTHER',
}

/**
 * A configured external service.
 */
export interface IExternalService {
    __typename: 'ExternalService'

    /**
     * The external service's unique ID.
     */
    id: ID

    /**
     * The kind of external service.
     */
    kind: ExternalServiceKind

    /**
     * The display name of the external service.
     */
    displayName: string

    /**
     * The JSON configuration of the external service.
     */
    config: JSONCString

    /**
     * When the external service was created.
     */
    createdAt: DateTime

    /**
     * When the external service was last updated.
     */
    updatedAt: DateTime

    /**
     * The namespace this external service belongs to.
     */
    namespace: ID | null

    /**
     * The number of repos synced by the external service.
     */
    repoCount: number

    /**
     * An optional URL that will be populated when webhooks have been configured for the external service.
     */
    webhookURL: string | null

    /**
     * This is an optional field that's populated when we ran into errors on the
     * backend side when trying to create/update an ExternalService, but the
     * create/update still succeeded.
     * It is a field on ExternalService instead of a separate thing in order to
     * not break the API and stay backwards compatible.
     */
    warning: string | null

    /**
     * External services are synced with code hosts in the background. This optional field
     * will contain any errors that occured during the most recent completed sync.
     */
    lastSyncError: string | null

    /**
     * LastSyncAt is the time the last sync job was run for this code host
     */
    lastSyncAt: DateTime

    /**
     * The timestamp of the next sync job
     */
    nextSyncAt: DateTime
}

/**
 * A list of repositories.
 */
export interface IRepositoryConnection {
    __typename: 'RepositoryConnection'

    /**
     * A list of repositories.
     */
    nodes: IRepository[]

    /**
     * The total count of repositories in the connection. This total count may be larger
     * than the number of nodes in this object when the result is paginated.
     * This requires admin permissions and will return null for all non-admin users.
     * In some cases, the total count can't be computed quickly; if so, it is null. Pass
     * precise: true to always compute total counts even if it takes a while.
     */
    totalCount: number | null

    /**
     * Pagination information.
     */
    pageInfo: IPageInfo
}

export interface ITotalCountOnRepositoryConnectionArguments {
    /**
     * @default false
     */
    precise?: boolean | null
}

/**
 * A repository is a Git source control repository that is mirrored from some origin code host.
 */
export interface IRepository {
    __typename: 'Repository'

    /**
     * The repository's unique ID.
     */
    id: ID

    /**
     * The repository's name, as a path with one or more components. It conventionally consists of
     * the repository's hostname and path (joined by "/"), minus any suffixes (such as ".git").
     * Examples:
     * - github.com/foo/bar
     * - my-code-host.example.com/myrepo
     * - myrepo
     */
    name: string

    /**
     * DEPRECATED: Use name.
     * @deprecated "Use name."
     */
    uri: string

    /**
     * The repository's description.
     */
    description: string

    /**
     * The primary programming language in the repository.
     */
    language: string

    /**
     * DEPRECATED: This field is unused in known clients.
     * The date when this repository was created on Sourcegraph.
     */
    createdAt: DateTime

    /**
     * DEPRECATED: This field is unused in known clients.
     * The date when this repository's metadata was last updated on Sourcegraph.
     */
    updatedAt: DateTime | null

    /**
     * Returns information about the given commit in the repository, or null if no commit exists with the given rev.
     */
    commit: IGitCommit | null

    /**
     * Information and status related to mirroring, if this repository is a mirror of another repository (e.g., on
     * some code host). In this case, the remote source repository is external to Sourcegraph and the mirror is
     * maintained by the Sourcegraph site (not the other way around).
     */
    mirrorInfo: IMirrorRepositoryInfo

    /**
     * Information about this repository from the external service that it originates from (such as GitHub, GitLab,
     * Phabricator, etc.).
     */
    externalRepository: IExternalRepository

    /**
     * Whether the repository is a fork.
     */
    isFork: boolean

    /**
     * Whether the repository has been archived.
     */
    isArchived: boolean

    /**
     * Whether the repository is private.
     */
    isPrivate: boolean

    /**
     * Lists all external services which yield this repository.
     */
    externalServices: IExternalServiceConnection

    /**
     * Whether the repository is currently being cloned.
     * @deprecated "use Repository.mirrorInfo.cloneInProgress instead"
     */
    cloneInProgress: boolean

    /**
     * Information about the text search index for this repository, or null if text search indexing
     * is not enabled or supported for this repository.
     */
    textSearchIndex: IRepositoryTextSearchIndex | null

    /**
     * The URL to this repository.
     */
    url: string

    /**
     * The URLs to this repository on external services associated with it.
     */
    externalURLs: IExternalLink[]

    /**
     * The repository's default Git branch (HEAD symbolic ref). If the repository is currently being cloned or is
     * empty, this field will be null.
     */
    defaultBranch: IGitRef | null

    /**
     * The repository's Git refs.
     */
    gitRefs: IGitRefConnection

    /**
     * The repository's Git branches.
     */
    branches: IGitRefConnection

    /**
     * The repository's Git tags.
     */
    tags: IGitRefConnection

    /**
     * A Git comparison in this repository between a base and head commit.
     */
    comparison: IRepositoryComparison

    /**
     * The repository's contributors.
     */
    contributors: IRepositoryContributorConnection

    /**
     * Whether the viewer has admin privileges on this repository.
     */
    viewerCanAdminister: boolean

    /**
     * Base64 data uri to an icon.
     */
    icon: string

    /**
     * A markdown string that is rendered prominently.
     */
    label: IMarkdown

    /**
     * A markdown string of that is rendered less prominently.
     */
    detail: IMarkdown

    /**
     * The result previews of the result.
     */
    matches: ISearchResultMatch[]

    /**
     * (experimental) The LSIF API may change substantially in the near future as we
     * continue to adjust it for our use cases. Changes will not be documented in the
     * CHANGELOG during this time.
     * The repository's LSIF uploads.
     */
    lsifUploads: ILSIFUploadConnection

    /**
     * (experimental) The LSIF API may change substantially in the near future as we
     * continue to adjust it for our use cases. Changes will not be documented in the
     * CHANGELOG during this time.
     * The repository's LSIF uploads.
     */
    lsifIndexes: ILSIFIndexConnection

    /**
     * Gets the indexing configuration associated with the repository.
     */
    indexConfiguration: IIndexConfiguration | null

    /**
     * Information and status related to the commit graph of this repository calculated
     * for use by code intelligence features.
     */
    codeIntelligenceCommitGraph: ICodeIntelligenceCommitGraph

    /**
     * A list of authorized users to access this repository with the given permission.
     * This API currently only returns permissions from the Sourcegraph provider, i.e.
     * "permissions.userMapping" in site configuration.
     */
    authorizedUsers: IUserConnection

    /**
     * The permissions information of the repository for the authenticated user.
     * It is null when there is no permissions data stored for the repository.
     */
    permissionsInfo: IPermissionsInfo | null
}

export interface ICommitOnRepositoryArguments {
    /**
     * The Git revision specifier (revspec) for the commit.
     */
    rev: string

    /**
     * Optional input revspec used to construct non-canonical URLs and other "friendly" field values. Used by
     * clients that must ensure consistency of revision resolution within a session/request (so they use full
     * SHAs) but also preserve the user input rev (for user friendliness).
     */
    inputRevspec?: string | null
}

export interface IExternalServicesOnRepositoryArguments {
    /**
     * Returns the first n external services from the list.
     */
    first?: number | null
}

export interface IGitRefsOnRepositoryArguments {
    /**
     * Returns the first n Git refs from the list.
     */
    first?: number | null

    /**
     * Return Git refs whose names match the query.
     */
    query?: string | null

    /**
     * Return only Git refs of the given type.
     * Known issue: It is only supported to retrieve Git branch and tag refs, not
     * other Git refs.
     */
    type?: GitRefType | null

    /**
     * Ordering for Git refs in the list.
     */
    orderBy?: GitRefOrder | null

    /**
     * Ordering is an expensive operation that doesn't scale for lots of
     * references. If this is true we fallback on not ordering. This should
     * never be false in interactive API requests.
     * @default true
     */
    interactive?: boolean | null
}

export interface IBranchesOnRepositoryArguments {
    /**
     * Returns the first n Git branches from the list.
     */
    first?: number | null

    /**
     * Return Git branches whose names match the query.
     */
    query?: string | null

    /**
     * Ordering for Git branches in the list.
     */
    orderBy?: GitRefOrder | null

    /**
     * Ordering is an expensive operation that doesn't scale for lots of
     * references. If this is true we fallback on not ordering. This should
     * never be false in interactive API requests.
     * @default true
     */
    interactive?: boolean | null
}

export interface ITagsOnRepositoryArguments {
    /**
     * Returns the first n Git tags from the list.
     */
    first?: number | null

    /**
     * Return Git tags whose names match the query.
     */
    query?: string | null
}

export interface IComparisonOnRepositoryArguments {
    /**
     * The base of the diff ("old" or "left-hand side"), or "HEAD" if not specified.
     */
    base?: string | null

    /**
     * The head of the diff ("new" or "right-hand side"), or "HEAD" if not specified.
     */
    head?: string | null

    /**
     * Attempt to fetch missing revisions from remote if they are not found
     * @default true
     */
    fetchMissing?: boolean | null
}

export interface IContributorsOnRepositoryArguments {
    /**
     * The Git revision range to compute contributors in.
     */
    revisionRange?: string | null

    /**
     * The date after which to count contributions.
     */
    after?: string | null

    /**
     * Return contributors to files in this path.
     */
    path?: string | null

    /**
     * Returns the first n contributors from the list.
     */
    first?: number | null
}

export interface ILsifUploadsOnRepositoryArguments {
    /**
     * An (optional) search query that searches over the state, repository name,
     * commit, root, and indexer properties.
     */
    query?: string | null

    /**
     * The state of returned uploads.
     */
    state?: LSIFUploadState | null

    /**
     * When specified, shows only uploads that are latest for the given repository.
     */
    isLatestForRepo?: boolean | null

    /**
     * When specified, indicates that this request should be paginated and
     * the first N results (relative to the cursor) should be returned. i.e.
     * how many results to return per page. It must be in the range of 0-5000.
     */
    first?: number | null

    /**
     * When specified, indicates that this request should be paginated and
     * to fetch results starting at this cursor.
     * A future request can be made for more results by passing in the
     * 'LSIFUploadConnection.pageInfo.endCursor' that is returned.
     */
    after?: string | null
}

export interface ILsifIndexesOnRepositoryArguments {
    /**
     * An (optional) search query that searches over the state, repository name,
     * and commit properties.
     */
    query?: string | null

    /**
     * The state of returned uploads.
     */
    state?: LSIFIndexState | null

    /**
     * When specified, indicates that this request should be paginated and
     * the first N results (relative to the cursor) should be returned. i.e.
     * how many results to return per page. It must be in the range of 0-5000.
     */
    first?: number | null

    /**
     * When specified, indicates that this request should be paginated and
     * to fetch results starting at this cursor.
     * A future request can be made for more results by passing in the
     * 'LSIFIndexConnection.pageInfo.endCursor' that is returned.
     */
    after?: string | null
}

export interface IAuthorizedUsersOnRepositoryArguments {
    /**
     * Permission that the user has on this repository.
     * @default "READ"
     */
    permission?: RepositoryPermission | null

    /**
     * Number of users to return after the given cursor.
     */
    first: number

    /**
     * Opaque pagination cursor.
     */
    after?: string | null
}

/**
 * Information and status related to the commit graph of this repository calculated
 * for use by code intelligence features.
 */
export interface ICodeIntelligenceCommitGraph {
    __typename: 'CodeIntelligenceCommitGraph'

    /**
     * Whether or not the commit graph needs to be updated.
     */
    stale: boolean

    /**
     * When, if ever, the commit graph was last refreshed.
     */
    updatedAt: DateTime | null
}

/**
 * Permissions information of a repository or a user.
 */
export interface IPermissionsInfo {
    __typename: 'PermissionsInfo'

    /**
     * The permission levels that a user has on the repository.
     */
    permissions: RepositoryPermission[]

    /**
     * The last complete synced time, the value is updated only after a user- or repo-
     * centric sync of permissions. It is null when the complete sync never happened.
     */
    syncedAt: DateTime | null

    /**
     * The last updated time of permissions, the value is updated whenever there is a
     * change to the database row (i.e. incremental update).
     */
    updatedAt: DateTime
}

/**
 * A reference to another Sourcegraph instance.
 */
export interface IRedirect {
    __typename: 'Redirect'

    /**
     * The URL of the other Sourcegraph instance.
     */
    url: string
}

/**
 * A repository or a link to another Sourcegraph instance location where this repository may be located.
 */
export type RepositoryRedirect = IRepository | IRedirect

/**
 * A URL to a resource on an external service, such as the URL to a repository on its external (origin) code host.
 */
export interface IExternalLink {
    __typename: 'ExternalLink'

    /**
     * The URL to the resource.
     */
    url: string

    /**
     * The kind of external service, such as "GITHUB", or null if unknown/unrecognized. This is used solely for
     * displaying an icon that represents the service.
     */
    serviceKind: ExternalServiceKind | null

    /**
     * The type of external service, such as "github", or null if unknown/unrecognized. This is used solely for
     * displaying an icon that represents the service.
     * @deprecated "use name serviceKind instead"
     */
    serviceType: string | null
}

/**
 * Information and status about the mirroring of a repository. In this case, the remote source repository
 * is external to Sourcegraph and the mirror is maintained by the Sourcegraph site (not the other way
 * around).
 */
export interface IMirrorRepositoryInfo {
    __typename: 'MirrorRepositoryInfo'

    /**
     * The URL of the remote source repository.
     */
    remoteURL: string

    /**
     * Whether the clone of the repository has begun but not yet completed.
     */
    cloneInProgress: boolean

    /**
     * A single line of text that contains progress information for the running clone command.
     * The format of the progress text is not specified.
     * It is intended to be displayed directly to a user.
     * e.g.
     * "Receiving objects:  95% (2041/2148), 292.01 KiB | 515.00 KiB/s"
     * "Resolving deltas:   9% (117/1263)"
     */
    cloneProgress: string | null

    /**
     * Whether the repository has ever been successfully cloned.
     */
    cloned: boolean

    /**
     * When the repository was last successfully updated from the remote source repository..
     */
    updatedAt: DateTime | null

    /**
     * The state of this repository in the update schedule.
     */
    updateSchedule: IUpdateSchedule | null

    /**
     * The state of this repository in the update queue.
     */
    updateQueue: IUpdateQueue | null
}

/**
 * The state of a repository in the update schedule.
 */
export interface IUpdateSchedule {
    __typename: 'UpdateSchedule'

    /**
     * The interval that was used when scheduling the current due time.
     */
    intervalSeconds: number

    /**
     * The next time that the repo will be inserted into the update queue.
     */
    due: DateTime

    /**
     * The index of the repo in the schedule.
     */
    index: number

    /**
     * The total number of repos in the schedule.
     */
    total: number
}

/**
 * The state of a repository in the update queue.
 */
export interface IUpdateQueue {
    __typename: 'UpdateQueue'

    /**
     * The index of the repo in the update queue.
     * Updating repos are placed at the end of the queue until they finish updating
     * so don't display this if updating is true.
     */
    index: number

    /**
     * True if the repo is currently updating.
     */
    updating: boolean

    /**
     * The total number of repos in the update queue (including updating repos).
     */
    total: number
}

/**
 * A repository on an external service (such as GitHub, GitLab, Phabricator, etc.).
 */
export interface IExternalRepository {
    __typename: 'ExternalRepository'

    /**
     * The repository's ID on the external service.
     * Example: For GitHub, this is the GitHub GraphQL API's node ID for the repository.
     */
    id: string

    /**
     * The type of external service where this repository resides.
     * Example: "github", "gitlab", etc.
     */
    serviceType: string

    /**
     * The particular instance of the external service where this repository resides. Its value is
     * opaque but typically consists of the canonical base URL to the service.
     * Example: For GitHub.com, this is "https://github.com/".
     */
    serviceID: string
}

/**
 * (experimental) A version context. Used to change the set of default repository and revisions searched.
 * Note: We do not expose the list of repositories and revisions in the version context. This is intentional. However, if a need arises we can add it in.
 */
export interface IVersionContext {
    __typename: 'VersionContext'

    /**
     * The name of the version context.
     */
    name: string

    /**
     * The description of the version context.
     */
    description: string
}

/**
 * (experimental) A search context. Specifies a set of repositories to be searched.
 */
export interface ISearchContext {
    __typename: 'SearchContext'

    /**
     * The unique id of the search context.
     */
    id: ID

    /**
     * The description of the search context.
     */
    description: string

    /**
     * Fully-qualified search context spec for use when querying.
     * Examples: global, @username, @username/ctx, and @org/ctx.
     */
    spec: string

    /**
     * Whether the search context is autodefined by Sourcegraph. Current examples include:
     * global search context ("global"), default user search context ("@user"), and
     * default organization search context ("@org").
     */
    autoDefined: boolean
}

/**
 * Information about a repository's text search index.
 */
export interface IRepositoryTextSearchIndex {
    __typename: 'RepositoryTextSearchIndex'

    /**
     * The indexed repository.
     */
    repository: IRepository

    /**
     * The status of the text search index, if available.
     */
    status: IRepositoryTextSearchIndexStatus | null

    /**
     * Git refs in the repository that are configured for text search indexing.
     */
    refs: IRepositoryTextSearchIndexedRef[]
}

/**
 * The status of a repository's text search index.
 */
export interface IRepositoryTextSearchIndexStatus {
    __typename: 'RepositoryTextSearchIndexStatus'

    /**
     * The date that the index was last updated.
     */
    updatedAt: DateTime

    /**
     * The byte size of the original content.
     */
    contentByteSize: number

    /**
     * The number of files in the original content.
     */
    contentFilesCount: number

    /**
     * The byte size of the index.
     */
    indexByteSize: number

    /**
     * The number of index shards.
     */
    indexShardsCount: number

    /**
     * EXPERIMENTAL: The number of newlines appearing in the index.
     */
    newLinesCount: number

    /**
     * EXPERIMENTAL: The number of newlines in the default branch.
     */
    defaultBranchNewLinesCount: number

    /**
     * EXPERIMENTAL: The number of newlines in the other branches.
     */
    otherBranchesNewLinesCount: number
}

/**
 * A Git ref (usually a branch) in a repository that is configured to be indexed for text search.
 */
export interface IRepositoryTextSearchIndexedRef {
    __typename: 'RepositoryTextSearchIndexedRef'

    /**
     * The Git ref (usually a branch) that is configured to be indexed for text search. To find the specific commit
     * SHA that was indexed, use RepositoryTextSearchIndexedRef.indexedCommit; this field's ref target resolves to
     * the current target, not the target at the time of indexing.
     */
    ref: IGitRef

    /**
     * Whether a text search index exists for this ref.
     */
    indexed: boolean

    /**
     * Whether the text search index is of the current commit for the Git ref. If false, the index is stale.
     */
    current: boolean

    /**
     * The indexed Git commit (which may differ from the ref's current target if the index is out of date). If
     * indexed is false, this field's value is null.
     */
    indexedCommit: IGitObject | null
}

/**
 * A list of Git refs.
 */
export interface IGitRefConnection {
    __typename: 'GitRefConnection'

    /**
     * A list of Git refs.
     */
    nodes: IGitRef[]

    /**
     * The total count of Git refs in the connection. This total count may be larger
     * than the number of nodes in this object when the result is paginated.
     */
    totalCount: number

    /**
     * Pagination information.
     */
    pageInfo: IPageInfo
}

/**
 * Either a preview or an actual repository comparison.
 */
export type RepositoryComparisonInterface = IRepositoryComparison | IPreviewRepositoryComparison

/**
 * A not-yet-committed preview of a diff on a repository.
 */
export interface IPreviewRepositoryComparison {
    __typename: 'PreviewRepositoryComparison'

    /**
     * The repository that is the base (left-hand side) of this comparison.
     */
    baseRepository: IRepository

    /**
     * The file diffs for each changed file.
     */
    fileDiffs: IFileDiffConnection
}

export interface IFileDiffsOnPreviewRepositoryComparisonArguments {
    /**
     * Return the first n file diffs from the list.
     */
    first?: number | null

    /**
     * Return file diffs after the given cursor.
     */
    after?: string | null
}

/**
 * The differences between two concrete Git commits in a repository.
 */
export interface IRepositoryComparison {
    __typename: 'RepositoryComparison'

    /**
     * The repository that is the base (left-hand side) of this comparison.
     */
    baseRepository: IRepository

    /**
     * The repository that is the head (right-hand side) of this comparison. Cross-repository
     * comparisons are not yet supported, so this is always equal to
     * RepositoryComparison.baseRepository.
     */
    headRepository: IRepository

    /**
     * The range that this comparison represents.
     */
    range: IGitRevisionRange

    /**
     * The commits in the comparison range, excluding the base and including the head.
     */
    commits: IGitCommitConnection

    /**
     * The file diffs for each changed file.
     */
    fileDiffs: IFileDiffConnection
}

export interface ICommitsOnRepositoryComparisonArguments {
    /**
     * Return the first n commits from the list.
     */
    first?: number | null
}

export interface IFileDiffsOnRepositoryComparisonArguments {
    /**
     * Return the first n file diffs from the list.
     */
    first?: number | null

    /**
     * Return file diffs after the given cursor.
     */
    after?: string | null
}

/**
 * A list of file diffs.
 */
export interface IFileDiffConnection {
    __typename: 'FileDiffConnection'

    /**
     * A list of file diffs.
     */
    nodes: IFileDiff[]

    /**
     * The total count of file diffs in the connection, if available. This total count may be larger than the number
     * of nodes in this object when the result is paginated.
     */
    totalCount: number | null

    /**
     * Pagination information.
     */
    pageInfo: IPageInfo

    /**
     * The diff stat for the file diffs in this object, which may be a subset of the entire diff if the result is
     * paginated.
     */
    diffStat: IDiffStat

    /**
     * The raw diff for the file diffs in this object, which may be a subset of the entire diff if the result is
     * paginated.
     */
    rawDiff: string
}

/**
 * A diff for a single file.
 */
export interface IFileDiff {
    __typename: 'FileDiff'

    /**
     * The old (original) path of the file, or null if the file was added.
     */
    oldPath: string | null

    /**
     * The old file, or null if the file was created (oldFile.path == oldPath).
     */
    oldFile: File2 | null

    /**
     * The new (changed) path of the file, or null if the file was deleted.
     */
    newPath: string | null

    /**
     * The new file, or null if the file was deleted (newFile.path == newPath).
     */
    newFile: File2 | null

    /**
     * The old file (if the file was deleted) and otherwise the new file. This file field is typically used by
     * clients that want to show a "View" link to the file.
     */
    mostRelevantFile: File2

    /**
     * Hunks that were changed from old to new.
     */
    hunks: IFileDiffHunk[]

    /**
     * The diff stat for the whole file.
     */
    stat: IDiffStat

    /**
     * FOR INTERNAL USE ONLY.
     * An identifier for the file diff that is unique among all other file diffs in the list that
     * contains it.
     */
    internalID: string
}

/**
 * The type of content in a hunk line.
 */
export enum DiffHunkLineType {
    /**
     * Added line.
     */
    ADDED = 'ADDED',

    /**
     * Unchanged line.
     */
    UNCHANGED = 'UNCHANGED',

    /**
     * Deleted line.
     */
    DELETED = 'DELETED',
}

/**
 * A single highlighted line, including the kind of line.
 */
export interface IHighlightedDiffHunkLine {
    __typename: 'HighlightedDiffHunkLine'

    /**
     * The HTML containing the syntax-highlighted line of code.
     */
    html: string

    /**
     * The operation that happened on this line, in patches it is prefixed with '+', '-', ' '.
     * Can be either add, delete, or no change.
     */
    kind: DiffHunkLineType
}

/**
 * A highlighted hunk, consisting of all its lines.
 */
export interface IHighlightedDiffHunkBody {
    __typename: 'HighlightedDiffHunkBody'

    /**
     * Whether highlighting was aborted.
     */
    aborted: boolean

    /**
     * The highlighted lines.
     */
    lines: IHighlightedDiffHunkLine[]
}

/**
 * A specific highlighted line range to fetch.
 */
export interface IHighlightLineRange {
    /**
     * The first line to fetch (0-indexed, inclusive). Values outside the bounds of the file will
     * automatically be clamped within the valid range.
     */
    startLine: number

    /**
     * The last line to fetch (0-indexed, inclusive). Values outside the bounds of the file will
     * automatically be clamped within the valid range.
     */
    endLine: number
}

/**
 * A changed region ("hunk") in a file diff.
 */
export interface IFileDiffHunk {
    __typename: 'FileDiffHunk'

    /**
     * The range of the old file that the hunk applies to.
     */
    oldRange: IFileDiffHunkRange

    /**
     * Whether the old file had a trailing newline.
     */
    oldNoNewlineAt: boolean

    /**
     * The range of the new file that the hunk applies to.
     */
    newRange: IFileDiffHunkRange

    /**
     * The diff hunk section heading, if any.
     */
    section: string | null

    /**
     * The hunk body, with lines prefixed with '-', '+', or ' '.
     */
    body: string

    /**
     * Highlight the hunk.
     */
    highlight: IHighlightedDiffHunkBody
}

export interface IHighlightOnFileDiffHunkArguments {
    disableTimeout: boolean
    isLightTheme: boolean

    /**
     * If highlightLongLines is true, lines which are longer than 2000 bytes are highlighted.
     * 2000 bytes is enabled. This may produce a significant amount of HTML
     * which some browsers (such as Chrome, but not Firefox) may have trouble
     * rendering efficiently.
     * @default false
     */
    highlightLongLines?: boolean | null
}

/**
 * A hunk range in one side (old/new) of a diff.
 */
export interface IFileDiffHunkRange {
    __typename: 'FileDiffHunkRange'

    /**
     * The first line that the hunk applies to.
     */
    startLine: number

    /**
     * The number of lines that the hunk applies to.
     */
    lines: number
}

/**
 * Statistics about a diff.
 */
export interface IDiffStat {
    __typename: 'DiffStat'

    /**
     * Number of additions.
     */
    added: number

    /**
     * Number of changes.
     */
    changed: number

    /**
     * Number of deletions.
     */
    deleted: number
}

/**
 * A list of contributors to a repository.
 */
export interface IRepositoryContributorConnection {
    __typename: 'RepositoryContributorConnection'

    /**
     * A list of contributors to a repository.
     */
    nodes: IRepositoryContributor[]

    /**
     * The total count of contributors in the connection, if available. This total count may be larger than the
     * number of nodes in this object when the result is paginated.
     */
    totalCount: number

    /**
     * Pagination information.
     */
    pageInfo: IPageInfo
}

/**
 * A contributor to a repository.
 */
export interface IRepositoryContributor {
    __typename: 'RepositoryContributor'

    /**
     * The personal information for the contributor.
     */
    person: IPerson

    /**
     * The number of contributions made by this contributor.
     */
    count: number

    /**
     * The repository in which the contributions occurred.
     */
    repository: IRepository

    /**
     * Commits by the contributor.
     */
    commits: IGitCommitConnection
}

export interface ICommitsOnRepositoryContributorArguments {
    /**
     * Return the first n commits.
     */
    first?: number | null
}

/**
 * A code symbol (e.g., a function, variable, type, class, etc.).
 * It is derived from DocumentSymbol as defined in the Language Server Protocol (see https://microsoft.github.io/language-server-protocol/specifications/specification-3-14/#textDocument_documentSymbol).
 */
export interface ISymbol {
    __typename: 'Symbol'

    /**
     * The name of the symbol.
     */
    name: string

    /**
     * The name of the symbol that contains this symbol, if any. This field's value is not guaranteed to be
     * structured in such a way that callers can infer a hierarchy of symbols.
     */
    containerName: string | null

    /**
     * The kind of the symbol.
     */
    kind: SymbolKind

    /**
     * The programming language of the symbol.
     */
    language: string

    /**
     * The location where this symbol is defined.
     */
    location: ILocation

    /**
     * The URL to this symbol (using the input revision specifier, which may not be immutable).
     */
    url: string

    /**
     * The canonical URL to this symbol (using an immutable revision specifier).
     */
    canonicalURL: string

    /**
     * Whether or not the symbol is local to the file it's defined in.
     */
    fileLocal: boolean
}

/**
 * A location inside a resource (in a repository at a specific commit).
 */
export interface ILocation {
    __typename: 'Location'

    /**
     * The file that this location refers to.
     */
    resource: IGitBlob

    /**
     * The range inside the file that this location refers to.
     */
    range: IRange | null

    /**
     * The URL to this location (using the input revision specifier, which may not be immutable).
     */
    url: string

    /**
     * The canonical URL to this location (using an immutable revision specifier).
     */
    canonicalURL: string
}

/**
 * A range inside a file. The start position is inclusive, and the end position is exclusive.
 */
export interface IRange {
    __typename: 'Range'

    /**
     * The start position of the range (inclusive).
     */
    start: IPosition

    /**
     * The end position of the range (exclusive).
     */
    end: IPosition
}

/**
 * A zero-based position inside a file.
 */
export interface IPosition {
    __typename: 'Position'

    /**
     * The line number (zero-based) of the position.
     */
    line: number

    /**
     * The character offset (zero-based) in the line of the position.
     */
    character: number
}

/**
 * A list of diagnostics.
 */
export interface IDiagnosticConnection {
    __typename: 'DiagnosticConnection'

    /**
     * A list of diagnostics.
     */
    nodes: IDiagnostic[]

    /**
     * The total count of diagnostics (which may be larger than nodes.length if the connection is paginated).
     */
    totalCount: number

    /**
     * Pagination information.
     */
    pageInfo: IPageInfo
}

/**
 * Represents a diagnostic, such as a compiler error or warning.
 */
export interface IDiagnostic {
    __typename: 'Diagnostic'

    /**
     * The location at which the message applies.
     */
    location: ILocation

    /**
     * The diagnostic's severity.
     */
    severity: DiagnosticSeverity | null

    /**
     * The diagnostic's code as provided by the tool.
     */
    code: string | null

    /**
     * A human-readable string describing the source of this
     * diagnostic, e.g. "typescript" or "super lint".
     */
    source: string | null

    /**
     * The diagnostic's message.
     */
    message: string | null
}

/**
 * Represents the severity level of a diagnostic.
 */
export enum DiagnosticSeverity {
    ERROR = 'ERROR',
    WARNING = 'WARNING',
    INFORMATION = 'INFORMATION',
    HINT = 'HINT',
}

/**
 * All possible kinds of symbols. This set matches that of the Language Server Protocol
 * (https://microsoft.github.io/language-server-protocol/specification#workspace_symbol).
 */
export enum SymbolKind {
    UNKNOWN = 'UNKNOWN',
    FILE = 'FILE',
    MODULE = 'MODULE',
    NAMESPACE = 'NAMESPACE',
    PACKAGE = 'PACKAGE',
    CLASS = 'CLASS',
    METHOD = 'METHOD',
    PROPERTY = 'PROPERTY',
    FIELD = 'FIELD',
    CONSTRUCTOR = 'CONSTRUCTOR',
    ENUM = 'ENUM',
    INTERFACE = 'INTERFACE',
    FUNCTION = 'FUNCTION',
    VARIABLE = 'VARIABLE',
    CONSTANT = 'CONSTANT',
    STRING = 'STRING',
    NUMBER = 'NUMBER',
    BOOLEAN = 'BOOLEAN',
    ARRAY = 'ARRAY',
    OBJECT = 'OBJECT',
    KEY = 'KEY',
    NULL = 'NULL',
    ENUMMEMBER = 'ENUMMEMBER',
    STRUCT = 'STRUCT',
    EVENT = 'EVENT',
    OPERATOR = 'OPERATOR',
    TYPEPARAMETER = 'TYPEPARAMETER',
}

/**
 * A list of symbols.
 */
export interface ISymbolConnection {
    __typename: 'SymbolConnection'

    /**
     * A list of symbols.
     */
    nodes: ISymbol[]

    /**
     * Pagination information.
     */
    pageInfo: IPageInfo
}

/**
 * A Git ref.
 */
export interface IGitRef {
    __typename: 'GitRef'

    /**
     * The globally addressable ID for the Git ref.
     */
    id: ID

    /**
     * The full ref name (e.g., "refs/heads/mybranch" or "refs/tags/mytag").
     */
    name: string

    /**
     * An unambiguous short name for the ref.
     */
    abbrevName: string

    /**
     * The display name of the ref. For branches ("refs/heads/foo"), this is the branch
     * name ("foo").
     * As a special case, for GitHub pull request refs of the form refs/pull/NUMBER/head,
     * this is "NUMBER".
     */
    displayName: string

    /**
     * The prefix of the ref, either "", "refs/", "refs/heads/", "refs/pull/", or
     * "refs/tags/". This prefix is always a prefix of the ref's name.
     */
    prefix: string

    /**
     * The type of this Git ref.
     */
    type: GitRefType

    /**
     * The object that the ref points to.
     */
    target: IGitObject

    /**
     * The associated repository.
     */
    repository: IRepository

    /**
     * The URL to this Git ref.
     */
    url: string
}

/**
 * All possible types of Git refs.
 */
export enum GitRefType {
    /**
     * A Git branch (in refs/heads/).
     */
    GIT_BRANCH = 'GIT_BRANCH',

    /**
     * A Git tag (in refs/tags/).
     */
    GIT_TAG = 'GIT_TAG',

    /**
     * A Git ref that is neither a branch nor tag.
     */
    GIT_REF_OTHER = 'GIT_REF_OTHER',
}

/**
 * Ordering options for Git refs.
 */
export enum GitRefOrder {
    /**
     * By the authored or committed at date, whichever is more recent.
     */
    AUTHORED_OR_COMMITTED_AT = 'AUTHORED_OR_COMMITTED_AT',
}

/**
 * A Git object.
 */
export interface IGitObject {
    __typename: 'GitObject'

    /**
     * This object's OID.
     */
    oid: GitObjectID

    /**
     * The abbreviated form of this object's OID.
     */
    abbreviatedOID: string

    /**
     * The commit object, if it is a commit and it exists; otherwise null.
     */
    commit: IGitCommit | null

    /**
     * The Git object's type.
     */
    type: GitObjectType
}

/**
 * All possible types of Git objects.
 */
export enum GitObjectType {
    /**
     * A Git commit object.
     */
    GIT_COMMIT = 'GIT_COMMIT',

    /**
     * A Git tag object.
     */
    GIT_TAG = 'GIT_TAG',

    /**
     * A Git tree object.
     */
    GIT_TREE = 'GIT_TREE',

    /**
     * A Git blob object.
     */
    GIT_BLOB = 'GIT_BLOB',

    /**
     * A Git object of unknown type.
     */
    GIT_UNKNOWN = 'GIT_UNKNOWN',
}

/**
 * A Git revspec expression that (possibly) resolves to a Git revision.
 */
export interface IGitRevSpecExpr {
    __typename: 'GitRevSpecExpr'

    /**
     * The original Git revspec expression.
     */
    expr: string

    /**
     * The Git object that the revspec resolves to, or null otherwise.
     */
    object: IGitObject | null
}

/**
 * A Git revspec.
 */
export type GitRevSpec = IGitRef | IGitRevSpecExpr | IGitObject

/**
 * A Git revision range of the form "base..head" or "base...head". Other revision
 * range formats are not supported.
 */
export interface IGitRevisionRange {
    __typename: 'GitRevisionRange'

    /**
     * The Git revision range expression of the form "base..head" or "base...head".
     */
    expr: string

    /**
     * The base (left-hand side) of the range.
     */
    base: GitRevSpec

    /**
     * The base's revspec as an expression.
     */
    baseRevSpec: IGitRevSpecExpr

    /**
     * The head (right-hand side) of the range.
     */
    head: GitRevSpec

    /**
     * The head's revspec as an expression.
     */
    headRevSpec: IGitRevSpecExpr

    /**
     * The merge-base of the base and head revisions, if this is a "base...head"
     * revision range. If this is a "base..head" revision range, then this field is null.
     */
    mergeBase: IGitObject | null
}

/**
 * A Phabricator repository.
 */
export interface IPhabricatorRepo {
    __typename: 'PhabricatorRepo'

    /**
     * The canonical repo name (e.g. "github.com/gorilla/mux").
     */
    name: string

    /**
     * An alias for name.
     * @deprecated "use name instead"
     */
    uri: string

    /**
     * The unique Phabricator identifier for the repo, like "MUX"
     */
    callsign: string

    /**
     * The URL to the phabricator instance (e.g. http://phabricator.sgdev.org)
     */
    url: string
}

/**
 * Pagination information. See https://facebook.github.io/relay/graphql/connections.htm#sec-undefined.PageInfo.
 */
export interface IPageInfo {
    __typename: 'PageInfo'

    /**
     * When paginating forwards, the cursor to continue.
     */
    endCursor: string | null

    /**
     * When paginating forwards, are there more items?
     */
    hasNextPage: boolean
}

/**
 * A list of Git commits.
 */
export interface IGitCommitConnection {
    __typename: 'GitCommitConnection'

    /**
     * A list of Git commits.
     */
    nodes: IGitCommit[]

    /**
     * The total number of Git commits in the connection. If the GitCommitConnection is paginated
     * (e.g., because a "first" parameter was provided to the field that produced it), this field is
     * null to avoid it taking unexpectedly long to compute the total count. Remove the pagination
     * parameters to obtain a non-null value for this field.
     */
    totalCount: number | null

    /**
     * Pagination information.
     */
    pageInfo: IPageInfo
}

/**
 * Statistics about a language's usage.
 */
export interface ILanguageStatistics {
    __typename: 'LanguageStatistics'

    /**
     * The name of the language.
     */
    name: string

    /**
     * The total bytes in the language.
     */
    totalBytes: number

    /**
     * The total number of lines in the language.
     */
    totalLines: number
}

/**
 * A Git commit.
 */
export interface IGitCommit {
    __typename: 'GitCommit'

    /**
     * The globally addressable ID for this commit.
     */
    id: ID

    /**
     * The repository that contains this commit.
     */
    repository: IRepository

    /**
     * This commit's Git object ID (OID), a 40-character SHA-1 hash.
     */
    oid: GitObjectID

    /**
     * The abbreviated form of this commit's OID.
     */
    abbreviatedOID: string

    /**
     * This commit's author.
     */
    author: ISignature

    /**
     * This commit's committer, if any.
     */
    committer: ISignature | null

    /**
     * The full commit message.
     */
    message: string

    /**
     * The first line of the commit message.
     */
    subject: string

    /**
     * The contents of the commit message after the first line.
     */
    body: string | null

    /**
     * Parent commits of this commit.
     */
    parents: IGitCommit[]

    /**
     * The URL to this commit (using the input revision specifier, which may not be immutable).
     */
    url: string

    /**
     * The canonical URL to this commit (using an immutable revision specifier).
     */
    canonicalURL: string

    /**
     * The URLs to this commit on its repository's external services.
     */
    externalURLs: IExternalLink[]

    /**
     * The Git tree in this commit at the given path.
     */
    tree: IGitTree | null

    /**
     * The Git blob in this commit at the given path.
     */
    blob: IGitBlob | null

    /**
     * The file at the given path for this commit.
     * See "File" documentation for the difference between this field and the "blob" field.
     */
    file: File2 | null

    /**
     * Lists the programming languages present in the tree at this commit.
     */
    languages: string[]

    /**
     * List statistics for each language present in the repository.
     */
    languageStatistics: ILanguageStatistics[]

    /**
     * The log of commits consisting of this commit and its ancestors.
     */
    ancestors: IGitCommitConnection

    /**
     * Returns the number of commits that this commit is behind and ahead of revspec.
     */
    behindAhead: IBehindAheadCounts

    /**
     * Symbols defined as of this commit. (All symbols, not just symbols that were newly defined in this commit.)
     */
    symbols: ISymbolConnection
}

export interface ITreeOnGitCommitArguments {
    /**
     * The path of the tree.
     * @default ""
     */
    path?: string | null

    /**
     * Whether to recurse into sub-trees. If true, it overrides the value of the "recursive" parameter on all of
     * GitTree's fields.
     * DEPRECATED: Use the "recursive" parameter on GitTree's fields instead.
     * @default false
     */
    recursive?: boolean | null
}

export interface IBlobOnGitCommitArguments {
    path: string
}

export interface IFileOnGitCommitArguments {
    path: string
}

export interface IAncestorsOnGitCommitArguments {
    /**
     * Returns the first n commits from the list.
     */
    first?: number | null

    /**
     * Return commits that match the query.
     */
    query?: string | null

    /**
     * Return commits that affect the path.
     */
    path?: string | null

    /**
     * Return commits more recent than the specified date.
     */
    after?: string | null
}

export interface IBehindAheadOnGitCommitArguments {
    revspec: string
}

export interface ISymbolsOnGitCommitArguments {
    /**
     * Returns the first n symbols from the list.
     */
    first?: number | null

    /**
     * Return symbols matching the query.
     */
    query?: string | null

    /**
     * A list of regular expressions, all of which must match all
     * file paths returned in the list.
     */
    includePatterns?: string[] | null
}

/**
 * A set of Git behind/ahead counts for one commit relative to another.
 */
export interface IBehindAheadCounts {
    __typename: 'BehindAheadCounts'

    /**
     * The number of commits behind the other commit.
     */
    behind: number

    /**
     * The number of commits ahead of the other commit.
     */
    ahead: number
}

/**
 * A signature.
 */
export interface ISignature {
    __typename: 'Signature'

    /**
     * The person.
     */
    person: IPerson

    /**
     * The date.
     */
    date: string
}

/**
 * A person.
 */
export interface IPerson {
    __typename: 'Person'

    /**
     * The name.
     */
    name: string

    /**
     * The email.
     */
    email: string

    /**
     * The name if set; otherwise the email username.
     */
    displayName: string

    /**
     * The avatar URL, if known.
     */
    avatarURL: string | null

    /**
     * The corresponding user account for this person, if one exists.
     */
    user: IUser | null
}

/**
 * A Git submodule
 */
export interface ISubmodule {
    __typename: 'Submodule'

    /**
     * The remote repository URL of the submodule.
     */
    url: string

    /**
     * The commit of the submodule.
     */
    commit: string

    /**
     * The path to which the submodule is checked out.
     */
    path: string
}

/**
 * A file, directory, or other tree entry.
 */
export type TreeEntry = IGitTree | IGitBlob

/**
 * A file, directory, or other tree entry.
 */
export interface ITreeEntry {
    __typename: 'TreeEntry'

    /**
     * The full path (relative to the repository root) of this tree entry.
     */
    path: string

    /**
     * The base name (i.e., file name only) of this tree entry.
     */
    name: string

    /**
     * Whether this tree entry is a directory.
     */
    isDirectory: boolean

    /**
     * The URL to this tree entry (using the input revision specifier, which may not be immutable).
     */
    url: string

    /**
     * The canonical URL to this tree entry (using an immutable revision specifier).
     */
    canonicalURL: string

    /**
     * The URLs to this tree entry on external services.
     */
    externalURLs: IExternalLink[]

    /**
     * Symbols defined in this file or directory.
     */
    symbols: ISymbolConnection

    /**
     * Submodule metadata if this tree points to a submodule
     */
    submodule: ISubmodule | null

    /**
     * Whether this tree entry is a single child
     */
    isSingleChild: boolean

    /**
     * (experimental) The LSIF API may change substantially in the near future as we
     * continue to adjust it for our use cases. Changes will not be documented in the
     * CHANGELOG during this time.
     * LSIF data for this tree entry.
     */
    lsif: TreeEntryLSIFData | null
}

export interface ISymbolsOnTreeEntryArguments {
    /**
     * Returns the first n symbols from the list.
     */
    first?: number | null

    /**
     * Return symbols matching the query.
     */
    query?: string | null
}

export interface IIsSingleChildOnTreeEntryArguments {
    /**
     * Returns the first n files in the tree.
     */
    first?: number | null

    /**
     * Recurse into sub-trees.
     * @default false
     */
    recursive?: boolean | null

    /**
     * Recurse into sub-trees of single-child directories
     * @default false
     */
    recursiveSingleChild?: boolean | null
}

export interface ILsifOnTreeEntryArguments {
    /**
     * An optional filter for the name of the tool that produced the upload data.
     */
    toolName?: string | null
}

/**
 * A Git tree in a repository.
 */
export interface IGitTree {
    __typename: 'GitTree'

    /**
     * The full path (relative to the root) of this tree.
     */
    path: string

    /**
     * Whether this tree is the root (top-level) tree.
     */
    isRoot: boolean

    /**
     * The base name (i.e., last path component only) of this tree.
     */
    name: string

    /**
     * True because this is a directory. (The value differs for other TreeEntry interface implementations, such as
     * File.)
     */
    isDirectory: boolean

    /**
     * The Git commit containing this tree.
     */
    commit: IGitCommit

    /**
     * The repository containing this tree.
     */
    repository: IRepository

    /**
     * The URL to this tree (using the input revision specifier, which may not be immutable).
     */
    url: string

    /**
     * The canonical URL to this tree (using an immutable revision specifier).
     */
    canonicalURL: string

    /**
     * The URLs to this tree on external services.
     */
    externalURLs: IExternalLink[]

    /**
     * The URL to this entry's raw contents as a Zip archive.
     */
    rawZipArchiveURL: string

    /**
     * Submodule metadata if this tree points to a submodule
     */
    submodule: ISubmodule | null

    /**
     * A list of directories in this tree.
     */
    directories: IGitTree[]

    /**
     * A list of files in this tree.
     */
    files: IFile[]

    /**
     * A list of entries in this tree.
     */
    entries: TreeEntry[]

    /**
     * Symbols defined in this tree.
     */
    symbols: ISymbolConnection

    /**
     * Whether this tree entry is a single child
     */
    isSingleChild: boolean

    /**
     * (experimental) The LSIF API may change substantially in the near future as we
     * continue to adjust it for our use cases. Changes will not be documented in the
     * CHANGELOG during this time.
     * LSIF data for this tree entry.
     */
    lsif: TreeEntryLSIFData | null
}

export interface IDirectoriesOnGitTreeArguments {
    /**
     * Returns the first n files in the tree.
     */
    first?: number | null

    /**
     * Recurse into sub-trees.
     * @default false
     */
    recursive?: boolean | null
}

export interface IFilesOnGitTreeArguments {
    /**
     * Returns the first n files in the tree.
     */
    first?: number | null

    /**
     * Recurse into sub-trees.
     * @default false
     */
    recursive?: boolean | null
}

export interface IEntriesOnGitTreeArguments {
    /**
     * Returns the first n files in the tree.
     */
    first?: number | null

    /**
     * Recurse into sub-trees. If true, implies recursiveSingleChild.
     * @default false
     */
    recursive?: boolean | null

    /**
     * Recurse into sub-trees of single-child directories. If true, we return a flat list of
     * every directory that is a single child, and any directories or files that are
     * nested in a single child.
     * @default false
     */
    recursiveSingleChild?: boolean | null
}

export interface ISymbolsOnGitTreeArguments {
    /**
     * Returns the first n symbols from the list.
     */
    first?: number | null

    /**
     * Return symbols matching the query.
     */
    query?: string | null
}

export interface IIsSingleChildOnGitTreeArguments {
    /**
     * Returns the first n files in the tree.
     */
    first?: number | null

    /**
     * Recurse into sub-trees.
     * @default false
     */
    recursive?: boolean | null

    /**
     * Recurse into sub-trees of single-child directories
     * @default false
     */
    recursiveSingleChild?: boolean | null
}

export interface ILsifOnGitTreeArguments {
    /**
     * An optional filter for the name of the tool that produced the upload data.
     */
    toolName?: string | null
}

/**
 * A file.
 * In a future version of Sourcegraph, a repository's files may be distinct from a repository's blobs
 * (for example, to support searching/browsing generated files that aren't committed and don't exist
 * as Git blobs). Clients should generally use the GitBlob concrete type and GitCommit.blobs (not
 * GitCommit.files), unless they explicitly want to opt-in to different behavior in the future.
 * INTERNAL: This is temporarily named File2 during a migration. Do not refer to the name File2 in
 * any API clients as the name will change soon.
 */
export type File2 = IVirtualFile | IGitBlob

/**
 * A file.
 * In a future version of Sourcegraph, a repository's files may be distinct from a repository's blobs
 * (for example, to support searching/browsing generated files that aren't committed and don't exist
 * as Git blobs). Clients should generally use the GitBlob concrete type and GitCommit.blobs (not
 * GitCommit.files), unless they explicitly want to opt-in to different behavior in the future.
 * INTERNAL: This is temporarily named File2 during a migration. Do not refer to the name File2 in
 * any API clients as the name will change soon.
 */
export interface IFile2 {
    __typename: 'File2'

    /**
     * The full path (relative to the root) of this file.
     */
    path: string

    /**
     * The base name (i.e., file name only) of this file.
     */
    name: string

    /**
     * False because this is a file, not a directory.
     */
    isDirectory: boolean

    /**
     * The content of this file.
     */
    content: string

    /**
     * The file size in bytes.
     */
    byteSize: number

    /**
     * Whether or not it is binary.
     */
    binary: boolean

    /**
     * The file rendered as rich HTML, or an empty string if it is not a supported
     * rich file type.
     * This HTML string is already escaped and thus is always safe to render.
     */
    richHTML: string

    /**
     * The URL to this file (using the input revision specifier, which may not be immutable).
     */
    url: string

    /**
     * The canonical URL to this file (using an immutable revision specifier).
     */
    canonicalURL: string

    /**
     * The URLs to this file on external services.
     */
    externalURLs: IExternalLink[]

    /**
     * Highlight the file.
     */
    highlight: IHighlightedFile
}

export interface IHighlightOnFile2Arguments {
    disableTimeout: boolean
    isLightTheme: boolean

    /**
     * If highlightLongLines is true, lines which are longer than 2000 bytes are highlighted.
     * 2000 bytes is enabled. This may produce a significant amount of HTML
     * which some browsers (such as Chrome, but not Firefox) may have trouble
     * rendering efficiently.
     * @default false
     */
    highlightLongLines?: boolean | null
}

/**
 * A virtual file is an arbitrary file that is generated in memory.
 */
export interface IVirtualFile {
    __typename: 'VirtualFile'

    /**
     * The full path (relative to the root) of this file.
     */
    path: string

    /**
     * The base name (i.e., file name only) of this file.
     */
    name: string

    /**
     * False because this is a file, not a directory.
     */
    isDirectory: boolean

    /**
     * The content of this file.
     */
    content: string

    /**
     * The file size in bytes.
     */
    byteSize: number

    /**
     * Whether or not it is binary.
     */
    binary: boolean

    /**
     * The file rendered as rich HTML, or an empty string if it is not a supported
     * rich file type.
     * This HTML string is already escaped and thus is always safe to render.
     */
    richHTML: string

    /**
     * Not implemented.
     */
    url: string

    /**
     * Not implemented.
     */
    canonicalURL: string

    /**
     * Not implemented.
     */
    externalURLs: IExternalLink[]

    /**
     * Highlight the file.
     */
    highlight: IHighlightedFile
}

export interface IHighlightOnVirtualFileArguments {
    disableTimeout: boolean
    isLightTheme: boolean

    /**
     * If highlightLongLines is true, lines which are longer than 2000 bytes are highlighted.
     * 2000 bytes is enabled. This may produce a significant amount of HTML
     * which some browsers (such as Chrome, but not Firefox) may have trouble
     * rendering efficiently.
     * @default false
     */
    highlightLongLines?: boolean | null
}

/**
 * File is temporarily preserved for backcompat with browser extension search API client code.
 */
export interface IFile {
    __typename: 'File'

    /**
     * The full path (relative to the repository root) of this file.
     */
    path: string

    /**
     * The base name (i.e., file name only) of this file's path.
     */
    name: string

    /**
     * Whether this is a directory.
     */
    isDirectory: boolean

    /**
     * The URL to this file on Sourcegraph.
     */
    url: string

    /**
     * The repository that contains this file.
     */
    repository: IRepository
}

/**
 * A Git blob in a repository.
 */
export interface IGitBlob {
    __typename: 'GitBlob'

    /**
     * The full path (relative to the repository root) of this blob.
     */
    path: string

    /**
     * The base name (i.e., file name only) of this blob's path.
     */
    name: string

    /**
     * False because this is a blob (file), not a directory.
     */
    isDirectory: boolean

    /**
     * The content of this blob.
     */
    content: string

    /**
     * The file size in bytes.
     */
    byteSize: number

    /**
     * Whether or not it is binary.
     */
    binary: boolean

    /**
     * The blob contents rendered as rich HTML, or an empty string if it is not a supported
     * rich file type.
     * This HTML string is already escaped and thus is always safe to render.
     */
    richHTML: string

    /**
     * The Git commit containing this blob.
     */
    commit: IGitCommit

    /**
     * The repository containing this Git blob.
     */
    repository: IRepository

    /**
     * The URL to this blob (using the input revision specifier, which may not be immutable).
     */
    url: string

    /**
     * The canonical URL to this blob (using an immutable revision specifier).
     */
    canonicalURL: string

    /**
     * The URLs to this blob on its repository's external services.
     */
    externalURLs: IExternalLink[]

    /**
     * Blame the blob.
     */
    blame: IHunk[]

    /**
     * Highlight the blob contents.
     */
    highlight: IHighlightedFile

    /**
     * Submodule metadata if this tree points to a submodule
     */
    submodule: ISubmodule | null

    /**
     * Symbols defined in this blob.
     */
    symbols: ISymbolConnection

    /**
     * Always false, since a blob is a file, not directory.
     */
    isSingleChild: boolean

    /**
     * (experimental) The LSIF API may change substantially in the near future as we
     * continue to adjust it for our use cases. Changes will not be documented in the
     * CHANGELOG during this time.
     * A wrapper around LSIF query methods. If no LSIF upload can be used to answer code
     * intelligence queries for this path-at-revision, this resolves to null.
     */
    lsif: IGitBlobLSIFData | null
}

export interface IBlameOnGitBlobArguments {
    startLine: number
    endLine: number
}

export interface IHighlightOnGitBlobArguments {
    disableTimeout: boolean
    isLightTheme: boolean

    /**
     * If highlightLongLines is true, lines which are longer than 2000 bytes are highlighted.
     * 2000 bytes is enabled. This may produce a significant amount of HTML
     * which some browsers (such as Chrome, but not Firefox) may have trouble
     * rendering efficiently.
     * @default false
     */
    highlightLongLines?: boolean | null
}

export interface ISymbolsOnGitBlobArguments {
    /**
     * Returns the first n symbols from the list.
     */
    first?: number | null

    /**
     * Return symbols matching the query.
     */
    query?: string | null
}

export interface IIsSingleChildOnGitBlobArguments {
    /**
     * Returns the first n files in the tree.
     */
    first?: number | null

    /**
     * Recurse into sub-trees.
     * @default false
     */
    recursive?: boolean | null

    /**
     * Recurse into sub-trees of single-child directories
     * @default false
     */
    recursiveSingleChild?: boolean | null
}

export interface ILsifOnGitBlobArguments {
    /**
     * An optional filter for the name of the tool that produced the upload data.
     */
    toolName?: string | null
}

/**
 * LSIF data available for a tree entry.
 */
export type TreeEntryLSIFData = IGitBlobLSIFData

/**
 * LSIF data available for a tree entry.
 */
export interface ITreeEntryLSIFData {
    __typename: 'TreeEntryLSIFData'

    /**
     * Code diagnostics provided through LSIF.
     */
    diagnostics: IDiagnosticConnection
}

export interface IDiagnosticsOnTreeEntryLSIFDataArguments {
    first?: number | null
}

/**
 * A wrapper object around LSIF query methods for a particular path-at-revision. When this node is
 * null, no LSIF data is available for containing git blob.
 */
export interface IGitBlobLSIFData {
    __typename: 'GitBlobLSIFData'

    /**
     * (experimental) The LSIF API may change substantially in the near future as we
     * continue to adjust it for our use cases. Changes will not be documented in the
     * CHANGELOG during this time.
     * Get aggregated local code intelligence for all ranges that fall in the window
     * indicated by the given zero-based start (inclusive) and end (exclusive) lines.
     * The associated data for each range is "local", in that the locations and hover
     * must also be defined in the same index as the source range. To get cross-repository
     * and cross-bundle results, you must query the definitions, references, and hovers
     * of that range explicitly.
     */
    ranges: ICodeIntelligenceRangeConnection | null

    /**
     * (experimental) The LSIF API may change substantially in the near future as we
     * continue to adjust it for our use cases. Changes will not be documented in the
     * CHANGELOG during this time.
     * A list of definitions of the symbol under the given document position.
     */
    definitions: ILocationConnection

    /**
     * (experimental) The LSIF API may change substantially in the near future as we
     * continue to adjust it for our use cases. Changes will not be documented in the
     * CHANGELOG during this time.
     * A list of references of the symbol under the given document position.
     */
    references: ILocationConnection

    /**
     * (experimental) The LSIF API may change substantially in the near future as we
     * continue to adjust it for our use cases. Changes will not be documented in the
     * CHANGELOG during this time.
     * The hover result of the symbol under the given document position.
     */
    hover: IHover | null

    /**
     * (experimental) The LSIF API may change substantially in the near future as we
     * continue to adjust it for our use cases. Changes will not be documented in the
     * CHANGELOG during this time.
     * Code diagnostics provided through LSIF.
     */
    diagnostics: IDiagnosticConnection
}

export interface IRangesOnGitBlobLSIFDataArguments {
    startLine: number
    endLine: number
}

export interface IDefinitionsOnGitBlobLSIFDataArguments {
    /**
     * The line on which the symbol occurs (zero-based, inclusive).
     */
    line: number

    /**
     * The character (not byte) of the start line on which the symbol occurs (zero-based, inclusive).
     */
    character: number
}

export interface IReferencesOnGitBlobLSIFDataArguments {
    /**
     * The line on which the symbol occurs (zero-based, inclusive).
     */
    line: number

    /**
     * The character (not byte) of the start line on which the symbol occurs (zero-based, inclusive).
     */
    character: number

    /**
     * When specified, indicates that this request should be paginated and
     * to fetch results starting at this cursor.
     * A future request can be made for more results by passing in the
     * 'LocationConnection.pageInfo.endCursor' that is returned.
     */
    after?: string | null

    /**
     * When specified, indicates that this request should be paginated and
     * the first N results (relative to the cursor) should be returned. i.e.
     * how many results to return per page.
     */
    first?: number | null
}

export interface IHoverOnGitBlobLSIFDataArguments {
    /**
     * The line on which the symbol occurs (zero-based, inclusive).
     */
    line: number

    /**
     * The character (not byte) of the start line on which the symbol occurs (zero-based, inclusive).
     */
    character: number
}

export interface IDiagnosticsOnGitBlobLSIFDataArguments {
    first?: number | null
}

/**
 * A highlighted file.
 */
export interface IHighlightedFile {
    __typename: 'HighlightedFile'

    /**
     * Whether or not it was aborted.
     */
    aborted: boolean

    /**
     * The HTML table that can be used to display the highlighted file.
     */
    html: string

    /**
     * A list of the desired line ranges. Each list is a list of lines, where each element is an HTML
     * table row '<tr>...</tr>' string. This is useful if you only need to display specific subsets of
     * the file.
     */
    lineRanges: string[][]
}

export interface ILineRangesOnHighlightedFileArguments {
    ranges: IHighlightLineRange[]
}

/**
 * A file match.
 */
export interface IFileMatch {
    __typename: 'FileMatch'

    /**
     * The file containing the match.
     * KNOWN ISSUE: This file's "commit" field contains incomplete data.
     * KNOWN ISSUE: This field's type should be File! not GitBlob!.
     */
    file: IGitBlob

    /**
     * The repository containing the file match.
     */
    repository: IRepository

    /**
     * The revspec of the revision that contains this match. If no revspec was given (such as when no
     * repository filter or revspec is specified in the search query), it is null.
     */
    revSpec: GitRevSpec | null

    /**
     * The resource.
     * @deprecated "use the file field instead"
     */
    resource: string

    /**
     * The symbols found in this file that match the query.
     */
    symbols: ISymbol[]

    /**
     * The line matches.
     */
    lineMatches: ILineMatch[]

    /**
     * Whether or not the limit was hit.
     */
    limitHit: boolean
}

/**
 * A line match.
 */
export interface ILineMatch {
    __typename: 'LineMatch'

    /**
     * The preview.
     */
    preview: string

    /**
     * The line number. 0-based. The first line will have lineNumber 0. Note: A
     * UI will normally display line numbers 1-based.
     */
    lineNumber: number

    /**
     * Tuples of [offset, length] measured in characters (not bytes).
     */
    offsetAndLengths: number[][]

    /**
     * Whether or not the limit was hit.
     */
    limitHit: boolean
}

/**
 * A hunk.
 */
export interface IHunk {
    __typename: 'Hunk'

    /**
     * The startLine.
     */
    startLine: number

    /**
     * The endLine.
     */
    endLine: number

    /**
     * The startByte.
     */
    startByte: number

    /**
     * The endByte.
     */
    endByte: number

    /**
     * The rev.
     */
    rev: string

    /**
     * The author.
     */
    author: ISignature

    /**
     * The message.
     */
    message: string

    /**
     * The commit that contains the hunk.
     */
    commit: IGitCommit
}

/**
 * A namespace is a container for certain types of data and settings, such as a user or organization.
 */
export type Namespace = IUser | IOrg

/**
 * A namespace is a container for certain types of data and settings, such as a user or organization.
 */
export interface INamespace {
    __typename: 'Namespace'

    /**
     * The globally unique ID of this namespace.
     */
    id: ID

    /**
     * The name of this namespace's component. For a user, this is the username. For an organization,
     * this is the organization name.
     */
    namespaceName: string

    /**
     * The URL to this namespace.
     */
    url: string
}

/**
 * A list of users.
 */
export interface IUserConnection {
    __typename: 'UserConnection'

    /**
     * A list of users.
     */
    nodes: IUser[]

    /**
     * The total count of users in the connection. This total count may be larger
     * than the number of nodes in this object when the result is paginated.
     */
    totalCount: number

    /**
     * Pagination information.
     */
    pageInfo: IPageInfo
}

/**
 * A user.
 */
export interface IUser {
    __typename: 'User'

    /**
     * The unique ID for the user.
     */
    id: ID

    /**
     * The user's username.
     */
    username: string

    /**
     * The user's primary email address.
     * Only the user and site admins can access this field.
     * @deprecated "use emails instead"
     */
    email: string

    /**
     * The display name chosen by the user.
     */
    displayName: string | null

    /**
     * The URL of the user's avatar image.
     */
    avatarURL: string | null

    /**
     * The URL to the user's profile on Sourcegraph.
     */
    url: string

    /**
     * The URL to the user's settings.
     */
    settingsURL: string | null

    /**
     * The date when the user account was created on Sourcegraph.
     */
    createdAt: DateTime

    /**
     * The date when the user account was last updated on Sourcegraph.
     */
    updatedAt: DateTime | null

    /**
     * Whether the user is a site admin.
     * Only the user and site admins can access this field.
     */
    siteAdmin: boolean

    /**
     * Whether the user account uses built in auth.
     */
    builtinAuth: boolean

    /**
     * The latest settings for the user.
     * Only the user and site admins can access this field.
     */
    latestSettings: ISettings | null

    /**
     * All settings for this user, and the individual levels in the settings cascade (global > organization > user)
     * that were merged to produce the final merged settings.
     * Only the user and site admins can access this field.
     */
    settingsCascade: ISettingsCascade

    /**
     * DEPRECATED
     * @deprecated "Use settingsCascade instead. This field is a deprecated alias for it and will be removed in a future release."
     */
    configurationCascade: IConfigurationCascade

    /**
     * The organizations that this user is a member of.
     */
    organizations: IOrgConnection

    /**
     * This user's organization memberships.
     */
    organizationMemberships: IOrganizationMembershipConnection

    /**
     * Tags associated with the user. These are used for internal site management and feature selection.
     * Only the user and site admins can access this field.
     */
    tags: string[]

    /**
     * The user's usage statistics on Sourcegraph.
     */
    usageStatistics: IUserUsageStatistics

    /**
     * The user's events on Sourcegraph.
     */
    eventLogs: IEventLogsConnection

    /**
     * The user's email addresses.
     * Only the user and site admins can access this field.
     */
    emails: IUserEmail[]

    /**
     * The user's access tokens (which grant to the holder the privileges of the user). This consists
     * of all access tokens whose subject is this user.
     * Only the user and site admins can access this field.
     */
    accessTokens: IAccessTokenConnection

    /**
     * A list of external accounts that are associated with the user.
     */
    externalAccounts: IExternalAccountConnection

    /**
     * The user's currently active session.
     * Only the currently authenticated user can access this field. Site admins are not able to access sessions for
     * other users.
     */
    session: ISession

    /**
     * Whether the viewer has admin privileges on this user. The user has admin privileges on their own user, and
     * site admins have admin privileges on all users.
     */
    viewerCanAdminister: boolean

    /**
     * Whether the viewer can change the username of this user.
     * The user can change their username unless auth.disableUsernameChanges is set.
     * Site admins can always change the username of any user.
     */
    viewerCanChangeUsername: boolean

    /**
     * The user's survey responses.
     * Only the user and site admins can access this field.
     */
    surveyResponses: ISurveyResponse[]

    /**
     * The URL to view this user's customer information (for Sourcegraph.com site admins).
     * Only Sourcegraph.com site admins may query this field.
     * FOR INTERNAL USE ONLY.
     */
    urlForSiteAdminBilling: string | null

    /**
     * The unique numeric ID for the user.
     * FOR INTERNAL USE ONLY.
     */
    databaseID: number

    /**
     * The name of this user namespace's component. For users, this is the username.
     */
    namespaceName: string

    /**
     * The permissions information of the user over repositories.
     * It is null when there is no permissions data stored for the user.
     */
    permissionsInfo: IPermissionsInfo | null

    /**
     * A list of monitors owned by the user or her organization.
     */
    monitors: IMonitorConnection

    /**
     * Repositories from external services owned by this user.
     */
    repositories: IRepositoryConnection

    /**
     * publicRepositories returns the repos listed in user_public_repos for this user
     */
    publicRepositories: IRepository[]

    /**
     * A list of campaigns applied under this user's namespace.
     * @deprecated "campaigns have been renamed to batch changes. Use batchChanges instead."
     */
    campaigns: ICampaignConnection

    /**
     * Returns a connection of configured external services accessible by this user, for usage with campaigns.
     * These are all code hosts configured on the Sourcegraph instance that are supported by campaigns. They are
     * connected to CampaignCredential resources, if one has been created for the code host connection before.
     * @deprecated "campaigns have been renamed to batch changes. Use batchChanges instead."
     */
    campaignsCodeHosts: ICampaignsCodeHostConnection

    /**
     * A list of batch changes applied under this user's namespace.
     */
    batchChanges: IBatchChangeConnection

    /**
     * Returns a connection of configured external services accessible by this user, for usage with batch changes.
     * These are all code hosts configured on the Sourcegraph instance that are supported by batch changes. They are
     * connected to BatchChangesCredential resources, if one has been created for the code host connection before.
     */
    batchChangesCodeHosts: IBatchChangesCodeHostConnection
}

export interface IEventLogsOnUserArguments {
    /**
     * Returns the first n event logs from the list.
     */
    first?: number | null

    /**
     * Only return events matching this event name
     */
    eventName?: string | null
}

export interface IAccessTokensOnUserArguments {
    /**
     * Returns the first n access tokens from the list.
     */
    first?: number | null
}

export interface IExternalAccountsOnUserArguments {
    /**
     * Returns the first n external accounts from the list.
     */
    first?: number | null
}

export interface IMonitorsOnUserArguments {
    /**
     * Returns the first n monitors from the list.
     * @default 50
     */
    first?: number | null

    /**
     * Opaque pagination cursor.
     */
    after?: string | null
}

export interface IRepositoriesOnUserArguments {
    /**
     * Returns the first n repositories from the list.
     */
    first?: number | null

    /**
     * Return repositories whose names match the query.
     */
    query?: string | null

    /**
     * An opaque cursor that is used for pagination.
     */
    after?: string | null

    /**
     * Include cloned repositories.
     * @default true
     */
    cloned?: boolean | null

    /**
     * Include repositories that are not yet cloned and for which cloning is not in progress.
     * @default true
     */
    notCloned?: boolean | null

    /**
     * Include repositories that have a text search index.
     * @default true
     */
    indexed?: boolean | null

    /**
     * Include repositories that do not have a text search index.
     * @default true
     */
    notIndexed?: boolean | null

    /**
     * Only include repositories from this external service.
     */
    externalServiceID?: ID | null
}

export interface ICampaignsOnUserArguments {
    /**
     * Returns the first n campaigns from the list.
     * @default 50
     */
    first?: number | null

    /**
     * Opaque pagination cursor.
     */
    after?: string | null

    /**
     * Only return campaigns in this state.
     */
    state?: CampaignState | null

    /**
     * Only include campaigns that the viewer can administer.
     */
    viewerCanAdminister?: boolean | null
}

export interface ICampaignsCodeHostsOnUserArguments {
    /**
     * Returns the first n code hosts from the list.
     * @default 50
     */
    first?: number | null

    /**
     * Opaque pagination cursor.
     */
    after?: string | null
}

export interface IBatchChangesOnUserArguments {
    /**
     * Returns the first n batch changes from the list.
     * @default 50
     */
    first?: number | null

    /**
     * Opaque pagination cursor.
     */
    after?: string | null

    /**
     * Only return batch changes in this state.
     */
    state?: BatchChangeState | null

    /**
     * Only include batch changes that the viewer can administer.
     */
    viewerCanAdminister?: boolean | null
}

export interface IBatchChangesCodeHostsOnUserArguments {
    /**
     * Returns the first n code hosts from the list.
     * @default 50
     */
    first?: number | null

    /**
     * Opaque pagination cursor.
     */
    after?: string | null
}

/**
 * An access token that grants to the holder the privileges of the user who created it.
 */
export interface IAccessToken {
    __typename: 'AccessToken'

    /**
     * The unique ID for the access token.
     */
    id: ID

    /**
     * The user whose privileges the access token grants.
     */
    subject: IUser

    /**
     * The scopes that define the allowed set of operations that can be performed using this access token.
     */
    scopes: string[]

    /**
     * A user-supplied descriptive note for the access token.
     */
    note: string

    /**
     * The user who created the access token. This is either the subject user (if the access token
     * was created by the same user) or a site admin (who can create access tokens for any user).
     */
    creator: IUser

    /**
     * The date when the access token was created.
     */
    createdAt: DateTime

    /**
     * The date when the access token was last used to authenticate a request.
     */
    lastUsedAt: DateTime | null
}

/**
 * A list of access tokens.
 */
export interface IAccessTokenConnection {
    __typename: 'AccessTokenConnection'

    /**
     * A list of access tokens.
     */
    nodes: IAccessToken[]

    /**
     * The total count of access tokens in the connection. This total count may be larger than the number of nodes
     * in this object when the result is paginated.
     */
    totalCount: number

    /**
     * Pagination information.
     */
    pageInfo: IPageInfo
}

/**
 * A list of authentication providers.
 */
export interface IAuthProviderConnection {
    __typename: 'AuthProviderConnection'

    /**
     * A list of authentication providers.
     */
    nodes: IAuthProvider[]

    /**
     * The total count of authentication providers in the connection. This total count may be larger than the number of nodes
     * in this object when the result is paginated.
     */
    totalCount: number

    /**
     * Pagination information.
     */
    pageInfo: IPageInfo
}

/**
 * A provider of user authentication, such as an external single-sign-on service (e.g., using OpenID Connect or
 * SAML). The provider information in this type is visible to all viewers and does not contain any secret values.
 */
export interface IAuthProvider {
    __typename: 'AuthProvider'

    /**
     * The type of the auth provider.
     */
    serviceType: string

    /**
     * An identifier for the service that the auth provider represents.
     */
    serviceID: string

    /**
     * An identifier for the client of the service that the auth provider represents.
     */
    clientID: string

    /**
     * The human-readable name of the provider.
     */
    displayName: string

    /**
     * Whether this auth provider is the builtin username-password auth provider.
     */
    isBuiltin: boolean

    /**
     * A URL that, when visited, initiates the authentication process for this auth provider.
     */
    authenticationURL: string | null
}

/**
 * A list of external accounts.
 */
export interface IExternalAccountConnection {
    __typename: 'ExternalAccountConnection'

    /**
     * A list of external accounts.
     */
    nodes: IExternalAccount[]

    /**
     * The total count of external accounts in the connection. This total count may be larger than the number of nodes
     * in this object when the result is paginated.
     */
    totalCount: number

    /**
     * Pagination information.
     */
    pageInfo: IPageInfo
}

/**
 * An external account associated with a user.
 */
export interface IExternalAccount {
    __typename: 'ExternalAccount'

    /**
     * The unique ID for the external account.
     */
    id: ID

    /**
     * The user on Sourcegraph.
     */
    user: IUser

    /**
     * The type of the external service where the external account resides.
     */
    serviceType: string

    /**
     * An identifier for the external service where the external account resides.
     */
    serviceID: string

    /**
     * An identifier for the client of the external service where the external account resides. This distinguishes
     * among multiple authentication providers that access the same service with different parameters.
     */
    clientID: string

    /**
     * An identifier for the external account (typically equal to or derived from the ID on the external service).
     */
    accountID: string

    /**
     * The creation date of this external account on Sourcegraph.
     */
    createdAt: DateTime

    /**
     * The last-updated date of this external account on Sourcegraph.
     */
    updatedAt: DateTime

    /**
     * A URL that, when visited, re-initiates the authentication process.
     */
    refreshURL: string | null

    /**
     * Provider-specific data about the external account.
     * Only site admins may query this field.
     */
    accountData: any | null
}

/**
 * An active user session.
 */
export interface ISession {
    __typename: 'Session'

    /**
     * Whether the user can sign out of this session on Sourcegraph.
     */
    canSignOut: boolean
}

/**
 * An organization membership.
 */
export interface IOrganizationMembership {
    __typename: 'OrganizationMembership'

    /**
     * The organization.
     */
    organization: IOrg

    /**
     * The user.
     */
    user: IUser

    /**
     * The time when this was created.
     */
    createdAt: DateTime

    /**
     * The time when this was updated.
     */
    updatedAt: DateTime
}

/**
 * A list of organization memberships.
 */
export interface IOrganizationMembershipConnection {
    __typename: 'OrganizationMembershipConnection'

    /**
     * A list of organization memberships.
     */
    nodes: IOrganizationMembership[]

    /**
     * The total count of organization memberships in the connection. This total count may be larger than the number
     * of nodes in this object when the result is paginated.
     */
    totalCount: number
}

/**
 * A user's email address.
 */
export interface IUserEmail {
    __typename: 'UserEmail'

    /**
     * The email address.
     */
    email: string

    /**
     * Whether the email address is the user's primary email address. Currently this is defined as the earliest
     * email address associated with the user, preferring verified emails to unverified emails.
     */
    isPrimary: boolean

    /**
     * Whether the email address has been verified by the user.
     */
    verified: boolean

    /**
     * Whether the email address is pending verification.
     */
    verificationPending: boolean

    /**
     * The user associated with this email address.
     */
    user: IUser

    /**
     * Whether the viewer has privileges to manually mark this email address as verified (without the user going
     * through the normal verification process). Only site admins have this privilege.
     */
    viewerCanManuallyVerify: boolean
}

/**
 * A list of organizations.
 */
export interface IOrgConnection {
    __typename: 'OrgConnection'

    /**
     * A list of organizations.
     */
    nodes: IOrg[]

    /**
     * The total count of organizations in the connection. This total count may be larger
     * than the number of nodes in this object when the result is paginated.
     */
    totalCount: number
}

/**
 * An organization, which is a group of users.
 */
export interface IOrg {
    __typename: 'Org'

    /**
     * The unique ID for the organization.
     */
    id: ID

    /**
     * The organization's name. This is unique among all organizations on this Sourcegraph site.
     */
    name: string

    /**
     * The organization's chosen display name.
     */
    displayName: string | null

    /**
     * The date when the organization was created.
     */
    createdAt: DateTime

    /**
     * A list of users who are members of this organization.
     */
    members: IUserConnection

    /**
     * The latest settings for the organization.
     * Only organization members and site admins can access this field.
     */
    latestSettings: ISettings | null

    /**
     * All settings for this organization, and the individual levels in the settings cascade (global > organization)
     * that were merged to produce the final merged settings.
     * Only organization members and site admins can access this field.
     */
    settingsCascade: ISettingsCascade

    /**
     * DEPRECATED
     * @deprecated "Use settingsCascade instead. This field is a deprecated alias for it and will be removed in a future release."
     */
    configurationCascade: IConfigurationCascade

    /**
     * A pending invitation for the viewer to join this organization, if any.
     */
    viewerPendingInvitation: IOrganizationInvitation | null

    /**
     * Whether the viewer has admin privileges on this organization. Currently, all of an organization's members
     * have admin privileges on the organization.
     */
    viewerCanAdminister: boolean

    /**
     * Whether the viewer is a member of this organization.
     */
    viewerIsMember: boolean

    /**
     * The URL to the organization.
     */
    url: string

    /**
     * The URL to the organization's settings.
     */
    settingsURL: string | null

    /**
     * The name of this user namespace's component. For organizations, this is the organization's name.
     */
    namespaceName: string

    /**
     * A list of campaigns initially applied in this organization.
     * @deprecated "campaigns have been renamed to batch changes. Use batchChanges instead."
     */
    campaigns: ICampaignConnection

    /**
     * A list of batch changes initially applied in this organization.
     */
    batchChanges: IBatchChangeConnection
}

export interface ICampaignsOnOrgArguments {
    /**
     * Returns the first n campaigns from the list.
     * @default 50
     */
    first?: number | null

    /**
     * Opaque pagination cursor.
     */
    after?: string | null

    /**
     * Only return campaigns in this state.
     */
    state?: CampaignState | null

    /**
     * Only include campaigns that the viewer can administer.
     */
    viewerCanAdminister?: boolean | null
}

export interface IBatchChangesOnOrgArguments {
    /**
     * Returns the first n batch changes from the list.
     * @default 50
     */
    first?: number | null

    /**
     * Opaque pagination cursor.
     */
    after?: string | null

    /**
     * Only return batch changes in this state.
     */
    state?: BatchChangeState | null

    /**
     * Only include batch changes that the viewer can administer.
     */
    viewerCanAdminister?: boolean | null
}

/**
 * The result of Mutation.inviteUserToOrganization.
 */
export interface IInviteUserToOrganizationResult {
    __typename: 'InviteUserToOrganizationResult'

    /**
     * Whether an invitation email was sent. If emails are not enabled on this site or if the user has no verified
     * email address, an email will not be sent.
     */
    sentInvitationEmail: boolean

    /**
     * The URL that the invited user can visit to accept or reject the invitation.
     */
    invitationURL: string
}

/**
 * An invitation to join an organization as a member.
 */
export interface IOrganizationInvitation {
    __typename: 'OrganizationInvitation'

    /**
     * The ID of the invitation.
     */
    id: ID

    /**
     * The organization that the invitation is for.
     */
    organization: IOrg

    /**
     * The user who sent the invitation.
     */
    sender: IUser

    /**
     * The user who received the invitation.
     */
    recipient: IUser

    /**
     * The date when this invitation was created.
     */
    createdAt: DateTime

    /**
     * The most recent date when a notification was sent to the recipient about this invitation.
     */
    notifiedAt: DateTime | null

    /**
     * The date when this invitation was responded to by the recipient.
     */
    respondedAt: DateTime | null

    /**
     * The recipient's response to this invitation, or no response (null).
     */
    responseType: OrganizationInvitationResponseType | null

    /**
     * The URL where the recipient can respond to the invitation when pending, or null if not pending.
     */
    respondURL: string | null

    /**
     * The date when this invitation was revoked.
     */
    revokedAt: DateTime | null
}

/**
 * The recipient's possible responses to an invitation to join an organization as a member.
 */
export enum OrganizationInvitationResponseType {
    /**
     * The invitation was accepted by the recipient.
     */
    ACCEPT = 'ACCEPT',

    /**
     * The invitation was rejected by the recipient.
     */
    REJECT = 'REJECT',
}

/**
 * RepositoryOrderBy enumerates the ways a repositories list can be ordered.
 */
export enum RepositoryOrderBy {
    REPOSITORY_NAME = 'REPOSITORY_NAME',
    REPO_CREATED_AT = 'REPO_CREATED_AT',

    /**
     * deprecated (use the equivalent REPOSITORY_CREATED_AT)
     */
    REPOSITORY_CREATED_AT = 'REPOSITORY_CREATED_AT',
}

/**
 * The default settings for the Sourcegraph instance. This is hardcoded in
 * Sourcegraph, but may change from release to release.
 */
export interface IDefaultSettings {
    __typename: 'DefaultSettings'

    /**
     * The opaque GraphQL ID.
     */
    id: ID

    /**
     * The latest default settings (this never changes).
     */
    latestSettings: ISettings | null

    /**
     * The URL to the default settings. This URL does not exist because you
     * cannot edit or directly view default settings.
     */
    settingsURL: string | null

    /**
     * Whether the viewer can modify the subject's settings. Always false for
     * default settings.
     */
    viewerCanAdminister: boolean

    /**
     * The default settings, and the final merged settings.
     * All viewers can access this field.
     */
    settingsCascade: ISettingsCascade

    /**
     * DEPRECATED
     * @deprecated "Use settingsCascade instead. This field is a deprecated alias for it and will be removed in a future release."
     */
    configurationCascade: IConfigurationCascade
}

/**
 * A site is an installation of Sourcegraph that consists of one or more
 * servers that share the same configuration and database.
 * The site is a singleton; the API only ever returns the single global site.
 */
export interface ISite {
    __typename: 'Site'

    /**
     * The site's opaque GraphQL ID. This is NOT the "site ID" as it is referred to elsewhere;
     * use the siteID field for that. (GraphQL node types conventionally have an id field of type
     * ID! that globally identifies the node.)
     */
    id: ID

    /**
     * The site ID.
     */
    siteID: string

    /**
     * The site's configuration. Only visible to site admins.
     */
    configuration: ISiteConfiguration

    /**
     * The site's latest site-wide settings (which are the second-lowest-precedence
     * in the configuration cascade for a user).
     */
    latestSettings: ISettings | null

    /**
     * The global settings for this site, and the final merged settings.
     * All viewers can access this field.
     */
    settingsCascade: ISettingsCascade

    /**
     * DEPRECATED
     * @deprecated "Use settingsCascade instead. This field is a deprecated alias for it and will be removed in a future release."
     */
    configurationCascade: IConfigurationCascade

    /**
     * The URL to the site's settings.
     */
    settingsURL: string | null

    /**
     * Whether the viewer can reload the site (with the reloadSite mutation).
     */
    canReloadSite: boolean

    /**
     * Whether the viewer can modify the subject's settings.
     */
    viewerCanAdminister: boolean

    /**
     * A list of all access tokens on this site.
     */
    accessTokens: IAccessTokenConnection

    /**
     * A list of all authentication providers. This information is visible to all viewers and does not contain any
     * secret information.
     */
    authProviders: IAuthProviderConnection

    /**
     * A list of all user external accounts on this site.
     */
    externalAccounts: IExternalAccountConnection

    /**
     * The build version of the Sourcegraph software that is running on this site (of the form
     * NNNNN_YYYY-MM-DD_XXXXX, like 12345_2018-01-01_abcdef).
     */
    buildVersion: string

    /**
     * The product version of the Sourcegraph software that is running on this site.
     */
    productVersion: string

    /**
     * Information about software updates for the version of Sourcegraph that this site is running.
     */
    updateCheck: IUpdateCheck

    /**
     * Whether the site needs to be configured to add repositories.
     */
    needsRepositoryConfiguration: boolean

    /**
     * Whether the site is over the limit for free user accounts, and a warning needs to be shown to all users.
     * Only applies if the site does not have a valid license.
     */
    freeUsersExceeded: boolean

    /**
     * Alerts to display to the viewer.
     */
    alerts: IAlert[]

    /**
     * BACKCOMPAT: Always returns true.
     */
    hasCodeIntelligence: boolean

    /**
     * Whether we want to show built-in searches on the saved searches page
     */
    disableBuiltInSearches: boolean

    /**
     * Whether the server sends emails to users to verify email addresses. If false, then site admins must manually
     * verify users' email addresses.
     */
    sendsEmailVerificationEmails: boolean

    /**
     * Information about this site's product subscription status.
     */
    productSubscription: IProductSubscriptionStatus

    /**
     * Usage statistics for this site.
     */
    usageStatistics: ISiteUsageStatistics

    /**
     * Monitoring overview for this site.
     * Note: This is primarily used for displaying recently-fired alerts in the web app. If your intent
     * is to monitor Sourcegraph, it is better to configure alerting or query Prometheus directly in
     * order to ensure that if the frontend goes down you still recieve alerts:
     * Configure alerting: https://docs.sourcegraph.com/admin/observability/alerting
     * Query Prometheus directly: https://docs.sourcegraph.com/admin/observability/alerting_custom_consumption
     */
    monitoringStatistics: IMonitoringStatistics
}

export interface IAccessTokensOnSiteArguments {
    /**
     * Returns the first n access tokens from the list.
     */
    first?: number | null
}

export interface IExternalAccountsOnSiteArguments {
    /**
     * Returns the first n external accounts from the list.
     */
    first?: number | null

    /**
     * Include only external accounts associated with this user.
     */
    user?: ID | null

    /**
     * Include only external accounts with this service type.
     */
    serviceType?: string | null

    /**
     * Include only external accounts with this service ID.
     */
    serviceID?: string | null

    /**
     * Include only external accounts with this client ID.
     */
    clientID?: string | null
}

export interface IUsageStatisticsOnSiteArguments {
    /**
     * Days of history (based on current UTC time).
     */
    days?: number | null

    /**
     * Weeks of history (based on current UTC time).
     */
    weeks?: number | null

    /**
     * Months of history (based on current UTC time).
     */
    months?: number | null
}

export interface IMonitoringStatisticsOnSiteArguments {
    /**
     * Days of history (based on current UTC time).
     */
    days?: number | null
}

/**
 * The configuration for a site.
 */
export interface ISiteConfiguration {
    __typename: 'SiteConfiguration'

    /**
     * The unique identifier of this site configuration version.
     */
    id: number

    /**
     * The effective configuration JSON.
     */
    effectiveContents: JSONCString

    /**
     * Messages describing validation problems or usage of deprecated configuration in the configuration JSON.
     * This includes both JSON Schema validation problems and other messages that perform more advanced checks
     * on the configuration (that can't be expressed in the JSON Schema).
     */
    validationMessages: string[]
}

/**
 * Information about software updates for Sourcegraph.
 */
export interface IUpdateCheck {
    __typename: 'UpdateCheck'

    /**
     * Whether an update check is currently in progress.
     */
    pending: boolean

    /**
     * When the last update check was completed, or null if no update check has
     * been completed (or performed) yet.
     */
    checkedAt: DateTime | null

    /**
     * If an error occurred during the last update check, this message describes
     * the error.
     */
    errorMessage: string | null

    /**
     * If an update is available, the version string of the updated version.
     */
    updateVersionAvailable: string | null
}

/**
 * The possible types of alerts (Alert.type values).
 */
export enum AlertType {
    INFO = 'INFO',
    WARNING = 'WARNING',
    ERROR = 'ERROR',
}

/**
 * An alert message shown to the viewer.
 */
export interface IAlert {
    __typename: 'Alert'

    /**
     * The type of this alert.
     */
    type: AlertType

    /**
     * The message body of this alert. Markdown is supported.
     */
    message: string

    /**
     * If set, this alert is dismissible. After being dismissed, no other alerts with the same
     * isDismissibleWithKey value will be shown. If null, this alert is not dismissible.
     */
    isDismissibleWithKey: string | null
}

/**
 * SettingsSubject is something that can have settings: a site ("global settings", which is different from "site
 * configuration"), an organization, or a user.
 */
export type SettingsSubject = IUser | IOrg | IDefaultSettings | ISite

/**
 * SettingsSubject is something that can have settings: a site ("global settings", which is different from "site
 * configuration"), an organization, or a user.
 */
export interface ISettingsSubject {
    __typename: 'SettingsSubject'

    /**
     * The ID.
     */
    id: ID

    /**
     * The latest settings.
     */
    latestSettings: ISettings | null

    /**
     * The URL to the settings.
     */
    settingsURL: string | null

    /**
     * Whether the viewer can modify the subject's settings.
     */
    viewerCanAdminister: boolean

    /**
     * All settings for this subject, and the individual levels in the settings cascade (global > organization > user)
     * that were merged to produce the final merged settings.
     */
    settingsCascade: ISettingsCascade

    /**
     * DEPRECATED
     * @deprecated "Use settingsCascade instead. This field is a deprecated alias for it and will be removed in a future release."
     */
    configurationCascade: IConfigurationCascade
}

/**
 * The configurations for all of the relevant settings subjects, plus the merged settings.
 */
export interface ISettingsCascade {
    __typename: 'SettingsCascade'

    /**
     * The other settings subjects that are applied with lower precedence than this subject to
     * form the final merged settings. For example, a user in 2 organizations would have the following
     * settings subjects: site (global settings), org 1, org 2, and the user.
     */
    subjects: SettingsSubject[]

    /**
     * The effective final merged settings as (stringified) JSON, merged from all of the subjects.
     */
    final: string

    /**
     * DEPRECATED: This field will be removed in a future release.
     * The effective final merged settings, merged from all of the subjects.
     * @deprecated "use final instead"
     */
    merged: IConfiguration
}

/**
 * DEPRECATED: Renamed to SettingsCascade.
 */
export interface IConfigurationCascade {
    __typename: 'ConfigurationCascade'

    /**
     * DEPRECATED
     * @deprecated "use SettingsCascade.subjects instead"
     */
    subjects: SettingsSubject[]

    /**
     * DEPRECATED
     * @deprecated "use SettingsCascade.final instead"
     */
    merged: IConfiguration
}

/**
 * Settings is a version of a configuration settings file.
 */
export interface ISettings {
    __typename: 'Settings'

    /**
     * The ID.
     */
    id: number

    /**
     * The subject that these settings are for.
     */
    subject: SettingsSubject

    /**
     * The author, or null if there is no author or the authoring user was deleted.
     */
    author: IUser | null

    /**
     * The time when this was created.
     */
    createdAt: DateTime

    /**
     * The stringified JSON contents of the settings. The contents may include "//"-style comments and trailing
     * commas in the JSON.
     */
    contents: JSONCString

    /**
     * DEPRECATED: This field will be removed in a future release.
     * The configuration.
     * @deprecated "use the contents field instead"
     */
    configuration: IConfiguration
}

/**
 * DEPRECATED: Use the contents field on the parent type instead. This type will be removed in a future release.
 */
export interface IConfiguration {
    __typename: 'Configuration'

    /**
     * DEPRECATED: This field will be removed in a future release.
     * The raw JSON contents, encoded as a string.
     * @deprecated "use the contents field on the parent type instead"
     */
    contents: JSONCString

    /**
     * DEPRECATED: This field is always empty. It will be removed in a future release.
     * @deprecated "use client-side JSON Schema validation instead"
     */
    messages: string[]
}

/**
 * UserUsageStatistics describes a user's usage statistics.
 * This information is visible to all viewers.
 */
export interface IUserUsageStatistics {
    __typename: 'UserUsageStatistics'

    /**
     * The number of search queries that the user has performed.
     */
    searchQueries: number

    /**
     * The number of page views that the user has performed.
     */
    pageViews: number

    /**
     * The number of code intelligence actions that the user has performed.
     */
    codeIntelligenceActions: number

    /**
     * The number of find-refs actions that the user has performed.
     */
    findReferencesActions: number

    /**
     * The last time the user was active (any action, any platform).
     */
    lastActiveTime: string | null

    /**
     * The last time the user was active on a code host integration.
     */
    lastActiveCodeHostIntegrationTime: string | null
}

/**
 * A user event.
 */
export enum UserEvent {
    PAGEVIEW = 'PAGEVIEW',
    SEARCHQUERY = 'SEARCHQUERY',
    CODEINTEL = 'CODEINTEL',
    CODEINTELREFS = 'CODEINTELREFS',
    CODEINTELINTEGRATION = 'CODEINTELINTEGRATION',
    CODEINTELINTEGRATIONREFS = 'CODEINTELINTEGRATIONREFS',

    /**
     * Product stages
     */
    STAGEMANAGE = 'STAGEMANAGE',
    STAGEPLAN = 'STAGEPLAN',
    STAGECODE = 'STAGECODE',
    STAGEREVIEW = 'STAGEREVIEW',
    STAGEVERIFY = 'STAGEVERIFY',
    STAGEPACKAGE = 'STAGEPACKAGE',
    STAGEDEPLOY = 'STAGEDEPLOY',
    STAGECONFIGURE = 'STAGECONFIGURE',
    STAGEMONITOR = 'STAGEMONITOR',
    STAGESECURE = 'STAGESECURE',
    STAGEAUTOMATE = 'STAGEAUTOMATE',
}

/**
 * A period of time in which a set of users have been active.
 */
export enum UserActivePeriod {
    /**
     * Since today at 00:00 UTC.
     */
    TODAY = 'TODAY',

    /**
     * Since the latest Monday at 00:00 UTC.
     */
    THIS_WEEK = 'THIS_WEEK',

    /**
     * Since the first day of the current month at 00:00 UTC.
     */
    THIS_MONTH = 'THIS_MONTH',

    /**
     * All time.
     */
    ALL_TIME = 'ALL_TIME',
}

/**
 * SiteUsageStatistics describes a site's aggregate usage statistics.
 * This information is visible to all viewers.
 */
export interface ISiteUsageStatistics {
    __typename: 'SiteUsageStatistics'

    /**
     * Recent daily active users.
     */
    daus: ISiteUsagePeriod[]

    /**
     * Recent weekly active users.
     */
    waus: ISiteUsagePeriod[]

    /**
     * Recent monthly active users.
     */
    maus: ISiteUsagePeriod[]
}

/**
 * SiteUsagePeriod describes a site's usage statistics for a given timespan.
 * This information is visible to all viewers.
 */
export interface ISiteUsagePeriod {
    __typename: 'SiteUsagePeriod'

    /**
     * The time when this started.
     */
    startTime: string

    /**
     * The user count.
     */
    userCount: number

    /**
     * The registered user count.
     */
    registeredUserCount: number

    /**
     * The anonymous user count.
     */
    anonymousUserCount: number

    /**
     * The count of registered users that have been active on a code host integration.
     * Excludes anonymous users.
     */
    integrationUserCount: number
}

/**
 * Monitoring overview.
 */
export interface IMonitoringStatistics {
    __typename: 'MonitoringStatistics'

    /**
     * Alerts fired in this time span.
     */
    alerts: IMonitoringAlert[]
}

/**
 * A high-level monitoring alert, for details see https://docs.sourcegraph.com/admin/observability/metrics#high-level-alerting-metrics
 */
export interface IMonitoringAlert {
    __typename: 'MonitoringAlert'

    /**
     * End time of this event, which describes the past 12h of recorded data.
     */
    timestamp: DateTime

    /**
     * Name of alert that the service fired.
     */
    name: string

    /**
     * Name of the service that fired the alert.
     */
    serviceName: string

    /**
     * Owner of the fired alert.
     */
    owner: string

    /**
     * Average percentage of time (between [0, 1]) that the event was firing over the 12h of recorded data. e.g.
     * 1.0 if it was firing 100% of the time on average during that 12h window, 0.5 if it was firing 50% of the
     * time on average, etc.
     */
    average: number
}

/**
 * A list of survey responses
 */
export interface ISurveyResponseConnection {
    __typename: 'SurveyResponseConnection'

    /**
     * A list of survey responses.
     */
    nodes: ISurveyResponse[]

    /**
     * The total count of survey responses in the connection. This total count may be larger
     * than the number of nodes in this object when the result is paginated.
     */
    totalCount: number

    /**
     * The count of survey responses submitted since 30 calendar days ago at 00:00 UTC.
     */
    last30DaysCount: number

    /**
     * The average score of survey responses in the connection submitted since 30 calendar days ago at 00:00 UTC.
     */
    averageScore: number

    /**
     * The net promoter score (NPS) of survey responses in the connection submitted since 30 calendar days ago at 00:00 UTC.
     * Return value is a signed integer, scaled from -100 (all detractors) to +100 (all promoters).
     * See https://en.wikipedia.org/wiki/Net_Promoter for explanation.
     */
    netPromoterScore: number
}

/**
 * An individual response to a user satisfaction (NPS) survey.
 */
export interface ISurveyResponse {
    __typename: 'SurveyResponse'

    /**
     * The unique ID of the survey response
     */
    id: ID

    /**
     * The user who submitted the survey (if they were authenticated at the time).
     */
    user: IUser | null

    /**
     * The email that the user manually entered (if they were NOT authenticated at the time).
     */
    email: string | null

    /**
     * User's likelihood of recommending Sourcegraph to a friend, from 0-10.
     */
    score: number

    /**
     * The answer to "What is the most important reason for the score you gave".
     */
    reason: string | null

    /**
     * The answer to "What can Sourcegraph do to provide a better product"
     */
    better: string | null

    /**
     * The time when this response was created.
     */
    createdAt: DateTime
}

/**
 * Information about this site's product subscription (which enables access to and renewals of a product license).
 */
export interface IProductSubscriptionStatus {
    __typename: 'ProductSubscriptionStatus'

    /**
     * The full name of the product in use, such as "Sourcegraph Enterprise".
     */
    productNameWithBrand: string

    /**
     * The max number of user accounts that have been active on this Sourcegraph site for the current license.
     * If no license is in use, returns zero.
     */
    actualUserCount: number

    /**
     * The date and time when the max number of user accounts that have been active on this Sourcegraph site for
     * the current license was reached. If no license is in use, returns an empty string.
     */
    actualUserCountDate: string

    /**
     * The number of users allowed. If there is a license, this is equal to ProductLicenseInfo.userCount. Otherwise,
     * it is the user limit for instances without a license, or null if there is no limit.
     */
    maximumAllowedUserCount: number | null

    /**
     * The number of free users allowed on a site without a license before a warning is shown to all users, or null
     * if a valid license is in use.
     */
    noLicenseWarningUserCount: number | null

    /**
     * The product license associated with this subscription, if any.
     */
    license: IProductLicenseInfo | null
}

/**
 * Information about this site's product license (which activates certain Sourcegraph features).
 */
export interface IProductLicenseInfo {
    __typename: 'ProductLicenseInfo'

    /**
     * The full name of the product that this license is for. To get the product name for the current
     * Sourcegraph site, use ProductSubscriptionStatus.productNameWithBrand instead (to handle cases where there is
     * no license).
     */
    productNameWithBrand: string

    /**
     * Tags indicating the product plan and features activated by this license.
     */
    tags: string[]

    /**
     * The number of users allowed by this license.
     */
    userCount: number

    /**
     * The date when this license expires.
     */
    expiresAt: DateTime
}

/**
 * An extension registry.
 */
export interface IExtensionRegistry {
    __typename: 'ExtensionRegistry'

    /**
     * Find an extension by its extension ID (which is the concatenation of the publisher name, a slash ("/"), and the
     * extension name).
     * To find an extension by its GraphQL ID, use Query.node.
     */
    extension: IRegistryExtension | null

    /**
     * A list of extensions published in the extension registry.
     */
    extensions: IRegistryExtensionConnection

    /**
     * A list of publishers with at least 1 extension in the registry.
     */
    publishers: IRegistryPublisherConnection

    /**
     * A list of publishers that the viewer may publish extensions as.
     */
    viewerPublishers: RegistryPublisher[]

    /**
     * The extension ID prefix for extensions that are published in the local extension registry. This is the
     * hostname (and port, if non-default HTTP/HTTPS) of the Sourcegraph "externalURL" site configuration property.
     * It is null if extensions published on this Sourcegraph site do not have an extension ID prefix.
     * Examples: "sourcegraph.example.com/", "sourcegraph.example.com:1234/"
     */
    localExtensionIDPrefix: string | null
}

export interface IExtensionOnExtensionRegistryArguments {
    extensionID: string
}

export interface IExtensionsOnExtensionRegistryArguments {
    /**
     * Returns the first n extensions from the list.
     */
    first?: number | null

    /**
     * Returns only extensions from this publisher.
     */
    publisher?: ID | null

    /**
     * Returns only extensions matching the query.
     * The following keywords are supported:
     * - category:"C" - include only extensions in the given category.
     * - tag:"T" - include only extensions in the given tag.
     * The following keywords are ignored by the server (so that the frontend can post-process the result set to
     * implement the keywords):
     * - installed - include only installed extensions.
     * - enabled - include only enabled extensions.
     * - disabled - include only disabled extensions.
     */
    query?: string | null

    /**
     * Include extensions from the local registry.
     * @default true
     */
    local?: boolean | null

    /**
     * Include extensions from remote registries.
     * @default true
     */
    remote?: boolean | null

    /**
     * Sorts the list of extension results such that the extensions with these IDs are first in the result set.
     * Typically, the client passes the list of added and enabled extension IDs in this parameter so that the
     * results include those extensions first (which is typically what the user prefers).
     */
    prioritizeExtensionIDs?: string[] | null
}

export interface IPublishersOnExtensionRegistryArguments {
    /**
     * Return the first n publishers from the list.
     */
    first?: number | null
}

/**
 * A publisher of a registry extension.
 */
export type RegistryPublisher = IUser | IOrg

/**
 * A list of publishers of extensions in the registry.
 */
export interface IRegistryPublisherConnection {
    __typename: 'RegistryPublisherConnection'

    /**
     * A list of publishers.
     */
    nodes: RegistryPublisher[]

    /**
     * The total count of publishers in the connection. This total count may be larger than the number of
     * nodes in this object when the result is paginated.
     */
    totalCount: number

    /**
     * Pagination information.
     */
    pageInfo: IPageInfo
}

/**
 * Mutations for the extension registry.
 */
export interface IExtensionRegistryMutation {
    __typename: 'ExtensionRegistryMutation'

    /**
     * Create a new extension in the extension registry.
     */
    createExtension: IExtensionRegistryCreateExtensionResult

    /**
     * Update an extension in the extension registry.
     * Only authorized extension publishers may perform this mutation.
     */
    updateExtension: IExtensionRegistryUpdateExtensionResult

    /**
     * Delete an extension from the extension registry.
     * Only authorized extension publishers may perform this mutation.
     */
    deleteExtension: IEmptyResponse

    /**
     * Publish an extension in the extension registry, creating it (if it doesn't yet exist) or updating it (if it
     * does).
     * This is a helper that wraps multiple other GraphQL mutations to expose a single API for publishing an
     * extension.
     */
    publishExtension: IExtensionRegistryPublishExtensionResult
}

export interface ICreateExtensionOnExtensionRegistryMutationArguments {
    /**
     * The ID of the extension's publisher (a user or organization).
     */
    publisher: ID

    /**
     * The name of the extension.
     */
    name: string
}

export interface IUpdateExtensionOnExtensionRegistryMutationArguments {
    /**
     * The extension to update.
     */
    extension: ID

    /**
     * The new name for the extension, or null to leave unchanged.
     */
    name?: string | null
}

export interface IDeleteExtensionOnExtensionRegistryMutationArguments {
    /**
     * The ID of the extension to delete.
     */
    extension: ID
}

export interface IPublishExtensionOnExtensionRegistryMutationArguments {
    /**
     * The extension ID of the extension to publish. If a host prefix (e.g., "sourcegraph.example.com/") is
     * needed and it is not included, it is automatically prepended.
     * Examples: "alice/myextension", "acmecorp/myextension"
     */
    extensionID: string

    /**
     * The extension manifest (as JSON).
     */
    manifest: string

    /**
     * The bundled JavaScript source of the extension.
     */
    bundle?: string | null

    /**
     * The source map of the extension's JavaScript bundle, if any.
     * The JavaScript bundle's "//# sourceMappingURL=" directive, if any, is ignored. When the bundle is served,
     * the source map provided here is referenced instead.
     */
    sourceMap?: string | null

    /**
     * Force publish even if there are warnings (such as invalid JSON warnings).
     * @default false
     */
    force?: boolean | null
}

/**
 * The result of Mutation.extensionRegistry.createExtension.
 */
export interface IExtensionRegistryCreateExtensionResult {
    __typename: 'ExtensionRegistryCreateExtensionResult'

    /**
     * The newly created extension.
     */
    extension: IRegistryExtension
}

/**
 * The result of Mutation.extensionRegistry.updateExtension.
 */
export interface IExtensionRegistryUpdateExtensionResult {
    __typename: 'ExtensionRegistryUpdateExtensionResult'

    /**
     * The newly updated extension.
     */
    extension: IRegistryExtension
}

/**
 * The result of Mutation.extensionRegistry.publishExtension.
 */
export interface IExtensionRegistryPublishExtensionResult {
    __typename: 'ExtensionRegistryPublishExtensionResult'

    /**
     * The extension that was just published.
     */
    extension: IRegistryExtension
}

/**
 * An extension's listing in the extension registry.
 */
export interface IRegistryExtension {
    __typename: 'RegistryExtension'

    /**
     * The unique, opaque, permanent ID of the extension. Do not display this ID to the user; display
     * RegistryExtension.extensionID instead (it is friendlier and still unique, but it can be renamed).
     */
    id: ID

    /**
     * The UUID of the extension. This identifies the extension externally (along with the origin). The UUID maps
     * 1-to-1 to RegistryExtension.id.
     */
    uuid: string

    /**
     * The publisher of the extension. If this extension is from a remote registry, the publisher may be null.
     */
    publisher: RegistryPublisher | null

    /**
     * The qualified, unique name that refers to this extension, consisting of the registry name (if non-default),
     * publisher's name, and the extension's name, all joined by "/" (for example, "acme-corp/my-extension-name").
     */
    extensionID: string

    /**
     * The extension ID without the registry name.
     */
    extensionIDWithoutRegistry: string

    /**
     * The name of the extension (not including the publisher's name).
     */
    name: string

    /**
     * The extension manifest, or null if none is set.
     */
    manifest: IExtensionManifest | null

    /**
     * The date when this extension was created on the registry.
     */
    createdAt: DateTime | null

    /**
     * The date when this extension was last updated on the registry (including updates to its metadata only, not
     * publishing new releases).
     */
    updatedAt: DateTime | null

    /**
     * The date when a release of this extension was most recently published, or null if there are no releases.
     */
    publishedAt: DateTime | null

    /**
     * The URL to the extension on this Sourcegraph site.
     */
    url: string

    /**
     * The URL to the extension on the extension registry where it lives (if this is a remote
     * extension). If this extension is local, then this field's value is null.
     */
    remoteURL: string | null

    /**
     * The name of this extension's registry.
     */
    registryName: string

    /**
     * Whether the registry extension is published on this Sourcegraph site.
     */
    isLocal: boolean

    /**
     * Whether the extension is marked as a work-in-progress extension by the extension author.
     */
    isWorkInProgress: boolean

    /**
     * Whether the viewer has admin privileges on this registry extension.
     */
    viewerCanAdminister: boolean
}

/**
 * A description of the extension, how to run or access it, and when to activate it.
 */
export interface IExtensionManifest {
    __typename: 'ExtensionManifest'

    /**
     * The raw JSON contents of the manifest.
     */
    raw: string

    /**
     * The description specified in the manifest, if any.
     */
    description: string | null

    /**
     * The URL to the bundled JavaScript source code for the extension, if any.
     */
    bundleURL: string | null
}

/**
 * A list of registry extensions.
 */
export interface IRegistryExtensionConnection {
    __typename: 'RegistryExtensionConnection'

    /**
     * A list of registry extensions.
     */
    nodes: IRegistryExtension[]

    /**
     * The total count of registry extensions in the connection. This total count may be larger than the number of
     * nodes in this object when the result is paginated.
     */
    totalCount: number

    /**
     * Pagination information.
     */
    pageInfo: IPageInfo

    /**
     * The URL to this list, or null if none exists.
     */
    url: string | null

    /**
     * Errors that occurred while communicating with remote registries to obtain the list of extensions.
     * In order to be able to return local extensions even when the remote registry is unreachable, errors are
     * recorded here instead of in the top-level GraphQL errors list.
     */
    error: string | null
}

/**
 * Aggregate local code intelligence for all ranges that fall between a window of lines in a document.
 */
export interface ICodeIntelligenceRangeConnection {
    __typename: 'CodeIntelligenceRangeConnection'

    /**
     * Aggregate local code intelligence grouped by range.
     */
    nodes: ICodeIntelligenceRange[]
}

/**
 * Aggregate code intelligence for a particular range within a document.
 */
export interface ICodeIntelligenceRange {
    __typename: 'CodeIntelligenceRange'

    /**
     * The range this code intelligence applies to.
     */
    range: IRange

    /**
     * A list of definitions of the symbol occurring within the range.
     */
    definitions: ILocationConnection

    /**
     * A list of references of the symbol occurring within the range.
     */
    references: ILocationConnection

    /**
     * The hover result of the symbol occurring within the range.
     */
    hover: IHover | null
}

/**
 * A list of locations within a file.
 */
export interface ILocationConnection {
    __typename: 'LocationConnection'

    /**
     * A list of locations within a file.
     */
    nodes: ILocation[]

    /**
     * Pagination information.
     */
    pageInfo: IPageInfo
}

/**
 * Hover range and markdown content.
 */
export interface IHover {
    __typename: 'Hover'

    /**
     * A markdown string containing the contents of the hover.
     */
    markdown: IMarkdown

    /**
     * The range to highlight.
     */
    range: IRange
}

/**
 * The state an LSIF upload can be in.
 */
export enum LSIFUploadState {
    /**
     * This upload is being processed.
     */
    PROCESSING = 'PROCESSING',

    /**
     * This upload failed to be processed.
     */
    ERRORED = 'ERRORED',

    /**
     * This upload was processed successfully.
     */
    COMPLETED = 'COMPLETED',

    /**
     * This upload is queued to be processed later.
     */
    QUEUED = 'QUEUED',

    /**
     * This upload is currently being transferred to Sourcegraph.
     */
    UPLOADING = 'UPLOADING',
}

/**
 * Metadata and status about an LSIF upload.
 */
export interface ILSIFUpload {
    __typename: 'LSIFUpload'

    /**
     * The ID.
     */
    id: ID

    /**
     * The project for which this upload provides code intelligence.
     */
    projectRoot: IGitTree | null

    /**
     * The original 40-character commit commit supplied at upload time.
     */
    inputCommit: string

    /**
     * The original root supplied at upload time.
     */
    inputRoot: string

    /**
     * The original indexer name supplied at upload time.
     */
    inputIndexer: string

    /**
     * The upload's current state.
     */
    state: LSIFUploadState

    /**
     * The time the upload was uploaded.
     */
    uploadedAt: DateTime

    /**
     * The time the upload was processed.
     */
    startedAt: DateTime | null

    /**
     * The time the upload compelted or errored.
     */
    finishedAt: DateTime | null

    /**
     * The processing error message (not set if state is not ERRORED).
     */
    failure: string | null

    /**
     * Whether or not this upload provides intelligence for the tip of the default branch. Find reference
     * queries will return symbols from remote repositories only when this property is true. This property
     * is updated asynchronously and is eventually consistent with the git data known by the Sourcegraph
     * instance.
     */
    isLatestForRepo: boolean

    /**
     * The rank of this upload in the queue. The value of this field is null if the upload has been processed.
     */
    placeInQueue: number | null
}

/**
 * A list of LSIF uploads.
 */
export interface ILSIFUploadConnection {
    __typename: 'LSIFUploadConnection'

    /**
     * A list of LSIF uploads.
     */
    nodes: ILSIFUpload[]

    /**
     * The total number of uploads in this result set.
     */
    totalCount: number | null

    /**
     * Pagination information.
     */
    pageInfo: IPageInfo
}

/**
 * The state an LSIF index can be in.
 */
export enum LSIFIndexState {
    /**
     * This index is being processed.
     */
    PROCESSING = 'PROCESSING',

    /**
     * This index failed to be processed.
     */
    ERRORED = 'ERRORED',

    /**
     * This index was processed successfully.
     */
    COMPLETED = 'COMPLETED',

    /**
     * This index is queued to be processed later.
     */
    QUEUED = 'QUEUED',
}

/**
 * Metadata and status about an LSIF index.
 */
export interface ILSIFIndex {
    __typename: 'LSIFIndex'

    /**
     * The ID.
     */
    id: ID

    /**
     * The project for which this upload provides code intelligence.
     */
    projectRoot: IGitTree | null

    /**
     * The original 40-character commit commit supplied at index time.
     */
    inputCommit: string

    /**
     * The original root supplied at index schedule time.
     */
    inputRoot: string

    /**
     * The name of the target indexer Docker image (e.g., sourcegraph/lsif-go@sha256:...).
     */
    inputIndexer: string

    /**
     * The index's current state.
     */
    state: LSIFIndexState

    /**
     * The time the index was queued.
     */
    queuedAt: DateTime

    /**
     * The time the index was processed.
     */
    startedAt: DateTime | null

    /**
     * The time the index compelted or errored.
     */
    finishedAt: DateTime | null

    /**
     * The processing error message (not set if state is not ERRORED).
     */
    failure: string | null

    /**
     * The configuration and execution summary (if completed or errored) of this index job.
     */
    steps: IIndexSteps

    /**
     * The rank of this index in the queue. The value of this field is null if the index has been processed.
     */
    placeInQueue: number | null
}

/**
 * Configuration and execution summary of an index job.
 */
export interface IIndexSteps {
    __typename: 'IndexSteps'

    /**
     * Execution log entries related to setting up the indexing workspace.
     */
    setup: IExecutionLogEntry[]

    /**
     * Configuration and execution summary (if completed or errored) of steps to be performed prior to indexing.
     */
    preIndex: IPreIndexStep[]

    /**
     * Configuration and execution summary (if completed or errored) of the indexer.
     */
    index: IIndexStep

    /**
     * Execution log entry related to uploading the dump produced by the indexing step.
     * This field be missing if the upload step had not been executed.
     */
    upload: IExecutionLogEntry | null

    /**
     * Execution log entries related to tearing down the indexing workspace.
     */
    teardown: IExecutionLogEntry[]
}

/**
 * The configuration and execution summary of a step to be performed prior to indexing.
 */
export interface IPreIndexStep {
    __typename: 'PreIndexStep'

    /**
     * The working directory relative to the cloned repository root.
     */
    root: string

    /**
     * The name of the Docker image to run.
     */
    image: string

    /**
     * The arguments to supply to the Docker container's entrypoint.
     */
    commands: string[]

    /**
     * The execution summary (if completed or errored) of the docker command.
     */
    logEntry: IExecutionLogEntry | null
}

/**
 * The configuration and execution summary of the indexer.
 */
export interface IIndexStep {
    __typename: 'IndexStep'

    /**
     * The arguments to supply to the indexer container.
     */
    indexerArgs: string[]

    /**
     * The path to the index file relative to the root directory (dump.lsif by default).
     */
    outfile: string | null

    /**
     * The execution summary (if completed or errored) of the index command.
     */
    logEntry: IExecutionLogEntry | null
}

/**
 * A description of a command run inside the executor to during processing of the parent record.
 */
export interface IExecutionLogEntry {
    __typename: 'ExecutionLogEntry'

    /**
     * An internal tag used to correlate this log entry with other records.
     */
    key: string

    /**
     * The arguments of the command run inside the executor.
     */
    command: string[]

    /**
     * The date when this command started.
     */
    startTime: DateTime

    /**
     * The exit code of the command.
     */
    exitCode: number

    /**
     * The combined stdout and stderr logs of the command.
     */
    out: string

    /**
     * The duration in milliseconds of the command.
     */
    durationMilliseconds: number
}

/**
 * A list of LSIF indexes.
 */
export interface ILSIFIndexConnection {
    __typename: 'LSIFIndexConnection'

    /**
     * A list of LSIF indexes.
     */
    nodes: ILSIFIndex[]

    /**
     * The total number of indexes in this result set.
     */
    totalCount: number | null

    /**
     * Pagination information.
     */
    pageInfo: IPageInfo
}

/**
 * Explicit configuration for indexing a repository.
 */
export interface IIndexConfiguration {
    __typename: 'IndexConfiguration'

    /**
     * The raw JSON-encoded index configuration.
     */
    configuration: string | null
}

/**
 * Mutations that are only used on Sourcegraph.com.
 * FOR INTERNAL USE ONLY.
 */
export interface IDotcomMutation {
    __typename: 'DotcomMutation'

    /**
     * Set or unset a user's associated billing information.
     * Only Sourcegraph.com site admins may perform this mutation.
     * FOR INTERNAL USE ONLY.
     */
    setUserBilling: IEmptyResponse

    /**
     * Creates new product subscription for an account.
     * Only Sourcegraph.com site admins may perform this mutation.
     * FOR INTERNAL USE ONLY.
     */
    createProductSubscription: IProductSubscription

    /**
     * Set or unset a product subscription's associated billing system subscription.
     * Only Sourcegraph.com site admins may perform this mutation.
     * FOR INTERNAL USE ONLY.
     */
    setProductSubscriptionBilling: IEmptyResponse

    /**
     * Generates and signs a new product license and associates it with an existing product subscription. The
     * product license key is signed with Sourcegraph.com's private key and is verifiable with the corresponding
     * public key.
     * Only Sourcegraph.com site admins may perform this mutation.
     * FOR INTERNAL USE ONLY.
     */
    generateProductLicenseForSubscription: IProductLicense

    /**
     * Creates a new product subscription and bills the associated payment method.
     * Only Sourcegraph.com authenticated users may perform this mutation.
     * FOR INTERNAL USE ONLY.
     */
    createPaidProductSubscription: ICreatePaidProductSubscriptionResult

    /**
     * Updates a new product subscription and credits or debits the associated payment method.
     * Only Sourcegraph.com site admins and the subscription's account owner may perform this
     * mutation.
     * FOR INTERNAL USE ONLY.
     */
    updatePaidProductSubscription: IUpdatePaidProductSubscriptionResult

    /**
     * Archives an existing product subscription.
     * Only Sourcegraph.com site admins may perform this mutation.
     * FOR INTERNAL USE ONLY.
     */
    archiveProductSubscription: IEmptyResponse
}

export interface ISetUserBillingOnDotcomMutationArguments {
    /**
     * The user to update.
     */
    user: ID

    /**
     * The billing customer ID (on the billing system) to associate this user with. If null, the association is
     * removed (i.e., the user is unlinked from the billing customer record).
     */
    billingCustomerID?: string | null
}

export interface ICreateProductSubscriptionOnDotcomMutationArguments {
    /**
     * The ID of the user (i.e., customer) to whom this product subscription is assigned.
     */
    accountID: ID
}

export interface ISetProductSubscriptionBillingOnDotcomMutationArguments {
    /**
     * The product subscription to update.
     */
    id: ID

    /**
     * The billing subscription ID (on the billing system) to associate this product subscription with. If null,
     * the association is removed (i.e., the subscription is unlinked from billing).
     */
    billingSubscriptionID?: string | null
}

export interface IGenerateProductLicenseForSubscriptionOnDotcomMutationArguments {
    /**
     * The product subscription to associate with the license.
     */
    productSubscriptionID: ID

    /**
     * The license to generate.
     */
    license: IProductLicenseInput
}

export interface ICreatePaidProductSubscriptionOnDotcomMutationArguments {
    /**
     * The ID of the user (i.e., customer) to whom the product subscription is assigned.
     * Only Sourcegraph.com site admins may perform this mutation for an accountID != the user ID of the
     * authenticated user.
     */
    accountID: ID

    /**
     * The details of the product subscription.
     */
    productSubscription: IProductSubscriptionInput

    /**
     * The token that represents the payment method used to purchase this product subscription,
     * or null if no payment is required.
     */
    paymentToken?: string | null
}

export interface IUpdatePaidProductSubscriptionOnDotcomMutationArguments {
    /**
     * The subscription to update.
     */
    subscriptionID: ID

    /**
     * The updated details of the product subscription. All fields of the input type must be set
     * (i.e., it does not support passing a null value to mean "do not update this field's
     * value").
     */
    update: IProductSubscriptionInput

    /**
     * The token that represents the payment method used to pay for (or receive credit for) this
     * product subscription update, or null if no payment is required.
     */
    paymentToken?: string | null
}

export interface IArchiveProductSubscriptionOnDotcomMutationArguments {
    id: ID
}

/**
 * Mutations that are only used on Sourcegraph.com.
 * FOR INTERNAL USE ONLY.
 */
export interface IDotcomQuery {
    __typename: 'DotcomQuery'

    /**
     * The product subscription with the given UUID. An error is returned if no such product
     * subscription exists.
     * Only Sourcegraph.com site admins and the account owners of the product subscription may
     * perform this query.
     * FOR INTERNAL USE ONLY.
     */
    productSubscription: IProductSubscription

    /**
     * A list of product subscriptions.
     * FOR INTERNAL USE ONLY.
     */
    productSubscriptions: IProductSubscriptionConnection

    /**
     * The invoice that would be generated for a new or updated subscription. This is used to show
     * users a preview of the credits, debits, and other billing information before creating or
     * updating a subscription.
     * Performing this query does not mutate any data or cause any billing changes to be made.
     */
    previewProductSubscriptionInvoice: IProductSubscriptionPreviewInvoice

    /**
     * A list of product licenses.
     * Only Sourcegraph.com site admins may perform this query.
     * FOR INTERNAL USE ONLY.
     */
    productLicenses: IProductLicenseConnection

    /**
     * A list of product pricing plans for Sourcegraph.
     */
    productPlans: IProductPlan[]
}

export interface IProductSubscriptionOnDotcomQueryArguments {
    uuid: string
}

export interface IProductSubscriptionsOnDotcomQueryArguments {
    /**
     * Returns the first n product subscriptions from the list.
     */
    first?: number | null

    /**
     * Returns only product subscriptions for the given account.
     * Only Sourcegraph.com site admins may perform this query with account == null.
     */
    account?: ID | null

    /**
     * Returns product subscriptions from users with usernames or email addresses that match the query.
     */
    query?: string | null
}

export interface IPreviewProductSubscriptionInvoiceOnDotcomQueryArguments {
    /**
     * The customer account (user) for whom this preview invoice will be generated, or null if there is none.
     */
    account?: ID | null

    /**
     * If non-null, preview the invoice for an update to the existing product subscription. The
     * product subscription's billing customer must match the account parameter. If null, preview
     * the invoice for a new subscription.
     */
    subscriptionToUpdate?: ID | null

    /**
     * The parameters for the product subscription to preview. All fields of the input type must
     * be set (i.e., it does not support passing a null value to mean "do not update this field's
     * value" when updating an existing subscription).
     */
    productSubscription: IProductSubscriptionInput
}

export interface IProductLicensesOnDotcomQueryArguments {
    /**
     * Returns the first n product subscriptions from the list.
     */
    first?: number | null

    /**
     * Returns only product subscriptions whose license key contains this substring.
     */
    licenseKeySubstring?: string | null

    /**
     * Returns only product licenses associated with the given subscription
     */
    productSubscriptionID?: ID | null
}

/**
 * A product subscription that was created on Sourcegraph.com.
 * FOR INTERNAL USE ONLY.
 */
export interface IProductSubscription {
    __typename: 'ProductSubscription'

    /**
     * The unique ID of this product subscription.
     */
    id: ID

    /**
     * The unique UUID of this product subscription. Unlike ProductSubscription.id, this does not
     * encode the type and is not a GraphQL node ID.
     */
    uuid: string

    /**
     * A name for the product subscription derived from its ID. The name is not guaranteed to be unique.
     */
    name: string

    /**
     * The user (i.e., customer) to whom this subscription is granted, or null if the account has been deleted.
     */
    account: IUser | null

    /**
     * The information that determines the price of this subscription, or null if there is no billing
     * information associated with this subscription.
     */
    invoiceItem: IProductSubscriptionInvoiceItem | null

    /**
     * A list of billing-related events related to this product subscription.
     */
    events: IProductSubscriptionEvent[]

    /**
     * The currently active product license associated with this product subscription, if any.
     */
    activeLicense: IProductLicense | null

    /**
     * A list of product licenses associated with this product subscription.
     * Only Sourcegraph.com site admins may list inactive product licenses (other viewers should use
     * ProductSubscription.activeLicense).
     */
    productLicenses: IProductLicenseConnection

    /**
     * The date when this product subscription was created.
     */
    createdAt: DateTime

    /**
     * Whether this product subscription was archived.
     */
    isArchived: boolean

    /**
     * The URL to view this product subscription.
     */
    url: string

    /**
     * The URL to view this product subscription in the site admin area.
     * Only Sourcegraph.com site admins may query this field.
     */
    urlForSiteAdmin: string | null

    /**
     * The URL to view this product subscription's billing information (for site admins).
     * Only Sourcegraph.com site admins may query this field.
     */
    urlForSiteAdminBilling: string | null
}

export interface IProductLicensesOnProductSubscriptionArguments {
    /**
     * Returns the first n product licenses from the list.
     */
    first?: number | null
}

/**
 * A list of product subscriptions.
 * FOR INTERNAL USE ONLY.
 */
export interface IProductSubscriptionConnection {
    __typename: 'ProductSubscriptionConnection'

    /**
     * A list of product subscriptions.
     */
    nodes: IProductSubscription[]

    /**
     * The total count of product subscriptions in the connection. This total count may be larger than the number of
     * nodes in this object when the result is paginated.
     */
    totalCount: number

    /**
     * Pagination information.
     */
    pageInfo: IPageInfo
}

/**
 * A preview of an invoice that would be generated for a new or updated product subscription.
 * FOR INTERNAL USE ONLY.
 */
export interface IProductSubscriptionPreviewInvoice {
    __typename: 'ProductSubscriptionPreviewInvoice'

    /**
     * The net price for this invoice, in USD cents. If this invoice represents an update to a
     * subscription, this is the difference between the existing price and the updated price.
     */
    price: number

    /**
     * For updates to existing subscriptions, the effective date for which this preview invoice was
     * calculated, expressed as the number of seconds since the epoch. For new subscriptions, this is
     * null.
     */
    prorationDate: string | null

    /**
     * Whether this invoice requires manual intervention.
     */
    isDowngradeRequiringManualIntervention: boolean

    /**
     * The "before" state of the product subscription (i.e., the existing subscription), prior to the update that this preview
     * represents, or null if the preview is for a new subscription.
     */
    beforeInvoiceItem: IProductSubscriptionInvoiceItem | null

    /**
     * The "after" state of the product subscription, with the update applied to the subscription.
     * For new subscriptions, this is just the invoice item for the subscription that will be
     * created.
     */
    afterInvoiceItem: IProductSubscriptionInvoiceItem
}

/**
 * An input type that describes a product license to be generated and signed.
 * FOR INTERNAL USE ONLY.
 */
export interface IProductLicenseInput {
    /**
     * The tags that indicate which features are activated by this license.
     */
    tags: string[]

    /**
     * The number of users for which this product subscription is valid.
     */
    userCount: number

    /**
     * The expiration date of this product license, expressed as the number of seconds since the epoch.
     */
    expiresAt: number
}

/**
 * A product license that was created on Sourcegraph.com.
 * FOR INTERNAL USE ONLY.
 */
export interface IProductLicense {
    __typename: 'ProductLicense'

    /**
     * The unique ID of this product license.
     */
    id: ID

    /**
     * The product subscription associated with this product license.
     */
    subscription: IProductSubscription

    /**
     * Information about this product license.
     */
    info: IProductLicenseInfo | null

    /**
     * The license key.
     */
    licenseKey: string

    /**
     * The date when this product license was created.
     */
    createdAt: DateTime
}

/**
 * A list of product licenses.
 * FOR INTERNAL USE ONLY.
 */
export interface IProductLicenseConnection {
    __typename: 'ProductLicenseConnection'

    /**
     * A list of product licenses.
     */
    nodes: IProductLicense[]

    /**
     * The total count of product licenses in the connection. This total count may be larger than the number of
     * nodes in this object when the result is paginated.
     */
    totalCount: number

    /**
     * Pagination information.
     */
    pageInfo: IPageInfo
}

/**
 * A product pricing plan for Sourcegraph.
 * FOR INTERNAL USE ONLY.
 */
export interface IProductPlan {
    __typename: 'ProductPlan'

    /**
     * The billing system's unique ID for the plan.
     */
    billingPlanID: string

    /**
     * The unique ID for the product.
     */
    productPlanID: string

    /**
     * The name of the product plan (e.g., "Enterprise Starter"). This is displayed to the user and
     * should be human-readable.
     */
    name: string

    /**
     * The name with the brand (e.g., "Sourcegraph Enterprise Starter").
     */
    nameWithBrand: string

    /**
     * The price (in USD cents) for one user for a year.
     */
    pricePerUserPerYear: number

    /**
     * The minimum quantity (user count) that can be purchased. Only applies when using tiered pricing.
     */
    minQuantity: number | null

    /**
     * The maximum quantity (user count) that can be purchased. Only applies when using tiered pricing.
     */
    maxQuantity: number | null

    /**
     * Defines if the tiering price should be graduated or volume based.
     */
    tiersMode: string

    /**
     * The tiered pricing for the plan.
     */
    planTiers: IPlanTier[]
}

/**
 * The information about a plan's tier.
 * FOR INTERNAL USE ONLY.
 */
export interface IPlanTier {
    __typename: 'PlanTier'

    /**
     * The per-user amount for this tier.
     */
    unitAmount: number

    /**
     * The maximum number of users that this tier applies to.
     */
    upTo: number

    /**
     * The flat fee for this tier.
     */
    flatAmount: number
}

/**
 * The information about a product subscription that determines its price.
 * FOR INTERNAL USE ONLY.
 */
export interface IProductSubscriptionInvoiceItem {
    __typename: 'ProductSubscriptionInvoiceItem'

    /**
     * The product plan for the subscription.
     */
    plan: IProductPlan

    /**
     * This subscription's user count.
     */
    userCount: number

    /**
     * The date when the subscription expires.
     */
    expiresAt: DateTime
}

/**
 * An input type that describes a product subscription to be purchased. Corresponds to
 * ProductSubscriptionInvoiceItem.
 * FOR INTERNAL USE ONLY.
 */
export interface IProductSubscriptionInput {
    /**
     * The billing plan ID for the subscription (ProductPlan.billingPlanID). This also specifies the
     * billing product, because a plan is associated with its product in the billing system.
     */
    billingPlanID: string

    /**
     * This subscription's user count.
     */
    userCount: number
}

/**
 * The result of Mutation.dotcom.createPaidProductSubscription.
 * FOR INTERNAL USE ONLY.
 */
export interface ICreatePaidProductSubscriptionResult {
    __typename: 'CreatePaidProductSubscriptionResult'

    /**
     * The newly created product subscription.
     */
    productSubscription: IProductSubscription
}

/**
 * The result of Mutation.dotcom.updatePaidProductSubscription.
 * FOR INTERNAL USE ONLY.
 */
export interface IUpdatePaidProductSubscriptionResult {
    __typename: 'UpdatePaidProductSubscriptionResult'

    /**
     * The updated product subscription.
     */
    productSubscription: IProductSubscription
}

/**
 * An event related to a product subscription.
 * FOR INTERNAL USE ONLY.
 */
export interface IProductSubscriptionEvent {
    __typename: 'ProductSubscriptionEvent'

    /**
     * The unique ID of the event.
     */
    id: string

    /**
     * The date when the event occurred.
     */
    date: string

    /**
     * The title of the event.
     */
    title: string

    /**
     * A description of the event.
     */
    description: string | null

    /**
     * A URL where the user can see more information about the event.
     */
    url: string | null
}

/**
 * FOR INTERNAL USE ONLY: A status message produced when repositories are being
 * cloned
 */
export interface ICloningProgress {
    __typename: 'CloningProgress'

    /**
     * The message of this status message
     */
    message: string
}

/**
 * FOR INTERNAL USE ONLY: A status message produced when repositories could not
 * be synced from an external service
 */
export interface IExternalServiceSyncError {
    __typename: 'ExternalServiceSyncError'

    /**
     * The message of this status message
     */
    message: string

    /**
     * The external service that failed to sync
     */
    externalService: IExternalService
}

/**
 * FOR INTERNAL USE ONLY: A status message produced when repositories could not
 * be synced
 */
export interface ISyncError {
    __typename: 'SyncError'

    /**
     * The message of this status message
     */
    message: string
}

/**
 * FOR INTERNAL USE ONLY: A status message
 */
export type StatusMessage = ICloningProgress | IExternalServiceSyncError | ISyncError

/**
 * FOR INTERNAL USE ONLY: A repository statistic
 */
export interface IRepositoryStats {
    __typename: 'RepositoryStats'

    /**
     * The amount of bytes stored in .git directories
     */
    gitDirBytes: any

    /**
     * The number of lines indexed
     */
    indexedLinesCount: any
}

/**
 * Different repository permission levels.
 */
export enum RepositoryPermission {
    READ = 'READ',
}

/**
 * A single user event that has been logged.
 */
export interface IEventLog {
    __typename: 'EventLog'

    /**
     * The name of the event.
     */
    name: string

    /**
     * The user who executed the event, if one exists.
     */
    user: IUser | null

    /**
     * The randomly generated unique user ID stored in a browser cookie.
     */
    anonymousUserID: string

    /**
     * The URL when the event was logged.
     */
    url: string

    /**
     * The source of the event.
     */
    source: EventSource

    /**
     * The additional argument information.
     */
    argument: string | null

    /**
     * The Sourcegraph version when the event was logged.
     */
    version: string

    /**
     * The timestamp when the event was logged.
     */
    timestamp: DateTime
}

/**
 * A list of event logs.
 */
export interface IEventLogsConnection {
    __typename: 'EventLogsConnection'

    /**
     * A list of event logs.
     */
    nodes: IEventLog[]

    /**
     * The total count of event logs in the connection. This total count may be larger than the number of nodes
     * in this object when the result is paginated.
     */
    totalCount: number

    /**
     * Pagination information.
     */
    pageInfo: IPageInfo
}

/**
 * A list of code host repositories
 */
export interface ICodeHostRepositoryConnection {
    __typename: 'CodeHostRepositoryConnection'

    /**
     * A list of repositories affiliated with a code host.
     */
    nodes: ICodeHostRepository[]
}

/**
 * A repository returned directly from a code host
 */
export interface ICodeHostRepository {
    __typename: 'CodeHostRepository'

    /**
     * The Name "owner/reponame" of the repo
     */
    name: string

    /**
     * The code host the repo came from
     */
    codeHost: IExternalService | null

    /**
     * Is the repo private
     */
    private: boolean
}

/**
 * The state of the campaign
 */
export enum CampaignState {
    OPEN = 'OPEN',
    CLOSED = 'CLOSED',
}

/**
 * A campaign is a set of related changes to apply to code across one or more repositories.
 */
export interface ICampaign {
    __typename: 'Campaign'

    /**
     * The unique ID for the campaign.
     */
    id: ID

    /**
     * The namespace where this campaign is defined.
     */
    namespace: Namespace

    /**
     * The name of the campaign.
     */
    name: string

    /**
     * The description (as Markdown).
     */
    description: string | null

    /**
     * The user that created the initial spec. In an org, this will be different from the namespace, or null if the user was deleted.
     */
    specCreator: IUser | null

    /**
     * The user who created the campaign initially by applying the spec for the first time, or null if the user was deleted.
     */
    initialApplier: IUser | null

    /**
     * The user who last updated the campaign by applying a spec to this campaign.
     * If the campaign hasn't been updated, the lastApplier is the initialApplier, or null if the user was deleted.
     */
    lastApplier: IUser | null

    /**
     * Whether the current user can edit or delete this campaign.
     */
    viewerCanAdminister: boolean

    /**
     * The URL to this campaign.
     */
    url: string

    /**
     * The date and time when the campaign was created.
     */
    createdAt: DateTime

    /**
     * The date and time when the campaign was updated. That can be by applying a spec, or by an internal process.
     * For reading the time the campaign spec was changed last, see lastAppliedAt.
     */
    updatedAt: DateTime

    /**
     * The date and time when the campaign was last updated with a new spec.
     */
    lastAppliedAt: DateTime

    /**
     * The date and time when the campaign was closed. If set, applying a spec for this campaign will fail with an error.
     */
    closedAt: DateTime | null

    /**
     * Stats on all the changesets that are tracked in this campaign.
     */
    changesetsStats: IChangesetsStats

    /**
     * The changesets in this campaign that already exist on the code host.
     */
    changesets: IChangesetConnection

    /**
     * The changeset counts over time, in 1-day intervals backwards from the point in time given in
     * the "to" parameter.
     */
    changesetCountsOverTime: IChangesetCounts[]

    /**
     * The diff stat for all the changesets in the campaign.
     */
    diffStat: IDiffStat

    /**
     * The current campaign spec this campaign reflects.
     */
    currentSpec: ICampaignSpec
}

export interface IChangesetsOnCampaignArguments {
    /**
     * @default 50
     */
    first?: number | null

    /**
     * Opaque pagination cursor.
     */
    after?: string | null

    /**
     * Only include changesets with the given state.
     */
    state?: ChangesetState | null

    /**
     * Only include changesets with the given review state.
     */
    reviewState?: ChangesetReviewState | null

    /**
     * Only include changesets with the given check state.
     */
    checkState?: ChangesetCheckState | null

    /**
     * Only return changesets that have been published by this campaign. Imported changesets will be omitted.
     */
    onlyPublishedByThisCampaign?: boolean | null

    /**
     * Search for changesets matching this query. Queries may include quoted substrings to match phrases, and words may be preceded by - to negate them.
     */
    search?: string | null
}

export interface IChangesetCountsOverTimeOnCampaignArguments {
    /**
     * Only include changeset counts up to this point in time (inclusive). Defaults to Campaign.createdAt.
     */
    from?: DateTime | null

    /**
     * Only include changeset counts up to this point in time (inclusive). Defaults to the
     * current time.
     */
    to?: DateTime | null
}

/**
 * The counts of changesets in certain states at a specific point in time.
 */
export interface IChangesetCounts {
    __typename: 'ChangesetCounts'

    /**
     * The point in time these counts were recorded.
     */
    date: DateTime

    /**
     * The total number of changesets.
     */
    total: number

    /**
     * The number of merged changesets.
     */
    merged: number

    /**
     * The number of closed changesets.
     */
    closed: number

    /**
     * The number of draft changesets (independent of review state).
     */
    draft: number

    /**
     * The number of open changesets (independent of review state).
     */
    open: number

    /**
     * The number of changesets that are both open and approved.
     */
    openApproved: number

    /**
     * The number of changesets that are both open and have requested changes.
     */
    openChangesRequested: number

    /**
     * The number of changesets that are both open and are pending review.
     */
    openPending: number
}

/**
 * A list of campaigns.
 */
export interface ICampaignConnection {
    __typename: 'CampaignConnection'

    /**
     * A list of campaigns.
     */
    nodes: ICampaign[]

    /**
     * The total number of campaigns in the connection.
     */
    totalCount: number

    /**
     * Pagination information.
     */
    pageInfo: IPageInfo
}

/**
 * The publication state of a changeset on Sourcegraph
 */
export enum ChangesetPublicationState {
    /**
     * The changeset has not yet been created on the code host.
     */
    UNPUBLISHED = 'UNPUBLISHED',

    /**
     * The changeset has been created on the code host.
     */
    PUBLISHED = 'PUBLISHED',
}

/**
 * The reconciler state of a changeset on Sourcegraph
 */
export enum ChangesetReconcilerState {
    /**
     * The changeset is enqueued for the reconciler to process it.
     */
    QUEUED = 'QUEUED',

    /**
     * The changeset reconciler is currently computing the delta between the
     * If a delta exists, the reconciler tries to update the state of the
     * changeset on the code host and on Sourcegraph to the desired state.
     */
    PROCESSING = 'PROCESSING',

    /**
     * The changeset reconciler ran into a problem while processing the
     * changeset and will retry it for a number of retries.
     */
    ERRORED = 'ERRORED',

    /**
     * The changeset reconciler ran into a problem while processing the
     * changeset that can't be fixed by retrying.
     */
    FAILED = 'FAILED',

    /**
     * The changeset is not enqueued for processing.
     */
    COMPLETED = 'COMPLETED',
}

/**
 * The state of a changeset on the code host on which it's hosted.
 */
export enum ChangesetExternalState {
    DRAFT = 'DRAFT',
    OPEN = 'OPEN',
    CLOSED = 'CLOSED',
    MERGED = 'MERGED',
    DELETED = 'DELETED',
}

/**
 * The review state of a changeset.
 */
export enum ChangesetReviewState {
    APPROVED = 'APPROVED',
    CHANGES_REQUESTED = 'CHANGES_REQUESTED',
    PENDING = 'PENDING',
    COMMENTED = 'COMMENTED',
    DISMISSED = 'DISMISSED',
}

/**
 * The state of checks (e.g., for continuous integration) on a changeset.
 */
export enum ChangesetCheckState {
    PENDING = 'PENDING',
    PASSED = 'PASSED',
    FAILED = 'FAILED',
}

/**
 * A label attached to a changeset on a code host.
 */
export interface IChangesetLabel {
    __typename: 'ChangesetLabel'

    /**
     * The label's text.
     */
    text: string

    /**
     * The label's color, as a hex color code without the . For example: "93ba13".
     */
    color: string

    /**
     * An optional description of the label.
     */
    description: string | null
}

/**
 * The visual state a changeset is currently in.
 */
export enum ChangesetState {
    /**
     * The changeset has not been marked as to be published.
     */
    UNPUBLISHED = 'UNPUBLISHED',

    /**
     * The changeset reconciler ran into a problem while processing the
     * changeset that can't be fixed by retrying.
     */
    FAILED = 'FAILED',

    /**
     * The changeset reconciler ran into a problem while processing the
     * changeset and will retry it for a number of retries.
     */
    RETRYING = 'RETRYING',

    /**
     * The changeset reconciler is currently computing the delta between the
     * If a delta exists, the reconciler tries to update the state of the
     * changeset on the code host and on Sourcegraph to the desired state.
     */
    PROCESSING = 'PROCESSING',

    /**
     * The changeset is published, not being reconciled and open on the code host.
     */
    OPEN = 'OPEN',

    /**
     * The changeset is published, not being reconciled and in draft state on the code host.
     */
    DRAFT = 'DRAFT',

    /**
     * The changeset is published, not being reconciled and closed on the code host.
     */
    CLOSED = 'CLOSED',

    /**
     * The changeset is published, not being reconciled and merged on the code host.
     */
    MERGED = 'MERGED',

    /**
     * The changeset is published, not being reconciled and has been deleted on the code host.
     */
    DELETED = 'DELETED',
}

/**
 * A changeset on a codehost.
 */
export type Changeset = IHiddenExternalChangeset | IExternalChangeset

/**
 * A changeset on a codehost.
 */
export interface IChangeset {
    __typename: 'Changeset'

    /**
     * The unique ID for the changeset.
     */
    id: ID

    /**
     * The campaigns that contain this changeset.
     * @deprecated "campaigns have been renamed to batch changes. Use batchChanges instead."
     */
    campaigns: ICampaignConnection

    /**
     * The batch changes that contain this changeset.
     */
    batchChanges: IBatchChangeConnection

    /**
     * The publication state of the changeset.
     * @deprecated "Use state instead. This field is deprecated and will be removed in a future release."
     */
    publicationState: ChangesetPublicationState

    /**
     * The reconciler state of the changeset.
     * @deprecated "Use state instead. This field is deprecated and will be removed in a future release."
     */
    reconcilerState: ChangesetReconcilerState

    /**
     * The external state of the changeset, or null when not yet published to the code host.
     * @deprecated "Use state instead. This field is deprecated and will be removed in a future release."
     */
    externalState: ChangesetExternalState | null

    /**
     * The state of the changeset.
     */
    state: ChangesetState

    /**
     * The date and time when the changeset was created.
     */
    createdAt: DateTime

    /**
     * The date and time when the changeset was updated.
     */
    updatedAt: DateTime

    /**
     * The date and time when the next changeset sync is scheduled, or null if none is scheduled.
     */
    nextSyncAt: DateTime | null
}

export interface ICampaignsOnChangesetArguments {
    /**
     * Returns the first n campaigns from the list.
     * @default 50
     */
    first?: number | null

    /**
     * Opaque pagination cursor.
     */
    after?: string | null

    /**
     * Only return campaigns in this state.
     */
    state?: CampaignState | null

    /**
     * Only include campaigns that the viewer can administer.
     */
    viewerCanAdminister?: boolean | null
}

export interface IBatchChangesOnChangesetArguments {
    /**
     * Returns the first n batch changes from the list.
     * @default 50
     */
    first?: number | null

    /**
     * Opaque pagination cursor.
     */
    after?: string | null

    /**
     * Only return batch changes in this state.
     */
    state?: BatchChangeState | null

    /**
     * Only include batch changes that the viewer can administer.
     */
    viewerCanAdminister?: boolean | null
}

/**
 * A changeset on a code host that the user does not have access to.
 */
export interface IHiddenExternalChangeset {
    __typename: 'HiddenExternalChangeset'

    /**
     * The unique ID for the changeset.
     */
    id: ID

    /**
     * The campaigns that contain this changeset.
     * @deprecated "campaigns have been renamed to batch changes. Use batchChanges instead."
     */
    campaigns: ICampaignConnection

    /**
     * The batch changes that contain this changeset.
     */
    batchChanges: IBatchChangeConnection

    /**
     * The publication state of the changeset.
     * @deprecated "Use state instead. This field is deprecated and will be removed in a future release."
     */
    publicationState: ChangesetPublicationState

    /**
     * The reconciler state of the changeset.
     * @deprecated "Use state instead. This field is deprecated and will be removed in a future release."
     */
    reconcilerState: ChangesetReconcilerState

    /**
     * The external state of the changeset, or null when not yet published to the code host.
     * @deprecated "Use state instead. This field is deprecated and will be removed in a future release."
     */
    externalState: ChangesetExternalState | null

    /**
     * The state of the changeset.
     */
    state: ChangesetState

    /**
     * The date and time when the changeset was created.
     */
    createdAt: DateTime

    /**
     * The date and time when the changeset was updated.
     */
    updatedAt: DateTime

    /**
     * The date and time when the next changeset sync is scheduled, or null if none is scheduled.
     */
    nextSyncAt: DateTime | null
}

export interface ICampaignsOnHiddenExternalChangesetArguments {
    /**
     * Returns the first n campaigns from the list.
     * @default 50
     */
    first?: number | null

    /**
     * Opaque pagination cursor.
     */
    after?: string | null

    /**
     * Only return campaigns in this state.
     */
    state?: CampaignState | null

    /**
     * Only include campaigns that the viewer can administer.
     */
    viewerCanAdminister?: boolean | null
}

export interface IBatchChangesOnHiddenExternalChangesetArguments {
    /**
     * Returns the first n batch changes from the list.
     * @default 50
     */
    first?: number | null

    /**
     * Opaque pagination cursor.
     */
    after?: string | null

    /**
     * Only return batch changes in this state.
     */
    state?: BatchChangeState | null

    /**
     * Only include batch changes that the viewer can administer.
     */
    viewerCanAdminister?: boolean | null
}

/**
 * A changeset on a code host (e.g., a pull request on GitHub).
 */
export interface IExternalChangeset {
    __typename: 'ExternalChangeset'

    /**
     * The unique ID for the changeset.
     */
    id: ID

    /**
     * The external ID that uniquely identifies this ExternalChangeset on the
     * code host. For example, on GitHub this is the pull request number. This is only set once the changeset is published on the code host.
     */
    externalID: string | null

    /**
     * The repository changed by this changeset.
     */
    repository: IRepository

    /**
     * The campaigns that contain this changeset.
     * @deprecated "campaigns have been renamed to batch changes. Use batchChanges instead."
     */
    campaigns: ICampaignConnection

    /**
     * The batch changes that contain this changeset.
     */
    batchChanges: IBatchChangeConnection

    /**
     * The events belonging to this changeset.
     */
    events: IChangesetEventConnection

    /**
     * The date and time when the changeset was created.
     */
    createdAt: DateTime

    /**
     * The date and time when the changeset was updated.
     */
    updatedAt: DateTime

    /**
     * The date and time when the next changeset sync is scheduled, or null if none is scheduled or when the initial sync hasn't happened.
     */
    nextSyncAt: DateTime | null

    /**
     * The title of the changeset, or null if the data hasn't been synced from the code host yet.
     */
    title: string | null

    /**
     * The body of the changeset, or null if the data hasn't been synced from the code host yet.
     */
    body: string | null

    /**
     * The author of the changeset, or null if the data hasn't been synced from the code host yet,
     * or the changeset has not yet been published.
     */
    author: IPerson | null

    /**
     * The publication state of the changeset.
     * @deprecated "Use state instead. This field is deprecated and will be removed in a future release."
     */
    publicationState: ChangesetPublicationState

    /**
     * The reconciler state of the changeset.
     * @deprecated "Use state instead. This field is deprecated and will be removed in a future release."
     */
    reconcilerState: ChangesetReconcilerState

    /**
     * The external state of the changeset, or null when not yet published to the code host.
     * @deprecated "Use state instead. This field is deprecated and will be removed in a future release."
     */
    externalState: ChangesetExternalState | null

    /**
     * The state of the changeset.
     */
    state: ChangesetState

    /**
     * The labels attached to the changeset on the code host.
     */
    labels: IChangesetLabel[]

    /**
     * The external URL of the changeset on the code host. Not set when changeset state is UNPUBLISHED, externalState is DELETED, or the changeset's data hasn't been synced yet.
     */
    externalURL: IExternalLink | null

    /**
     * The review state of this changeset. This is only set once the changeset is published on the code host.
     */
    reviewState: ChangesetReviewState | null

    /**
     * The diff of this changeset, or null if the changeset is closed (without merging) or is already merged.
     */
    diff: RepositoryComparisonInterface | null

    /**
     * The diffstat of this changeset, or null if the changeset is closed
     * (without merging) or is already merged. This data is also available
     * indirectly through the diff field above, but if only the diffStat is
     * required, this field is cheaper to access.
     */
    diffStat: IDiffStat | null

    /**
     * The state of the checks (e.g., for continuous integration) on this changeset, or null if no
     * checks have been configured.
     */
    checkState: ChangesetCheckState | null

    /**
     * An error that has occurred when publishing or updating the changeset. This is only set when the changeset state is ERRORED and the viewer can administer this changeset.
     */
    error: string | null

    /**
     * An error that has occured during the last sync of the changeset. Null, if was successful.
     */
    syncerError: string | null

    /**
     * The current changeset spec for this changeset.
     *
     * Null if the changeset was only imported.
     */
    currentSpec: IVisibleChangesetSpec | null
}

export interface ICampaignsOnExternalChangesetArguments {
    /**
     * Returns the first n campaigns from the list.
     * @default 50
     */
    first?: number | null

    /**
     * Opaque pagination cursor.
     */
    after?: string | null

    /**
     * Only return campaigns in this state.
     */
    state?: CampaignState | null

    /**
     * Only include campaigns that the viewer can administer.
     */
    viewerCanAdminister?: boolean | null
}

export interface IBatchChangesOnExternalChangesetArguments {
    /**
     * Returns the first n batch changes from the list.
     * @default 50
     */
    first?: number | null

    /**
     * Opaque pagination cursor.
     */
    after?: string | null

    /**
     * Only return batch changes in this state.
     */
    state?: BatchChangeState | null

    /**
     * Only include batch changes that the viewer can administer.
     */
    viewerCanAdminister?: boolean | null
}

export interface IEventsOnExternalChangesetArguments {
    /**
     * @default 50
     */
    first?: number | null
    after?: string | null
}

/**
 * Used in the campaign page for the overview component.
 */
export interface IChangesetsStats {
    __typename: 'ChangesetsStats'

    /**
     * The count of unpublished changesets.
     */
    unpublished: number

    /**
     * The count of externalState: DRAFT changesets.
     */
    draft: number

    /**
     * The count of externalState: OPEN changesets.
     */
    open: number

    /**
     * The count of externalState: MERGED changesets.
     */
    merged: number

    /**
     * The count of externalState: CLOSED changesets.
     */
    closed: number

    /**
     * The count of externalState: DELETED changesets.
     */
    deleted: number

    /**
     * The count of changesets in retrying state.
     */
    retrying: number

    /**
     * The count of changesets in failed state.
     */
    failed: number

    /**
     * The count of changesets that are currently processing or enqueued to be.
     */
    processing: number

    /**
     * The count of all changesets.
     */
    total: number
}

/**
 * A list of changesets.
 */
export interface IChangesetConnection {
    __typename: 'ChangesetConnection'

    /**
     * A list of changesets.
     */
    nodes: Changeset[]

    /**
     * The total number of changesets in the connection.
     */
    totalCount: number

    /**
     * Pagination information.
     */
    pageInfo: IPageInfo
}

/**
 * A changeset event in a code host (e.g., a comment on a pull request on GitHub).
 */
export interface IChangesetEvent {
    __typename: 'ChangesetEvent'

    /**
     * The unique ID for the changeset event.
     */
    id: ID

    /**
     * The changeset this event belongs to.
     */
    changeset: IExternalChangeset

    /**
     * The date and time when the changeset was created.
     */
    createdAt: DateTime
}

/**
 * A list of changeset events.
 */
export interface IChangesetEventConnection {
    __typename: 'ChangesetEventConnection'

    /**
     * A list of changeset events.
     */
    nodes: IChangesetEvent[]

    /**
     * The total number of changeset events in the connection.
     */
    totalCount: number

    /**
     * Pagination information.
     */
    pageInfo: IPageInfo
}

/**
 * A connection of all code hosts usable with campaigns and accessible by the user
 * this is requested on.
 */
export interface ICampaignsCodeHostConnection {
    __typename: 'CampaignsCodeHostConnection'

    /**
     * A list of code hosts.
     */
    nodes: ICampaignsCodeHost[]

    /**
     * The total number of configured external services in the connection.
     */
    totalCount: number

    /**
     * Pagination information.
     */
    pageInfo: IPageInfo
}

/**
 * A code host usable with campaigns. This service is accessible by the user it belongs to.
 */
export interface ICampaignsCodeHost {
    __typename: 'CampaignsCodeHost'

    /**
     * The kind of external service.
     */
    externalServiceKind: ExternalServiceKind

    /**
     * The URL of the external service.
     */
    externalServiceURL: string

    /**
     * The configured credential, if any.
     */
    credential: ICampaignsCredential | null

    /**
     * If true, some of the repositories on this code host require
     * an SSH key to be configured.
     */
    requiresSSH: boolean
}

/**
 * A user token configured for campaigns use on the specified code host.
 */
export interface ICampaignsCredential {
    __typename: 'CampaignsCredential'

    /**
     * A globally unique identifier.
     */
    id: ID

    /**
     * The kind of external service.
     */
    externalServiceKind: ExternalServiceKind

    /**
     * The URL of the external service.
     */
    externalServiceURL: string

    /**
     * The public key to use on the external service for SSH keypair authentication.
     * Not set if the credential doesn't support SSH access.
     */
    sshPublicKey: string | null

    /**
     * The date and time this token has been created at.
     */
    createdAt: DateTime
}

/**
 * This enum declares all operations supported by the reconciler.
 */
export enum ChangesetSpecOperation {
    /**
     * Push a new commit to the code host.
     */
    PUSH = 'PUSH',

    /**
     * Update the existing changeset on the codehost. This is purely the changeset resource on the code host,
     * not the git commit. For updates to the commit, see 'PUSH'.
     */
    UPDATE = 'UPDATE',

    /**
     * Move the existing changeset out of being a draft.
     */
    UNDRAFT = 'UNDRAFT',

    /**
     * Publish a changeset to the codehost.
     */
    PUBLISH = 'PUBLISH',

    /**
     * Publish a changeset to the codehost as a draft changeset. (Only on supported code hosts).
     */
    PUBLISH_DRAFT = 'PUBLISH_DRAFT',

    /**
     * Sync the changeset with the current state on the codehost.
     */
    SYNC = 'SYNC',

    /**
     * Import an existing changeset from the code host with the ExternalID from the spec.
     */
    IMPORT = 'IMPORT',

    /**
     * Close the changeset on the codehost.
     */
    CLOSE = 'CLOSE',

    /**
     * Reopen the changeset on the codehost.
     */
    REOPEN = 'REOPEN',

    /**
     * Internal operation to get around slow code host updates.
     */
    SLEEP = 'SLEEP',

    /**
     * The changeset is removed from some of the associated campaigns.
     */
    DETACH = 'DETACH',
}

/**
 * Description of the current changeset state vs the changeset spec desired state.
 */
export interface IChangesetSpecDelta {
    __typename: 'ChangesetSpecDelta'

    /**
     * When run, the title of the changeset will be updated.
     */
    titleChanged: boolean

    /**
     * When run, the body of the changeset will be updated.
     */
    bodyChanged: boolean

    /**
     * When run, the changeset will be taken out of draft mode.
     */
    undraft: boolean

    /**
     * When run, the target branch of the changeset will be updated.
     */
    baseRefChanged: boolean

    /**
     * When run, a new commit will be created on the branch of the changeset.
     */
    diffChanged: boolean

    /**
     * When run, a new commit will be created on the branch of the changeset.
     */
    commitMessageChanged: boolean

    /**
     * When run, a new commit in the name of the specified author will be created on the branch of the changeset.
     */
    authorNameChanged: boolean

    /**
     * When run, a new commit in the name of the specified author will be created on the branch of the changeset.
     */
    authorEmailChanged: boolean
}

/**
 * The type of the changeset spec.
 */
export enum ChangesetSpecType {
    /**
     * References an existing changeset on a code host to be imported.
     */
    EXISTING = 'EXISTING',

    /**
     * References a branch and a patch to be applied to create the changeset from.
     */
    BRANCH = 'BRANCH',
}

/**
 * A changeset spec is an immutable description of the desired state of a changeset in a campaign. To
 * create a changeset spec, use the createChangesetSpec mutation.
 */
export type ChangesetSpec = IHiddenChangesetSpec | IVisibleChangesetSpec

/**
 * A changeset spec is an immutable description of the desired state of a changeset in a campaign. To
 * create a changeset spec, use the createChangesetSpec mutation.
 */
export interface IChangesetSpec {
    __typename: 'ChangesetSpec'

    /**
     * The unique ID for a changeset spec.
     *
     * The ID is unguessable (i.e., long and randomly generated, not sequential). This is important
     * even though repository permissions also apply to viewers of changeset specs, because being
     * allowed to view a repository should not entitle a person to view all not-yet-published
     * changesets for that repository. Consider a campaign to fix a security vulnerability: the
     * campaign author may prefer to prepare all of the changesets in private so that the window
     * between revealing the problem and merging the fixes is as short as possible.
     */
    id: ID

    /**
     * The type of changeset spec.
     */
    type: ChangesetSpecType

    /**
     * The date, if any, when this changeset spec expires and is automatically purged. A changeset
     * spec never expires (and this field is null) if its campaign spec has been applied.
     */
    expiresAt: DateTime | null
}

/**
 * A changeset spec is an immutable description of the desired state of a changeset in a campaign. To
 * create a changeset spec, use the createChangesetSpec mutation.
 */
export interface IHiddenChangesetSpec {
    __typename: 'HiddenChangesetSpec'

    /**
     * The unique ID for a changeset spec.
     *
     * The ID is unguessable (i.e., long and randomly generated, not sequential). This is important
     * even though repository permissions also apply to viewers of changeset specs, because being
     * allowed to view a repository should not entitle a person to view all not-yet-published
     * changesets for that repository. Consider a campaign to fix a security vulnerability: the
     * campaign author may prefer to prepare all of the changesets in private so that the window
     * between revealing the problem and merging the fixes is as short as possible.
     */
    id: ID

    /**
     * The type of changeset spec.
     */
    type: ChangesetSpecType

    /**
     * The date, if any, when this changeset spec expires and is automatically purged. A changeset
     * spec never expires (and this field is null) if its campaign spec has been applied.
     */
    expiresAt: DateTime | null
}

/**
 * A changeset spec is an immutable description of the desired state of a changeset in a campaign. To
 * create a changeset spec, use the createChangesetSpec mutation.
 */
export interface IVisibleChangesetSpec {
    __typename: 'VisibleChangesetSpec'

    /**
     * The unique ID for a changeset spec.
     *
     * The ID is unguessable (i.e., long and randomly generated, not sequential). This is important
     * even though repository permissions also apply to viewers of changeset specs, because being
     * allowed to view a repository should not entitle a person to view all not-yet-published
     * changesets for that repository. Consider a campaign to fix a security vulnerability: the
     * campaign author may prefer to prepare all of the changesets in private so that the window
     * between revealing the problem and merging the fixes is as short as possible.
     */
    id: ID

    /**
     * The type of changeset spec.
     */
    type: ChangesetSpecType

    /**
     * The description of the changeset.
     */
    description: ChangesetDescription

    /**
     * The date, if any, when this changeset spec expires and is automatically purged. A changeset
     * spec never expires (and this field is null) if its campaign spec has been applied.
     */
    expiresAt: DateTime | null
}

/**
 * All possible types of changesets that can be specified in a changeset spec.
 */
export type ChangesetDescription = IExistingChangesetReference | IGitBranchChangesetDescription

/**
 * A reference to a changeset that already exists on a code host (and was not created by the
 * campaign).
 */
export interface IExistingChangesetReference {
    __typename: 'ExistingChangesetReference'

    /**
     * The repository that contains the existing changeset on the code host.
     */
    baseRepository: IRepository

    /**
     * The ID that uniquely identifies the existing changeset on the code host.
     *
     * For GitHub and Bitbucket Server, this is the pull request number (as a string) in the
     * base repository. For example, "1234" for PR 1234.
     */
    externalID: string
}

/**
 * A description of a changeset that represents the proposal to merge one branch into another.
 * This is used to describe a pull request (on GitHub and Bitbucket Server).
 */
export interface IGitBranchChangesetDescription {
    __typename: 'GitBranchChangesetDescription'

    /**
     * The repository that this changeset spec is proposing to change.
     */
    baseRepository: IRepository

    /**
     * The full name of the Git ref in the base repository that this changeset is based on (and is
     * proposing to be merged into). This ref must exist on the base repository. For example,
     * "refs/heads/master" or "refs/heads/main".
     */
    baseRef: string

    /**
     * The base revision this changeset is based on. It is the latest commit in
     * baseRef at the time when the changeset spec was created.
     * For example: "4095572721c6234cd72013fd49dff4fb48f0f8a4"
     */
    baseRev: string

    /**
     * The repository that contains the branch with this changeset's changes.
     *
     * Fork repositories and cross-repository changesets are not yet supported. Therefore,
     * headRepository must be equal to baseRepository.
     */
    headRepository: IRepository

    /**
     * The full name of the Git ref that holds the changes proposed by this changeset. This ref will
     * be created or updated with the commits. For example, "refs/heads/fix-foo" (for
     * the Git branch "fix-foo").
     */
    headRef: string

    /**
     * The title of the changeset on the code host.
     *
     * On Bitbucket Server or GitHub this is the title of the pull request.
     */
    title: string

    /**
     * The body of the changeset on the code host.
     *
     * On Bitbucket Server or GitHub this is the body/description of the pull request.
     */
    body: string

    /**
     * The Git commits with the proposed changes. These commits are pushed to the head ref.
     *
     * Only 1 commit is supported.
     */
    commits: IGitCommitDescription[]

    /**
     * The total diff of the changeset diff.
     */
    diff: IPreviewRepositoryComparison

    /**
     * The diffstat of this changeset spec. This data is also available
     * indirectly through the diff field above, but if only the diffStat is
     * required, this field is cheaper to access.
     */
    diffStat: IDiffStat

    /**
     * Whether or not the changeset described here should be created right after
     * applying the ChangesetSpec this description belongs to.
     *
     * If this is false, the changeset will only be created on Sourcegraph and
     * can be previewed.
     *
     * Another ChangesetSpec with the same description, but "published: true",
     * can later be applied publish the changeset.
     */
    published: any
}

/**
 * A description of a Git commit.
 */
export interface IGitCommitDescription {
    __typename: 'GitCommitDescription'

    /**
     * The full commit message.
     */
    message: string

    /**
     * The first line of the commit message.
     */
    subject: string

    /**
     * The contents of the commit message after the first line.
     */
    body: string | null

    /**
     * The Git commit author.
     */
    author: IPerson

    /**
     * The commit diff (in unified diff format).
     *
     * The filenames must not be prefixed (e.g., with 'a/' and 'b/'). Tip: use 'git diff --no-prefix'
     * to omit the prefix.
     */
    diff: string
}

/**
 * A list of changeset specs.
 */
export interface IChangesetSpecConnection {
    __typename: 'ChangesetSpecConnection'

    /**
     * The total number of changeset specs in the connection.
     */
    totalCount: number

    /**
     * Pagination information.
     */
    pageInfo: IPageInfo

    /**
     * A list of changeset specs.
     */
    nodes: ChangesetSpec[]
}

/**
 * A preview for which actions applyCampaign would result in when called at the point of time this preview was created at.
 */
export type ChangesetApplyPreview = IVisibleChangesetApplyPreview | IHiddenChangesetApplyPreview

/**
 * A preview entry to a repository to which the user has access.
 */
export type VisibleApplyPreviewTargets =
    | IVisibleApplyPreviewTargetsAttach
    | IVisibleApplyPreviewTargetsUpdate
    | IVisibleApplyPreviewTargetsDetach

/**
 * A preview entry where no changeset existed before matching the changeset spec.
 */
export interface IVisibleApplyPreviewTargetsAttach {
    __typename: 'VisibleApplyPreviewTargetsAttach'

    /**
     * The changeset spec from this entry.
     */
    changesetSpec: IVisibleChangesetSpec
}

/**
 * A preview entry where a changeset matches the changeset spec.
 */
export interface IVisibleApplyPreviewTargetsUpdate {
    __typename: 'VisibleApplyPreviewTargetsUpdate'

    /**
     * The changeset spec from this entry.
     */
    changesetSpec: IVisibleChangesetSpec

    /**
     * The changeset from this entry.
     */
    changeset: IExternalChangeset
}

/**
 * A preview entry where no changeset spec exists for the changeset currently in
 * the target campaign.
 */
export interface IVisibleApplyPreviewTargetsDetach {
    __typename: 'VisibleApplyPreviewTargetsDetach'

    /**
     * The changeset from this entry.
     */
    changeset: IExternalChangeset
}

/**
 * A preview entry to a repository to which the user has no access.
 */
export type HiddenApplyPreviewTargets =
    | IHiddenApplyPreviewTargetsAttach
    | IHiddenApplyPreviewTargetsUpdate
    | IHiddenApplyPreviewTargetsDetach

/**
 * A preview entry where no changeset existed before matching the changeset spec.
 */
export interface IHiddenApplyPreviewTargetsAttach {
    __typename: 'HiddenApplyPreviewTargetsAttach'

    /**
     * The changeset spec from this entry.
     */
    changesetSpec: IHiddenChangesetSpec
}

/**
 * A preview entry where a changeset matches the changeset spec.
 */
export interface IHiddenApplyPreviewTargetsUpdate {
    __typename: 'HiddenApplyPreviewTargetsUpdate'

    /**
     * The changeset spec from this entry.
     */
    changesetSpec: IHiddenChangesetSpec

    /**
     * The changeset from this entry.
     */
    changeset: IHiddenExternalChangeset
}

/**
 * A preview entry where no changeset spec exists for the changeset currently in
 * the target campaign.
 */
export interface IHiddenApplyPreviewTargetsDetach {
    __typename: 'HiddenApplyPreviewTargetsDetach'

    /**
     * The changeset from this entry.
     */
    changeset: IHiddenExternalChangeset
}

/**
 * One preview entry in the list of all previews against a campaign spec. Each mapping
 * between changeset specs and current changesets yields one of these. It describes
 * which operations are taken against which changeset spec and changeset to ensure the
 * desired state is met.
 */
export interface IHiddenChangesetApplyPreview {
    __typename: 'HiddenChangesetApplyPreview'

    /**
     * The operations to take to achieve the desired state.
     */
    operations: ChangesetSpecOperation[]

    /**
     * The delta between the current changeset state and what the new changeset spec
     * envisions the changeset to look like.
     */
    delta: IChangesetSpecDelta

    /**
     * The target entities in this preview entry.
     */
    targets: HiddenApplyPreviewTargets
}

/**
 * One preview entry in the list of all previews against a campaign spec. Each mapping
 * between changeset specs and current changesets yields one of these. It describes
 * which operations are taken against which changeset spec and changeset to ensure the
 * desired state is met.
 */
export interface IVisibleChangesetApplyPreview {
    __typename: 'VisibleChangesetApplyPreview'

    /**
     * The operations to take to achieve the desired state.
     */
    operations: ChangesetSpecOperation[]

    /**
     * The delta between the current changeset state and what the new changeset spec
     * envisions the changeset to look like.
     */
    delta: IChangesetSpecDelta

    /**
     * The target entities in this preview entry.
     */
    targets: VisibleApplyPreviewTargets
}

/**
 * Aggregated stats on nodes in this connection.
 */
export interface IChangesetApplyPreviewConnectionStats {
    __typename: 'ChangesetApplyPreviewConnectionStats'

    /**
     * Push a new commit to the code host.
     */
    push: number

    /**
     * Update the existing changeset on the codehost. This is purely the changeset resource on the code host,
     * not the git commit. For updates to the commit, see 'PUSH'.
     */
    update: number

    /**
     * Move the existing changeset out of being a draft.
     */
    undraft: number

    /**
     * Publish a changeset to the codehost.
     */
    publish: number

    /**
     * Publish a changeset to the codehost as a draft changeset. (Only on supported code hosts).
     */
    publishDraft: number

    /**
     * Sync the changeset with the current state on the codehost.
     */
    sync: number

    /**
     * Import an existing changeset from the code host with the ExternalID from the spec.
     */
    import: number

    /**
     * Close the changeset on the codehost.
     */
    close: number

    /**
     * Reopen the changeset on the codehost.
     */
    reopen: number

    /**
     * Internal operation to get around slow code host updates.
     */
    sleep: number

    /**
     * The changeset is removed from some of the associated campaigns.
     */
    detach: number

    /**
     * The amount of changesets that are added to the campaign in this operation.
     */
    added: number

    /**
     * The amount of changesets that are already attached to the campaign and modified in this operation.
     */
    modified: number

    /**
     * The amount of changesets that are disassociated from the campaign in this operation.
     */
    removed: number
}

/**
 * A list of preview entries.
 */
export interface IChangesetApplyPreviewConnection {
    __typename: 'ChangesetApplyPreviewConnection'

    /**
     * The total number of entries in the connection.
     */
    totalCount: number

    /**
     * Pagination information.
     */
    pageInfo: IPageInfo

    /**
     * A list of preview entries.
     */
    nodes: ChangesetApplyPreview[]

    /**
     * Stats on the elements in this connnection. Does not respect pagination parameters.
     */
    stats: IChangesetApplyPreviewConnectionStats
}

/**
 * A CampaignDescription describes a campaign.
 */
export interface ICampaignDescription {
    __typename: 'CampaignDescription'

    /**
     * The name as parsed from the input.
     */
    name: string

    /**
     * The description as parsed from the input.
     */
    description: string
}

/**
 * A campaign spec is an immutable description of the desired state of a campaign. To create a
 * campaign spec, use the createCampaignSpec mutation.
 */
export interface ICampaignSpec {
    __typename: 'CampaignSpec'

    /**
     * The unique ID for a campaign spec.
     *
     * The ID is unguessable (i.e., long and randomly generated, not sequential).
     * Consider a campaign to fix a security vulnerability: the campaign author may prefer
     * to prepare the campaign, including the description in private so that the window
     * between revealing the problem and merging the fixes is as short as possible.
     */
    id: ID

    /**
     * The original YAML or JSON input that was used to create this campaign spec.
     */
    originalInput: string

    /**
     * The parsed JSON value of the original input. If the original input was YAML, the YAML is
     * converted to the equivalent JSON.
     */
    parsedInput: any

    /**
     * The CampaignDescription that describes this campaign.
     */
    description: ICampaignDescription

    /**
     * Generates a preview what operations would be performed if the campaign spec would be applied.
     * This preview is not a guarantee, since the state of the changesets can change between the time
     * the preview is generated and when the campaign spec is applied.
     */
    applyPreview: IChangesetApplyPreviewConnection

    /**
     * The specs for changesets associated with this campaign.
     */
    changesetSpecs: IChangesetSpecConnection

    /**
     * The user who created this campaign spec (or null if the user no longer exists).
     */
    creator: IUser | null

    /**
     * The date when this campaign spec was created.
     */
    createdAt: DateTime

    /**
     * The namespace (either a user or organization) of the campaign spec.
     */
    namespace: Namespace

    /**
     * The date, if any, when this campaign spec expires and is automatically purged. A campaign spec
     * never expires if it has been applied.
     */
    expiresAt: DateTime | null

    /**
     * The URL of a web page that allows applying this campaign spec and
     * displays a preview of which changesets will be created by applying it.
     */
    applyURL: string

    /**
     * When true, the viewing user can apply this spec.
     */
    viewerCanAdminister: boolean

    /**
     * The diff stat for all the changeset specs in the campaign spec.
     */
    diffStat: IDiffStat

    /**
     * The campaign this spec will update when applied. If it's null, the
     * campaign doesn't yet exist.
     */
    appliesToCampaign: ICampaign | null

    /**
     * The newest version of this campaign spec, as identified by its namespace
     * and name. If this is the newest version, this field will be null.
     */
    supersedingCampaignSpec: ICampaignSpec | null

    /**
     * The code host connections required for applying this spec. Includes the credentials of the current user.
     */
    viewerCampaignsCodeHosts: ICampaignsCodeHostConnection
}

export interface IApplyPreviewOnCampaignSpecArguments {
    /**
     * Returns the first n entries from the list.
     * @default 50
     */
    first?: number | null

    /**
     * Opaque pagination cursor.
     */
    after?: string | null

    /**
     * Search for changesets matching this query. Queries may include quoted substrings to match phrases, and words may be preceded by - to negate them.
     */
    search?: string | null

    /**
     * Search for changesets that are currently in this state.
     */
    currentState?: ChangesetState | null

    /**
     * Search for changesets that will have the given action performed.
     */
    action?: ChangesetSpecOperation | null
}

export interface IChangesetSpecsOnCampaignSpecArguments {
    /**
     * @default 50
     */
    first?: number | null
    after?: string | null
}

export interface IViewerCampaignsCodeHostsOnCampaignSpecArguments {
    /**
     * Returns the first n code hosts from the list.
     * @default 50
     */
    first?: number | null

    /**
     * Opaque pagination cursor.
     */
    after?: string | null

    /**
     * Only returns the code hosts for which the viewer doesn't have credentials.
     * @default false
     */
    onlyWithoutCredential?: boolean | null
}

/**
 * The state of the batch change.
 */
export enum BatchChangeState {
    OPEN = 'OPEN',
    CLOSED = 'CLOSED',
}

/**
 * A batch change is a set of related changes to apply to code across one or more repositories.
 */
export interface IBatchChange {
    __typename: 'BatchChange'

    /**
     * The unique ID for the batch change.
     */
    id: ID

    /**
     * The namespace where this batch change is defined.
     */
    namespace: Namespace

    /**
     * The name of the batch change.
     */
    name: string

    /**
     * The description (as Markdown).
     */
    description: string | null

    /**
     * The user that created the initial spec. In an org, this will be different from the namespace, or null if the user was deleted.
     */
    specCreator: IUser | null

    /**
     * The user who created the batch change initially by applying the spec for the first time, or null if the user was deleted.
     */
    initialApplier: IUser | null

    /**
     * The user who last updated the batch change by applying a spec to this batch change.
     * If the batch change hasn't been updated, the lastApplier is the initialApplier, or null if the user was deleted.
     */
    lastApplier: IUser | null

    /**
     * Whether the current user can edit or delete this batch change.
     */
    viewerCanAdminister: boolean

    /**
     * The URL to this batch change.
     */
    url: string

    /**
     * The date and time when the batch change was created.
     */
    createdAt: DateTime

    /**
     * The date and time when the batch change was updated. That can be by applying a spec, or by an internal process.
     * For reading the time the batch change spec was changed last, see lastAppliedAt.
     */
    updatedAt: DateTime

    /**
     * The date and time when the batch change was last updated with a new spec.
     */
    lastAppliedAt: DateTime

    /**
     * The date and time when the batch change was closed. If set, applying a spec for this batch change will fail with an error.
     */
    closedAt: DateTime | null

    /**
     * Stats on all the changesets that are tracked in this batch change.
     */
    changesetsStats: IChangesetsStats

    /**
     * The changesets in this batch change that already exist on the code host.
     */
    changesets: IChangesetConnection

    /**
     * The changeset counts over time, in 1-day intervals backwards from the point in time given in
     * the "to" parameter.
     */
    changesetCountsOverTime: IChangesetCounts[]

    /**
     * The diff stat for all the changesets in the batch change.
     */
    diffStat: IDiffStat

    /**
     * The current change spec this change reflects.
     */
    currentSpec: IBatchSpec
}

export interface IChangesetsOnBatchChangeArguments {
    /**
     * @default 50
     */
    first?: number | null

    /**
     * Opaque pagination cursor.
     */
    after?: string | null

    /**
     * Only include changesets with the given state.
     */
    state?: ChangesetState | null

    /**
     * Only include changesets with the given review state.
     */
    reviewState?: ChangesetReviewState | null

    /**
     * Only include changesets with the given check state.
     */
    checkState?: ChangesetCheckState | null

    /**
     * Only return changesets that have been published by this batch change. Imported changesets will be omitted.
     */
    onlyPublishedByThisBatchChange?: boolean | null

    /**
     * Search for changesets matching this query. Queries may include quoted substrings to match phrases, and words may be preceded by - to negate them.
     */
    search?: string | null
}

export interface IChangesetCountsOverTimeOnBatchChangeArguments {
    /**
     * Only include changeset counts up to this point in time (inclusive). Defaults to BatchChange.createdAt.
     */
    from?: DateTime | null

    /**
     * Only include changeset counts up to this point in time (inclusive). Defaults to the
     * current time.
     */
    to?: DateTime | null
}

/**
 * A batch spec is an immutable description of the desired state of a campaign. To create a
 * batch spec, use the createBatchSpec mutation.
 */
export interface IBatchSpec {
    __typename: 'BatchSpec'

    /**
     * The unique ID for a batch spec.
     *
     * The ID is unguessable (i.e., long and randomly generated, not sequential).
     * Consider a campaign to fix a security vulnerability: the campaign author may prefer
     * to prepare the campaign, including the description in private so that the window
     * between revealing the problem and merging the fixes is as short as possible.
     */
    id: ID

    /**
     * The original YAML or JSON input that was used to create this batch spec.
     */
    originalInput: string

    /**
     * The parsed JSON value of the original input. If the original input was YAML, the YAML is
     * converted to the equivalent JSON.
     */
    parsedInput: any

    /**
     * The BatchChangeDescription that describes this batch change.
     */
    description: IBatchChangeDescription

    /**
     * Generates a preview what operations would be performed if the batch spec would be applied.
     * This preview is not a guarantee, since the state of the changesets can change between the time
     * the preview is generated and when the batch spec is applied.
     */
    applyPreview: IChangesetApplyPreviewConnection

    /**
     * The specs for changesets associated with this campaign.
     */
    changesetSpecs: IChangesetSpecConnection

    /**
     * The user who created this batch spec (or null if the user no longer exists).
     */
    creator: IUser | null

    /**
     * The date when this batch spec was created.
     */
    createdAt: DateTime

    /**
     * The namespace (either a user or organization) of the batch spec.
     */
    namespace: Namespace

    /**
     * The date, if any, when this batch spec expires and is automatically purged. A batch spec
     * never expires if it has been applied.
     */
    expiresAt: DateTime | null

    /**
     * The URL of a web page that allows applying this batch spec and
     * displays a preview of which changesets will be created by applying it.
     */
    applyURL: string

    /**
     * When true, the viewing user can apply this spec.
     */
    viewerCanAdminister: boolean

    /**
     * The diff stat for all the changeset specs in the batch spec.
     */
    diffStat: IDiffStat

    /**
     * The campaign this spec will update when applied. If it's null, the
     * campaign doesn't yet exist.
     */
    appliesToBatchChange: IBatchChange | null

    /**
     * The newest version of this batch spec, as identified by its namespace
     * and name. If this is the newest version, this field will be null.
     */
    supersedingBatchSpec: IBatchSpec | null

    /**
     * The code host connections required for applying this spec. Includes the credentials of the current user.
     */
    viewerBatchChangesCodeHosts: IBatchChangesCodeHostConnection
}

export interface IApplyPreviewOnBatchSpecArguments {
    /**
     * Returns the first n entries from the list.
     * @default 50
     */
    first?: number | null

    /**
     * Opaque pagination cursor.
     */
    after?: string | null

    /**
     * Search for changesets matching this query. Queries may include quoted substrings to match phrases, and words may be preceded by - to negate them.
     */
    search?: string | null

    /**
     * Search for changesets that are currently in this state.
     */
    currentState?: ChangesetState | null

    /**
     * Search for changesets that will have the given action performed.
     */
    action?: ChangesetSpecOperation | null
}

export interface IChangesetSpecsOnBatchSpecArguments {
    /**
     * @default 50
     */
    first?: number | null
    after?: string | null
}

export interface IViewerBatchChangesCodeHostsOnBatchSpecArguments {
    /**
     * Returns the first n code hosts from the list.
     * @default 50
     */
    first?: number | null

    /**
     * Opaque pagination cursor.
     */
    after?: string | null

    /**
     * Only returns the code hosts for which the viewer doesn't have credentials.
     * @default false
     */
    onlyWithoutCredential?: boolean | null
}

/**
 * A list of batch changes.
 */
export interface IBatchChangeConnection {
    __typename: 'BatchChangeConnection'

    /**
     * A list of batch changes.
     */
    nodes: IBatchChange[]

    /**
     * The total number of batch changes in the connection.
     */
    totalCount: number

    /**
     * Pagination information.
     */
    pageInfo: IPageInfo
}

/**
 * A connection of all code hosts usable with batch changes and accessible by the user
 * this is requested on.
 */
export interface IBatchChangesCodeHostConnection {
    __typename: 'BatchChangesCodeHostConnection'

    /**
     * A list of code hosts.
     */
    nodes: IBatchChangesCodeHost[]

    /**
     * The total number of configured external services in the connection.
     */
    totalCount: number

    /**
     * Pagination information.
     */
    pageInfo: IPageInfo
}

/**
 * A code host usable with batch changes. This service is accessible by the user it belongs to.
 */
export interface IBatchChangesCodeHost {
    __typename: 'BatchChangesCodeHost'

    /**
     * The kind of external service.
     */
    externalServiceKind: ExternalServiceKind

    /**
     * The URL of the external service.
     */
    externalServiceURL: string

    /**
     * The configured credential, if any.
     */
    credential: IBatchChangesCredential | null

    /**
     * If true, some of the repositories on this code host require
     * an SSH key to be configured.
     */
    requiresSSH: boolean
}

/**
 * A user token configured for batch changes use on the specified code host.
 */
export interface IBatchChangesCredential {
    __typename: 'BatchChangesCredential'

    /**
     * A globally unique identifier.
     */
    id: ID

    /**
     * The kind of external service.
     */
    externalServiceKind: ExternalServiceKind

    /**
     * The URL of the external service.
     */
    externalServiceURL: string

    /**
     * The public key to use on the external service for SSH keypair authentication.
     * Not set if the credential doesn't support SSH access.
     */
    sshPublicKey: string | null

    /**
     * The date and time this token has been created at.
     */
    createdAt: DateTime
}

/**
 * A BatchChangeDescription describes a batch change.
 */
export interface IBatchChangeDescription {
    __typename: 'BatchChangeDescription'

    /**
     * The name as parsed from the input.
     */
    name: string

    /**
     * The description as parsed from the input.
     */
    description: string
}
