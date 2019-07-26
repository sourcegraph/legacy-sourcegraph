// Package shared contains the frontend command implementation shared
package shared

import (
	"fmt"
	"os"

	"sourcegraph.com/cmd/frontend/internal/cli"
	"sourcegraph.com/pkg/env"
)

// Main is the main function that runs the frontend process.
//
// It is exposed as function in a package so that it can be called by other
// main package implementations such as Sourcegraph Enterprise, which import
// proprietary/private code.
func Main() {
	env.Lock()
	err := cli.Main()
	if err != nil {
		fmt.Fprintln(os.Stderr, "fatal:", err)
		os.Exit(1)
	}
}
