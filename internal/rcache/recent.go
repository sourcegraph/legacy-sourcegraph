package rcache

import (
	"context"
	"fmt"
	"unicode/utf8"

	"github.com/gomodule/redigo/redis"
	"github.com/inconshreveable/log15"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// RecentCache holds the most recently inserted items, discarding older ones if the total item count goes over the configured size.
type RecentCache struct {
	key  string
	size int
}

// NewRecent returns a RecentCache, storing only a fixed amount of elements, discarding old ones if needed.
func NewRecent(key string, size int) *RecentCache {
	return &RecentCache{
		key:  key,
		size: size,
	}
}

// Insert b in the cache and drops the last recently inserted item if the size exceeds the configured limit.
func (q *RecentCache) Insert(b []byte) {
	c := pool.Get()
	defer c.Close()

	if !utf8.Valid(b) {
		log15.Error("rcache: keys must be valid utf8", "key", b)
	}
	key := q.globalPrefixKey()

	// O(1) because we're just adding a single element.
	_, err := c.Do("LPUSH", key, b)
	if err != nil {
		log15.Warn("failed to execute redis command", "cmd", "LPUSH", "error", err)
	}

	// O(1) because the average case if just about dropping the last element.
	_, err = c.Do("LTRIM", key, 0, q.size-1)
	if err != nil {
		log15.Warn("failed to execute redis command", "cmd", "LTRIM", "error", err)
	}
}

// All return all items stored in the RecentCache.
//
// This a O(n) operation, where n is the list size.
func (q *RecentCache) All(ctx context.Context) ([][]byte, error) {
	c, err := pool.GetContext(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "get redis conn")
	}
	defer c.Close()
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	key := q.globalPrefixKey()
	res, err := redis.Values(c.Do("LRANGE", key, 0, -1))
	if err != nil {
		return nil, err
	}
	bs, err := redis.ByteSlices(res, nil)
	if err != nil {
		return nil, err
	}
	if len(bs) > q.size {
		bs = bs[:q.size]
	}
	return bs, nil
}

func (q *RecentCache) globalPrefixKey() string {
	return fmt.Sprintf("%s:%s", globalPrefix, q.key)
}
