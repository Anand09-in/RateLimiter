package cache

import (
	"context"
	"database/sql"
	"log/slog"
	"sync"
	"time"
)

type Rule struct {
	ID          string
	Name        string
	Algorithm   string
	LimitCount  int
	WindowSecs  int
	Endpoint    string
	IsActive    bool
}

type Client struct {
	ID          string
	Name        string
	APIKey      string
	RuleID      string
	CustomLimit *int
	IsActive    bool
}

type RuleCache struct {
	mu       sync.RWMutex
	rules    map[string]*Rule   // key: rule_id
	clients  map[string]*Client // key: api_key
	lastSync time.Time
}

func New() *RuleCache {
	return &RuleCache{
		rules:   make(map[string]*Rule),
		clients: make(map[string]*Client),
	}
}

func (c *RuleCache) GetByAPIKey(key string) (*Client, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	cl, ok := c.clients[key]
	return cl, ok
}

func (c *RuleCache) GetRule(ruleID string) (*Rule, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	r, ok := c.rules[ruleID]
	return r, ok
}

func (c *RuleCache) LastSync() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lastSync
}

func (c *RuleCache) Refresh(db *sql.DB) error {
	rows, err := db.Query(`SELECT id, name, algorithm, limit_count, window_secs, endpoint, is_active FROM rules WHERE is_active = true`)
	if err != nil {
		return err
	}
	defer rows.Close()

	rules := make(map[string]*Rule)
	for rows.Next() {
		r := &Rule{}
		if err := rows.Scan(&r.ID, &r.Name, &r.Algorithm, &r.LimitCount, &r.WindowSecs, &r.Endpoint, &r.IsActive); err != nil {
			return err
		}
		rules[r.ID] = r
	}

	crows, err := db.Query(`SELECT id, name, api_key, rule_id, custom_limit, is_active FROM clients WHERE is_active = true`)
	if err != nil {
		return err
	}
	defer crows.Close()

	clients := make(map[string]*Client)
	for crows.Next() {
		cl := &Client{}
		if err := crows.Scan(&cl.ID, &cl.Name, &cl.APIKey, &cl.RuleID, &cl.CustomLimit, &cl.IsActive); err != nil {
			return err
		}
		clients[cl.APIKey] = cl
	}

	c.mu.Lock()
	c.rules = rules
	c.clients = clients
	c.lastSync = time.Now()
	c.mu.Unlock()
	return nil
}

func (c *RuleCache) StartRefresh(ctx context.Context, db *sql.DB, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				t0 := time.Now()
				if err := c.Refresh(db); err != nil {
					slog.Warn("cache refresh failed, keeping stale cache", "error", err)
				} else {
					c.mu.RLock()
					slog.Info("cache refreshed",
						"rules_count", len(c.rules),
						"clients_count", len(c.clients),
						"duration_ms", time.Since(t0).Milliseconds(),
					)
					c.mu.RUnlock()
				}
			}
		}
	}()
}
