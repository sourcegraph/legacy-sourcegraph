package db

var (
	AccessTokens     = &AccessTokenStore{}
	ExternalServices = &ExternalServiceStore{}
	DefaultRepos     = &DefaultRepoStore{}
	Repos            = &RepoStore{}
	Phabricator      = &phabricator{}
	QueryRunnerState = &QueryRunnerStateStore{}
	Namespaces       = &NamespaceStore{}
	Orgs             = &OrgStore{}
	OrgMembers       = &orgMembers{}
	SavedSearches    = &savedSearches{}
	Settings         = &settings{}
	Users            = &UserStore{}
	UserCredentials  = &userCredentials{}
	UserEmails       = &userEmails{}
	EventLogs        = &eventLogs{}

	SurveyResponses = &surveyResponses{}

	ExternalAccounts = &userExternalAccounts{}

	OrgInvitations = &orgInvitations{}

	Authz AuthzStore = &authzStore{}

	Secrets = &secrets{}
)
