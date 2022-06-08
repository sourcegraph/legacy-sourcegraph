package dependencies

import (
	"context"
	"os"
	"time"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/check"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/usershell"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const (
	depsHomebrew      = "Homebrew"
	depsBaseUtilities = "Base utilities"
)

// Mac declares Mac dependencies.
var Mac = []category{
	{
		Name: depsHomebrew,
		Checks: []*dependency{
			{
				Name:        "brew",
				Check:       checkAction(check.InPath("brew")),
				Fix:         cmdAction(`eval $(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)`),
				Description: `We depend on having the Homebrew package manager available on macOS: https://brew.sh`,
			},
		},
	},
	{
		Name:      depsBaseUtilities,
		DependsOn: []string{depsHomebrew},
		Checks: []*dependency{
			{
				Name:  "git",
				Check: checkAction(check.Combine(check.InPath("git"), checkGitVersion(">= 2.34.1"))),
				Fix:   cmdAction(`brew install git`),
			},
			{
				Name:  "gnu-sed",
				Check: checkAction(check.InPath("gsed")),
				Fix:   cmdAction("brew install gnu-sed"),
			},
			{
				Name:  "findutils",
				Check: checkAction(check.InPath("gfind")),
				Fix:   cmdAction("brew install findutils"),
			},
			{
				Name:  "comby",
				Check: checkAction(check.InPath("comby")),
				Fix:   cmdAction("brew install comby"),
			},
			{
				Name:  "pcre",
				Check: checkAction(check.InPath("pcregrep")),
				Fix:   cmdAction(`brew install pcre`),
			},
			{
				Name:  "sqlite",
				Check: checkAction(check.InPath("sqlite3")),
				Fix:   cmdAction(`brew install sqlite`),
			},
			{
				Name:  "jq",
				Check: checkAction(check.InPath("jq")),
				Fix:   cmdAction(`brew install jq`),
			},
			{
				Name:  "bash",
				Check: checkAction(check.CommandOutputContains("bash --version", "version 5")),
				Fix:   cmdAction(`brew install bash`)},
			{
				Name: "rosetta",
				Check: checkAction(
					check.Any(
						// will return true on non-m1 macs
						check.CommandOutputContains("uname -m", "x86_64"),
						// oahd is the process running rosetta
						check.CommandExitCode("pgrep oahd", 0)),
				),
				Fix: cmdAction(`softwareupdate --install-rosetta --agree-to-license`),
			},
			{
				Name: "docker",
				Enabled: func(ctx context.Context, args CheckArgs) error {
					// Docker is quite funky in CI
					if os.Getenv("CI") == "true" {
						return errors.New("skipping Docker in CI")
					}
					return nil
				},
				Check: checkAction(check.Combine(
					check.WrapErrMessage(check.InPath("docker"),
						"if Docker is installed and the check fails, you might need to restart terminal and 'sg setup'"),
				)),
				Fix: func(ctx context.Context, cio check.IO, args CheckArgs) error {
					// TODO stream lines
					if err := usershell.Cmd(ctx, `brew install --cask docker`).Run(); err != nil {
						return err
					}

					cio.Verbose("Docker installed - attempting to start docker")

					return usershell.Cmd(ctx, "open --hide --background /Applications/Docker.app").Run()
				},
			},
			{
				Name:  "asdf",
				Check: checkAction(check.CommandOutputContains("asdf", "version")),
				Fix: func(ctx context.Context, cio check.IO, args CheckArgs) error {
					// Uses `&&` to avoid appending the shell config on failed installations attempts.
					cmd := `brew install asdf && echo ". ${HOMEBREW_PREFIX:-/usr/local}/opt/asdf/libexec/asdf.sh" >> ` + usershell.ShellConfigPath(ctx)
					return usershell.Cmd(ctx, cmd).Run()
				},
			},
		},
	},
	categoryCloneRepositories(),
	{
		Name:      "Programming languages & tooling",
		DependsOn: []string{depsHomebrew, depsBaseUtilities},
		Checks: []*check.Check[CheckArgs]{
			{
				Name:  "go",
				Check: checkGoVersion,
				Fix: cmdsAction(
					"asdf plugin-add golang https://github.com/kennyp/asdf-golang.git",
					"asdf install golang",
				),
			},
			{
				Name:  "yarn",
				Check: checkYarnVersion,
				Fix: cmdsAction(
					"brew install gpg",
					"asdf plugin-add yarn",
					"asdf install yarn",
				),
			},
			{
				Name:  "node",
				Check: checkNodeVersion,
				Fix: cmdsAction(
					"asdf plugin add nodejs https://github.com/asdf-vm/asdf-nodejs.git",
					`grep -s "legacy_version_file = yes" ~/.asdfrc >/dev/null || echo 'legacy_version_file = yes' >> ~/.asdfrc`,
					"asdf install nodejs",
				),
			},
			{
				Name:  "rust",
				Check: checkRustVersion,
				Fix: cmdsAction(
					"asdf plugin-add rust https://github.com/asdf-community/asdf-rust.git",
					"asdf install rust",
				),
			},
		},
	},
	{
		Name: "Postgres database",
		Checks: []*dependency{
			{
				Name: "Install Postgres",
				Description: `psql, the PostgreSQL CLI client, needs to be available in your $PATH.

If you've installed PostgreSQL with Homebrew that should be the case.

If you used another method, make sure psql is available.`,
				Check: checkAction(check.InPath("psql")),
				Fix:   cmdAction("brew install postgresql"),
			},
			{
				Name: "Start Postgres",
				// In the eventuality of the user using a non standard configuration and having
				// set it up appropriately in its configuration, we can bypass the standard postgres
				// check and directly check for the sourcegraph database.
				//
				// Because only the latest error is returned, it's better to finish with the real check
				// for error message clarity.
				Check: func(ctx context.Context, cio check.IO, args CheckArgs) error {
					if err := checkSourcegraphDatabase(ctx, cio, args); err == nil {
						return nil
					}
					return checkPostgresConnection(ctx)
				},
				Description: `Sourcegraph requires the PostgreSQL database to be running.

We recommend installing it with Homebrew and starting it as a system service.
If you know what you're doing, you can also install PostgreSQL another way.
For example: you can use https://postgresapp.com/

If you're not sure: use the recommended commands to install PostgreSQL.`,
				Fix: func(ctx context.Context, cio check.IO, args CheckArgs) error {
					err := usershell.Cmd(ctx, "brew services start postgresql").Run()
					if err != nil {
						return err
					}

					// Wait for startup
					time.Sleep(5 * time.Second)

					// Doesn't matter if this succeeds
					_ = usershell.Cmd(ctx, "createdb").Run()
					return nil
				},
			},
			{
				Name:  "Connection to 'sourcegraph' database",
				Check: checkSourcegraphDatabase,
				Description: `` +
					`Once PostgreSQL is installed and running, we need to set up Sourcegraph database itself and a
specific user.`,
				Fix: cmdsAction(
					"createuser --superuser sourcegraph || true",
					`psql -c "ALTER USER sourcegraph WITH PASSWORD 'sourcegraph';"`,
					`createdb --owner=sourcegraph --encoding=UTF8 --template=template0 sourcegraph`,
				),
			},
		},
	},
}
