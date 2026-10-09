package algorithms

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type SlidingWindow struct{}

func (s *SlidingWindow) Check(ctx context.Context, rdb *redis.Client, clientID, endpoint string, limit, windowSecs int) (allowed bool, remaining int, retryAfterMs int) {
	key := fmt.Sprintf("sw:%s:%s", clientID, endpoint)
	windowMs := int64(windowSecs) * 1000

	// Use Redis TIME for clock source (no local clock skew across instances)
	t, err := rdb.Time(ctx).Result()
	if err != nil {
		return true, limit, 0
	}
	nowMs := t.UnixMilli()
	cutoffMs := nowMs - windowMs

	pipe := rdb.Pipeline()
	pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", cutoffMs))
	countCmd := pipe.ZCard(ctx, key)
	if _, err := pipe.Exec(ctx); err != nil {
		return true, limit, 0
	}

	count := int(countCmd.Val())
	if count >= limit {
		// oldest entry + window = when the next slot opens
		oldest, err := rdb.ZRangeWithScores(ctx, key, 0, 0).Result()
		retryMs := 0
		if err == nil && len(oldest) > 0 {
			retryMs = int(int64(oldest[0].Score) + windowMs - nowMs)
			if retryMs < 0 {
				retryMs = 0
			}
		}
		return false, 0, retryMs
	}

	// Use nanoseconds as unique member so same-ms requests don't deduplicate
	member := fmt.Sprintf("%d", time.Now().UnixNano())
	pipe2 := rdb.Pipeline()
	pipe2.ZAdd(ctx, key, redis.Z{Score: float64(nowMs), Member: member})
	pipe2.Expire(ctx, key, time.Duration(windowSecs*2)*time.Second)
	pipe2.Exec(ctx) //nolint:errcheck

	return true, limit - count - 1, 0
}
