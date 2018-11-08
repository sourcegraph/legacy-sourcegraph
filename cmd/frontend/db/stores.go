package db

import (
	"database/sql"

	"github.com/sourcegraph/sourcegraph/pkg/conf/confdb"
	"github.com/sourcegraph/sourcegraph/pkg/dbconn"
)

var (
	AccessTokens               = &accessTokens{}
	DiscussionThreads          = &discussionThreads{}
	DiscussionComments         = &discussionComments{}
	DiscussionMailReplyTokens  = &discussionMailReplyTokens{}
	Repos                      = &repos{}
	Phabricator                = &phabricator{}
	SavedQueries               = &savedQueries{}
	Orgs                       = &orgs{}
	OrgMembers                 = &orgMembers{}
	Settings                   = &settings{}
	Users                      = &users{}
	UserEmails                 = &userEmails{}
	SiteIDInfo                 = &siteIDInfo{}
	CertCache                  = &certCache{}
	CoreSiteConfigurationFiles = &confdb.CoreSiteConfigurationFiles{Conn: func() *sql.DB { return dbconn.Global }}

	SurveyResponses = &surveyResponses{}

	ExternalAccounts = &userExternalAccounts{}

	OrgInvitations = &orgInvitations{}

	// GlobalDeps is a stub implementation of a global dependency index
	GlobalDeps GlobalDepsProvider = &globalDeps{}

	// Pkgs is a stub implementation of a global package metadata index
	Pkgs PkgsProvider = &pkgs{}
)
