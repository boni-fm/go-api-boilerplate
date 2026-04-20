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

	// API-007: configure the global timezone so that time.Now() returns the
	// correct localized time throughout the entire process.
	//
	// Priority (highest → lowest):
	//   1. TZ environment variable   — standard Unix/production override
	//   2. Timezone key in appsettings.ini — project-level default
	//   3. UTC                            — safe fallback
	//
	// Using the TZ env var allows production deployments (Docker, K8s) to inject
	// the timezone without touching appsettings.ini. The ini key is kept for local
	// development convenience.
	tzName := os.Getenv("TZ")
	if tzName == "" {
		tzName = cfg.Timezone
	}
	if tzName != "" {
		loc, err := time.LoadLocation(tzName)
		if err != nil {
			log_.Warnf("Invalid timezone %q: %v — falling back to UTC", tzName, err)
		} else {
			time.Local = loc
			log_.Infof("Timezone set to %s", tzName)
		}
	}

	// Wire server, middleware, and database.
	fiberCfg := server.NewFiberConfig(cfg)
	srv := server.NewServer(cfg, fiberCfg)
	middlewareDeps := middleware.NewMiddlewareDependencies(log_, srv.App, cfg.IsDevelopment)
	srv.SetMiddlewareDeps(middlewareDeps)

	// API-001 + API-003: initialise all tenant DB connections and inject the
	// registry so MultiTenantMiddleware can route each request to the correct
	// database based on the X-Kunci header.
	registry := database.InitDatabases(cfg.Kunci, log_)
	srv.SetRegistry(registry)

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
		// 1. Stop accepting new HTTP connections and let in-flight requests
		//    finish (they may still submit background jobs to the pool).
		if err := srv.App.ShutdownWithTimeout(10 * time.Second); err != nil {
			log_.Errorf("Forced shutdown after timeout: %v", err)
			os.Exit(1)
		}
		// 2. Drain the worker pool so no background tasks are abandoned.
		srv.Pool.Stop()
		// 3. ARC-004: close all database connections to drain pgx pools
		//    and avoid leaving idle connections on PostgreSQL.
		if errs := registry.Close(); len(errs) > 0 {
			for _, e := range errs {
				log_.Errorf("DB close error: %v", e)
			}
		}
		log_.Info("Server exited gracefully.")
	}
}
