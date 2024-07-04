package graphql

import (
	"context"
	"github.com/sourcegraph/scip/bindings/go/scip"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const DefaultDefinitionsPageSize = 100

// Definitions returns the list of source locations that define the symbol at the given position.
func (r *gitBlobLSIFDataResolver) Definitions(ctx context.Context, args *resolverstubs.LSIFQueryPositionArgs) (_ resolverstubs.LocationConnectionResolver, err error) {
	requestArgs := codenav.OccurrenceRequestArgs{
		RepositoryID: r.requestState.RepositoryID,
		Commit:       r.requestState.Commit,
		Path:         r.requestState.Path,
		Limit:        DefaultDefinitionsPageSize,
		// Cursor is zero value as this API has historically not supported pagination.
		RawCursor: "",
		Matcher:   codenav.NewStartPositionMatcher(scip.Position{Line: args.Line, Character: args.Character}),
	}
	ctx, _, endObservation := observeResolver(ctx, &err, r.operations.definitions, time.Second, observation.Args{Attrs: requestArgs.Attrs()})
	defer endObservation()

	// NOTE: We don't support pagination for definitions in the GraphQL API.
	def, _, err := r.codeNavSvc.GetDefinitions(ctx, requestArgs, r.requestState, codenav.Cursor{})
	if err != nil {
		return nil, errors.Wrap(err, "codeNavSvc.GetDefinitions")
	}

	if args.Filter != nil && *args.Filter != "" {
		filtered := def[:0]
		for _, loc := range def {
			if strings.Contains(loc.Path.RawValue(), *args.Filter) {
				filtered = append(filtered, loc)
			}
		}
		def = filtered
	}

	return newLocationConnectionResolver(def, nil, r.locationResolver), nil
}
