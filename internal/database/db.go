// Package database nyediain layer koneksi database,
// termasuk registry multi-tenant dan helper buat context request.
package database

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/boni-fm/go-libsd3/pkg/db/postgres"
	"github.com/boni-fm/go-libsd3/pkg/log"
)

// contextKey adalah tipe unexported buat nyimpen *postgres.Database
// di dalam context.Context, biar gak bentrok sama package lain.
type contextKey struct{}

// Registry nyimpen satu atau lebih koneksi database untuk multi-DC/multi-tenant.
// Semua method aman dipanggil dari banyak goroutine sekaligus.
//
// Koneksi pertama yang didaftarin otomatis jadi default —
// dipake kalau request gak spesifiin tenant key (header X-Kunci).
type Registry struct {
	mu          sync.RWMutex
	connections map[string]*postgres.Database
	defaultKey  string
}

// NewRegistry bikin Registry baru yang kosong dan siap dipakai.
func NewRegistry() *Registry {
	return &Registry{connections: make(map[string]*postgres.Database)}
}

// Register nambahin koneksi database dengan nama tertentu.
// Koneksi pertama yang didaftarin otomatis jadi default.
func (r *Registry) Register(kunci string, db *postgres.Database) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.connections[kunci] = db
	if r.defaultKey == "" {
		r.defaultKey = kunci
	}
}

// Get ngambil koneksi berdasarkan key. Return value kedua nunjukin
// apakah key-nya ketemu atau enggak.
func (r *Registry) Get(kunci string) (*postgres.Database, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	db, ok := r.connections[kunci]
	return db, ok
}

// Default ngambil koneksi default (yang pertama didaftarin).
// Kalau belum ada yang didaftarin, return nil.
func (r *Registry) Default() *postgres.Database {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.connections[r.defaultKey]
}

// DefaultKey ngambil key dari koneksi default.
func (r *Registry) DefaultKey() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.defaultKey
}

// Close nutup semua koneksi yang terdaftar — dipanggil pas graceful shutdown
// supaya connection pool ke PostgreSQL bersih gak ada yang ngambang.
// Error dikumpulin tapi gak fatal, toh prosesnya mau mati juga.
func (r *Registry) Close() []error {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var errs []error
	for k, db := range r.connections {
		if err := db.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close %s: %w", k, err))
		}
	}
	return errs
}

// dbConnectTimeout batas waktu maksimal buat establish koneksi database waktu startup.
// Kalau PostgreSQL gak bisa direach, proses langsung gagal — gak nunggu lama.
const dbConnectTimeout = 15 * time.Second

// InitDatabases konek ke semua kunci yang ada di string kunci (pisah koma,
// contoh: "g009sim,g010sim") terus daftarin ke Registry baru.
// Panic lewat log.Panicf kalau ada koneksi yang gagal.
func InitDatabases(kunci string, log_ *log.Logger) *Registry {
	registry := NewRegistry()
	for _, k := range strings.Split(kunci, ",") {
		k = strings.TrimSpace(k)
		if k == "" {
			continue
		}
		db := initSingle(k, log_)
		registry.Register(k, db)
		log_.Infof("Database connected: %s", k)
	}
	return registry
}

// initSingle buka satu koneksi PostgreSQL untuk kunci tertentu.
// Ada timeout biar gak nge-hang pas startup kalau DB-nya mati.
func initSingle(kunci string, log_ *log.Logger) *postgres.Database {
	ctx, cancel := context.WithTimeout(context.Background(), dbConnectTimeout)
	defer cancel()

	dbcfg := postgres.Config{KodeDC: kunci}
	db, err := postgres.NewDatabase(ctx, &dbcfg)
	if err != nil {
		log_.Panicf("Failed to connect to database [%s]: %v", kunci, err)
	}
	return db
}

// WithDB nyimpen koneksi database ke dalam context,
// biar bisa diambil lagi di repository via DBFromContext.
func WithDB(ctx context.Context, db *postgres.Database) context.Context {
	return context.WithValue(ctx, contextKey{}, db)
}

// DBFromContext ngambil *postgres.Database yang udah disimpen sama WithDB.
// Return nil kalau belum pernah di-set.
func DBFromContext(ctx context.Context) *postgres.Database {
	db, _ := ctx.Value(contextKey{}).(*postgres.Database)
	return db
}

// ErrNoDB dikembaliin sama repository kalau context-nya gak ada koneksi DB-nya.
var ErrNoDB = fmt.Errorf("no database in request context")
