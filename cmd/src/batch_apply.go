package main

import (
	"context"
	"errors"
	"flag"
	"fmt"

	"github.com/sourcegraph/sourcegraph/lib/output"
)

func init() {
	usage := `
'src batch apply' is used to apply a batch spec on a Sourcegraph instance,
creating or updating the described batch change if necessary.

Usage:

    src batch apply -f FILE [command options]

Examples:

    $ src batch apply -f batch.spec.yaml
  
    $ src batch apply -f batch.spec.yaml -namespace myorg

`

	flagSet := flag.NewFlagSet("apply", flag.ExitOnError)
	flags := newBatchExecuteFlags(flagSet, batchDefaultCacheDir(), batchDefaultTempDirPrefix())

	handler := func(args []string) error {
		if err := flagSet.Parse(args); err != nil {
			return err
		}

		if len(flagSet.Args()) != 0 {
			return &usageError{errors.New("additional arguments not allowed")}
		}

		ctx, cancel := contextCancelOnInterrupt(context.Background())
		defer cancel()

		var ui batchExecUI
		if flags.textOnly {
			ui = &batchExecJSONLinesUI{}
		} else {
			out := output.NewOutput(flagSet.Output(), output.OutputOpts{Verbose: *verbose})
			ui = &batchExecTUI{out: out}
		}

		err := executeBatchSpec(ctx, executeBatchSpecOpts{
			flags:  flags,
			client: cfg.apiClient(flags.api, flagSet.Output()),

			applyBatchSpec: true,

			ui: ui,
		})
		if err != nil {
			return &exitCodeError{nil, 1}
		}

		return nil
	}

	batchCommands = append(batchCommands, &command{
		flagSet: flagSet,
		handler: handler,
		usageFunc: func() {
			fmt.Fprintf(flag.CommandLine.Output(), "Usage of 'src batch %s':\n", flagSet.Name())
			flagSet.PrintDefaults()
			fmt.Println(usage)
		},
	})
}
