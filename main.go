package main

import (
	"fmt"
	"go-api-boilerplate/config"
	"go-api-boilerplate/internal/middleware"
	"go-api-boilerplate/internal/server"

	"github.com/boni-fm/go-libsd3/pkg/log"
)

func main() {
	cfg := config.LoadConfigIni()
	log_ := log.NewLoggerWithFilename(cfg.AppName)

	// init semuanya!
	fiberCfg := server.NewFiberConfig(cfg)
	srv := server.NewServer(cfg, fiberCfg)
	middlewareDeps := middleware.NewMiddlewareDepedencies(log_, srv.App, cfg.IsDevelopment)
	srv.SetMiddlewareDeps(middlewareDeps)

	// Mulai ~ 🤩
	fmt.Println("Service started ~~ ༼ つ ◕_◕ ༽つ")
	fmt.Println(`
      ┌──────────────────────────────────────────┐
      │ IT SOFTWARE DEVELOPMENT 3 GO-API SERVICE │
      └──────────────────────────────────────────┘
	`)

	srv.Start()
}
