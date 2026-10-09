package algorithms

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/redis/go-redis/v9"
)

func newTestRedis(t *testing.T) *redis.Client {
	t.Helper()
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		t.Skipf("redis not available: %v", err)
	}
	return rdb
}

func flushKeys(t *testing.T, rdb *redis.Client, pattern string) {
	t.Helper()
	ctx := context.Background()
	keys, _ := rdb.Keys(ctx, pattern).Result()
	if len(keys) > 0 {
		rdb.Del(ctx, keys...)
	}
}

func TestFixedWindow_Basic(t *testing.T) {
	rdb := newTestRedis(t)
	defer rdb.Close()
	flushKeys(t, rdb, "fw:test-fw:*")

	fw := &FixedWindow{}
	ctx := context.Background()
	limit := 5

	for i := 1; i <= 5; i++ {
		allowed, remaining, _ := fw.Check(ctx, rdb, "test-fw", "/test", limit, 60)
		if !allowed {
			t.Errorf("request %d: expected allowed, got denied", i)
		}
		if remaining != limit-i {
			t.Errorf("request %d: expected remaining=%d, got %d", i, limit-i, remaining)
		}
	}

	for i := 6; i <= 10; i++ {
		allowed, remaining, retryMs := fw.Check(ctx, rdb, "test-fw", "/test", limit, 60)
		if allowed {
			t.Errorf("request %d: expected denied, got allowed", i)
		}
		if remaining != 0 {
			t.Errorf("request %d: expected remaining=0, got %d", i, remaining)
		}
		if retryMs <= 0 {
			t.Errorf("request %d: expected retry_after_ms > 0, got %d", i, retryMs)
		}
	}
}

func TestSlidingWindow_Basic(t *testing.T) {
	rdb := newTestRedis(t)
	defer rdb.Close()
	flushKeys(t, rdb, "sw:test-sw:*")

	sw := &SlidingWindow{}
	ctx := context.Background()
	limit := 5

	for i := 1; i <= 5; i++ {
		allowed, _, _ := sw.Check(ctx, rdb, "test-sw", "/test", limit, 60)
		if !allowed {
			t.Errorf("request %d: expected allowed, got denied", i)
		}
	}

	for i := 6; i <= 10; i++ {
		allowed, remaining, retryMs := sw.Check(ctx, rdb, "test-sw", "/test", limit, 60)
		if allowed {
			t.Errorf("request %d: expected denied, got allowed", i)
		}
		if remaining != 0 {
			t.Errorf("request %d: expected remaining=0, got %d", i, remaining)
		}
		if retryMs <= 0 {
			t.Errorf("request %d: expected retry_after_ms > 0, got %d", i, retryMs)
		}
	}
}

func TestTokenBucket_Basic(t *testing.T) {
	rdb := newTestRedis(t)
	defer rdb.Close()
	flushKeys(t, rdb, "tb:test-tb-basic:*")

	tb := NewTokenBucket()
	ctx := context.Background()
	if err := tb.LoadScript(ctx, rdb); err != nil {
		t.Fatalf("LoadScript: %v", err)
	}

	capacity := 5

	for i := 1; i <= 5; i++ {
		allowed, _, _ := tb.Check(ctx, rdb, "test-tb-basic", "/test", capacity, 1, 60)
		if !allowed {
			t.Errorf("request %d: expected allowed, got denied", i)
		}
	}

	allowed, _, retryMs := tb.Check(ctx, rdb, "test-tb-basic", "/test", capacity, 1, 60)
	if allowed {
		t.Errorf("request 6: expected denied, got allowed")
	}
	if retryMs <= 0 {
		t.Errorf("request 6: expected retry_after_ms > 0, got %d", retryMs)
	}
}

func TestTokenBucket_NoConcurrentRaceCondition(t *testing.T) {
	rdb := newTestRedis(t)
	defer rdb.Close()
	flushKeys(t, rdb, "tb:test-race:*")

	tb := NewTokenBucket()
	ctx := context.Background()
	if err := tb.LoadScript(ctx, rdb); err != nil {
		t.Fatalf("LoadScript: %v", err)
	}

	capacity := 10
	goroutines := 50

	var wg sync.WaitGroup
	allowedCount := 0
	var mu sync.Mutex

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			allowed, _, _ := tb.Check(ctx, rdb,
				"test-race", fmt.Sprintf("/test-%d", id%1), // same key
				capacity, 1, 3600)
			if allowed {
				mu.Lock()
				allowedCount++
				mu.Unlock()
			}
		}(i)
	}

	wg.Wait()

	if allowedCount > capacity {
		t.Errorf("race condition: %d requests allowed, capacity was %d", allowedCount, capacity)
	}
	t.Logf("allowed %d/%d requests (capacity=%d) — no race condition", allowedCount, goroutines, capacity)
}
