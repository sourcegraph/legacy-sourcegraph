package graphqlbackend

import (
	"context"
	"errors"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
	"github.com/sourcegraph/sourcegraph/internal/wrexec"
)

type RecordedCommandsArgs struct {
	RepoName string
}

func (r *schemaResolver) RecordedCommands(ctx context.Context, args RecordedCommandsArgs) ([]RecordedCommandResolver, error) {
	repoName := args.RepoName
	if strings.TrimSpace(repoName) == "" {
		return nil, errors.New("empty string provided as repository name")
	}
	recordingConf := conf.Get().SiteConfig().GitRecorder
	store := rcache.NewFIFOList(wrexec.GetFIFOListKey(repoName), recordingConf.Size)
	empty, err := store.IsEmpty()
	if err != nil {
		return nil, err
	}
	if empty {
		return nil, nil
	}
	raws, err := store.All(ctx)
	if err != nil {
		return nil, err
	}
	resolvers := make([]RecordedCommandResolver, len(raws))
	for i, raw := range raws {
		command, err := wrexec.UnmarshalCommand(raw)
		if err != nil {
			return nil, err
		}
		resolvers[i] = NewRecordedCommandResolver(command)
	}
	return resolvers, nil
}

type RecordedCommandResolver interface {
	Start() gqlutil.DateTime
	Duration() float64
	Command() string
	Dir() string
	Path() string
}

type recordedCommandResolver struct {
	command wrexec.RecordedCommand
}

func NewRecordedCommandResolver(command wrexec.RecordedCommand) RecordedCommandResolver {
	return &recordedCommandResolver{command: command}
}

func (r *recordedCommandResolver) Start() gqlutil.DateTime {
	return *gqlutil.FromTime(r.command.Start)
}

func (r *recordedCommandResolver) Duration() float64 {
	return r.command.Duration
}

func (r *recordedCommandResolver) Command() string {
	return strings.Join(r.command.Args, " ")
}

func (r *recordedCommandResolver) Dir() string {
	return r.command.Dir
}

func (r *recordedCommandResolver) Path() string {
	return r.command.Path
}
