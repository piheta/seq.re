package tests

import (
	"encoding/base64"
	"errors"
	"strings"
	"testing"
	"time"

	badger "github.com/dgraph-io/badger/v4"
	"github.com/piheta/seq.re/internal/features/paste"
)

func TestPasteCreation(t *testing.T) {
	db := SetupTestDB(t)
	repo := paste.NewPasteRepo(db)
	service := paste.NewPasteService(repo)

	content := "package main\n\nfunc main() {\n\tprintln(\"Hello, World!\")\n}"
	language := "go"

	created, err := service.CreatePaste(content, language, false, false)

	if err != nil {
		t.Fatalf("failed to create paste: %v", err)
	}

	if created == nil {
		t.Fatal("expected paste to be created, got nil")
	}

	if created.Content != content {
		t.Errorf("expected content %s, got %s", content, created.Content)
	}

	if created.Language != language {
		t.Errorf("expected language %s, got %s", language, created.Language)
	}

	if created.Short == "" {
		t.Error("expected short code to be generated, got empty string")
	}

	if len(created.Short) != 6 {
		t.Errorf("expected short code length 6, got %d", len(created.Short))
	}

	if created.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}

	if created.ExpiresAt.IsZero() {
		t.Error("expected ExpiresAt to be set")
	}

	// Verify expiry is 7 days from now (with 1 second tolerance)
	expectedExpiry := time.Now().Add(7 * 24 * time.Hour)
	diff := created.ExpiresAt.Sub(expectedExpiry).Abs()
	if diff > time.Second {
		t.Errorf("expected expiry ~7 days from now, got %v", created.ExpiresAt)
	}
}

func TestPasteCreationWithoutLanguage(t *testing.T) {
	db := SetupTestDB(t)
	repo := paste.NewPasteRepo(db)
	service := paste.NewPasteService(repo)

	content := "Just some plain text without a language"

	created, err := service.CreatePaste(content, "", false, false)

	if err != nil {
		t.Fatalf("failed to create paste: %v", err)
	}

	if created.Language != "" {
		t.Errorf("expected empty language, got %s", created.Language)
	}

	if created.Content != content {
		t.Errorf("expected content %s, got %s", content, created.Content)
	}
}

func TestPasteRetrieval(t *testing.T) {
	db := SetupTestDB(t)
	repo := paste.NewPasteRepo(db)
	service := paste.NewPasteService(repo)

	content := "console.log('Hello, World!');"
	language := "javascript"

	created, err := service.CreatePaste(content, language, false, false)
	if err != nil {
		t.Fatalf("failed to create paste: %v", err)
	}

	// Retrieve the paste
	retrieved, err := service.GetPaste(created.Short)
	if err != nil {
		t.Fatalf("failed to retrieve paste: %v", err)
	}

	if retrieved == nil {
		t.Fatal("expected paste to be retrieved, got nil")
	}

	if retrieved.Short != created.Short {
		t.Errorf("expected short %s, got %s", created.Short, retrieved.Short)
	}

	if retrieved.Content != content {
		t.Errorf("expected content %s, got %s", content, retrieved.Content)
	}

	if retrieved.Language != language {
		t.Errorf("expected language %s, got %s", language, retrieved.Language)
	}
}

func TestPasteOneTime(t *testing.T) {
	db := SetupTestDB(t)
	repo := paste.NewPasteRepo(db)
	service := paste.NewPasteService(repo)

	content := "This is a one-time paste"
	created, err := service.CreatePaste(content, "plain", false, true)
	if err != nil {
		t.Fatalf("failed to create paste: %v", err)
	}

	if !created.OneTime {
		t.Error("expected OneTime to be true")
	}

	// First retrieval should succeed
	retrieved, err := service.GetPaste(created.Short)
	if err != nil {
		t.Fatalf("failed to retrieve paste on first attempt: %v", err)
	}

	if retrieved == nil {
		t.Fatal("expected paste to be retrieved, got nil")
	}

	if retrieved.Content != content {
		t.Errorf("expected content %s, got %s", content, retrieved.Content)
	}

	// Second retrieval should fail (paste deleted after first view)
	retrieved, err = service.GetPaste(created.Short)
	if err == nil {
		t.Fatal("expected error on second retrieval, got nil")
	}

	if !errors.Is(err, badger.ErrKeyNotFound) {
		t.Errorf("expected ErrKeyNotFound, got %v", err)
	}

	if retrieved != nil {
		t.Errorf("expected nil paste after deletion, got %v", retrieved)
	}
}

func TestPasteEncryptedPersistence(t *testing.T) {
	db := SetupTestDB(t)
	repo := paste.NewPasteRepo(db)
	service := paste.NewPasteService(repo)

	// Simulate client sending already base64-encoded encrypted content
	plainContent := "encrypted content here"
	base64Content := base64.StdEncoding.EncodeToString([]byte(plainContent))

	// Create encrypted paste WITHOUT onetime flag
	created, err := service.CreatePaste(base64Content, "", true, false)
	if err != nil {
		t.Fatalf("failed to create paste: %v", err)
	}

	if !created.Encrypted {
		t.Error("expected Encrypted to be true")
	}

	if created.OneTime {
		t.Error("expected OneTime to be false")
	}

	// First retrieval should succeed
	retrieved, err := service.GetPaste(created.Short)
	if err != nil {
		t.Fatalf("failed to retrieve encrypted paste: %v", err)
	}

	if retrieved == nil {
		t.Fatal("expected paste to be retrieved, got nil")
	}

	if !retrieved.Encrypted {
		t.Error("expected Encrypted flag to be true")
	}

	// Content should remain base64 encoded (not double-encoded)
	if retrieved.Content != base64Content {
		t.Errorf("expected content to remain as %s, got %s", base64Content, retrieved.Content)
	}

	// Should be valid base64
	if _, err := base64.StdEncoding.DecodeString(retrieved.Content); err != nil {
		t.Errorf("expected valid base64 content: %v", err)
	}

	// Second retrieval should succeed (encrypted without onetime should persist)
	retrieved, err = service.GetPaste(created.Short)
	if err != nil {
		t.Error("expected encrypted non-onetime paste to be retrievable multiple times")
	}

	if retrieved == nil {
		t.Error("expected paste to still exist")
	}
}

func TestPasteNotFound(t *testing.T) {
	db := SetupTestDB(t)
	repo := paste.NewPasteRepo(db)
	service := paste.NewPasteService(repo)

	// Try to retrieve non-existent paste
	retrieved, err := service.GetPaste("abc123")

	if err == nil {
		t.Fatal("expected error for non-existent paste, got nil")
	}

	if !errors.Is(err, badger.ErrKeyNotFound) {
		t.Errorf("expected ErrKeyNotFound, got %v", err)
	}

	if retrieved != nil {
		t.Errorf("expected nil paste for non-existent key, got %v", retrieved)
	}
}

func TestPasteExpiry(t *testing.T) {
	db := SetupTestDB(t)
	repo := paste.NewPasteRepo(db)
	service := paste.NewPasteService(repo)

	content := "expiring paste"
	created, err := service.CreatePaste(content, "", false, false)
	if err != nil {
		t.Fatalf("failed to create paste: %v", err)
	}

	// Verify paste exists immediately
	retrieved, err := service.GetPaste(created.Short)
	if err != nil {
		t.Fatalf("failed to retrieve paste: %v", err)
	}

	if retrieved == nil {
		t.Fatal("expected paste to exist, got nil")
	}

	// Create a paste with very short TTL for testing
	shortLivedPaste := paste.Paste{
		Short:     "testex",
		Content:   "short lived content",
		Language:  "plain",
		Encrypted: false,
		OneTime:   false,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(1 * time.Second),
	}

	err = repo.Create(&shortLivedPaste)
	if err != nil {
		t.Fatalf("failed to create short-lived paste: %v", err)
	}

	// Wait for expiry
	time.Sleep(2 * time.Second)

	// Verify it's gone (TTL expired)
	retrieved, err = repo.GetByShort("testex")
	if err == nil {
		t.Error("expected error after paste expiry, got nil")
	}

	if !errors.Is(err, badger.ErrKeyNotFound) {
		t.Errorf("expected ErrKeyNotFound after expiry, got %v", err)
	}

	if retrieved != nil {
		t.Errorf("expected nil paste after expiry, got %v", retrieved)
	}
}

func TestMultiplePasteCreation(t *testing.T) {
	db := SetupTestDB(t)
	repo := paste.NewPasteRepo(db)
	service := paste.NewPasteService(repo)

	pasteData := []struct {
		content  string
		language string
	}{
		{"print('Python')", "python"},
		{"console.log('JS')", "javascript"},
		{"fmt.Println('Go')", "go"},
		{"puts 'Ruby'", "ruby"},
		{"echo 'PHP'", "php"},
	}

	pastes := make([]*paste.Paste, len(pasteData))
	for i, data := range pasteData {
		created, err := service.CreatePaste(data.content, data.language, false, false)
		if err != nil {
			t.Fatalf("failed to create paste %d: %v", i, err)
		}
		pastes[i] = created
	}

	// Verify all pastes can be retrieved
	for i, p := range pastes {
		retrieved, err := service.GetPaste(p.Short)
		if err != nil {
			t.Fatalf("failed to retrieve paste %d: %v", i, err)
		}

		if retrieved.Content != pasteData[i].content {
			t.Errorf("paste %d: expected content %s, got %s", i, pasteData[i].content, retrieved.Content)
		}

		if retrieved.Language != pasteData[i].language {
			t.Errorf("paste %d: expected language %s, got %s", i, pasteData[i].language, retrieved.Language)
		}
	}
}

func TestPasteShortCodeUniqueness(t *testing.T) {
	db := SetupTestDB(t)
	repo := paste.NewPasteRepo(db)
	service := paste.NewPasteService(repo)

	// Create 100 pastes and verify all have unique short codes
	shortCodes := make(map[string]bool)
	for i := range 100 {
		created, err := service.CreatePaste("content"+string(rune(i)), "", false, false)
		if err != nil {
			t.Fatalf("failed to create paste %d: %v", i, err)
		}

		if shortCodes[created.Short] {
			t.Errorf("duplicate short code generated: %s", created.Short)
		}
		shortCodes[created.Short] = true
	}

	if len(shortCodes) != 100 {
		t.Errorf("expected 100 unique short codes, got %d", len(shortCodes))
	}
}

func TestPasteContentValidation(t *testing.T) {
	db := SetupTestDB(t)
	repo := paste.NewPasteRepo(db)
	service := paste.NewPasteService(repo)

	tests := []struct {
		name     string
		content  string
		language string
	}{
		{"Empty string", "", "plain"},
		{"Short content", "abc", "plain"},
		{"Long content", strings.Repeat("a", 10000), "plain"},
		{"Special characters", "!@#$%^&*()_+{}|:<>?", "plain"},
		{"Unicode", "🚀 Hello 世界", "plain"},
		{"Multi-line", "line1\nline2\nline3", "plain"},
		{"Tabs and spaces", "\t\tindented\n  spaces", "python"},
		{"JSON", `{"key": "value", "nested": {"foo": "bar"}}`, "json"},
		{"Code with comments", "// This is a comment\nfunc main() {}", "go"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			created, err := service.CreatePaste(tt.content, tt.language, false, false)
			if err != nil {
				t.Fatalf("failed to create paste: %v", err)
			}

			retrieved, err := service.GetPaste(created.Short)
			if err != nil {
				t.Fatalf("failed to retrieve paste: %v", err)
			}

			if retrieved.Content != tt.content {
				t.Errorf("expected content %q, got %q", tt.content, retrieved.Content)
			}

			if retrieved.Language != tt.language {
				t.Errorf("expected language %s, got %s", tt.language, retrieved.Language)
			}
		})
	}
}

func TestPasteDeletion(t *testing.T) {
	db := SetupTestDB(t)
	repo := paste.NewPasteRepo(db)
	service := paste.NewPasteService(repo)

	content := "test content"
	created, err := service.CreatePaste(content, "", false, false)
	if err != nil {
		t.Fatalf("failed to create paste: %v", err)
	}

	// Verify paste exists
	retrieved, err := service.GetPaste(created.Short)
	if err != nil {
		t.Fatalf("failed to retrieve paste: %v", err)
	}

	if retrieved == nil {
		t.Fatal("expected paste to exist, got nil")
	}

	// Delete the paste
	err = service.DeletePaste(created.Short)
	if err != nil {
		t.Fatalf("failed to delete paste: %v", err)
	}

	// Verify paste is deleted
	retrieved, err = service.GetPaste(created.Short)
	if err == nil {
		t.Fatal("expected error when retrieving deleted paste, got nil")
	}

	if !errors.Is(err, badger.ErrKeyNotFound) {
		t.Errorf("expected ErrKeyNotFound, got %v", err)
	}

	if retrieved != nil {
		t.Errorf("expected nil, got %v", retrieved)
	}
}

func TestPasteTimestamps(t *testing.T) {
	db := SetupTestDB(t)
	repo := paste.NewPasteRepo(db)
	service := paste.NewPasteService(repo)

	beforeCreate := time.Now()
	created, err := service.CreatePaste("timestamp test", "plain", false, false)
	afterCreate := time.Now()

	if err != nil {
		t.Fatalf("failed to create paste: %v", err)
	}

	// Verify CreatedAt is between before and after
	if created.CreatedAt.Before(beforeCreate) || created.CreatedAt.After(afterCreate) {
		t.Errorf("CreatedAt %v not between %v and %v", created.CreatedAt, beforeCreate, afterCreate)
	}

	// Verify ExpiresAt is roughly 7 days after CreatedAt
	expectedExpiry := created.CreatedAt.Add(7 * 24 * time.Hour)
	diff := created.ExpiresAt.Sub(expectedExpiry).Abs()
	if diff > time.Second {
		t.Errorf("ExpiresAt %v not ~7 days after CreatedAt %v", created.ExpiresAt, created.CreatedAt)
	}

	// Verify ExpiresAt is after CreatedAt
	if !created.ExpiresAt.After(created.CreatedAt) {
		t.Error("ExpiresAt should be after CreatedAt")
	}
}

func TestPasteLanguageOptions(t *testing.T) {
	db := SetupTestDB(t)
	repo := paste.NewPasteRepo(db)
	service := paste.NewPasteService(repo)

	languages := []string{
		"go", "python", "javascript", "typescript", "rust",
		"java", "c", "cpp", "ruby", "php", "swift", "kotlin",
		"bash", "shell", "sql", "json", "yaml", "markdown",
		"html", "css", "xml", "dockerfile", "plain",
	}

	for _, lang := range languages {
		created, err := service.CreatePaste("test content", lang, false, false)
		if err != nil {
			t.Fatalf("failed to create paste with language %s: %v", lang, err)
		}

		if created.Language != lang {
			t.Errorf("expected language %s, got %s", lang, created.Language)
		}

		retrieved, err := service.GetPaste(created.Short)
		if err != nil {
			t.Fatalf("failed to retrieve paste with language %s: %v", lang, err)
		}

		if retrieved.Language != lang {
			t.Errorf("retrieved: expected language %s, got %s", lang, retrieved.Language)
		}
	}
}

func TestPasteEncryptedAndOneTime(t *testing.T) {
	db := SetupTestDB(t)
	repo := paste.NewPasteRepo(db)
	service := paste.NewPasteService(repo)

	// Simulate client sending already base64-encoded encrypted content
	plainContent := "super secret content"
	base64Content := base64.StdEncoding.EncodeToString([]byte(plainContent))

	created, err := service.CreatePaste(base64Content, "", true, true)
	if err != nil {
		t.Fatalf("failed to create paste: %v", err)
	}

	if !created.Encrypted {
		t.Error("expected Encrypted to be true")
	}

	if !created.OneTime {
		t.Error("expected OneTime to be true")
	}

	// Retrieve should return base64 content and delete
	retrieved, err := service.GetPaste(created.Short)
	if err != nil {
		t.Fatalf("failed to retrieve paste: %v", err)
	}

	// Content should remain base64 encoded (not double-encoded)
	if retrieved.Content != base64Content {
		t.Errorf("expected content to remain as %s, got %s", base64Content, retrieved.Content)
	}

	// Should be valid base64
	if _, err := base64.StdEncoding.DecodeString(retrieved.Content); err != nil {
		t.Errorf("expected valid base64 content: %v", err)
	}

	// Second retrieval should fail
	retrieved, err = service.GetPaste(created.Short)
	if err == nil {
		t.Fatal("expected error on second retrieval, got nil")
	}

	if !errors.Is(err, badger.ErrKeyNotFound) {
		t.Errorf("expected ErrKeyNotFound, got %v", err)
	}
}
