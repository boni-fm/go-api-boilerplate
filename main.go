package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-api-boilerplate/config"
	"go-api-boilerplate/internal/database"
	"go-api-boilerplate/internal/middleware"
	"go-api-boilerplate/internal/server"

	"github.com/boni-fm/go-libsd3/pkg/log"
)

func main() {
	cfg := config.LoadConfigIni()
	log_ := log.NewLoggerWithFilename(cfg.AppName)

	// Wire server, middleware, and database.
	fiberCfg := server.NewFiberConfig(cfg)
	srv := server.NewServer(cfg, fiberCfg)
	middlewareDeps := middleware.NewMiddlewareDependencies(log_, srv.App, cfg.IsDevelopment)
	srv.SetMiddlewareDeps(middlewareDeps)

	database.InitDatabase(cfg.Kunci, log_)

	fmt.Println("Service started ~~ ༼ つ ◕_◕ ༽つ")
	fmt.Println(`
      ┌──────────────────────────────────────────┐
      │ IT SOFTWARE DEVELOPMENT 3 GO-API SERVICE │
      └──────────────────────────────────────────┘
	`)

	// Graceful shutdown: listen for SIGTERM / SIGINT in the background while
	// the server runs. When a signal arrives the server is given 10 seconds to
	// finish in-flight requests before the process exits.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Start()
	}()

	select {
	case err := <-errCh:
		if err != nil {
			log_.Errorf("Server failed to start: %v", err)
			panic(err)
		}
	case sig := <-quit:
		log_.Infof("Received signal %v — initiating graceful shutdown...", sig)
		if err := srv.App.ShutdownWithTimeout(10 * time.Second); err != nil {
			log_.Errorf("Forced shutdown after timeout: %v", err)
			os.Exit(1)
		}
		log_.Info("Server exited gracefully.")
	}
}
