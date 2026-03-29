package database

import (
	"context"

	"github.com/boni-fm/go-libsd3/pkg/db/postgres"
	"github.com/boni-fm/go-libsd3/pkg/log"
)

// Db is the application-wide database connection pool.
// It is initialised once at startup via InitDatabase and is safe
// for concurrent use after that point.
var Db *postgres.Database

// InitDatabase opens a PostgreSQL connection pool using the provided key.
// It panics (via log.Panicf) if the connection cannot be established,
// since the application cannot function without a database.
func InitDatabase(kunci string, log *log.Logger) {
	dbcfg := postgres.Config{
		KodeDC: kunci,
	}
	db, err := postgres.NewDatabase(context.Background(), &dbcfg)
	if err != nil {
		// Logger embeds *logrus.Logger; logrus.Panicf logs at PanicLevel and
		// then calls panic() internally — no separate panic() call is needed.
		// See: github.com/sirupsen/logrus entry.go, log() method, PanicLevel branch.
		log.Panicf("Failed to connect to database: %v", err)
	}
	Db = db
}

// GetDatabase returns the shared database connection pool.
// Callers that need the pool directly (e.g. repository implementations)
// may use this helper instead of accessing the package-level variable.
func GetDatabase() *postgres.Database {
	return Db
}
