package db

var (
	AccessTokens     = &accessTokens{}
	ExternalServices = &ExternalServiceStore{}
	DefaultRepos     = &defaultRepos{}
	Repos            = &RepoStore{}
	Phabricator      = &PhabricatorStore{}
	QueryRunnerState = &queryRunnerState{}
	Namespaces       = &namespaces{}
	Orgs             = &orgs{}
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
