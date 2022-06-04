package parser

import (
	"context"
)

type ParserFactory func() (SimpleParser, error)

type ParserPool interface {
	Get(ctx context.Context) (SimpleParser, error)
	Done(parser SimpleParser)
}

type parserPool struct {
	newParser ParserFactory
	pool      chan SimpleParser
}

func NewParserPool(newParser ParserFactory, numParserProcesses int) (ParserPool, error) {
	pool := make(chan SimpleParser, numParserProcesses)
	for i := 0; i < numParserProcesses; i++ {
		parser, err := newParser()
		if err != nil {
			return nil, err
		}
		pool <- parser
	}

	return &parserPool{
		newParser: newParser,
		pool:      pool,
	}, nil
}

// Get a parser from the pool. Once this parser is no longer in use, the Done method
// MUST be called with either the parser or a nil value (when countering an error).
// Nil values will be recreated on-demand via the factory supplied when constructing
// the pool. This method always returns a non-nil parser with a nil error value.
//
// This method blocks until a parser is available or the given context is canceled.
func (p *parserPool) Get(ctx context.Context) (SimpleParser, error) {
	select {
	case parser := <-p.pool:
		if parser != nil {
			return parser, nil
		}

		return p.newParser()

	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (p *parserPool) Done(parser SimpleParser) {
	p.pool <- parser
}
