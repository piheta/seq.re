package tests

import (
	"errors"
	"testing"
	"time"

	badger "github.com/dgraph-io/badger/v4"
	"github.com/piheta/seq.re/internal/features/link"
)

func TestLinkCreation(t *testing.T) {
	db := SetupTestDB(t)
	repo := link.NewLinkRepo(db)
	service := link.NewLinkService(repo)

	url := "https://example.com"
	created, err := service.CreateLink(url)

	if err != nil {
		t.Fatalf("failed to create link: %v", err)
	}

	if created == nil {
		t.Fatal("expected link to be created, got nil")
	}

	if created.URL != url {
		t.Errorf("expected URL %s, got %s", url, created.URL)
	}

	if created.Short == "" {
		t.Error("expected short code to be generated, got empty string")
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

func TestLinkRetrieval(t *testing.T) {
	db := SetupTestDB(t)
	repo := link.NewLinkRepo(db)
	service := link.NewLinkService(repo)

	url := "https://example.com"
	created, err := service.CreateLink(url)
	if err != nil {
		t.Fatalf("failed to create link: %v", err)
	}

	// Retrieve the link
	retrieved, err := service.GetLinkByShort(created.Short)
	if err != nil {
		t.Fatalf("failed to retrieve link: %v", err)
	}

	if retrieved == nil {
		t.Fatal("expected link to be retrieved, got nil")
	}

	if retrieved.Short != created.Short {
		t.Errorf("expected short %s, got %s", created.Short, retrieved.Short)
	}

	if retrieved.URL != created.URL {
		t.Errorf("expected URL %s, got %s", created.URL, retrieved.URL)
	}
}

func TestLinkNotFound(t *testing.T) {
	db := SetupTestDB(t)
	repo := link.NewLinkRepo(db)
	service := link.NewLinkService(repo)

	// Try to retrieve non-existent link
	retrieved, err := service.GetLinkByShort("nonexistent")

	if err == nil {
		t.Fatal("expected error for non-existent link, got nil")
	}

	if !errors.Is(err, badger.ErrKeyNotFound) {
		t.Errorf("expected ErrKeyNotFound, got %v", err)
	}

	if retrieved != nil {
		t.Errorf("expected nil link for non-existent key, got %v", retrieved)
	}
}

func TestLinkExpiry(t *testing.T) {
	db := SetupTestDB(t)
	repo := link.NewLinkRepo(db)
	service := link.NewLinkService(repo)

	url := "https://example.com"
	created, err := service.CreateLink(url)
	if err != nil {
		t.Fatalf("failed to create link: %v", err)
	}

	// Verify link exists immediately
	retrieved, err := service.GetLinkByShort(created.Short)
	if err != nil {
		t.Fatalf("failed to retrieve link: %v", err)
	}

	if retrieved == nil {
		t.Fatal("expected link to exist, got nil")
	}

	// Create a link with very short TTL for testing
	shortLivedLink := link.Link{
		Short:     "testshort",
		URL:       "https://short-lived.com",
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(1 * time.Second),
	}

	err = repo.Create(&shortLivedLink)
	if err != nil {
		t.Fatalf("failed to create short-lived link: %v", err)
	}

	// Verify it exists
	retrieved, err = service.GetLinkByShort("testshort")
	if err != nil {
		t.Fatalf("failed to retrieve short-lived link: %v", err)
	}

	if retrieved == nil {
		t.Fatal("expected short-lived link to exist, got nil")
	}

	// Wait for expiry
	time.Sleep(2 * time.Second)

	// Verify it's gone
	retrieved, err = service.GetLinkByShort("testshort")
	if err == nil {
		t.Error("expected error after link expiry, got nil")
	}

	if !errors.Is(err, badger.ErrKeyNotFound) {
		t.Errorf("expected ErrKeyNotFound after expiry, got %v", err)
	}

	if retrieved != nil {
		t.Errorf("expected nil link after expiry, got %v", retrieved)
	}
}

func TestMultipleLinkCreation(t *testing.T) {
	db := SetupTestDB(t)
	repo := link.NewLinkRepo(db)
	service := link.NewLinkService(repo)

	urls := []string{
		"https://example1.com",
		"https://example2.com",
		"https://example3.com",
	}

	links := make([]*link.Link, len(urls))
	for i, url := range urls {
		created, err := service.CreateLink(url)
		if err != nil {
			t.Fatalf("failed to create link %d: %v", i, err)
		}
		links[i] = created
	}

	// Verify all links can be retrieved
	for i, l := range links {
		retrieved, err := service.GetLinkByShort(l.Short)
		if err != nil {
			t.Fatalf("failed to retrieve link %d: %v", i, err)
		}

		if retrieved.URL != urls[i] {
			t.Errorf("link %d: expected URL %s, got %s", i, urls[i], retrieved.URL)
		}
	}
}

func TestShortCodeUniqueness(t *testing.T) {
	db := SetupTestDB(t)
	repo := link.NewLinkRepo(db)
	service := link.NewLinkService(repo)

	// Create 100 links and verify all have unique short codes
	shortCodes := make(map[string]bool)
	for i := range 100 {
		created, err := service.CreateLink("https://example.com/" + string(rune(i)))
		if err != nil {
			t.Fatalf("failed to create link %d: %v", i, err)
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
