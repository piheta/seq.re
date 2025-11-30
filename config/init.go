package config

import (
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

type config struct {
	Host string
	Port string
}

var Config config

func InitEnv() {
	// Load env vars from .env file if it exists
	dotEnvLoaded := true
	if err := godotenv.Load(); err != nil {
		dotEnvLoaded = false
	}

	Config = config{
		Host: os.Getenv("HOST"), // http://localhost
		Port: os.Getenv("PORT"), // :8080
	}

	if !dotEnvLoaded {
		slog.Warn("No .env file found, Using default environment variables")
	}

	slog.With("host", Config.Host).Info("ENV Loaded")
}
