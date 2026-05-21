package main

import (
	"fmt"
	"go-api-boilerplate/internal/database"
	"go-api-boilerplate/internal/utility/injector"
	"go-api-boilerplate/internal/utility/tztime"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-api-boilerplate/config"
	"go-api-boilerplate/internal/middleware"
	"go-api-boilerplate/internal/server"

	"github.com/boni-fm/go-libsd3/pkg/log"
)

// @title           Go Api ~
// @version         1.x.x
// @description     ~

// @contact.name   Department IT SD 3
// @contact.email  sd3@indomaret.co.id

// // // Ini comment untuk naro token di header
// // // jadi, token auth nya gk ditaro didalam query param ...
// //@securityDefinitions.apikey  BearerAuth
// //@in                          header
// //@name                        Authorization
// //@description                 Enter your bearer token in the format: Bearer {token}
func main() {
	/*
		INITIALIZE AWAL APLIKASI

		------------
			*Note ::
				jika ingin ada perubahan, ... take it with your own risk ...
				jadi minimalisir perubahan di bagian core api ...
	*/

	fmt.Println("[Startup] Initializing configuration dependencies...")
	cfg := config.LoadConfigIni()
	log_ := log.NewLoggerWithFilename(cfg.AppName)

	// Disini setup timezone secara global
	// untuk resolved time.Now() sesuai dengan timezone aplikasinya
	// versi debug (isDevelopment = false), akan baca setting timezone sesuai dengan appsettings.ini
	tztime.SetupTimezone(log_, &cfg)

	// Wire server, middleware, and database.
	fiberCfg := server.NewFiberConfig(cfg)
	srv := server.NewServer(cfg, fiberCfg)

	dbManager := database.GetDcAdapter(cfg.AppName, log_)
	middlewareDeps := middleware.NewMiddlewareDependencies(
		log_,
		srv.App,
		cfg.IsDevelopment,
	)

	// Function dibawah dipakai untuk case kodedc dari appsetting.ini
	// jadi db nya akan di register/preconnect ke adapter
	// kemudian di inject lgsg kedalam repo ...
	//
	// Untuk menggunakan ini, pastikan kunci di appsettings.ini ada
	// dan pastikan handler tidak menggunakan middleware multidc
	// kalau ingin pakai fungsi diatas (kunci hardcode)
	//
	// static >> kunci dari appsettings.ini
	// locals >> dari middleware multidc
	//
	// Inject ke seluruh penjuru repo
	//dbInjector := injector.NewStaticInjector(dbManager, cfg.KodeDc) //jadi ditaro dalam memory ~

	// Ini untuk multidc
	// kalau ingin pakai fungsi diatas (kunci hardcode)
	// bisa comment atau hapus line dibawah
	// jadi db injector nya dalam bentuk static, bkn dari local
	dbInjector := injector.NewLocalsInjector()

	srv.SetMiddlewareDeps(middlewareDeps)
	srv.SetDbAdapter(dbManager)
	srv.SetInjector(dbInjector)

	fmt.Println("Service starting ~~ ༼ つ ◕_◕ ༽つ")
	fmt.Println(`
		 /$$$$$$ /$$$$$$$$        /$$$$$$  /$$$$$$$         /$$$$$$ 
		|_  $$_/|__  $$__/       /$$__  $$| $$__  $$       /$$__  $$
		  | $$     | $$         | $$  \__/| $$  \ $$      |__/  \ $$
		  | $$     | $$         |  $$$$$$ | $$  | $$         /$$$$$/
		  | $$     | $$          \____  $$| $$  | $$        |___  $$
		  | $$     | $$          /$$  \ $$| $$  | $$       /$$  \ $$
		 /$$$$$$   | $$         |  $$$$$$/| $$$$$$$/      |  $$$$$$/
		|______/   |__/          \______/ |_______/        \______/ 
	`)

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

		srv.Pool.Stop()
		dbManager.CloseAll()

		log_.Info("Server meninggal dengan elegan ~~")
	}
}
