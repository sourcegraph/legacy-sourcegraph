package perforce

import (
	"bytes"
	"context"
	"encoding/json"
	"os"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/gitserverfs"
	"github.com/sourcegraph/sourcegraph/internal/byteutils"
	p4types "github.com/sourcegraph/sourcegraph/internal/perforce"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// P4ProtectsForUserArguments are the arguments for P4ProtectsForUser.
type P4ProtectsForUserArguments struct {
	// P4PORT is the address of the Perforce server.
	P4Port string
	// P4User is the Perforce username to authenticate with.
	P4User string
	// P4Passwd is the Perforce password to authenticate with.
	P4Passwd string

	// Username is the username for which to get the protect definition for
	Username string
}

// P4ProtectsForUser returns all protect definitions that apply to the given username.
func P4ProtectsForUser(ctx context.Context, fs gitserverfs.FS, args P4ProtectsForUserArguments) ([]*p4types.Protect, error) {
	options := []P4OptionFunc{
		WithAuthentication(args.P4User, args.P4Passwd),
		WithHost(args.P4Port),
	}

	// -u User : Displays protection lines that apply to the named user. This option
	// requires super access.
	options = append(options, WithArguments("-Mj", "-ztag", "protects", "-u", args.Username))

	p4home, err := fs.P4HomeDir(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create p4home dir")
	}

	scratchDir, err := fs.TempDir(ctx, "p4-protects-")
	if err != nil {
		return nil, errors.Wrap(err, "could not create temp dir to invoke 'p4 protects'")
	}
	defer os.Remove(scratchDir)

	cmd := NewBaseCommand(ctx, p4home, scratchDir, options...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		if ctxerr := ctx.Err(); ctxerr != nil {
			err = errors.Wrap(ctxerr, "p4 protects context error")
		}

		if len(out) > 0 {
			err = errors.Wrapf(err, `failed to run command "p4 protects" (output follows)\n\n%s`, specifyCommandInErrorMessage(string(out), cmd.Unwrap()))
		}

		return nil, err
	}

	if len(out) == 0 {
		// no error, but also no protects.
		return nil, nil
	}

	if os.Getenv("SRC_GITSERVER_P4_BROKER_ENABLED") != "" {
		return parseP4BrokerProtects(out)
	} else {
		return parseP4Protects(out)
	}
}

type P4ProtectsForDepotArguments struct {
	// P4PORT is the address of the Perforce server.
	P4Port string
	// P4User is the Perforce username to authenticate with.
	P4User string
	// P4Passwd is the Perforce password to authenticate with.
	P4Passwd string

	// Depot is the depot to get the protect definition for.
	Depot string
}

// P4ProtectsForUser returns all protect definitions that apply to the given depot.
func P4ProtectsForDepot(ctx context.Context, fs gitserverfs.FS, args P4ProtectsForDepotArguments) ([]*p4types.Protect, error) {
	options := []P4OptionFunc{
		WithAuthentication(args.P4User, args.P4Passwd),
		WithHost(args.P4Port),
	}

	// -a : Displays protection lines for all users. This option requires super
	// access.
	options = append(options, WithArguments("-Mj", "-ztag", "protects", "-a", args.Depot))

	p4home, err := fs.P4HomeDir(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create p4home dir")
	}

	scratchDir, err := fs.TempDir(ctx, "p4-protects-")
	if err != nil {
		return nil, errors.Wrap(err, "could not create temp dir to invoke 'p4 protects'")
	}
	defer os.Remove(scratchDir)

	cmd := NewBaseCommand(ctx, p4home, scratchDir, options...)

	out, err := cmd.CombinedOutput()
	if err != nil {
		if ctxerr := ctx.Err(); ctxerr != nil {
			err = errors.Wrap(ctxerr, "p4 protects context error")
		}

		if len(out) > 0 {
			err = errors.Wrapf(err, `failed to run command "p4 protects" (output follows)\n\n%s`, specifyCommandInErrorMessage(string(out), cmd.Unwrap()))
		}

		return nil, err
	}

	if len(out) == 0 {
		// no error, but also no protects.
		return nil, nil
	}

	return parseP4Protects(out)
}

type perforceJSONProtect struct {
	DepotFile string  `json:"depotFile"`
	Host      string  `json:"host"`
	Line      string  `json:"line"`
	Perm      string  `json:"perm"`
	IsGroup   *string `json:"isgroup,omitempty"`
	Unmap     *string `json:"unmap,omitempty"`
	User      string  `json:"user"`
}

type perforceBrokerJSONProtect struct {
	Data []byte `json:"data"` // base64 encoded JSON
}

// parseP4BrokerProtects decodes a `p4 protects` message returned from a
// `p4broker` filter.
//
// It first decodes the "data" JSON field from the response, which should be
// a base64 JSON string from an actual `p4 protects` command.
// Once decoded, it calls the parseP4Protects function to
// parse the protects as normal.
func parseP4BrokerProtects(brokerProtects []byte) ([]*p4types.Protect, error) {
	var parsedBrokerResponse perforceBrokerJSONProtect
	if err := json.Unmarshal(brokerProtects, &parsedBrokerResponse); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal protect line")
	}

	if len(parsedBrokerResponse.Data) == 0 {
		return nil, errors.New("not a valid protects response")
	}

	return parseP4Protects(parsedBrokerResponse.Data)
}

// parseP4Protects expects output from a `p4 protects` command called with
// the `-Mj -ztag` flags, and returns a parsed list of protects.
func parseP4Protects(out []byte) ([]*p4types.Protect, error) {
	protects := make([]*p4types.Protect, 0)

	lr := byteutils.NewLineReader(out)
	for lr.Scan() {
		line := lr.Line()

		// Trim whitespace
		line = bytes.TrimSpace(line)

		var parsedLine perforceJSONProtect
		if err := json.Unmarshal(line, &parsedLine); err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal protect line")
		}

		if parsedLine.DepotFile == "" {
			return nil, errors.New("not a valid protects response")
		}

		entityType := "user"
		if parsedLine.IsGroup != nil {
			entityType = "group"
		}

		protects = append(protects, &p4types.Protect{
			Host:        parsedLine.Host,
			EntityType:  entityType,
			EntityName:  parsedLine.User,
			Match:       parsedLine.DepotFile,
			IsExclusion: parsedLine.Unmap != nil,
			Level:       parsedLine.Perm,
		})
	}

	return protects, nil
}
