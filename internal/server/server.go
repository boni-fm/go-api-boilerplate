package server

import (
	"context"
	"go-api-boilerplate/config"
	"go-api-boilerplate/internal/api/router"
	"go-api-boilerplate/internal/database"
	"go-api-boilerplate/internal/middleware"
	"go-api-boilerplate/internal/utility/fibererror"
	"go-api-boilerplate/internal/utility/swagger"
	"go-api-boilerplate/internal/worker"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/static"
)

// defaultWorkerCount is the number of goroutines in the background worker pool.
// Teams should tune this value based on their service's background-task volume.
const defaultWorkerCount = 4

// defaultWorkerCapacity is the maximum number of jobs that can queue up in the
// worker pool before new submissions are shed (dropped). 128 gives ~10 ms of
// buffer at 10 k tasks/sec without materialising as a memory concern.
const defaultWorkerCapacity = 128

type FiberConfig struct {
	AppName       string
	CaseSensitive bool
	ErrorHandler  fiber.ErrorHandler
	ReadTimeout   time.Duration
	WriteTimeout  time.Duration
	IdleTimeout   time.Duration
	BodyLimit     int
}

// Server owns the Fiber application, its configuration, the middleware
// dependency graph, the background worker pool, and the database registry.
// All fields should be treated as read-only after Start is called.
type Server struct {
	App            *fiber.App
	Cfg            *config.Config
	MiddlewareDeps *middleware.MiddlewareDependencies
	// Pool is the shared bounded worker pool for background tasks.
	Pool *worker.Pool
	// Registry holds the named database connections for multi-tenant routing.
	Registry *database.Registry
}

func NewFiberConfig(cfg config.Config) *FiberConfig {
	return &FiberConfig{
		AppName:       cfg.AppName,
		CaseSensitive: false,
		ErrorHandler:  fibererror.GlobalErrorHandler,
		ReadTimeout:   15 * time.Second,
		WriteTimeout:  15 * time.Second,
		IdleTimeout:   60 * time.Second,
		BodyLimit:     4 * 1024 * 1024, // 4 MB
	}
}

func NewServer(cfg config.Config, fiberCfg *FiberConfig) *Server {
	app := fiber.New(fiber.Config{
		AppName:       fiberCfg.AppName,
		CaseSensitive: fiberCfg.CaseSensitive,
		ErrorHandler:  fiberCfg.ErrorHandler,

		// Immutable must be true because handlers may pass context-derived
		// values (c.Params(), c.Query(), c.Get()) into the worker pool for
		// background processing. Without this flag, those values are backed
		// by mutable fasthttp byte slices that are recycled after the handler
		// returns, causing silent data corruption in background tasks.
		Immutable: true,

		// Setup timeouts and body limits
		ReadTimeout:  fiberCfg.ReadTimeout,
		WriteTimeout: fiberCfg.WriteTimeout,
		IdleTimeout:  fiberCfg.IdleTimeout,
		BodyLimit:    fiberCfg.BodyLimit,
	})

	return &Server{
		App:  app,
		Cfg:  &cfg,
		Pool: worker.New(defaultWorkerCount, defaultWorkerCapacity),
	}
}

func (s *Server) SetMiddlewareDeps(middlewareDeps *middleware.MiddlewareDependencies) {
	s.MiddlewareDeps = middlewareDeps
}

// SetRegistry stores the database registry so that middleware and routes can
// resolve tenant-specific connections at request time.
func (s *Server) SetRegistry(registry *database.Registry) {
	s.Registry = registry
}

func (s *Server) Start() error {
	// Launch background worker pool before accepting HTTP traffic so that
	// handlers can dispatch tasks from the very first request.
	s.Pool.Start(context.Background())

	s.MiddlewareDeps.InitAllMiddleware(s.Registry)
	router.SetupRoutes(s.MiddlewareDeps.Log, s.App, s.Pool, s.Registry, s.Cfg.IsDevelopment)

	if s.Cfg.IsDevelopment {
		swagger.SwaggerSetup()
	}

	// In Fiber v3 app.Static() was removed. Static files are served via the
	// static middleware registered as a catch-all route so that more specific
	// routes (API, swagger, health probes) take precedence.
	s.App.Get("/*", static.New("./static/public"))

	// DisableStartupMessage was moved from fiber.Config to ListenConfig in v3.
	return s.App.Listen(":"+s.Cfg.Port, fiber.ListenConfig{
		DisableStartupMessage: false,
	})
}
