package redispool

import (
	"context"
	"fmt"

	"github.com/gomodule/redigo/redis"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var (
	rateLimitScript          *redis.Script
	tokenBucketGlobablPrefix = "v2:rate_limiters"
	bucketCapacity           = 1000
	bucketRefillRate         = 1000
)

type RateLimiter interface {
	GetTokensFromBucket(ctx context.Context, tokenBucketName string, amount int) (allowed bool, remianingTokens int, err error)
}
type rateLimiter struct {
	pool *redis.Pool
}

func NewRateLimiter() (RateLimiter, error) {
	var err error
	pool, ok := Store.Pool()
	if !ok {
		err = errors.New("unable to get redis connection")
	}
	return &rateLimiter{
		pool: pool,
	}, err
}

func (r *rateLimiter) GetTokensFromBucket(ctx context.Context, tokenBucketName string, amount int) (allowed bool, remianingTokens int, err error) {
	fmt.Printf("Getting tokens for: %s, amount requested: %d\n", tokenBucketName, amount)
	err = loadRateLimitScript()
	if err != nil {
		return false, 0, errors.Wrapf(err, "unable to get tokens from bucket %s", tokenBucketName)
	}

	result, err := rateLimitScript.DoContext(ctx, r.pool.Get(), fmt.Sprintf("%s:%s", tokenBucketGlobablPrefix, tokenBucketName), bucketCapacity, bucketRefillRate, amount)
	if err != nil {
		return false, 0, errors.Wrapf(err, "error while getting tokens from bucket %s", tokenBucketName)
	}

	response, ok := result.([]interface{})
	if !ok || len(response) != 2 {
		return false, 0, errors.New("unexpected response from Lua script")
	}

	allwd, ok := response[0].(int64)
	if !ok {
		return false, 0, errors.New("unexpected response for allowed")
	}

	remTokens, ok := response[1].(int64)
	if !ok {
		return false, 0, errors.New("unexpected response for tokens left")
	}
	fmt.Printf("Finished getting tokens for: %s, amount remaining: %d, allowed: %+v\n", tokenBucketName, remTokens, allwd == 1)

	return allwd == 1, int(remTokens), nil
}

func loadRateLimitScript() error {
	loadScriptSHA := func(pool *redis.Pool) error {
		s := redis.NewScript(1, rateLimitLuaScript)
		err := s.Load(pool.Get())
		if err != nil {
			return errors.Wrap(err, "failed to load rate limit script in redis")
		}
		rateLimitScript = s
		return nil
	}

	pool, ok := Store.Pool()
	if !ok {
		return errors.New("unable to get redis connection")
	}

	if rateLimitScript == nil {
		return loadScriptSHA(pool)
	}

	return nil
}

const rateLimitLuaScript = `local bucket_key = KEYS[1]
local capacity = tonumber(ARGV[1])
local refill_rate = tonumber(ARGV[2])
local request_tokens = tonumber(ARGV[3])

-- Check if the bucket exists.
local bucket_exists = redis.call('EXISTS', bucket_key)

-- If the bucket does not exist or has expired, replenish the bucket and set the new expiration time.
if bucket_exists == 0 then
    redis.call('SET', bucket_key, capacity)
    redis.call('EXPIRE', bucket_key, 30)
end

-- Get the current token count in the bucket.
local current_tokens = tonumber(redis.call('GET', bucket_key))

-- Check if there are enough tokens for the current request.
local allowed = (current_tokens >= request_tokens)

-- If the request is allowed, decrement the tokens for this request.
if allowed then
    redis.call('DECRBY', bucket_key, request_tokens)
end

return {allowed and 1 or 0, tonumber(redis.call('GET', bucket_key))}`
