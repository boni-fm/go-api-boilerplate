// config/initialize.go
// Package config handles loading and parsing application configuration from
// the appsettings.ini file. Exposing LoadConfigIniFromPath allows callers
// (including tests) to load configuration from any INI file path.
package config

import (
	"log"

	"gopkg.in/ini.v1"
)

// Config holds all application configuration values loaded from appsettings.ini.
type Config struct {
	AppName       string
	IsDevelopment bool
	Kunci         string
	Port          string
}

// LoadConfigIni loads configuration from the default "appsettings.ini" file
// located in the current working directory.
func LoadConfigIni() Config {
	return LoadConfigIniFromPath("appsettings.ini")
}

// LoadConfigIniFromPath loads and parses configuration from the INI file at
// the given path. It is the primary implementation used by LoadConfigIni and
// can be called directly in tests with a temporary file path.
// The process is terminated via log.Fatalf if the file cannot be read.
func LoadConfigIniFromPath(path string) Config {
	cfg, err := ini.Load(path)
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
