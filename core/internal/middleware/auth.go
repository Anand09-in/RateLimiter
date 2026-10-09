package middleware

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/atharvanand/rate-limiter/core/internal/cache"
)

type contextKey string

const ClientContextKey contextKey = "client"

func Auth(c *cache.RuleCache) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			apiKey := r.Header.Get("X-API-Key")
			if apiKey == "" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]string{"error": "missing api key"})
				return
			}

			client, ok := c.GetByAPIKey(apiKey)
			if !ok || !client.IsActive {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]string{"error": "invalid api key"})
				return
			}

			ctx := context.WithValue(r.Context(), ClientContextKey, client)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
