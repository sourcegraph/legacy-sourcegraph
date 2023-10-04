package perforce

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/executil"
	"github.com/sourcegraph/sourcegraph/internal/wrexec"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// P4TestWithTrust attempts to test the Perforce server and performs a trust operation when needed.
func P4TestWithTrust(ctx context.Context, p4home, p4port, p4user, p4passwd string) error {
	// Attempt to check connectivity, may be prompted to trust.
	err := P4Test(ctx, p4home, p4port, p4user, p4passwd)
	if err == nil {
		return nil // The test worked, session still valid for the user
	}

	// If the output indicates that we have to run p4trust first, do that.
	if strings.Contains(err.Error(), "To allow connection use the 'p4 trust' command.") {
		err := P4Trust(ctx, p4home, p4port)
		if err != nil {
			return errors.Wrap(err, "trust")
		}
		// Now attempt to run p4test again.
		err = P4Test(ctx, p4home, p4port, p4user, p4passwd)
		if err != nil {
			return errors.Wrap(err, "testing connection after trust")
		}
		return nil
	}

	// Something unexpected happened, bubble up the error
	return err
}

// P4Trust blindly accepts fingerprint of the Perforce server.
func P4Trust(ctx context.Context, p4home, host string) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "p4", "trust", "-y", "-f")
	cmd.Env = append(os.Environ(),
		"P4PORT="+host,
		"HOME="+p4home,
	)

	out, err := executil.RunCommandCombinedOutput(ctx, wrexec.Wrap(ctx, log.NoOp(), cmd))
	if err != nil {
		if ctxerr := ctx.Err(); ctxerr != nil {
			err = ctxerr
		}
		if len(out) > 0 {
			err = errors.Errorf("%s (output follows)\n\n%s", err, out)
		}
		return err
	}
	return nil
}

// P4Test uses `p4 login -s` to test the Perforce connection: port, user, passwd.
func P4Test(ctx context.Context, p4home, p4port, p4user, p4passwd string) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// `p4 ping` requires extra-special access, so we want to avoid using it
	//
	// p4 login -s checks the connection and the credentials,
	// so it seems like the perfect alternative to `p4 ping`.
	cmd := exec.CommandContext(ctx, "p4", "login", "-s")
	cmd.Env = append(os.Environ(),
		"P4PORT="+p4port,
		"P4USER="+p4user,
		"P4PASSWD="+p4passwd,
		"HOME="+p4home,
	)

	out, err := executil.RunCommandCombinedOutput(ctx, wrexec.Wrap(ctx, log.NoOp(), cmd))
	if err != nil {
		if ctxerr := ctx.Err(); ctxerr != nil {
			err = errors.Wrap(ctxerr, "p4 login context error")
		}
		if len(out) > 0 {
			err = errors.Errorf("%s (output follows)\n\n%s", err, specifyCommandInErrorMessage(string(out), cmd))
		}
		return err
	}
	return nil
}
