package handlers

import (
	"go-api-boilerplate/internal/api/services"

	"github.com/boni-fm/go-libsd3/pkg/log"
)

// NewHandlersRegistryForTest constructs a HandlersRegistry with the given
// dependencies and is intended for use in tests only.
//
// Pass a *services.UserService built with a mock repository to exercise
// handler logic in isolation. Passing nil for userSvc is safe for tests
// that do not call any user-domain endpoints (e.g. PingPong tests).
// Pool is always nil in test registries; handlers guard pool dispatch with
// `if hr.Pool != nil`, so tests that don't exercise background tasks are safe.
func NewHandlersRegistryForTest(log_ *log.Logger, userSvc *services.UserService) *HandlersRegistry {
	return &HandlersRegistry{
		log_:        log_,
		SwaggerDoc:  nil, // not needed for most handler tests
		UserService: userSvc,
		Pool:        nil,
	}
}
