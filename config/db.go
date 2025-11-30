package config

import (
	"fmt"
	"log"
	"log/slog"

	badger "github.com/dgraph-io/badger/v4"
)

var DB *badger.DB

func ConnectDB() error {
	var err error
	var db *badger.DB

	db, err = badger.Open(badger.DefaultOptions("/tmp/badger"))
	if err != nil {
		log.Fatal(err)
	}

	slog.Info("Database connection successful")
	DB = db

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
