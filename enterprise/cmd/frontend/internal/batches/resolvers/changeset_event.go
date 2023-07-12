package resolvers

import (
	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/batches/store"
	btypes "github.com/sourcegraph/sourcegraph/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
)

type changesetEventResolver struct {
	store             *store.Store
	changesetResolver *changesetResolver
	*btypes.ChangesetEvent
}

const changesetEventIDKind = "ChangesetEvent"

func marshalChangesetEventID(id int64) graphql.ID {
	return relay.MarshalID(changesetEventIDKind, id)
}

func (r *changesetEventResolver) ID() graphql.ID {
	return marshalChangesetEventID(r.ChangesetEvent.ID)
}

func (r *changesetEventResolver) CreatedAt() gqlutil.DateTime {
	return gqlutil.DateTime{Time: r.ChangesetEvent.CreatedAt}
}

func (r *changesetEventResolver) Changeset() graphqlbackend.ExternalChangesetResolver {
	return r.changesetResolver
}
