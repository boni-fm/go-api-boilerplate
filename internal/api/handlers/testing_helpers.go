package handlers

import (
	"github.com/boni-fm/go-libsd3/pkg/log"
)

// NewHandlersRegistryForTest constructs a HandlersRegistry with the given
// dependencies and is intended for use in tests only. The userSvc parameter
// may be nil when the test does not exercise user-related handlers; a nil
// UserService will panic if those handlers are invoked.
func NewHandlersRegistryForTest(log_ *log.Logger, userSvc UserServiceIface) *HandlersRegistry {
	return &HandlersRegistry{
		log_:        log_,
		SwaggerDoc:  nil, // not needed for most handler tests
		UserService: userSvc,
	}
}
