// Package handlers provides HTTP request handlers and their shared dependency registry.
package handlers

import (
	"go-api-boilerplate/internal/api/repository"
	"go-api-boilerplate/internal/api/services"
	"go-api-boilerplate/internal/utility/swagger"
	"go-api-boilerplate/internal/worker"

	"github.com/boni-fm/go-libsd3/pkg/log"
)

// HandlersRegistry is the shared dependency container for all HTTP handlers.
// It is constructed once at startup and injected wherever handlers are wired.
//
// Concrete types are used directly (no service-level interfaces) to keep the
// dependency graph easy to follow for new engineers. Testability is preserved
// at the repository layer: inject a mock UserRepository into NewUserService.
type HandlersRegistry struct {
	log_        *log.Logger
	SwaggerDoc  *swagger.DocumentModifier
	UserService *services.UserService
	// Pool is the bounded background worker pool. Handlers may submit
	// fire-and-forget tasks (audit logs, metric flushes, email dispatch, etc.)
	// without blocking the HTTP response path. Pool may be nil in tests that
	// do not exercise background-task code paths.
	Pool *worker.Pool
}

// NewHandlersRegistry wires all handler dependencies:
//   - creates the concrete PostgreSQL repository
//   - wraps it in UserService (which owns bcrypt hashing)
//   - creates the SwaggerDoc document modifier for serving Swagger UI
//   - stores the shared worker pool for background task dispatch
func NewHandlersRegistry(log_ *log.Logger, pool *worker.Pool) *HandlersRegistry {
	repo := repository.NewPostgresUserRepository()
	return &HandlersRegistry{
		log_:        log_,
		SwaggerDoc:  swagger.NewDocumentModifier(),
		UserService: services.NewUserService(log_, repo),
		Pool:        pool,
	}
}
