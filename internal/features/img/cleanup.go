package img

import (
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

func (s *ImageService) StartCleanupWorker(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		slog.With("interval", interval).Info("Image cleanup worker started")

		for range ticker.C {
			s.cleanupOrphanedFiles()
		}
	}()
}

func (s *ImageService) cleanupOrphanedFiles() {
	entries, err := os.ReadDir(s.uploadDir)
	if err != nil {
		slog.With("error", err).Error("failed to read upload directory for cleanup")
		return
	}

	deletedCount := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()
		ext := filepath.Ext(filename)
		if ext == "" {
			continue
		}
		short := filename[:len(filename)-len(ext)]

		_, err := s.imageRepo.GetByShort(short)
		if err != nil {
			filePath := filepath.Join(s.uploadDir, filename)
			if err := os.Remove(filePath); err != nil {
				slog.With("error", err).With("file", filename).Warn("failed to delete orphaned file")
			} else {
				deletedCount++
			}
		}
	}

	if deletedCount > 0 {
		slog.With("count", deletedCount).Info("cleaned up orphaned image files")
	}
}
