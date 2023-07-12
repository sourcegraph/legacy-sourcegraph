package compute

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/search/result"
)

type Command interface {
	command()
	Run(context.Context, result.Match) (Result, error)
	ToSearchPattern() string
	String() string
}

type CommandPostRunHook interface {
	PostRunHook(*commandParser)
}

var (
	_ Command = (*MatchOnly)(nil)
	_ Command = (*Replace)(nil)
	_ Command = (*Output)(nil)

	_ CommandPostRunHook = (*Replace)(nil)
)

func (MatchOnly) command() {}
func (Replace) command()   {}
func (Output) command()    {}
