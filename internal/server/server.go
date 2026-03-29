package server

import (
	"context"
	"go-api-boilerplate/config"
	"go-api-boilerplate/internal/api/router"
	"go-api-boilerplate/internal/middleware"
	"go-api-boilerplate/internal/utility/fibererror"
	"go-api-boilerplate/internal/utility/swagger"
	"go-api-boilerplate/internal/worker"
	"time"

	"github.com/gofiber/fiber/v2"
)

// defaultWorkerCount is the number of goroutines in the background worker pool.
// Teams should tune this value based on their service's background-task volume.
const defaultWorkerCount = 4

// defaultWorkerCapacity is the maximum number of jobs that can queue up in the
// worker pool before new submissions are shed (dropped). 128 gives ~10 ms of
// buffer at 10 k tasks/sec without materialising as a memory concern.
const defaultWorkerCapacity = 128

type FiberConfig struct {
	AppName               string
	DisableStartupMessage bool
	CaseSensitive         bool
	ErrorHandler          fiber.ErrorHandler
	ReadTimeout           time.Duration
	WriteTimeout          time.Duration
	IdleTimeout           time.Duration
	BodyLimit             int
}

// Server owns the Fiber application, its configuration, the middleware
// dependency graph, and the background worker pool. All fields should be
// treated as read-only after Start is called.
type Server struct {
	App            *fiber.App
	Cfg            *config.Config
	MiddlewareDeps *middleware.MiddlewareDependencies
	// Pool is the shared bounded worker pool for background tasks.
	// Handlers receive an injected reference via HandlersRegistry; they must
	// not retain a direct pointer to Server.
	Pool *worker.Pool
}

func NewFiberConfig(cfg config.Config) *FiberConfig {
	return &FiberConfig{
		AppName:               cfg.AppName,
		DisableStartupMessage: false,
		CaseSensitive:         false,
		ErrorHandler:          fibererror.GlobalErrorHandler,
		ReadTimeout:           15 * time.Second,
		WriteTimeout:          15 * time.Second,
		IdleTimeout:           60 * time.Second,
		BodyLimit:             4 * 1024 * 1024, // 4 MB
	}
}

func NewServer(cfg config.Config, fiberCfg *FiberConfig) *Server {
	// start servernya
	app := fiber.New(fiber.Config{
		AppName:               fiberCfg.AppName,
		DisableStartupMessage: fiberCfg.DisableStartupMessage,
		CaseSensitive:         fiberCfg.CaseSensitive,
		ErrorHandler:          fiberCfg.ErrorHandler,

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

func (s *Server) Start() error {
	// Launch background worker pool before accepting HTTP traffic so that
	// handlers can dispatch tasks from the very first request.
	s.Pool.Start(context.Background())

	s.MiddlewareDeps.InitAllMiddleware()
	router.SetupRoutes(s.MiddlewareDeps.Log, s.App, s.Pool)

	if s.Cfg.IsDevelopment {
		swagger.SwaggerSetup()
	}

	s.App.Static("/", "./static/public")

	return s.App.Listen(":" + s.Cfg.Port)
}
