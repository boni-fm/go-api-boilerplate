// Package handlers provides HTTP request handlers and their shared dependency registry.
package handlers

import (
	"go-api-boilerplate/internal/api/repository"
	"go-api-boilerplate/internal/api/services"
	"go-api-boilerplate/internal/utility/swagger"

	"github.com/boni-fm/go-libsd3/pkg/log"
)

// HandlersRegistry is the shared dependency container for all HTTP handlers.
// It is constructed once at startup and injected wherever handlers are wired.
type HandlersRegistry struct {
	log_        *log.Logger
	SwaggerDoc  *swagger.DocumentModifier
	UserService UserServiceIface
}

// NewHandlersRegistry wires all handler dependencies:
//   - creates the concrete PostgreSQL repository
//   - wraps it in UserService (which owns bcrypt hashing)
//   - stores references to shared utilities
func NewHandlersRegistry(log_ *log.Logger) *HandlersRegistry {
	repo := repository.NewPostgresUserRepository()
	return &HandlersRegistry{
		log_:        log_,
		SwaggerDoc:  swagger.NewDocumentModifier(),
		UserService: services.NewUserService(log_, repo),
	}
}
