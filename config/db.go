package config

import (
	"encoding/hex"
	"fmt"
	"log/slog"

	badger "github.com/dgraph-io/badger/v4"
)

var DB *badger.DB

func ConnectDB(dbPath string) error {
	opts := badger.DefaultOptions(dbPath)
	opts = opts.WithLogger(&badgerLogger{logger: slog.Default()})

	encryptionKey := Config.DBEncryptionKey

	if encryptionKey != "" {
		key, err := hex.DecodeString(encryptionKey)
		if err != nil {
			return fmt.Errorf("invalid DB_ENCRYPTION_KEY format (must be hex): %w", err)
		}

		// Validate key length (must be 16, 24, or 32 bytes for AES-128, AES-192, or AES-256)
		keyLen := len(key)
		if keyLen != 16 && keyLen != 24 && keyLen != 32 {
			return fmt.Errorf("DB_ENCRYPTION_KEY must be 16, 24, or 32 bytes (32, 48, or 64 hex chars), got %d bytes", keyLen)
		}

		opts = opts.WithEncryptionKey(key)
		opts = opts.WithIndexCacheSize(100 << 20) // 100 MB cache recommended for encrypted DBs

		slog.Info("Database encryption enabled", slog.Int("keySize", keyLen*8))
	} else {
		slog.Warn("Database encryption disabled - set DB_ENCRYPTION_KEY to enable")
	}

	var err error
	DB, err = badger.Open(opts)
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

// SLOG Adapter for Badger DB
type badgerLogger struct {
	logger *slog.Logger
}

func (l *badgerLogger) Errorf(format string, args ...any) {
	l.logger.Error(fmt.Sprintf(format, args...))
}

func (l *badgerLogger) Warningf(format string, args ...any) {
	l.logger.Warn(fmt.Sprintf(format, args...))
}

func (l *badgerLogger) Infof(format string, args ...any) {
	l.logger.Info(fmt.Sprintf(format, args...))
}

func (l *badgerLogger) Debugf(format string, args ...any) {
	l.logger.Debug(fmt.Sprintf(format, args...))
}
