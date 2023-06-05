package store

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/keegancsmith/sqlf"

	ghtypes "github.com/sourcegraph/sourcegraph/enterprise/internal/github_apps/types"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	encryption "github.com/sourcegraph/sourcegraph/internal/encryption"
	"github.com/sourcegraph/sourcegraph/internal/encryption/keyring"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	itypes "github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// GitHubAppsStore handles storing and retrieving GitHub Apps from the database.
type GitHubAppsStore interface {
	// Create inserts a new GitHub App into the database.
	Create(ctx context.Context, app *ghtypes.GitHubApp) (int, error)

	// Delete removes a GitHub App from the database by ID.
	Delete(ctx context.Context, id int) error

	// Update updates a GitHub App in the database and returns the updated struct.
	Update(ctx context.Context, id int, app *ghtypes.GitHubApp) (*ghtypes.GitHubApp, error)

	// Install creates a new GitHub App installation in the database.
	Install(ctx context.Context, id, installationID int) error

	// GetLatestInstallID retrieves the latest GitHub App installation ID from the
	// database for the GitHub App with the provided appID.
	GetLatestInstallID(ctx context.Context, appID int) (int, error)

	// GetByID retrieves a GitHub App from the database by ID.
	GetByID(ctx context.Context, id int) (*ghtypes.GitHubApp, error)

	// GetByAppID retrieves a GitHub App from the database by appID and base url
	GetByAppID(ctx context.Context, appID int, baseURL string) (*ghtypes.GitHubApp, error)

	// GetBySlug retrieves a GitHub App from the database by slug and base url
	GetBySlug(ctx context.Context, slug string, baseURL string) (*ghtypes.GitHubApp, error)

	// GetByDomain retrieves a GitHub App from the database by domain and base url
	GetByDomain(ctx context.Context, domain itypes.GitHubAppDomain, baseURL string) (*ghtypes.GitHubApp, error)

	// WithEncryptionKey sets encryption key on store. Returns a new GitHubAppsStore
	WithEncryptionKey(key encryption.Key) GitHubAppsStore

	// Lists all GitHub Apps in the store and optionally filters by domain
	List(ctx context.Context, domain *itypes.GitHubAppDomain) ([]*ghtypes.GitHubApp, error)
}

// gitHubAppStore handles storing and retrieving GitHub Apps from the database.
type gitHubAppsStore struct {
	*basestore.Store

	key encryption.Key
}

func GitHubAppsWith(other *basestore.Store) GitHubAppsStore {
	return &gitHubAppsStore{
		Store: basestore.NewWithHandle(other.Handle()),
	}
}

// WithEncryptionKey sets encryption key on store. Returns a new GitHubAppsStore
func (s *gitHubAppsStore) WithEncryptionKey(key encryption.Key) GitHubAppsStore {
	return &gitHubAppsStore{Store: s.Store, key: key}
}

func (s *gitHubAppsStore) getEncryptionKey() encryption.Key {
	if s.key != nil {
		return s.key
	}
	return keyring.Default().GitHubAppKey
}

var scanIDAndTimes = basestore.NewFirstScanner(func(s dbutil.Scanner) (*ghtypes.GitHubApp, error) {
	var app ghtypes.GitHubApp

	err := s.Scan(
		&app.ID,
		&app.CreatedAt,
		&app.UpdatedAt)
	return &app, err
})

// Create inserts a new GitHub App into the database. The default domain for the App is "repos".
func (s *gitHubAppsStore) Create(ctx context.Context, app *ghtypes.GitHubApp) (int, error) {
	key := s.getEncryptionKey()
	clientSecret, _, err := encryption.MaybeEncrypt(ctx, key, app.ClientSecret)
	if err != nil {
		return -1, err
	}
	privateKey, keyID, err := encryption.MaybeEncrypt(ctx, key, app.PrivateKey)
	if err != nil {
		return -1, err
	}

	baseURL, err := url.Parse(app.BaseURL)
	if err != nil {
		return -1, errors.New(fmt.Sprintf("unable to parse base URL: %s", baseURL.String()))
	}
	baseURL = extsvc.NormalizeBaseURL(baseURL)
	domain := app.Domain
	if domain == "" {
		domain = itypes.ReposGitHubAppDomain
	}

	// We enforce that GitHub Apps created in the "batches" domain are for unique instance URLs.
	if domain == itypes.BatchesGitHubAppDomain {
		existingGHApp, err := s.GetByDomain(ctx, domain, baseURL.String())
		// An error is expected if no existing app was found, but we double check that
		// we didn't get a different, unrelated error
		if err != nil && !strings.Contains(err.Error(), "no app exists matching criteria") {
			return -1, errors.Wrap(err, "checking for existing batches app")
		}
		if existingGHApp != nil {
			return -1, errors.New("GitHub App already exists for this GitHub instance in the batches domain")
		}
	}

	query := sqlf.Sprintf(`INSERT INTO
	    github_apps (app_id, name, domain, slug, base_url, app_url, client_id, client_secret, private_key, encryption_key_id, logo)
    	VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
		RETURNING id`,
		app.AppID, app.Name, domain, app.Slug, baseURL.String(), app.AppURL, app.ClientID, clientSecret, privateKey, keyID, app.Logo)
	id, _, err := basestore.ScanFirstInt(s.Query(ctx, query))
	return id, err
}

// Delete removes a GitHub App from the database by ID.
func (s *gitHubAppsStore) Delete(ctx context.Context, id int) error {
	query := sqlf.Sprintf(`DELETE FROM github_apps WHERE id = %s`, id)
	return s.Exec(ctx, query)
}

func scanGitHubApp(s dbutil.Scanner) (*ghtypes.GitHubApp, error) {
	var app ghtypes.GitHubApp

	err := s.Scan(
		&app.ID,
		&app.AppID,
		&app.Name,
		&app.Domain,
		&app.Slug,
		&app.BaseURL,
		&app.AppURL,
		&app.ClientID,
		&app.ClientSecret,
		&app.WebhookID,
		&app.PrivateKey,
		&app.EncryptionKey,
		&app.Logo,
		&app.CreatedAt,
		&app.UpdatedAt)
	return &app, err
}

var (
	scanGitHubApps     = basestore.NewSliceScanner(scanGitHubApp)
	scanFirstGitHubApp = basestore.NewFirstScanner(scanGitHubApp)
)

func (s *gitHubAppsStore) decrypt(ctx context.Context, apps ...*ghtypes.GitHubApp) ([]*ghtypes.GitHubApp, error) {
	key := s.getEncryptionKey()

	for _, app := range apps {
		cs, err := encryption.MaybeDecrypt(ctx, key, app.ClientSecret, app.EncryptionKey)
		if err != nil {
			return nil, err
		}
		app.ClientSecret = cs
		pk, err := encryption.MaybeDecrypt(ctx, key, app.PrivateKey, app.EncryptionKey)
		if err != nil {
			return nil, err
		}
		app.PrivateKey = pk
	}

	return apps, nil
}

// Update updates a GitHub App in the database and returns the updated struct.
func (s *gitHubAppsStore) Update(ctx context.Context, id int, app *ghtypes.GitHubApp) (*ghtypes.GitHubApp, error) {
	key := s.getEncryptionKey()
	clientSecret, _, err := encryption.MaybeEncrypt(ctx, key, app.ClientSecret)
	if err != nil {
		return nil, err
	}
	privateKey, keyID, err := encryption.MaybeEncrypt(ctx, key, app.PrivateKey)
	if err != nil {
		return nil, err
	}

	query := sqlf.Sprintf(`UPDATE github_apps
             SET app_id = %s, name = %s, domain = %s, slug = %s, base_url = %s, app_url = %s, client_id = %s, client_secret = %s, webhook_id = %d, private_key = %s, encryption_key_id = %s, logo = %s, updated_at = NOW()
             WHERE id = %s
			 RETURNING id, app_id, name, domain, slug, base_url, app_url, client_id, client_secret, webhook_id, private_key, encryption_key_id, logo, created_at, updated_at`,
		app.AppID, app.Name, app.Domain, app.Slug, app.BaseURL, app.AppURL, app.ClientID, clientSecret, app.WebhookID, privateKey, keyID, app.Logo, id)
	app, ok, err := scanFirstGitHubApp(s.Query(ctx, query))
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.Newf("cannot update app with id: %d because no such app exists", id)
	}
	apps, err := s.decrypt(ctx, app)
	if err != nil {
		return nil, err
	}
	return apps[0], nil
}

// Install creates a new GitHub App installation in the database.
func (s *gitHubAppsStore) Install(ctx context.Context, id, installationID int) error {
	query := sqlf.Sprintf(`
		INSERT INTO github_app_installs (app_id, installation_id)
    	VALUES (%s, %s)
		ON CONFLICT DO NOTHING
		RETURNING id`,
		id, installationID)
	return s.Exec(ctx, query)
}

func (s *gitHubAppsStore) GetLatestInstallID(ctx context.Context, appID int) (int, error) {
	query := sqlf.Sprintf(`
		SELECT installation_id
		FROM github_app_installs
		JOIN github_apps ON github_app_installs.app_id = github_apps.id
		WHERE github_apps.app_id = %s
		ORDER BY github_app_installs.id DESC LIMIT 1
		`, appID)
	installID, _, err := basestore.ScanFirstInt(s.Query(ctx, query))
	return installID, err
}

func (s *gitHubAppsStore) get(ctx context.Context, where *sqlf.Query) (*ghtypes.GitHubApp, error) {
	selectQuery := `SELECT
		id,
		app_id,
		name,
		domain,
		slug,
		base_url,
		app_url,
		client_id,
		client_secret,
		webhook_id,
		private_key,
		encryption_key_id,
		logo,
		created_at,
		updated_at
	FROM github_apps
	WHERE %s`

	query := sqlf.Sprintf(selectQuery, where)
	app, ok, err := scanFirstGitHubApp(s.Query(ctx, query))
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.Newf("no app exists matching criteria: %v", *where)
	}

	apps, err := s.decrypt(ctx, app)
	if err != nil {
		return nil, err
	}
	return apps[0], nil
}

func (s *gitHubAppsStore) list(ctx context.Context, where *sqlf.Query) ([]*ghtypes.GitHubApp, error) {
	selectQuery := `SELECT
		id,
		app_id,
		name,
		domain,
		slug,
		base_url,
		app_url,
		client_id,
		client_secret,
		webhook_id,
		private_key,
		encryption_key_id,
		logo,
		created_at,
		updated_at
	FROM github_apps
	WHERE %s`

	query := sqlf.Sprintf(selectQuery, where)
	apps, err := scanGitHubApps(s.Query(ctx, query))
	if err != nil {
		return nil, err
	}

	return s.decrypt(ctx, apps...)
}

// GetByID retrieves a GitHub App from the database by ID.
func (s *gitHubAppsStore) GetByID(ctx context.Context, id int) (*ghtypes.GitHubApp, error) {
	return s.get(ctx, sqlf.Sprintf(`id = %s`, id))
}

// GetByAppID retrieves a GitHub App from the database by appID and base url
func (s *gitHubAppsStore) GetByAppID(ctx context.Context, appID int, baseURL string) (*ghtypes.GitHubApp, error) {
	return s.get(ctx, sqlf.Sprintf(`app_id = %s AND base_url = %s`, appID, baseURL))
}

// GetBySlug retrieves a GitHub App from the database by slug and base url
func (s *gitHubAppsStore) GetBySlug(ctx context.Context, slug string, baseURL string) (*ghtypes.GitHubApp, error) {
	return s.get(ctx, sqlf.Sprintf(`slug = %s AND base_url = %s`, slug, baseURL))
}

// GetByDomain retrieves a GitHub App from the database by domain and base url
func (s *gitHubAppsStore) GetByDomain(ctx context.Context, domain itypes.GitHubAppDomain, baseURL string) (*ghtypes.GitHubApp, error) {
	return s.get(ctx, sqlf.Sprintf(`domain = %s AND base_url = %s`, domain, baseURL))
}

// List lists all GitHub Apps in the store
func (s *gitHubAppsStore) List(ctx context.Context, domain *itypes.GitHubAppDomain) ([]*ghtypes.GitHubApp, error) {
	where := sqlf.Sprintf(`true`)
	if domain != nil {
		where = sqlf.Sprintf("domain = %s", *domain)
	}
	return s.list(ctx, where)
}
