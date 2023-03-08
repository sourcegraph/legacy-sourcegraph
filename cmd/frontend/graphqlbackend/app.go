package graphqlbackend

import (
	"context"
	"github.com/graph-gophers/graphql-go"

	"path/filepath"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/internal/service/servegit"
	"github.com/sourcegraph/sourcegraph/internal/singleprogram/filepicker"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

type LocalDirectoryArgs struct {
	Dir string
}

type AppResolver interface {
	LocalDirectoryPicker(ctx context.Context) (LocalDirectoryResolver, error)
	LocalDirectory(ctx context.Context, args *LocalDirectoryArgs) (LocalDirectoryResolver, error)
	LocalExternalServices(ctx context.Context) ([]LocalExternalServiceResolver, error)
}

type LocalDirectoryResolver interface {
	Path() string
	Repositories() ([]LocalRepositoryResolver, error)
}

type LocalRepositoryResolver interface {
	Name() string
	Path() string
}

type LocalExternalServiceResolver interface {
	ID() graphql.ID
	Path() string
	Autogenerated() bool
}

type appResolver struct {
	logger log.Logger
	db     database.DB
}

var _ AppResolver = &appResolver{}

func NewAppResolver(logger log.Logger, db database.DB) *appResolver {
	return &appResolver{
		logger: logger,
		db:     db,
	}
}

func (r *appResolver) checkLocalDirectoryAccess(ctx context.Context) error {
	return auth.CheckCurrentUserIsSiteAdmin(ctx, r.db)
}

func (r *appResolver) LocalDirectoryPicker(ctx context.Context) (LocalDirectoryResolver, error) {
	// 🚨 SECURITY: Only site admins on app may use API which accesses local filesystem.
	if err := r.checkLocalDirectoryAccess(ctx); err != nil {
		return nil, err
	}

	picker, ok := filepicker.Lookup(r.logger)
	if !ok {
		return nil, errors.New("filepicker is not available")
	}

	path, err := picker(ctx)
	if err != nil {
		return nil, err
	}

	return &localDirectoryResolver{path: path}, nil
}

func (r *appResolver) LocalDirectory(ctx context.Context, args *LocalDirectoryArgs) (LocalDirectoryResolver, error) {
	// 🚨 SECURITY: Only site admins on app may use API which accesses local filesystem.
	if err := r.checkLocalDirectoryAccess(ctx); err != nil {
		return nil, err
	}

	path, err := filepath.Abs(args.Dir)
	if err != nil {
		return nil, err
	}

	return &localDirectoryResolver{path: path}, nil
}

type localDirectoryResolver struct {
	path string
}

func (r *localDirectoryResolver) Path() string {
	return r.path
}

func (r *localDirectoryResolver) Repositories() ([]LocalRepositoryResolver, error) {
	// TODO(keegan) this should be injected from the global instance. For now
	// we are hardcoding the relevant defaults for ServeConfig.
	srv := &servegit.Serve{
		ServeConfig: servegit.ServeConfig{
			Timeout:  5 * time.Second,
			MaxDepth: 10,
		},
		Logger: log.Scoped("serve", ""),
	}

	repos, err := srv.Repos(r.path)
	if err != nil {
		return nil, err
	}

	local := make([]LocalRepositoryResolver, 0, len(repos))
	for _, repo := range repos {
		local = append(local, localRepositoryResolver{
			name: repo.Name,
			path: repo.AbsFilePath,
		})
	}

	return local, nil
}

type localRepositoryResolver struct {
	name string
	path string
}

func (r localRepositoryResolver) Name() string {
	return r.name
}

func (r localRepositoryResolver) Path() string {
	return r.path
}

func (r *appResolver) LocalExternalServices(ctx context.Context) ([]LocalExternalServiceResolver, error) {
	// 🚨 SECURITY: Only site admins on app may use API which accesses local filesystem.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	opt := database.ExternalServicesListOptions{
		Kinds: []string{extsvc.KindOther},
	}

	externalServices, err := r.db.ExternalServices().List(ctx, opt)

	if err != nil {
		return nil, err
	}

	localExternalServices := make([]LocalExternalServiceResolver, 0)
	for _, externalService := range externalServices {
		serviceConfig, err := externalService.Config.Decrypt(ctx)
		if err != nil {
			return nil, err
		}

		var otherConfig schema.OtherExternalServiceConnection
		if err = jsonc.Unmarshal(serviceConfig, &otherConfig); err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal service config JSON")
		}

		if len(otherConfig.Repos) == 1 && otherConfig.Repos[0] == "src-serve-local" {
			// Sourcegraph App Upserts() an external service with ID of ExtSVCID to the database and we
			// distinguish this in our returned results to discern which external services should not be modified
			// by users
			isAppAutogenerated := externalService.ID == servegit.ExtSVCID

			localExtSvc := localExternalServiceResolver{
				id:            MarshalExternalServiceID(externalService.ID),
				path:          otherConfig.Root,
				autogenerated: isAppAutogenerated,
			}
			localExternalServices = append(localExternalServices, localExtSvc)
		}
	}

	return localExternalServices, nil
}

type localExternalServiceResolver struct {
	id            graphql.ID
	path          string
	autogenerated bool
}

func (r localExternalServiceResolver) ID() graphql.ID {
	return r.id
}

func (r localExternalServiceResolver) Path() string {
	return r.path
}

func (r localExternalServiceResolver) Autogenerated() bool {
	return r.autogenerated
}
