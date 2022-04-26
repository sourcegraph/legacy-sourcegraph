// Command searcher is a simple service which exposes an API to text search a
// repo at a specific commit. See the searcher package for more information.
package main

import (
	"context"
	"io"
	stdlog "log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/searcher/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	connections "github.com/sourcegraph/sourcegraph/internal/database/connections/live"
	"github.com/sourcegraph/sourcegraph/internal/debugserver"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/hostname"
	"github.com/sourcegraph/sourcegraph/internal/logging"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/profiler"
	"github.com/sourcegraph/sourcegraph/internal/sentry"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
	"github.com/sourcegraph/sourcegraph/internal/tracer"
	"github.com/sourcegraph/sourcegraph/internal/version"
	"github.com/sourcegraph/sourcegraph/lib/log"
)

var (
	cacheDir    = env.Get("CACHE_DIR", "/tmp", "directory to store cached archives.")
	cacheSizeMB = env.Get("SEARCHER_CACHE_SIZE_MB", "100000", "maximum size of the on disk cache in megabytes")
)

const port = "3181"

func ensureFrontendDB(logger log.Logger) database.DB {
	dsn := conf.GetServiceConnectionValueAndRestartOnChange(func(serviceConnections conftypes.ServiceConnections) string {
		return serviceConnections.PostgresDSN
	})
	sqlDB, err := connections.EnsureNewFrontendDB(dsn, "searcher", &observation.TestContext)
	if err != nil {
		logger.Fatal("Failed to connect to frontend database", log.Error(err))
	}
	return database.NewDB(sqlDB), nil
}

func main() {
	env.Lock()
	env.HandleHelpFlag()
	stdlog.SetFlags(0)
	conf.Init()
	logging.Init()
	syncLogs := log.Init(log.Resource{
		Name:       env.MyName,
		Version:    version.Version(),
		InstanceID: hostname.Get(),
	})
	defer syncLogs()
	tracer.Init(conf.DefaultClient())
	sentry.Init(conf.DefaultClient())
	trace.Init()
	profiler.Init()

	logger := log.Scoped("service", "the searcher service")

	// Ready immediately
	ready := make(chan struct{})
	close(ready)
	go debugserver.NewServerRoutine(ready).Start()

	var cacheSizeBytes int64
	if i, err := strconv.ParseInt(cacheSizeMB, 10, 64); err != nil {
		logger.Fatal("invalid int for SEARCHER_CACHE_SIZE_MB",
			log.String("SEARCHER_CACHE_SIZE_MB", cacheSizeMB),
			log.Error(err))
	} else {
		cacheSizeBytes = i * 1000 * 1000
	}

	db := ensureFrontendDB(logger)
	git := gitserver.NewClient(db)

	service := &search.Service{
		Store: &search.Store{
			FetchTar: func(ctx context.Context, repo api.RepoName, commit api.CommitID) (io.ReadCloser, error) {
				return git.Archive(ctx, repo, gitserver.ArchiveOptions{
					Treeish: string(commit),
					Format:  "tar",
				})
			},
			FetchTarPaths: func(ctx context.Context, repo api.RepoName, commit api.CommitID, paths []string) (io.ReadCloser, error) {
				pathspecs := make([]gitserver.Pathspec, len(paths))
				for i, p := range paths {
					pathspecs[i] = gitserver.PathspecLiteral(p)
				}
				return git.Archive(ctx, repo, gitserver.ArchiveOptions{
					Treeish:   string(commit),
					Format:    "tar",
					Pathspecs: pathspecs,
				})
			},
			FilterTar:         search.NewFilter,
			Path:              filepath.Join(cacheDir, "searcher-archives"),
			MaxCacheSizeBytes: cacheSizeBytes,
			DB:                db,
		},
		Log: logger,
	}
	service.Store.Start()

	// Set up handler middleware
	handler := actor.HTTPMiddleware(service)
	handler = trace.HTTPMiddleware(handler, conf.DefaultClient())
	handler = ot.HTTPMiddleware(handler)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	g, ctx := errgroup.WithContext(ctx)

	host := ""
	if env.InsecureDev {
		host = "127.0.0.1"
	}
	addr := net.JoinHostPort(host, port)
	server := &http.Server{
		ReadTimeout:  75 * time.Second,
		WriteTimeout: 10 * time.Minute,
		Addr:         addr,
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// For cluster liveness and readiness probes
			if r.URL.Path == "/healthz" {
				w.WriteHeader(200)
				_, _ = w.Write([]byte("ok"))
				return
			}
			handler.ServeHTTP(w, r)
		}),
	}

	go func() {
		logger.Info("searcher: listening", log.String("addr", server.Addr))
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			logger.Fatal(err.Error())
		}
		return nil
	}()

	// Shutdown
	g.Go(func() error {
		return shutdownOnSignal(ctx, server)
	})

	return g.Wait()
}

	// Wait for at most for the configured shutdown timeout.
	ctx, cancel := context.WithTimeout(context.Background(), goroutine.GracefulShutdownTimeout)
	defer cancel()
	// Stop accepting requests.
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("shutting down http server", log.Error(err))
	}
}
