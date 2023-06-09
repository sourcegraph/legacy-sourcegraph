package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/httpapi"
	apirouter "github.com/sourcegraph/sourcegraph/cmd/frontend/internal/httpapi/router"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/migration"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/multiversion"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/runner"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/store"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/httpserver"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration/migrations"
	"github.com/sourcegraph/sourcegraph/internal/version"
	"github.com/sourcegraph/sourcegraph/internal/version/upgradestore"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

var buffer strings.Builder // :)

var shouldAutoUpgade = env.MustGetBool("SRC_AUTOUPGRADE", false, "blahblahblah")

func tryAutoUpgrade(ctx context.Context, obsvCtx *observation.Context, db database.DB, hook store.RegisterMigratorsUsingConfAndStoreFactoryFunc) (err error) {
	toVersion, ok := oobmigration.NewVersionFromString(version.Version())
	if !ok {
		return nil
	}
	var currentVersion oobmigration.Version

	upgradestore := upgradestore.New(db)

	_, doAutoUpgrade, err := upgradestore.GetAutoUpgrade(ctx)
	if err != nil {
		return errors.Wrap(err, "autoupgradestore.GetAutoUpgrade")
	}
	if !doAutoUpgrade && !shouldAutoUpgade {
		return nil
	}

	if err := upgradestore.EnsureUpgradeTable(ctx); err != nil {
		return errors.Wrap(err, "autoupgradestore.EnsureUpgradeTable")
	}

	// try to claim
	for {
		currentVersionStr, _, err := upgradestore.GetServiceVersion(ctx)
		if err != nil {
			return errors.Wrap(err, "autoupgradestore.GetServiceVersion")
		}

		currentVersion, ok = oobmigration.NewVersionFromString(currentVersionStr)
		if !ok {
			return errors.Newf("VERSION STRING BAD %s", currentVersion)
		}

		if cmp := oobmigration.CompareVersions(currentVersion, toVersion); cmp == oobmigration.VersionOrderAfter || cmp == oobmigration.VersionOrderEqual {
			return nil
		}

		claimed, err := upgradestore.ClaimAutoUpgrade(ctx, currentVersionStr, version.Version())
		if err != nil {
			return errors.Wrap(err, "autoupgradstore.ClaimAutoUpgrade")
		}

		if claimed {
			break
		}

		time.Sleep(time.Second * 10)
	}

	stopFunc, err := serveConfigurationServer(ctx, obsvCtx)
	if err != nil {
		return err
	}
	defer stopFunc()

	if err := runMigration(ctx, obsvCtx, currentVersion, toVersion, db, hook); err != nil {
		return err
	}

	if err := upgradestore.SetServiceVersion(ctx, toVersion.GitTag()); err != nil {
		return errors.Wrap(err, "autoupgradstore.SetServiceVersion")
	}

	return errors.New("MIGRATION SUCCEEDED, RESTARTING")
}

func runMigration(ctx context.Context,
	obsvCtx *observation.Context,
	from,
	to oobmigration.Version,
	db database.DB,
	enterpriseMigratorsHook store.RegisterMigratorsUsingConfAndStoreFactoryFunc,
) error {
	versionRange, err := oobmigration.UpgradeRange(from, to)
	if err != nil {
		return err
	}

	fmt.Printf("RANGE %+v %+v %+v\n", from, to, versionRange)

	interrupts, err := oobmigration.ScheduleMigrationInterrupts(from, to)
	if err != nil {
		return err
	}

	plan, err := multiversion.PlanMigration(from, to, versionRange, interrupts)
	if err != nil {
		return err
	}

	registerMigrators := store.ComposeRegisterMigratorsFuncs(
		migrations.RegisterOSSMigratorsUsingConfAndStoreFactory,
		enterpriseMigratorsHook,
	)

	tee := io.MultiWriter(&buffer, os.Stdout)
	out := output.NewOutput(tee, output.OutputOpts{})

	runnerFactory := func(schemaNames []string, schemas []*schemas.Schema) (*runner.Runner, error) {
		return migration.NewRunnerWithSchemas(
			obsvCtx,
			out,
			"frontend-autoupgrader", schemaNames, schemas,
		)
	}

	return multiversion.RunMigration(
		ctx,
		db,
		runnerFactory,
		plan,
		runner.ApplyPrivilegedMigrations,
		nil,
		true,
		true,
		false,
		true,
		false,
		registerMigrators,
		nil, // only needed for drift
		out,
	)
}

func serveConfigurationServer(ctx context.Context, obsvCtx *observation.Context) (context.CancelFunc, error) {
	serveMux := http.NewServeMux()
	router := mux.NewRouter().PathPrefix("/.internal").Subrouter()
	middleware := httpapi.JsonMiddleware(&httpapi.ErrorHandler{
		Logger:       obsvCtx.Logger,
		WriteErrBody: true,
	})
	router.Path("/configuration").Methods(http.MethodPost).Name(apirouter.Configuration)
	router.Get(apirouter.Configuration).Handler(middleware(func(w http.ResponseWriter, r *http.Request) error {
		configuration := conf.Unified{
			ServiceConnectionConfig: conftypes.ServiceConnections{
				PostgresDSN:          "lol",
				CodeIntelPostgresDSN: "lol",
				CodeInsightsDSN:      "lol",
			},
		}
		return json.NewEncoder(w).Encode(configuration)
	}))
	serveMux.Handle("/.internal/", router)
	h := http.Handler(serveMux)
	server := &http.Server{
		Handler:      h,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	listener, err := httpserver.NewListener(httpAddrInternal)
	if err != nil {
		return nil, err
	}
	confServer := httpserver.New(listener, server)

	goroutine.Go(func() {
		confServer.Start()
	})

	return confServer.Stop, nil
}

func serveUpgradeUI(ctx context.Context, logger log.Logger) (context.CancelFunc, error) {
	serveMux := http.NewServeMux()
	router := mux.NewRouter().PathPrefix("/.internal").Subrouter()
	middleware := httpapi.JsonMiddleware(&httpapi.ErrorHandler{
		Logger:       logger,
		WriteErrBody: true,
	})
	router.Get(apirouter.Configuration).Handler(middleware(func(w http.ResponseWriter, r *http.Request) error {
		configuration := conf.Unified{
			ServiceConnectionConfig: conftypes.ServiceConnections{
				PostgresDSN:          "lol",
				CodeIntelPostgresDSN: "lol",
				CodeInsightsDSN:      "lol",
			},
		}
		return json.NewEncoder(w).Encode(configuration)
	}))
	serveMux.Handle("/.internal/", router)
	h := http.Handler(serveMux)
	server := &http.Server{
		Handler:      h,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	listener, err := httpserver.NewListener(httpAddrInternal)
	if err != nil {
		return nil, err
	}
	confServer := httpserver.New(listener, server)

	goroutine.Go(func() {
		confServer.Start()
	})

	return confServer.Stop, nil
}
