package algorithms

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type FixedWindow struct{}

func (f *FixedWindow) Check(ctx context.Context, rdb *redis.Client, clientID, endpoint string, limit, windowSecs int) (allowed bool, remaining int, retryAfterMs int) {
	// stub — implemented in Phase 4
	return true, limit, 0
}
