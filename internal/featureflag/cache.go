package featureflag

import (
	"errors"
	"fmt"

	"github.com/gomodule/redigo/redis"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/redispool"
)

var (
	pool = redispool.Store
)

func getEvaluatedFlagSetFromCache(flags []*FeatureFlag, a *actor.Actor) FlagSet {
	flagSet := FlagSet{}

	c := pool.Get()
	defer c.Close()

	visitorID, err := getVisitorIDForActor(a)

	if err != nil {
		return flagSet
	}

	for _, flag := range flags {
		if value, err := redis.Bool(c.Do("HGET", getFlagCacheKey(flag.Name), visitorID)); err == nil {
			flagSet[flag.Name] = value
		}
	}

	return flagSet
}

func setEvaluatedFlagToCache(name string, a *actor.Actor, value bool) {
	c := pool.Get()
	defer c.Close()

	var visitorID string

	visitorID, err := getVisitorIDForActor(a)

	if err != nil {
		return
	}

	c.Do("HSET", getFlagCacheKey(name), visitorID, fmt.Sprintf("%v", value))
}

// TODO: discuss if we should clear when feature flag is updated?
// Maybe we can still keep in cache until it is re-evaluated by new request from user/client?
// However, from GQL api flag name can be changed as well, but in admin UI only value can change actually
func ClearFlagFromCache(name string) {
	c := pool.Get()
	defer c.Close()

	c.Do("DEL", getFlagCacheKey(name))
}

// TODO: discuss if we should clear when feature flag is updated?
// Maybe we can still keep in cache until it is re-evaluated by new request from user/client?
// Because, technically user is still using old flag, until makes new evaluate request
func ClearFlagForOverrideFromCache(name string, userIDs []*int32) {
	c := pool.Get()
	defer c.Close()

	for _, userID := range userIDs {
		c.Do("HDEL", getFlagCacheKey(name), fmt.Sprintf("uid_%v", userID))
	}
}

func getVisitorIDForActor(a *actor.Actor) (string, error) {
	if a.IsAuthenticated() {
		return fmt.Sprintf("uid_%v", a.UID), nil
	} else if a.AnonymousUID != "" {
		return "auid_" + a.AnonymousUID, nil
	} else {
		return "", errors.New("UID/AnonymousUID are emptry for the given actor.")
	}
}

func getFlagCacheKey(name string) string {
	return "ff_" + name
}
