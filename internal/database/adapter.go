package database

import (
	"context"
	"fmt"
	"sync"
	"time"

	pgsd3 "github.com/boni-fm/go-libsd3/pkg/db/postgres"
	logger "github.com/boni-fm/go-libsd3/pkg/log"
	lru "github.com/hashicorp/golang-lru/v2/expirable"
	"golang.org/x/sync/singleflight"
)

// ─────────────────────────────────────────────────────────────────
// AdapterConfig
// ─────────────────────────────────────────────────────────────────

type AdapterConfig struct {
	MaxDc            int           // LRU cap: max live tenant pool entries
	IdleTTL          time.Duration // TTL: evict entry idle longer than this
	EvictionInterval time.Duration
}

func DefaultAdapterConfig() AdapterConfig {
	return AdapterConfig{
		MaxDc:   15,
		IdleTTL: 20 * time.Minute,
	}
}

// registryKey builds the map key as "KodeDC:AppName".
// This ensures the same KodeDC used by two different AppNames
// gets two separate pool entries with correct application_name
// visible in pg_stat_activity.
func registryKey(kodeDc, appName string) string {
	return kodeDc + ":" + appName
}

// ─────────────────────────────────────────────────────────────────
// DcAdapter
// ─────────────────────────────────────────────────────────────────

type DcAdapter struct {
	mu sync.Mutex

	// sf deduplicates concurrent init calls for the same key.
	// Group is safe for concurrent use with no initialization needed.
	sf singleflight.Group

	// cache is the LRU+TTL store.
	// Key:   kodeDC string
	// Value: *pgsd3.Database
	//
	// expirable.LRU handles:
	//   - LRU eviction when size > MaxDc
	//   - TTL eviction when entry idle > IdleTTL
	//   - OnEvict callback to close the DB cleanly
	//   - thread-safe get/set/evict internally
	cache *lru.LRU[string, *pgsd3.Database]

	cm  *pgsd3.ConnectionManager
	cfg AdapterConfig

	appName string
	log     *logger.Logger
}

var (
	adapterInstance *DcAdapter
	adapterOnce     sync.Once
)

func GetDcAdapterWithCustomConfig(
	cfg AdapterConfig,
	appName string,
	log *logger.Logger,
) *DcAdapter {
	adapterOnce.Do(func() {
		a := &DcAdapter{
			cm:      pgsd3.NewConnectionManager(),
			cfg:     cfg,
			log:     log,
			appName: appName,
		}

		// OnEvict fires when an entry is removed for ANY reason:
		//   - LRU cap exceeded     (size > MaxDc)
		//   - TTL expired          (idle > IdleTTL)
		//   - manual Remove() call
		//   - cache.Purge()
		//
		// We close the connection through the pool so it is removed from
		// ConnectionPool.connections too. This is important: pool.Connect's
		// fast path checks cp.connections, so removing the entry there ensures
		// the next GetOrInit triggers a fresh NewDatabase() call.
		onEvict := func(key string, db *pgsd3.Database) {
			_ = pgsd3.GetConnectionPool().CloseConnection(key)
			fmt.Printf("🗑️  Evicted [%s] (LRU or TTL)\n", key)
		}

		a.cache = lru.NewLRU[string, *pgsd3.Database](
			cfg.MaxDc,
			onEvict,
			cfg.IdleTTL,
		)

		adapterInstance = a
	})
	return adapterInstance
}

func GetDcAdapter(
	appName string,
	log *logger.Logger,
) *DcAdapter {
	adapterOnce.Do(func() {
		a := &DcAdapter{
			cm:      pgsd3.NewConnectionManager(),
			cfg:     DefaultAdapterConfig(),
			log:     log,
			appName: appName,
		}

		onEvict := func(key string, db *pgsd3.Database) {
			_ = pgsd3.GetConnectionPool().CloseConnection(key)
			fmt.Printf("🗑️  Evicted [%s] (LRU or TTL)\n", key)
		}

		a.cache = lru.NewLRU[string, *pgsd3.Database](
			a.cfg.MaxDc,
			onEvict,
			a.cfg.IdleTTL,
		)

		adapterInstance = a
	})
	return adapterInstance
}

func (a *DcAdapter) DefaultDbConfig(kodeDc string) pgsd3.Config {
	return pgsd3.Config{
		KodeDC:  kodeDc,
		AppName: a.appName,
	}
}

// GetOrInit is the hot path.
//
// Flow:
//  1. check cache — hit with a live connection = return instantly
//  2. miss (or stale closed connection) = singleflight.Do(key, connectFn)
//     - first goroutine runs connectFn
//     - all concurrent goroutines for same key wait and share the result
//     - if connectFn errors → all get the error, next call retries fresh
//  3. on success = store in cache (starts LRU + TTL clock)
func (a *DcAdapter) GetOrInit(ctx context.Context, kodeDc string) (*pgsd3.Database, error) {
	// Fast path: cache hit — only return if the connection is still alive.
	// We must check IsClosed() because an eviction callback may have already
	// called db.Close() on this pointer while it was still in the cache
	// (race between TTL background goroutine and an incoming request).
	if db, ok := a.cache.Get(kodeDc); ok {
		if !db.IsClosed() {
			return db, nil
		}
		// Stale entry: evict it so the slow path creates a fresh connection.
		a.cache.Remove(kodeDc) // triggers onEvict → removes from ConnectionPool too
	}

	// Slow path: cache miss — deduplicate with singleflight
	result, err, _ := a.sf.Do(kodeDc, func() (interface{}, error) {
		// Register config into go-libsd3 — protected by mu
		// to prevent duplicate-key panic if two goroutines
		// race here for different keys simultaneously.
		// The error is intentionally ignored: RegisterConfig returns an error
		// if the config already exists, which is fine — it means a previous
		// successful registration is still in place.
		a.mu.Lock()
		_ = pgsd3.GetConnectionPool().RegisterConfig(
			kodeDc,
			a.DefaultDbConfig(kodeDc),
		)
		a.mu.Unlock()

		// Connect — this is the expensive operation.
		// pool.Connect handles the case where the config exists but the
		// connection was previously closed (creates a fresh *Database).
		db, err := a.cm.GetDB(ctx, kodeDc)
		if err != nil {
			return nil, fmt.Errorf("[%s] connect failed: %w", kodeDc, err)
		}

		// Store in cache INSIDE the singleflight fn so concurrent waiters
		// that call cache.Get() after this returns will find it immediately.
		a.cache.Add(kodeDc, db)
		fmt.Printf("✅ Connected [%s]\n", kodeDc)
		return db, nil
	})
	if err != nil {
		return nil, err
	}

	return result.(*pgsd3.Database), nil
}

// PreConnect connects at startup for known high-traffic KodeDCs.
// Call in main.go so the pool is warm before the first real request.
func (a *DcAdapter) PreConnect(ctx context.Context, kodeDc string) error {
	_, err := a.GetOrInit(ctx, kodeDc)
	return err
}

// Reset clears a failed or stale entry — next GetOrInit() retries.
func (a *DcAdapter) Reset(kodeDc, appName string) {
	key := registryKey(kodeDc, appName)
	a.cache.Remove(key) // triggers OnEvict → CloseConnection
	fmt.Printf("🔄 Reset [%s] — will reconnect on next request\n", key)
}

// Stats returns live pool stats from go-libsd3 for all active entries.
func (a *DcAdapter) Stats(ctx context.Context) map[string]interface{} {
	result := make(map[string]interface{})
	for _, key := range a.cache.Keys() {
		stats, err := pgsd3.GetConnectionPool().GetConnectionStats(ctx, key)
		if err != nil {
			result[key] = map[string]interface{}{"error": err.Error()}
		} else {
			result[key] = stats
		}
	}
	return result
}

// HealthCheck delegates to go-libsd3's HealthCheckAll.
func (a *DcAdapter) HealthCheck(ctx context.Context) map[string]error {
	return a.cm.HealthCheck(ctx)
}

// CloseAll graceful shutdown — closes every cached connection.
func (a *DcAdapter) CloseAll() {
	// Purge() calls OnEvict for every entry → CloseConnection for each.
	a.cache.Purge()
	_ = a.cm.CloseAllConnections()
	fmt.Println("🔌 All connections closed")
}
