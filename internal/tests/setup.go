package tests

import (
	"os"
	"testing"

	badger "github.com/dgraph-io/badger/v4"
)

// SetupTestDB creates a temporary Badger database for testing
func SetupTestDB(t *testing.T) *badger.DB {
	t.Helper()

	tempDir, err := os.MkdirTemp("", "badger-test-*")
	if err != nil {
		t.Fatalf("failed to create temp directory: %v", err)
	}

	opts := badger.DefaultOptions(tempDir).
		WithInMemory(false).
		WithLoggingLevel(badger.WARNING)

	db, err := badger.Open(opts)
	if err != nil {
		t.Fatalf("failed to open badger database: %v", err)
	}

	t.Cleanup(func() {
		_ = db.Close()
		_ = os.RemoveAll(tempDir)
	})

	return db
}
