package config

// disini buat load env dari config.ini

type Config struct {
	AppName       string
	IsDevelopment bool
	Kunci         string
	Port          string
}

func LoadConfigIni() Config {
	return Config{
		AppName:       "Go API Boilerplate",
		IsDevelopment: true,
		Kunci:         "kuncidcho",
		Port:          "8080",
	}
}
