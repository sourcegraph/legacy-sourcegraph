package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/run"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/stdout"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

var (
	startFlagSet       = flag.NewFlagSet("sg start", flag.ExitOnError)
	debugStartServices = startFlagSet.String("debug", "", "Comma separated list of services to set at debug log level.")
	infoStartServices  = startFlagSet.String("info", "", "Comma separated list of services to set at info log level.")
	warnStartServices  = startFlagSet.String("warn", "", "Comma separated list of services to set at warn log level.")
	errorStartServices = startFlagSet.String("error", "", "Comma separated list of services to set at error log level.")
	critStartServices  = startFlagSet.String("crit", "", "Comma separated list of services to set at crit log level.")

	startCommand = &ffcli.Command{
		Name:       "start",
		ShortUsage: "sg start [commandset]",
		ShortHelp:  "🌟Starts the given commandset. Without a commandset it starts the default Sourcegraph dev environment.",
		LongHelp:   constructStartCmdLongHelp(),

		FlagSet: startFlagSet,
		Exec:    startExec,
	}

	// run-set is the deprecated older version of `start`
	runSetFlagSet = flag.NewFlagSet("sg run-set", flag.ExitOnError)
	runSetCommand = &ffcli.Command{
		Name:       "run-set",
		ShortUsage: "sg run-set <commandset>",
		ShortHelp:  "DEPRECATED. Use 'sg start' instead. Run the given commandset.",
		FlagSet:    runSetFlagSet,
		Exec:       runSetExec,
	}
)

func constructStartCmdLongHelp() string {
	var out strings.Builder

	fmt.Fprintf(&out, `Runs the given commandset.

If no commandset is specified, it starts the commandset with the name 'default'.

Use this to start your Sourcegraph environment!
`)

	// Attempt to parse config to list available commands, but don't fail on
	// error, because we should never error when the user wants --help output.
	_, _ = parseConf(*configFlag, *overwriteConfigFlag)

	if globalConf != nil {
		fmt.Fprintf(&out, "\n")
		fmt.Fprintf(&out, "AVAILABLE COMMANDSETS IN %s%s%s\n", output.StyleBold, *configFlag, output.StyleReset)

		var names []string
		for name := range globalConf.Commandsets {
			switch name {
			case "enterprise-codeintel":
				names = append(names, fmt.Sprintf("  %s 🧠", name))
			case "batches":
				names = append(names, fmt.Sprintf("  %s 🦡", name))
			default:
				names = append(names, fmt.Sprintf("  %s", name))
			}
		}
		sort.Strings(names)
		fmt.Fprint(&out, strings.Join(names, "\n"))
	}

	return out.String()
}

func startExec(ctx context.Context, args []string) error {
	ok, errLine := parseConf(*configFlag, *overwriteConfigFlag)
	if !ok {
		out.WriteLine(errLine)
		os.Exit(1)
	}

	if len(args) > 2 {
		out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: too many arguments\n"))
		return flag.ErrHelp
	}

	if len(args) == 0 {
		args = append(args, "default")
	}

	set, ok := globalConf.Commandsets[args[0]]
	if !ok {
		out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: commandset %q not found :(\n", args[0]))
		return flag.ErrHelp
	}

	var checks []run.Check
	for _, name := range set.Checks {
		check, ok := globalConf.Checks[name]
		if !ok {
			out.WriteLine(output.Linef("", output.StyleWarning, "WARNING: check %s not found in config\n", name))
			continue
		}
		checks = append(checks, check)
	}

	ok, err := run.Checks(ctx, globalConf.Env, checks...)
	if err != nil {
		out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: checks could not be run: %s\n", err))
	}

	if !ok {
		out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: checks did not pass, aborting start of commandset %s\n", set.Name))
		return nil
	}

	cmds := make([]run.Command, 0, len(set.Commands))
	for _, name := range set.Commands {
		cmd, ok := globalConf.Commands[name]
		if !ok {
			return errors.Errorf("command %q not found in commandset %q", name, args[0])
		}

		cmds = append(cmds, cmd)
	}

	levelOverrides := logLevelOverrides()
	for _, cmd := range cmds {
		enrichWithLogLevels(&cmd, levelOverrides)
	}

	env := globalConf.Env
	for k, v := range set.Env {
		env[k] = v
	}

	return run.Commands(ctx, env, cmds...)
}

// logLevelOverrides builds a map of commands -> log level that should be overridden in the environment.
func logLevelOverrides() map[string]string {
	levelServices := make(map[string][]string)
	levelServices["debug"] = parseCsv(*debugStartServices)
	levelServices["info"] = parseCsv(*infoStartServices)
	levelServices["warn"] = parseCsv(*warnStartServices)
	levelServices["error"] = parseCsv(*errorStartServices)
	levelServices["crit"] = parseCsv(*critStartServices)

	overrides := make(map[string]string)
	for level, services := range levelServices {
		for _, service := range services {
			overrides[service] = level
		}
	}

	return overrides
}

// enrichWithLogLevels will add any logger level overrides to a given command if they have been specified.
func enrichWithLogLevels(cmd *run.Command, overrides map[string]string) {
	logLevelVariable := "SRC_LOG_LEVEL"

	if level, ok := overrides[cmd.Name]; ok {
		out.WriteLine(output.Linef("", output.StylePending, "Setting log level: %s for command %s.", level, cmd.Name))
		if cmd.Env == nil {
			cmd.Env = make(map[string]string, 1)
			cmd.Env[logLevelVariable] = level
		}
		cmd.Env[logLevelVariable] = level
	}
}

// parseCsv takes an input comma seperated string and returns a list of tokens each trimmed for whitespace
func parseCsv(input string) []string {
	tokens := strings.Split(input, ",")
	results := make([]string, 0, len(tokens))
	for _, token := range tokens {
		results = append(results, strings.TrimSpace(token))
	}
	return results
}

var deprecationStyle = output.CombineStyles(output.Fg256Color(255), output.Bg256Color(124))

func runSetExec(ctx context.Context, args []string) error {
	stdout.Out.WriteLine(output.Linef("", deprecationStyle, " _______________________________________________________________________ "))
	stdout.Out.WriteLine(output.Linef("", deprecationStyle, "/         `sg run-set` is deprecated - use `sg start` instead!          \\"))
	stdout.Out.WriteLine(output.Linef("", deprecationStyle, "!                                                                       !"))
	stdout.Out.WriteLine(output.Linef("", deprecationStyle, "!         Run `sg start -help` for usage information.                   !"))
	stdout.Out.WriteLine(output.Linef("", deprecationStyle, "\\_______________________________________________________________________/"))
	stdout.Out.WriteLine(output.Linef("", deprecationStyle, "                               !  !                                      "))
	stdout.Out.WriteLine(output.Linef("", deprecationStyle, "                               !  !                                      "))
	stdout.Out.WriteLine(output.Linef("", deprecationStyle, "                               L_ !                                      "))
	stdout.Out.WriteLine(output.Linef("", deprecationStyle, "                              / _)!                                      "))
	stdout.Out.WriteLine(output.Linef("", deprecationStyle, "                             / /__L                                      "))
	stdout.Out.WriteLine(output.Linef("", deprecationStyle, "                       _____/ (____)                                     "))
	stdout.Out.WriteLine(output.Linef("", deprecationStyle, "                              (____)                                     "))
	stdout.Out.WriteLine(output.Linef("", deprecationStyle, "                       _____  (____)                                     "))
	stdout.Out.WriteLine(output.Linef("", deprecationStyle, "                            \\_(____)                                     "))
	stdout.Out.WriteLine(output.Linef("", deprecationStyle, "                               !  !                                      "))
	stdout.Out.WriteLine(output.Linef("", deprecationStyle, "                               !  !                                      "))
	stdout.Out.WriteLine(output.Linef("", deprecationStyle, "                               \\__/                                      "))
	return startExec(ctx, args)
}
