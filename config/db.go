package config

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	badger "github.com/dgraph-io/badger/v4"
)

var DB *badger.DB

func ConnectDB() error {
	var err error
	DB, err = badger.Open(badger.DefaultOptions("/tmp/badger"))
	if err != nil {
		return err
	}

	// graceful shutdown
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
		slog.Info("Shutting down...")
		Close()
		os.Exit(0)
	}()

	slog.Info("Database connection successful")
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
