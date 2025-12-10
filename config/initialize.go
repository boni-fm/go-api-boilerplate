// config/initialize.go
package config

import (
	"log"

	"gopkg.in/ini.v1"
)

type Config struct {
	AppName       string
	IsDevelopment bool
	Kunci         string
	Port          string
}

func LoadConfigIni() Config {
	cfg, err := ini.Load("appsettings.ini")
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}

	return Config{
		AppName:       cfg.Section("CONFIG").Key("AppName").MustString("Go API Boilerplate"),
		IsDevelopment: cfg.Section("CONFIG").Key("IsDevelopment").MustBool(false),
		Kunci:         cfg.Section("CONFIG").Key("Kunci").String(),
		Port:          cfg.Section("CONFIG").Key("Port").MustString("8080"),
	}
}
