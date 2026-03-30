package handlers

import (
	"github.com/boni-fm/go-libsd3/pkg/log"
)

// NewHandlersRegistryForTest constructs a HandlersRegistry with the given
// dependencies and is intended for use in tests only.
//
// Both userSvc and the Pool field are nil-safe in the handlers that are
// already implemented: passing nil for userSvc will panic only if a handler
// that exercises the UserService is called, and passing a nil Pool (the
// default) will skip any worker-pool dispatch code paths that guard with
// `if hr.Pool != nil`.
//
// Inject a mock ProfileService after construction when testing profile handlers:
//
//	hr := handlers.NewHandlersRegistryForTest(l, nil)
//	hr.ProfileService = &mockProfileSvc{...}
func NewHandlersRegistryForTest(log_ *log.Logger, userSvc UserServiceIface) *HandlersRegistry {
	return &HandlersRegistry{
		log_:           log_,
		SwaggerDoc:     nil, // not needed for most handler tests
		UserService:    userSvc,
		ProfileService: nil, // inject explicitly when testing profile handlers
		Pool:           nil, // tests that don't exercise background tasks can safely pass nil
	}
}
