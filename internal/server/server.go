package server

import (
	"go-api-boilerplate/config"
	"go-api-boilerplate/internal/api/router"
	"go-api-boilerplate/internal/middleware"
	"go-api-boilerplate/internal/utility/fibererror"
	"go-api-boilerplate/internal/utility/swagger"
	"time"

	"github.com/gofiber/fiber/v2"
)

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

type Server struct {
	App            *fiber.App
	Cfg            *config.Config
	MiddlewareDeps *middleware.MiddlewareDependencies
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
		App: app,
		Cfg: &cfg,
	}
}

func (s *Server) SetMiddlewareDeps(middlewareDeps *middleware.MiddlewareDependencies) {
	s.MiddlewareDeps = middlewareDeps
}

func (s *Server) Start() error {
	s.MiddlewareDeps.InitAllMiddleware()
	router.SetupRoutes(s.MiddlewareDeps.Log, s.App)

	if s.Cfg.IsDevelopment {
		swagger.SwaggerSetup()
	}

	s.App.Static("/", "./static/public")

	return s.App.Listen(":" + s.Cfg.Port)
}
