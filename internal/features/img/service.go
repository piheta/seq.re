package img

import (
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/piheta/seq.re/internal/shared"
)

type ImageService struct {
	imageRepo *ImageRepo
	uploadDir string
}

func NewImageService(imageRepo *ImageRepo, uploadDir string) *ImageService {
	if err := os.MkdirAll(uploadDir, 0750); err != nil {
		slog.Error("Failed to create upload directory", "error", err)
	}
	return &ImageService{
		imageRepo: imageRepo,
		uploadDir: uploadDir,
	}
}

func (s *ImageService) CreateImage(fileData []byte, contentType string, encrypted bool, onetime bool) (*Image, error) {
	short := shared.CreateShort()

	ext := s.getFileExtension(contentType, encrypted)
	filename := fmt.Sprintf("%s%s", short, ext)
	filePath := filepath.Join(s.uploadDir, filename)

	if err := os.WriteFile(filePath, fileData, 0600); err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	image := Image{
		Short:       short,
		FilePath:    filePath,
		ContentType: contentType,
		Encrypted:   encrypted,
		OneTime:     onetime,
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(7 * 24 * time.Hour),
	}

	err := s.imageRepo.Create(&image)
	if err != nil {
		_ = os.Remove(filePath)
		return nil, err
	}

	return &image, nil
}

func (s *ImageService) GetImage(short string) (*Image, []byte, error) {
	image, err := s.imageRepo.GetByShort(short)
	if err != nil {
		return nil, nil, err
	}

	fileData, err := os.ReadFile(image.FilePath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read file: %w", err)
	}

	if image.Encrypted {
		fileData = []byte(base64.StdEncoding.EncodeToString(fileData))
	}

	if image.OneTime {
		if err := s.DeleteImage(short, image.FilePath); err != nil {
			slog.With("error", err).With("short", short).Error("failed to delete onetime image after retrieval")
			return nil, nil, errors.New("failed to delete image")
		}
	}

	return image, fileData, nil
}

func (s *ImageService) DeleteImage(short string, filePath string) error {
	if err := s.imageRepo.Delete(short); err != nil {
		return err
	}

	if err := os.Remove(filePath); err != nil {
		slog.With("error", err).With("path", filePath).Error("failed to delete file from disk")
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

//nolint:revive // encrypted flag is acceptable for control flow
func (s *ImageService) getFileExtension(contentType string, encrypted bool) string {
	if encrypted {
		return ".bin"
	}

	switch contentType {
	case "image/png":
		return ".png"
	case "image/jpeg":
		return ".jpg"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	default:
		return ".bin"
	}
}
