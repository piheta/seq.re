package config

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
	"github.com/lmittmann/tint"
)

type config struct {
	RedirectHost    string
	RedirectPort    string
	BehindProxy     bool
	DBPath          string
	DBEncryptionKey string
}

var Config config

func InitEnv() {
	// Load env vars from .env file if it exists
	dotEnvLoaded := true
	if err := godotenv.Load(); err != nil {
		dotEnvLoaded = false
	}

	Config = config{
		RedirectHost:    os.Getenv("REDIRECT_HOST"),          // http://localhost
		RedirectPort:    os.Getenv("REDIRECT_PORT"),          // :8080
		BehindProxy:     os.Getenv("BEHIND_PROXY") == "true", // Required in order to determine sender ip
		DBPath:          os.Getenv("DB_PATH"),
		DBEncryptionKey: os.Getenv("DB_ENCRYPTION_KEY"),
	}

	if !dotEnvLoaded {
		slog.Warn("No .env file found, Using default environment variables")
	}

	w := os.Stderr
	var handler slog.Handler

	handler = tint.NewHandler(w, &tint.Options{
		Level:      slog.LevelDebug,
		TimeFormat: "01/02 15:04:05",
	})

	logger := slog.New(handler)
	slog.SetDefault(logger)

	green := "\033[1;92m" // ANSI code for green text
	reset := "\033[0m"    // Reset to default color

	// nolint:forbidigo
	fmt.Println(green +
		" ___  ___  __ _   _ __ ___ \n" +
		"/ __|/ _ \\/ _` | | '__/ _ \\\n" +
		"\\__ \\  __/ (_| |_| | |  __/\n" +
		"|___|\\___|\\__, (_)_|  \\___|" +
		"\n             | |           " +
		"\n             |_|           " +
		"\n" + reset)

	slog.With("redirect_host", Config.RedirectHost).With("redirect_port", Config.RedirectPort).Info("ENV Loaded")
}
