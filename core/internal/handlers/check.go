package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Anand09-in/RateLimiter/core/internal/algorithms"
	"github.com/Anand09-in/RateLimiter/core/internal/cache"
	"github.com/Anand09-in/RateLimiter/core/internal/metrics"
	"github.com/Anand09-in/RateLimiter/core/internal/middleware"
	"github.com/redis/go-redis/v9"
)

type checkRequest struct {
	Endpoint string `json:"endpoint"`
}

type checkResponse struct {
	Allowed      bool `json:"allowed"`
	Remaining    int  `json:"remaining"`
	RetryAfterMs int  `json:"retry_after_ms"`
}

type CheckHandler struct {
	cache       *cache.RuleCache
	rdb         *redis.Client
	fixedWindow  *algorithms.FixedWindow
	slidingWindow *algorithms.SlidingWindow
	tokenBucket  *algorithms.TokenBucket
}

func NewCheckHandler(c *cache.RuleCache, rdb *redis.Client, tb *algorithms.TokenBucket) *CheckHandler {
	return &CheckHandler{
		cache:         c,
		rdb:           rdb,
		fixedWindow:   &algorithms.FixedWindow{},
		slidingWindow: &algorithms.SlidingWindow{},
		tokenBucket:   tb,
	}
}

func (h *CheckHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	client, _ := r.Context().Value(middleware.ClientContextKey).(*cache.Client)

	var req checkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Endpoint == "" {
		req.Endpoint = "*"
	}

	rule, ok := h.cache.GetRule(client.RuleID)
	if !ok {
		http.Error(w, `{"error":"rule not found"}`, http.StatusInternalServerError)
		return
	}

	limit := rule.LimitCount
	if client.CustomLimit != nil {
		limit = *client.CustomLimit
	}

	var allowed bool
	var remaining, retryAfterMs int

	switch rule.Algorithm {
	case "fixed":
		allowed, remaining, retryAfterMs = h.fixedWindow.Check(r.Context(), h.rdb, client.ID, req.Endpoint, limit, rule.WindowSecs)
	case "sliding":
		allowed, remaining, retryAfterMs = h.slidingWindow.Check(r.Context(), h.rdb, client.ID, req.Endpoint, limit, rule.WindowSecs)
	case "token":
		allowed, remaining, retryAfterMs = h.tokenBucket.Check(r.Context(), h.rdb, client.ID, req.Endpoint, limit, limit/rule.WindowSecs, rule.WindowSecs)
	default:
		allowed, remaining, retryAfterMs = true, limit, 0
	}

	result := "allowed"
	if !allowed {
		result = "denied"
	}

	duration := time.Since(start)
	metrics.RequestsTotal.WithLabelValues(client.ID, rule.Algorithm, result).Inc()
	metrics.CheckDuration.WithLabelValues(rule.Algorithm).Observe(duration.Seconds())

	statusCode := http.StatusOK
	if !allowed {
		statusCode = http.StatusTooManyRequests
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(checkResponse{
		Allowed:      allowed,
		Remaining:    remaining,
		RetryAfterMs: retryAfterMs,
	})
}
