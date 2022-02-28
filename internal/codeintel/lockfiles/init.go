package lockfiles

import (
	"context"
	"io"
	"sync"

	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

var (
	depSvc     *Service
	depSvcOnce sync.Once
)

func GetService(
	checker authz.SubRepoPermissionChecker,
	lsFiles func(context.Context, authz.SubRepoPermissionChecker, api.RepoName, api.CommitID, ...string) ([]string, error),
	archive func(context.Context, api.RepoName, gitserver.ArchiveOptions) (io.ReadCloser, error),
) *Service {
	depSvcOnce.Do(func() {
		observationContext := &observation.Context{
			Logger:     log15.Root(),
			Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
			Registerer: prometheus.DefaultRegisterer,
		}

		depSvc = newService(checker, lsFiles, archive, observationContext)
	})

	return depSvc
}

func TestService(
	checker authz.SubRepoPermissionChecker,
	lsFiles func(context.Context, authz.SubRepoPermissionChecker, api.RepoName, api.CommitID, ...string) ([]string, error),
	archive func(context.Context, api.RepoName, gitserver.ArchiveOptions) (io.ReadCloser, error),
) *Service {
	return newService(checker, lsFiles, archive, &observation.TestContext)
}
