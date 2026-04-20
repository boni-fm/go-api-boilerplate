// Package database provides the database connection layer, including a
// multi-tenant registry and request-scoped context helpers.
package database

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/boni-fm/go-libsd3/pkg/db/postgres"
	"github.com/boni-fm/go-libsd3/pkg/log"
)

// contextKey is the unexported type used to store a *postgres.Database
// in a context.Context, preventing collisions with other packages.
type contextKey struct{}

// Registry holds one or more named database connections for multi-DC /
// multi-tenant deployments. All methods are safe for concurrent use.
//
// The first connection registered becomes the default, which is used when
// the request does not specify a tenant key (X-Kunci header).
type Registry struct {
	mu         sync.RWMutex
	connections map[string]*postgres.Database
	defaultKey  string
}

// NewRegistry returns an empty, ready-to-use Registry.
func NewRegistry() *Registry {
	return &Registry{connections: make(map[string]*postgres.Database)}
}

// Register adds a named database connection. The first connection registered
// is automatically promoted to the default.
func (r *Registry) Register(kunci string, db *postgres.Database) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.connections[kunci] = db
	if r.defaultKey == "" {
		r.defaultKey = kunci
	}
}

// Get retrieves the connection for the given key. The second return value
// indicates whether the key was found in the registry.
func (r *Registry) Get(kunci string) (*postgres.Database, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	db, ok := r.connections[kunci]
	return db, ok
}

// Default returns the default (first-registered) database connection.
// Returns nil if no connections have been registered.
func (r *Registry) Default() *postgres.Database {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.connections[r.defaultKey]
}

// DefaultKey returns the key of the default connection.
func (r *Registry) DefaultKey() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.defaultKey
}

// InitDatabases connects to every kunci listed in the comma-separated
// kunci string (e.g. "g009sim,g010sim") and registers them in a new Registry.
// It panics via log.Panicf if any connection cannot be established.
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

// initSingle opens a single PostgreSQL connection pool for the given kunci.
func initSingle(kunci string, log_ *log.Logger) *postgres.Database {
	dbcfg := postgres.Config{KodeDC: kunci}
	db, err := postgres.NewDatabase(context.Background(), &dbcfg)
	if err != nil {
		log_.Panicf("Failed to connect to database [%s]: %v", kunci, err)
	}
	return db
}

// WithDB stores the given database connection in the context, making it
// available to repository methods via DBFromContext.
func WithDB(ctx context.Context, db *postgres.Database) context.Context {
	return context.WithValue(ctx, contextKey{}, db)
}

// DBFromContext retrieves the *postgres.Database previously stored by
// WithDB. Returns nil if no database has been set on the context.
func DBFromContext(ctx context.Context) *postgres.Database {
	db, _ := ctx.Value(contextKey{}).(*postgres.Database)
	return db
}

// ErrNoDB is returned by repository methods when no database connection
// is found in the request context (e.g. MultiTenantMiddleware was skipped).
var ErrNoDB = fmt.Errorf("no database in request context: ensure MultiTenantMiddleware is registered")
