package algorithms

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

const tokenBucketScript = `
local key = KEYS[1]
local capacity = tonumber(ARGV[1])
local refill_rate = tonumber(ARGV[2])
local now = tonumber(ARGV[3])

local bucket = redis.call('HMGET', key, 'tokens', 'last_refill')
local tokens = tonumber(bucket[1]) or capacity
local last_refill = tonumber(bucket[2]) or now

local elapsed = (now - last_refill) / 1000
local new_tokens = math.min(capacity, tokens + elapsed * refill_rate)

if new_tokens < 1 then
  return {0, 0, math.ceil((1 - new_tokens) / refill_rate * 1000)}
end

new_tokens = new_tokens - 1
redis.call('HMSET', key, 'tokens', new_tokens, 'last_refill', now)
redis.call('EXPIRE', key, 3600)
return {1, math.floor(new_tokens), 0}
`

type TokenBucket struct {
	scriptSHA string
}

func NewTokenBucket() *TokenBucket {
	return &TokenBucket{}
}

func (t *TokenBucket) LoadScript(ctx context.Context, rdb *redis.Client) error {
	sha, err := rdb.ScriptLoad(ctx, tokenBucketScript).Result()
	if err != nil {
		return err
	}
	t.scriptSHA = sha
	return nil
}

func (t *TokenBucket) Check(ctx context.Context, rdb *redis.Client, clientID, endpoint string, capacity, refillRate, windowSecs int) (allowed bool, remaining int, retryAfterMs int) {
	key := fmt.Sprintf("tb:%s:%s", clientID, endpoint)

	// Use Redis TIME for clock source
	ts, err := rdb.Time(ctx).Result()
	if err != nil {
		return true, capacity, 0
	}
	nowMs := ts.UnixMilli()

	var result []interface{}

	if t.scriptSHA != "" {
		res, err := rdb.EvalSha(ctx, t.scriptSHA, []string{key},
			capacity, refillRate, nowMs).Result()
		if err != nil && err.Error() == "NOSCRIPT No matching script. Please use EVAL." {
			// fallback: reload and retry
			t.LoadScript(ctx, rdb) //nolint:errcheck
			res, err = rdb.EvalSha(ctx, t.scriptSHA, []string{key},
				capacity, refillRate, nowMs).Result()
		}
		if err != nil {
			return true, capacity, 0
		}
		result, _ = res.([]interface{})
	} else {
		res, err := rdb.Eval(ctx, tokenBucketScript, []string{key},
			capacity, refillRate, nowMs).Result()
		if err != nil {
			return true, capacity, 0
		}
		result, _ = res.([]interface{})
	}

	if len(result) != 3 {
		return true, capacity, 0
	}

	allowedInt, _ := result[0].(int64)
	rem, _ := result[1].(int64)
	retry, _ := result[2].(int64)
	return allowedInt == 1, int(rem), int(retry)
}
