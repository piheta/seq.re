package paste

import (
	"errors"
	"log/slog"
	"time"

	"github.com/piheta/seq.re/internal/shared"
)

type PasteService struct {
	pasteRepo *PasteRepo
}

func NewPasteService(pasteRepo *PasteRepo) *PasteService {
	return &PasteService{
		pasteRepo: pasteRepo,
	}
}

func (s *PasteService) CreatePaste(content string, language string, encrypted bool, onetime bool) (*Paste, error) {
	paste := Paste{
		Short:     shared.CreateShort(),
		Content:   content,
		Language:  language,
		Encrypted: encrypted,
		OneTime:   onetime,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}

	err := s.pasteRepo.Create(&paste)
	if err != nil {
		return nil, err
	}

	return &paste, nil
}

func (s *PasteService) GetPaste(short string) (*Paste, error) {
	paste, err := s.pasteRepo.GetByShort(short)
	if err != nil {
		return nil, err
	}

	if paste.Encrypted {
		if err := s.DeletePaste(short); err != nil {
			slog.With("error", err).With("short", short).Error("failed to delete encrypted paste after retrieval")
			return nil, errors.New("failed to delete paste")
		}

		return paste, nil
	}

	if paste.OneTime {
		if err := s.DeletePaste(short); err != nil {
			slog.With("error", err).With("short", short).Error("failed to delete onetime paste after retrieval")
			return nil, errors.New("failed to delete paste")
		}
	}

	return paste, nil
}

func (s *PasteService) DeletePaste(short string) error {
	return s.pasteRepo.Delete(short)
}
