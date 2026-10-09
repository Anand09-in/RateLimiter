package algorithms

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type FixedWindow struct{}

func (f *FixedWindow) Check(ctx context.Context, rdb *redis.Client, clientID, endpoint string, limit, windowSecs int) (allowed bool, remaining int, retryAfterMs int) {
	now := time.Now().Unix()
	bucket := now / int64(windowSecs)
	key := fmt.Sprintf("fw:%s:%s:%d", clientID, endpoint, bucket)

	pipe := rdb.Pipeline()
	incrCmd := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, time.Duration(windowSecs)*time.Second)
	if _, err := pipe.Exec(ctx); err != nil {
		return true, limit, 0
	}

	count := int(incrCmd.Val())
	if count > limit {
		windowEnd := (bucket + 1) * int64(windowSecs)
		retryMs := int((windowEnd - now) * 1000)
		return false, 0, retryMs
	}
	return true, limit - count, 0
}
