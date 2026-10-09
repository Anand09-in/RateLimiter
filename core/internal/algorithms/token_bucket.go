package algorithms

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type TokenBucket struct {
	scriptSHA string
}

func (t *TokenBucket) LoadScript(ctx context.Context, rdb *redis.Client) error {
	// stub — implemented in Phase 4
	return nil
}

func (t *TokenBucket) Check(ctx context.Context, rdb *redis.Client, clientID, endpoint string, capacity, ratePer int, windowSecs int) (allowed bool, remaining int, retryAfterMs int) {
	// stub — implemented in Phase 4
	return true, capacity, 0
}
