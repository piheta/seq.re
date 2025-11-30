package config

import (
	"fmt"
	"log/slog"
	"os"

	badger "github.com/dgraph-io/badger/v4"
)

var DB *badger.DB

func ConnectDB() error {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "/tmp/badger"
	}

	var err error
	DB, err = badger.Open(badger.DefaultOptions(dbPath))
	if err != nil {
		return err
	}

	slog.Info("Database connection successful", slog.String("path", dbPath))
	return nil
}

func Close() error {
	if DB == nil {
		return nil
	}

	// Close the database connection
	if err := DB.Close(); err != nil {
		slog.With("error", err).Error("Failed to close DB connection")
		return fmt.Errorf("error closing database connection: %v", err)
	}

	return nil
}
