package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/atharvanand/rate-limiter/core/internal/algorithms"
	"github.com/atharvanand/rate-limiter/core/internal/cache"
	"github.com/atharvanand/rate-limiter/core/internal/handlers"
	"github.com/atharvanand/rate-limiter/core/internal/metrics"
	"github.com/atharvanand/rate-limiter/core/internal/middleware"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	redisAddr := getenv("REDIS_ADDR", "localhost:6379")
	postgresDSN := getenv("POSTGRES_DSN", "postgres://rluser:rlpass@localhost:5432/ratelimiter?sslmode=disable")
	refreshInterval := getenvDuration("CACHE_REFRESH_INTERVAL", 30*time.Second)

	rdb := redis.NewClient(&redis.Options{Addr: redisAddr})
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		slog.Error("redis connection failed", "error", err)
		os.Exit(1)
	}
	slog.Info("connected to redis", "addr", redisAddr)

	db, err := sql.Open("postgres", postgresDSN)
	if err != nil {
		slog.Error("postgres open failed", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		slog.Warn("postgres ping failed, cache will start empty", "error", err)
	} else {
		slog.Info("connected to postgres")
	}

	metrics.Register()

	ruleCache := cache.New()
	// best-effort initial load; fail-open if Postgres not ready
	if err := ruleCache.Refresh(db); err != nil {
		slog.Warn("initial cache load failed", "error", err)
	}

	cancelCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	ruleCache.StartRefresh(cancelCtx, db, refreshInterval)

	// update cache age metric every second
	go func() {
		t := time.NewTicker(time.Second)
		defer t.Stop()
		for {
			select {
			case <-cancelCtx.Done():
				return
			case <-t.C:
				age := time.Since(ruleCache.LastSync()).Seconds()
				metrics.CacheAgeSeconds.Set(age)
			}
		}
	}()

	tb := &algorithms.TokenBucket{}
	if err := tb.LoadScript(ctx, rdb); err != nil {
		slog.Warn("token bucket script load failed", "error", err)
	}

	r := mux.NewRouter()

	r.Handle("/metrics", promhttp.Handler()).Methods(http.MethodGet)

	r.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}).Methods(http.MethodGet)

	checkHandler := handlers.NewCheckHandler(ruleCache, rdb, tb)
	r.Handle("/check", middleware.Auth(ruleCache)(checkHandler)).Methods(http.MethodPost)

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		slog.Info("core service listening", "addr", ":8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	slog.Info("shutting down")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("graceful shutdown failed", "error", err)
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getenvDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		d, err := time.ParseDuration(v)
		if err == nil {
			return d
		}
	}
	return fallback
}
