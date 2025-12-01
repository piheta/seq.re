package config

import (
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

type config struct {
	RedirectHost string
	RedirectPort string
	BehindProxy  bool
}

var Config config

func InitEnv() {
	// Load env vars from .env file if it exists
	dotEnvLoaded := true
	if err := godotenv.Load(); err != nil {
		dotEnvLoaded = false
	}

	Config = config{
		RedirectHost: os.Getenv("REDIRECT_HOST"),          // http://localhost
		RedirectPort: os.Getenv("REDIRECT_PORT"),          // :8080
		BehindProxy:  os.Getenv("BEHIND_PROXY") == "true", // Required in order to determine sender ip
	}

	if !dotEnvLoaded {
		slog.Warn("No .env file found, Using default environment variables")
	}

	slog.With("redirect_host", Config.RedirectHost).With("redirect_port", Config.RedirectPort).Info("ENV Loaded")
}
