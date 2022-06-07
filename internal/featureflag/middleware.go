package featureflag

import (
	"context"
	"net/http"

	"github.com/sourcegraph/sourcegraph/internal/actor"
)

type flagContextKey struct{}

type Store interface {
	GetFeatureFlag(context.Context, string) (*FeatureFlag, error)
	GetFeatureFlags(context.Context) ([]*FeatureFlag, error)
	GetUserFlag(ctx context.Context, userID int32, flagName string) (*bool, error)
	GetAnonymousUserFlag(ctx context.Context, anonymousUID string, flagName string) (*bool, error)
	GetGlobalFeatureFlag(ctx context.Context, flagName string) (*bool, error)
	GetOrgOverrideForUser(ctx context.Context, uid int32, flag string) (*Override, error)
	GetUserOverride(ctx context.Context, uid int32, flag string) (*Override, error)
}

// Middleware evaluates the feature flags for the current user and adds the
// feature flags to the current context.
func Middleware(ffs Store, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Cookie")
		next.ServeHTTP(w, r.WithContext(WithFlags(r.Context(), ffs)))
	})
}

// flagSetFetcher is a lazy fetcher for a FlagSet. It will fetch the flags as
// required, once they're loaded from the context. This pattern prevents us
// from loading feature flags on every request, even when we don't end up using
// them.
type flagSetFetcher struct {
	ffs Store
}

func (f *flagSetFetcher) evaluateForActor(ctx context.Context, a *actor.Actor, flagName string) (flag *bool, err error) {
	if a.IsAuthenticated() {
		if flag, err = f.ffs.GetUserFlag(ctx, a.UID, flagName); flag != nil {
			setEvaluatedFlagToCache(flagName, a, *flag)
		}
		return flag, err
	}

	if a.AnonymousUID != "" {
		if flag, err = f.ffs.GetAnonymousUserFlag(ctx, a.AnonymousUID, flagName); flag != nil {
			setEvaluatedFlagToCache(flagName, a, *flag)
		}
		return flag, err
	}

	if flag, err = f.ffs.GetGlobalFeatureFlag(ctx, flagName); flag != nil {
		setEvaluatedFlagToCache(flagName, a, *flag)
	}
	return flag, err
}

// EvaluateForActorFromContext evaluates value for the flag name passed
// for the actor from the context. It requires the context to be wrapped
// with *WithFlags*, otherwise it will return false by default. It also
// set the evaluated flags to redis cache which later can be used to pass
// feature flags context to event logs.
func EvaluateForActorFromContext(ctx context.Context, flagName string) (result bool) {
	result = false
	if flags := ctx.Value(flagContextKey{}); flags != nil {
		if value, _ := flags.(*flagSetFetcher).evaluateForActor(ctx, actor.FromContext(ctx), flagName); value != nil {
			result = *value
		}
	}
	return result
}

// FromContext returns a map of already evaluated flags and their values
// for the actor from the context. It required the context to be wrapped
// with *WithFlags*, otherwise it will return an empty map by default.
func FromContext(ctx context.Context) FlagSet {
	if flags := ctx.Value(flagContextKey{}); flags != nil {
		if f, err := flags.(*flagSetFetcher).ffs.GetFeatureFlags(ctx); err == nil {
			return getEvaluatedFlagSetFromCache(f, actor.FromContext(ctx))
		}
	}

	return FlagSet{}
}

// WithFlags adds a flag fetcher to the context so consumers of the
// returned context can use FromContext.
func WithFlags(ctx context.Context, ffs Store) context.Context {
	fetcher := &flagSetFetcher{ffs: ffs}
	return context.WithValue(ctx, flagContextKey{}, fetcher)
}
