package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/garyburd/redigo/redis"
	"github.com/jackc/pgx/v4"
	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/db"
	"github.com/sourcegraph/sourcegraph/internal/database"
	connections "github.com/sourcegraph/sourcegraph/internal/database/connections/live"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/runner"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/store"
	"github.com/sourcegraph/sourcegraph/internal/database/postgresdsn"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

var (
	dbFlagSet             = flag.NewFlagSet("sg db", flag.ExitOnError)
	dbResetPGFlagSet      = flag.NewFlagSet("sg db reset-pg", flag.ExitOnError)
	dbDatabaseNameFlag    = dbResetPGFlagSet.String("db", db.DefaultDatabase.Name, "The target database instance.")
	dbRedisFlagSet        = flag.NewFlagSet("sg db reset-redis", flag.ExitOnError)
	dbAddUserFlagSet      = flag.NewFlagSet("sg db add-user", flag.ExitOnError)
	dbAddUserNameFlag     = dbAddUserFlagSet.String("name", "sourcegraph", "User name")
	dbAddUserPasswordFlag = dbAddUserFlagSet.String("password", "sourcegraph", "User password")

	dbCommand = &ffcli.Command{
		Name:       "db",
		ShortUsage: "",
		LongHelp:   "",
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
		Subcommands: []*ffcli.Command{
			{
				Name:       "reset-pg",
				ShortUsage: fmt.Sprintf("sg db reset-pg [-db=%s]", db.DefaultDatabase.Name),
				ShortHelp:  "Drops, recreates and migrates the specified Sourcegraph database.",
				LongHelp:   `Run 'sg db reset-pg' to drop and recreate Sourcegraph databases. If -db is not set, then the "frontend" database is used (what's set as PGDATABASE in env or the sg.config.yaml). If -db is set to "all" then all databases are reset and recreated.`,
				FlagSet:    dbResetPGFlagSet,
				Exec:       dbResetPGExec,
			},
			{
				Name:       "reset-redis",
				ShortUsage: fmt.Sprintf("sg db reset-redis [-db=%s]", db.DefaultDatabase.Name), // TODO edit flag
				ShortHelp:  "Drops, recreates and migrates the specified redis Sourcegraph database.",
				LongHelp:   `Run 'sg db reset-redis' to drop and recreate Sourcegraph redis databases. TODO`,
				FlagSet:    dbRedisFlagSet,
				Exec:       dbResetRedisExec,
			},
			{
				Name:       "add-user",
				ShortUsage: fmt.Sprintf("sg db add-user [-name=%s -password=%s]", "sourcegraph", "sourcegraph"),
				ShortHelp:  "Create an admin sourcegraph user",
				LongHelp:   `Run 'sg db add-user -name bob' to create an admin user whose email is bob@sourcegraph.com. The password will be printed if the operation succeeds`,
				FlagSet:    dbAddUserFlagSet,
				Exec:       dbAddUserExec,
			},
		},
	}
)

// We currently cannot override the CONFIGURATION_MODE at runtime, which by
// defaults trigger calls to the frontend application and will prevent database calls
// to succeed unless the front-end is running.
//
// This function forces the configuration mode by restarting the current process
// with an updated environment that forces the configuration mode.
//
// See https://github.com/sourcegraph/sourcegraph/issues/29222
func forceConfigurationMode() {
	mode, ok := os.LookupEnv("CONFIGURATION_MODE")
	if !ok || mode != "empty" {
		path, err := os.Executable()
		if err != nil {
			panic(err)
		}
		env := os.Environ()
		overridenEnv := make([]string, 0, len(env))
		for _, e := range env {
			if !strings.HasPrefix(e, "CONFIGURATION_MODE=") {
				overridenEnv = append(overridenEnv, e)
			}
		}
		overridenEnv = append(overridenEnv, "CONFIGURATION_MODE=empty")

		if err := syscall.Exec(path, os.Args, overridenEnv); err != nil {
			panic(err)
		}
	}
}

func dbAddUserExec(ctx context.Context, args []string) error {
	// If we detect CONFIGURATION_MODE to be different than "empty", force its value to empty.
	forceConfigurationMode()

	// Read the configuration.
	ok, _ := parseConf(*configFlag, *overwriteConfigFlag)
	if !ok {
		return errors.New("failed to read sg.config.yaml. This command needs to be run in the `sourcegraph` repository")
	}

	// Connect to the database.
	conn, err := connections.NewFrontendDB(postgresdsn.New("", "", getEnv), "frontend", true, &observation.TestContext)
	if err != nil {
		return err
	}
	db := database.NewDB(conn)

	// Create the user, generating an email based on the username.
	email := fmt.Sprintf("%s@sourcegraph.com", *dbAddUserNameFlag)
	user, err := db.Users().Create(ctx, database.NewUser{
		Username:        *dbAddUserNameFlag,
		Email:           email,
		EmailIsVerified: true,
		Password:        *dbAddUserPasswordFlag,
	})
	if err != nil {
		return err
	}

	// Make the user site admin.
	err = db.Users().SetIsSiteAdmin(ctx, user.ID, true)
	if err != nil {
		return err
	}

	// Report back the new user informations.
	writeFingerPointingLinef(
		"User %s%s%s (%s%s%s) has been created and its password is %s%s%s.",
		output.StyleOrange,
		*dbAddUserNameFlag,
		output.StyleReset,
		output.StyleOrange,
		email,
		output.StyleReset,
		output.StyleOrange,
		*dbAddUserPasswordFlag,
		output.StyleReset,
	)

	return nil
}

func dbResetRedisExec(ctx context.Context, args []string) error {
	// Read the configuration.
	ok, _ := parseConf(*configFlag, *overwriteConfigFlag)
	if !ok {
		return errors.New("failed to read sg.config.yaml. This command needs to be run in the `sourcegraph` repository")
	}

	// Connect to the redis database.
	endpoint := getEnv("REDIS_ENDPOINT")
	conn, err := redis.Dial("tcp", endpoint, redis.DialConnectTimeout(5*time.Second))
	if err != nil {
		return errors.Wrapf(err, "failed to connect to Redis at %s", endpoint)
	}

	// Drop everything in redis
	_, err = conn.Do("flushall")
	if err != nil {
		return errors.Wrap(err, "failed to run command on redis")
	}

	return nil
}

func dbResetPGExec(ctx context.Context, args []string) error {
	// If we detect CONFIGURATION_MODE to be different than "empty", force its value to empty.
	forceConfigurationMode()

	// Read the configuration.
	ok, _ := parseConf(*configFlag, *overwriteConfigFlag)
	if !ok {
		return errors.New("failed to read sg.config.yaml. This command needs to be run in the `sourcegraph` repository")
	}

	var (
		dsnMap      = map[string]string{}
		schemaNames []string
	)

	if *dbDatabaseNameFlag == "all" {
		schemaNames = schemas.SchemaNames
	} else {
		schemaNames = strings.Split(*dbDatabaseNameFlag, ",")
	}

	for _, name := range schemaNames {
		if name == "frontend" {
			dsnMap[name] = postgresdsn.New("", "", getEnv)
		} else {
			dsnMap[name] = postgresdsn.New(strings.ToUpper(name), "", getEnv)
		}
	}

	for name, dsn := range dsnMap {
		var (
			db  *pgx.Conn
			err error
		)

		db, err = pgx.Connect(ctx, dsn)
		if err != nil {
			return errors.Wrap(err, "failed to connect to Postgres database")
		}

		writeFingerPointingLinef("This will reset database %s%s%s. Are you okay with this?", output.StyleOrange, name, output.StyleReset)
		ok := getBool()
		if !ok {
			return nil
		}

		_, err = db.Exec(ctx, "DROP SCHEMA public CASCADE; CREATE SCHEMA public;")
		if err != nil {
			writeFailureLinef("Failed to drop schema 'public': %s", err)
			return err
		}

		if err := db.Close(ctx); err != nil {
			return err
		}
	}

	storeFactory := func(db *sql.DB, migrationsTable string) connections.Store {
		return store.NewWithDB(db, migrationsTable, store.NewOperations(&observation.TestContext))
	}

	options := runner.Options{
		Up:            true,
		NumMigrations: 0,
		SchemaNames:   schemaNames,
	}

	return connections.RunnerFromDSNs(dsnMap, "sg", storeFactory).Run(ctx, options)
}

func getEnv(key string) string {
	// First look into process env, emulating the logic in makeEnv used
	// in internal/run/run.go
	val, ok := os.LookupEnv(key)
	if ok {
		return val
	}
	// Otherwise check in globalConf.Env and *expand* the key, because a value might refer to another env var.
	return os.Expand(globalConf.Env[key], func(lookup string) string {
		if lookup == key {
			return os.Getenv(lookup)
		}

		if e, ok := globalConf.Env[lookup]; ok {
			return e
		}
		return os.Getenv(lookup)
	})
}
