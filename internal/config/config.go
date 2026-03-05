package config

import (
	"os"
)

type Config struct {
	AppPort  string
	DBDSN    string
	LogLevel string
}

func Load() Config {
	cfg := Config{
		AppPort:  getEnv("APP_PORT", "8080"),
		DBDSN:    getEnv("DB_DSN", ""),
		LogLevel: getEnv("LOG_LEVEL", "info"),
	}
	return cfg
}

func getEnv(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}
