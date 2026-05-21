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
	MaxDc            int           // maks jumlah tenant aktif di LRU
	IdleTTL          time.Duration // berapa lama idle sebelum di-evict
	EvictionInterval time.Duration
}

func DefaultAdapterConfig() AdapterConfig {
	return AdapterConfig{
		MaxDc:            15,
		IdleTTL:          20 * time.Minute,
		EvictionInterval: 5 * time.Minute,
	}
}

// ─────────────────────────────────────────────────────────────────
// DcAdapter
// ─────────────────────────────────────────────────────────────────

type DcAdapter struct {
	mu sync.Mutex

	// sf deduplikasi concurrent init buat key yang sama.
	sf singleflight.Group

	// cache adalah LRU+TTL store.
	// Key:   kodeDC string
	// Value: *pgsd3.Database
	//
	// expirable.LRU handle:
	//   - evict LRU kalau size > MaxDc
	//   - evict TTL kalau idle > IdleTTL
	//   - callback OnEvict buat nutup DB yang keluarkan
	//   - thread-safe secara internal
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

		// onEvict dipanggil tiap ada entry yang keluar — entah karena:
		//   - LRU penuh (size > MaxDc)
		//   - TTL habis (idle > IdleTTL)
		//   - dipanggil manual Remove() atau Purge()
		//
		// Nutup koneksi lewat pool biar entry-nya juga ilang dari ConnectionPool.
		// Ini penting supaya Connect() fast path gak balik pointer yang udah mati.
		onEvict := func(key string, db *pgsd3.Database) {
			_ = pgsd3.GetConnectionPool().CloseConnection(key)
			log.Infof("Evicted [%s] (LRU or TTL)", key)
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
			log.Infof("Evicted [%s] (LRU or TTL)", key)
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

// GetOrInit adalah hot path buat dapetin koneksi DB.
//
// Alurnya:
//  1. cek cache — kalau ada dan masih hidup, langsung return
//  2. miss atau koneksi mati → singleflight.Do(key, connectFn)
//     - goroutine pertama jalanin connectFn
//     - goroutine lain yang minta key sama nunggu dan numpang hasil yang sama
//     - kalau connectFn error → semua dapat error, request berikutnya retry fresh
//  3. sukses → simpen ke cache (mulai clock LRU + TTL)
func (a *DcAdapter) GetOrInit(ctx context.Context, kodeDc string) (*pgsd3.Database, error) {
	// Fast path: cache hit — cek dulu IsClosed() biar gak balik koneksi mati.
	// Bisa terjadi race antara TTL eviction goroutine dan request yang masuk.
	if db, ok := a.cache.Get(kodeDc); ok {
		if !db.IsClosed() {
			return db, nil
		}
		// Entry udah mati, buang — slow path bakal bikin koneksi baru.
		a.cache.Remove(kodeDc) // trigger onEvict → hapus dari ConnectionPool juga
	}

	// Slow path: cache miss — deduplikasi pakai singleflight
	result, err, _ := a.sf.Do(kodeDc, func() (interface{}, error) {
		// Daftarin config ke go-libsd3
		// supaya gak ada panic duplicate-key kalau dua goroutine
		// race buat key yang berbeda sekaligus.
		// Error diabaikan: RegisterConfig error kalau key udah ada — it's fine.
		a.mu.Lock()
		_ = pgsd3.GetConnectionPool().RegisterConfig(
			kodeDc,
			a.DefaultDbConfig(kodeDc),
		)
		a.mu.Unlock()

		// Connect, connect, connect ...
		// pool.Connect handle kasus di mana config udah ada tapi koneksinya mati.
		db, err := a.cm.GetDB(ctx, kodeDc)
		if err != nil {
			a.log.Errorf("Failed to connect [%s]: %v", kodeDc, err)
			return nil, fmt.Errorf("[%s] connect failed: %w", kodeDc, err)
		}

		// Simpen ke cache didalam singleflight fn supaya goroutine lain
		// yang nunggu langsung bisa nemu di cache setelah ini return.
		a.cache.Add(kodeDc, db)
		a.log.Infof("Connected [%s]", kodeDc)
		return db, nil
	})
	if err != nil {
		return nil, err
	}

	return result.(*pgsd3.Database), nil
}

// PreConnect konek di awal startup buat KodeDC yang udah ketebak bakal rame.
// Dipanggil di main.go biar pool udah warm sebelum request pertama masuk.
func (a *DcAdapter) PreConnect(ctx context.Context, kodeDc string) error {
	_, err := a.GetOrInit(ctx, kodeDc)
	return err
}

// Reset buang entry yang gagal atau stale — GetOrInit() berikutnya bakal retry.
func (a *DcAdapter) Reset(kodeDc string) {
	a.cache.Remove(kodeDc) // trigger OnEvict → CloseConnection
	a.log.Infof("Reset KodeDC [%s]", kodeDc)
}

// Stats ngambil statistik pool dari go-libsd3 buat semua entry yang aktif.
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

// HealthCheck delegasi ke go-libsd3's HealthCheckAll.
func (a *DcAdapter) HealthCheck(ctx context.Context) map[string]error {
	return a.cm.HealthCheck(ctx)
}

// CloseAll graceful shutdown — nutup semua koneksi yang ada di cache.
// Purge() manggil OnEvict tiap entry → CloseConnection masing-masing.
func (a *DcAdapter) CloseAll() {
	a.cache.Purge()
	_ = a.cm.CloseAllConnections()
	a.log.Infof("Database [%s] closed", a.appName)
}
