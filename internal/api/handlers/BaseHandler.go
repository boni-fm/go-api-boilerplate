package handlers

import (
	"go-api-boilerplate/config"
	"go-api-boilerplate/internal/api/services"
	"go-api-boilerplate/internal/database"
	"go-api-boilerplate/internal/utility/injector"
	"go-api-boilerplate/internal/utility/swagger"
	"go-api-boilerplate/internal/worker"

	"github.com/boni-fm/go-libsd3/pkg/log"
)

// HandlersRegistry
// HandlersRegistry is the shared dependency container for all HTTP handlers.
// It is constructed once at startup and injected wherever handlers are wired.
//
// Concrete types are used directly (no service-level interfaces) to keep the
// dependency graph easy to follow for new engineers. Testability is preserved
// at the repository layer: inject a mock UserRepository into NewUserService.
type HandlersRegistry struct {
	// Depedencies
	// Kalau ada lagi bisa ditambah disini
	log_      *log.Logger
	dbManager *database.DcAdapter
	cfg       *config.Config

	// hehe ini ngapain hayo~
	dbInjector injector.DBInjector

	SwaggerDoc  *swagger.DocumentModifier
	UserService *injector.ServiceFactory[services.UserService]

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
func NewHandlersRegistry(log_ *log.Logger, pool *worker.Pool, manager *database.DcAdapter, cfg *config.Config, dbInject injector.DBInjector) *HandlersRegistry {
	//repo := repository.NewPostgresUserRepository()
	return &HandlersRegistry{
		// Inject beberapa depedencies yang sekiranya
		// berguna didalam handler
		//
		// Note ::
		// jika ingin nambahin yang baru ke dalam service, perlu diperhatikan injector nya
		// perlu ditambahin ke dalam Service Factory nya ...
		log_:      log_,
		dbManager: manager,
		cfg:       cfg,

		SwaggerDoc: swagger.NewDocumentModifier(),
		Pool:       pool,

		// SERVICE COLLECTION nya ~~
		UserService: injector.NewServiceFactory[services.UserService](
			dbInject,
			log_,
			cfg,
			services.NewUserService,
		),
	}
}
