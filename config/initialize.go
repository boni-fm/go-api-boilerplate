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

	Timezone string
}

func LoadConfigIni() Config {
	return LoadConfigIniFromPath("appsettings.ini")
}

func LoadConfigIniFromPath(path string) Config {
	cfg, err := ini.Load(path)
	if err != nil {
		log.Fatalf("Gagal membaca config file appsettings.ini, pastikan file tersebut ada dalam directory \n err :: %v", err)
	}

	return Config{
		AppName:       cfg.Section("CONFIG").Key("AppName").MustString("GoAPIBoilerplate"),
		IsDevelopment: cfg.Section("CONFIG").Key("IsDevelopment").MustBool(false),
		Kunci:         cfg.Section("CONFIG").Key("Kunci").String(),
		Port:          cfg.Section("CONFIG").Key("Port").MustString("8080"),
		Timezone:      cfg.Section("CONFIG").Key("Timezone").MustString("Asia/Jakarta"),
	}
}
