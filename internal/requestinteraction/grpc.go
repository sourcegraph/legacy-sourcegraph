package requestinteraction

import (
	"context"

	"google.golang.org/grpc/metadata"

	internalgrpc "github.com/sourcegraph/sourcegraph/internal/grpc/propagator"
)

type Propagator struct{}

func (Propagator) FromContext(ctx context.Context) metadata.MD {
	interaction := FromContext(ctx)
	if interaction == nil {
		return metadata.New(nil)
	}

	return metadata.Pairs(
		headerKeyInteractionID, interaction.ID,
	)
}

func (Propagator) InjectContext(ctx context.Context, md metadata.MD) (context.Context, error) {
	if vals := md.Get(headerKeyInteractionID); len(vals) > 0 {
		id := vals[0]
		return WithClient(ctx, &Interaction{
			ID: id,
		}), nil
	}

	return ctx, nil
}

var _ internalgrpc.Propagator = Propagator{}
