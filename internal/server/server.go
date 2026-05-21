package server

import (
	"context"
	"go-api-boilerplate/config"
	"go-api-boilerplate/internal/api/router"
	"go-api-boilerplate/internal/database"
	"go-api-boilerplate/internal/middleware"
	"go-api-boilerplate/internal/utility/fibererror"
	"go-api-boilerplate/internal/utility/injector"
	"go-api-boilerplate/internal/utility/swagger"
	"go-api-boilerplate/internal/worker"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/static"
)

// jumlah goroutine buat worker pool background — tuning sesuai kebutuhan service
const defaultWorkerCount = 4

// kapasitas queue worker pool — kalau penuh, job baru langsung dibuang (shed)
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

// Server itu induknya — nyimpen app Fiber, config, middleware, worker pool, sama koneksi db.
// Setelah Start dipanggil, field-field di sini jangan diubah-ubah lagi.
// Kalau diubah, resiko ditanggung sendiri ~
type Server struct {
	App            *fiber.App
	Cfg            *config.Config
	MiddlewareDeps *middleware.MiddlewareDependencies

	// worker pool buat background task (audit log, metrik, dll)
	Pool *worker.Pool

	// adapter db buat multidc/singledc
	DcAdapter *database.DcAdapter
	Injector  injector.DBInjector
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

		// Immutable = true biar gak ada data corruption pas nilai dari context
		// (c.Params, c.Query, dll) diterusin ke worker pool background.
		// Tanpa ini, nilai-nilai itu bisa ke-recycle sama fasthttp sebelum sempet dipakai.
		Immutable: true,

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

func (s *Server) SetDbAdapter(adapter *database.DcAdapter) {
	s.DcAdapter = adapter
}

func (s *Server) SetInjector(inject injector.DBInjector) {
	s.Injector = inject
}

func (s *Server) Start() error {
	// jalanin worker pool duluan sebelum nerima request HTTP
	s.Pool.Start(context.Background())

	s.MiddlewareDeps.InitAllMiddleware()
	router.SetupRoutes(s.MiddlewareDeps.Log, s.App, s.Pool, s.DcAdapter, s.Cfg, s.Injector)

	if s.Cfg.IsDevelopment {
		swagger.SwaggerSetup()
	}

	// static file (HTML, gambar, dll) diserve lewat catch-all route "/*"
	// biar route API yang lebih spesifik tetap didahulukan
	s.App.Get("/*", static.New("./static/public"))

	return s.App.Listen(":"+s.Cfg.Port, fiber.ListenConfig{
		DisableStartupMessage: false,
	})
}
