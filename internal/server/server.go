package server

import (
	"go-api-boilerplate/config"
	"go-api-boilerplate/internal/api/router"
	"go-api-boilerplate/internal/middleware"
	"go-api-boilerplate/internal/utility/fibererror"
	"go-api-boilerplate/internal/utility/swagger"

	"github.com/gofiber/fiber/v2"
)

type FiberConfig struct {
	AppName               string
	DisableStartupMessage bool
	CaseSensitive         bool
	ErrorHandler          fiber.ErrorHandler
}

type Server struct {
	App            *fiber.App
	Cfg            *config.Config
	MiddlewareDeps *middleware.MiddlewareDepedencies
}

func NewFiberConfig(cfg config.Config) *FiberConfig {
	return &FiberConfig{
		AppName:               cfg.AppName,
		DisableStartupMessage: false,
		CaseSensitive:         false,
		ErrorHandler:          fibererror.GlobalErrorHandler,
	}
}

func NewServer(cfg config.Config, fiberCfg *FiberConfig) *Server {
	// start servernya
	app := fiber.New(fiber.Config{
		AppName:               fiberCfg.AppName,
		DisableStartupMessage: fiberCfg.DisableStartupMessage,
		CaseSensitive:         fiberCfg.CaseSensitive,
		ErrorHandler:          fiberCfg.ErrorHandler,
	})

	return &Server{
		App: app,
		Cfg: &cfg,
	}
}

func (s *Server) SetMiddlewareDeps(middlewareDeps *middleware.MiddlewareDepedencies) {
	s.MiddlewareDeps = middlewareDeps
}

func (s *Server) Start() error {
	router.SetupRoutes(s.App)
	swagger.SwaggerSetup(s.App)

	s.App.Static("/", "./static/public")
	s.MiddlewareDeps.InitAllMiddleware()

	return s.App.Listen(":" + s.Cfg.Port)
}
