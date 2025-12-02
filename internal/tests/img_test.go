package tests

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	badger "github.com/dgraph-io/badger/v4"
	"github.com/piheta/seq.re/internal/features/img"
)

func TestImageCreation(t *testing.T) {
	db := SetupTestDB(t)
	tempDir := t.TempDir()
	repo := img.NewImageRepo(db)
	service := img.NewImageService(repo, tempDir)

	imageData := []byte("fake image data")
	contentType := "image/png"

	created, err := service.CreateImage(imageData, contentType, false, false)
	if err != nil {
		t.Fatalf("failed to create image: %v", err)
	}

	if created == nil {
		t.Fatal("expected image to be created, got nil")
	}

	if created.Short == "" {
		t.Error("expected short code to be generated, got empty string")
	}

	if created.ContentType != contentType {
		t.Errorf("expected content type %s, got %s", contentType, created.ContentType)
	}

	if created.Encrypted {
		t.Error("expected Encrypted to be false")
	}

	if created.OneTime {
		t.Error("expected OneTime to be false")
	}

	if created.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}

	if created.ExpiresAt.IsZero() {
		t.Error("expected ExpiresAt to be set")
	}

	// Verify file was written to disk
	if _, err := os.Stat(created.FilePath); os.IsNotExist(err) {
		t.Errorf("expected file to exist at %s", created.FilePath)
	}

	// Verify file content
	content, err := os.ReadFile(created.FilePath)
	if err != nil {
		t.Fatalf("failed to read image file: %v", err)
	}

	if string(content) != string(imageData) {
		t.Errorf("expected file content %s, got %s", imageData, content)
	}
}

func TestEncryptedImageCreation(t *testing.T) {
	db := SetupTestDB(t)
	tempDir := t.TempDir()
	repo := img.NewImageRepo(db)
	service := img.NewImageService(repo, tempDir)

	imageData := []byte("encrypted image data")
	contentType := "application/octet-stream"

	created, err := service.CreateImage(imageData, contentType, true, false)
	if err != nil {
		t.Fatalf("failed to create encrypted image: %v", err)
	}

	if !created.Encrypted {
		t.Error("expected Encrypted to be true")
	}

	// Verify file extension is .bin for encrypted
	if filepath.Ext(created.FilePath) != ".bin" {
		t.Errorf("expected .bin extension for encrypted image, got %s", filepath.Ext(created.FilePath))
	}
}

func TestOneTimeImageCreation(t *testing.T) {
	db := SetupTestDB(t)
	tempDir := t.TempDir()
	repo := img.NewImageRepo(db)
	service := img.NewImageService(repo, tempDir)

	imageData := []byte("onetime image data")
	contentType := "image/jpeg"

	created, err := service.CreateImage(imageData, contentType, false, true)
	if err != nil {
		t.Fatalf("failed to create onetime image: %v", err)
	}

	if !created.OneTime {
		t.Error("expected OneTime to be true")
	}

	if created.Encrypted {
		t.Error("expected Encrypted to be false")
	}
}

func TestImageRetrieval(t *testing.T) {
	db := SetupTestDB(t)
	tempDir := t.TempDir()
	repo := img.NewImageRepo(db)
	service := img.NewImageService(repo, tempDir)

	imageData := []byte("test image data")
	contentType := "image/png"

	created, err := service.CreateImage(imageData, contentType, false, false)
	if err != nil {
		t.Fatalf("failed to create image: %v", err)
	}

	// Retrieve the image
	retrievedData, retrievedType, encrypted, err := service.GetImage(created.Short)
	if err != nil {
		t.Fatalf("failed to retrieve image: %v", err)
	}

	if string(retrievedData) != string(imageData) {
		t.Errorf("expected image data %s, got %s", imageData, retrievedData)
	}

	if retrievedType != contentType {
		t.Errorf("expected content type %s, got %s", contentType, retrievedType)
	}

	if encrypted {
		t.Error("expected encrypted to be false")
	}

	// Verify image still exists (not one-time)
	_, _, _, err = service.GetImage(created.Short)
	if err != nil {
		t.Error("expected image to still exist after first retrieval")
	}
}

func TestOneTimeImageDeletion(t *testing.T) {
	db := SetupTestDB(t)
	tempDir := t.TempDir()
	repo := img.NewImageRepo(db)
	service := img.NewImageService(repo, tempDir)

	imageData := []byte("onetime image")
	contentType := "image/png"

	created, err := service.CreateImage(imageData, contentType, false, true)
	if err != nil {
		t.Fatalf("failed to create onetime image: %v", err)
	}

	// Retrieve once
	_, _, _, err = service.GetImage(created.Short)
	if err != nil {
		t.Fatalf("failed to retrieve onetime image: %v", err)
	}

	// Verify file was deleted from disk
	if _, err := os.Stat(created.FilePath); !os.IsNotExist(err) {
		t.Error("expected file to be deleted from disk after retrieval")
	}

	// Try to retrieve again - should fail
	_, _, _, err = service.GetImage(created.Short)
	if err == nil {
		t.Error("expected error when retrieving onetime image again")
	}

	if !errors.Is(err, badger.ErrKeyNotFound) {
		t.Errorf("expected ErrKeyNotFound, got %v", err)
	}
}

func TestEncryptedImageDeletion(t *testing.T) {
	db := SetupTestDB(t)
	tempDir := t.TempDir()
	repo := img.NewImageRepo(db)
	service := img.NewImageService(repo, tempDir)

	imageData := []byte("encrypted image")
	contentType := "application/octet-stream"

	created, err := service.CreateImage(imageData, contentType, true, false)
	if err != nil {
		t.Fatalf("failed to create encrypted image: %v", err)
	}

	// Retrieve once
	retrievedData, _, encrypted, err := service.GetImage(created.Short)
	if err != nil {
		t.Fatalf("failed to retrieve encrypted image: %v", err)
	}

	if !encrypted {
		t.Error("expected encrypted flag to be true")
	}

	// Verify data is base64 encoded (different from original)
	if string(retrievedData) == string(imageData) {
		t.Error("expected base64 encoded data, got raw data")
	}

	// Verify file was deleted from disk
	if _, err := os.Stat(created.FilePath); !os.IsNotExist(err) {
		t.Error("expected encrypted file to be deleted from disk after retrieval")
	}

	// Try to retrieve again - should fail
	_, _, _, err = service.GetImage(created.Short)
	if err == nil {
		t.Error("expected error when retrieving encrypted image again")
	}
}

func TestImageNotFound(t *testing.T) {
	db := SetupTestDB(t)
	tempDir := t.TempDir()
	repo := img.NewImageRepo(db)
	service := img.NewImageService(repo, tempDir)

	// Try to retrieve non-existent image
	_, _, _, err := service.GetImage("nonexistent")

	if err == nil {
		t.Fatal("expected error for non-existent image, got nil")
	}

	if !errors.Is(err, badger.ErrKeyNotFound) {
		t.Errorf("expected ErrKeyNotFound, got %v", err)
	}
}

func TestMultipleImageCreation(t *testing.T) {
	db := SetupTestDB(t)
	tempDir := t.TempDir()
	repo := img.NewImageRepo(db)
	service := img.NewImageService(repo, tempDir)

	images := []struct {
		data        []byte
		contentType string
	}{
		{[]byte("image1"), "image/png"},
		{[]byte("image2"), "image/jpeg"},
		{[]byte("image3"), "image/gif"},
	}

	createdImages := make([]*img.Image, len(images))
	for i, imgData := range images {
		created, err := service.CreateImage(imgData.data, imgData.contentType, false, false)
		if err != nil {
			t.Fatalf("failed to create image %d: %v", i, err)
		}
		createdImages[i] = created
	}

	// Verify all images can be retrieved
	for i, created := range createdImages {
		retrieved, contentType, _, err := service.GetImage(created.Short)
		if err != nil {
			t.Fatalf("failed to retrieve image %d: %v", i, err)
		}

		if string(retrieved) != string(images[i].data) {
			t.Errorf("image %d: expected data %s, got %s", i, images[i].data, retrieved)
		}

		if contentType != images[i].contentType {
			t.Errorf("image %d: expected content type %s, got %s", i, images[i].contentType, contentType)
		}
	}
}

func TestImageExpiry(t *testing.T) {
	db := SetupTestDB(t)
	tempDir := t.TempDir()
	repo := img.NewImageRepo(db)
	service := img.NewImageService(repo, tempDir)

	imageData := []byte("expiring image")
	contentType := "image/png"

	created, err := service.CreateImage(imageData, contentType, false, false)
	if err != nil {
		t.Fatalf("failed to create image: %v", err)
	}

	// Verify expiry is 7 days from now (with 1 second tolerance)
	expectedExpiry := time.Now().Add(7 * 24 * time.Hour)
	diff := created.ExpiresAt.Sub(expectedExpiry).Abs()
	if diff > time.Second {
		t.Errorf("expected expiry ~7 days from now, got %v", created.ExpiresAt)
	}

	// Create a short-lived image for testing expiry
	shortLivedImage := img.Image{
		Short:       "testshort",
		FilePath:    filepath.Join(tempDir, "testshort.png"),
		ContentType: "image/png",
		Encrypted:   false,
		OneTime:     false,
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(1 * time.Second),
	}

	// Write file
	err = os.WriteFile(shortLivedImage.FilePath, []byte("test"), 0600)
	if err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	err = repo.Create(&shortLivedImage)
	if err != nil {
		t.Fatalf("failed to create short-lived image: %v", err)
	}

	// Verify it exists
	_, _, _, err = service.GetImage("testshort")
	if err != nil {
		t.Fatalf("failed to retrieve short-lived image: %v", err)
	}

	// Wait for expiry
	time.Sleep(2 * time.Second)

	// Verify it's gone from DB
	_, _, _, err = service.GetImage("testshort")
	if err == nil {
		t.Error("expected error after image expiry, got nil")
	}

	if !errors.Is(err, badger.ErrKeyNotFound) {
		t.Errorf("expected ErrKeyNotFound after expiry, got %v", err)
	}
}

func TestImageShortCodeUniqueness(t *testing.T) {
	db := SetupTestDB(t)
	tempDir := t.TempDir()
	repo := img.NewImageRepo(db)
	service := img.NewImageService(repo, tempDir)

	// Create 100 images and verify all have unique short codes
	shortCodes := make(map[string]bool)
	for i := range 100 {
		imageData := []byte("image" + string(rune(i)))
		created, err := service.CreateImage(imageData, "image/png", false, false)
		if err != nil {
			t.Fatalf("failed to create image %d: %v", i, err)
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

func TestFileExtensionMapping(t *testing.T) {
	db := SetupTestDB(t)
	tempDir := t.TempDir()
	repo := img.NewImageRepo(db)
	service := img.NewImageService(repo, tempDir)

	tests := []struct {
		contentType string
		wantExt     string
	}{
		{"image/png", ".png"},
		{"image/jpeg", ".jpg"},
		{"image/gif", ".gif"},
		{"image/webp", ".webp"},
		{"application/octet-stream", ".bin"},
	}

	for _, tt := range tests {
		t.Run(tt.contentType, func(t *testing.T) {
			created, err := service.CreateImage([]byte("test"), tt.contentType, false, false)
			if err != nil {
				t.Fatalf("failed to create image: %v", err)
			}

			gotExt := filepath.Ext(created.FilePath)
			if gotExt != tt.wantExt {
				t.Errorf("expected extension %s for %s, got %s", tt.wantExt, tt.contentType, gotExt)
			}
		})
	}
}

func TestEncryptedAndOneTime(t *testing.T) {
	db := SetupTestDB(t)
	tempDir := t.TempDir()
	repo := img.NewImageRepo(db)
	service := img.NewImageService(repo, tempDir)

	imageData := []byte("encrypted and onetime")
	contentType := "application/octet-stream"

	// Create image that is both encrypted and onetime
	created, err := service.CreateImage(imageData, contentType, true, true)
	if err != nil {
		t.Fatalf("failed to create image: %v", err)
	}

	if !created.Encrypted {
		t.Error("expected Encrypted to be true")
	}

	if !created.OneTime {
		t.Error("expected OneTime to be true")
	}

	// Retrieve once - should delete because encrypted flag triggers deletion
	_, _, _, err = service.GetImage(created.Short)
	if err != nil {
		t.Fatalf("failed to retrieve image: %v", err)
	}

	// Verify it's deleted
	_, _, _, err = service.GetImage(created.Short)
	if err == nil {
		t.Error("expected error when retrieving deleted image")
	}
}
