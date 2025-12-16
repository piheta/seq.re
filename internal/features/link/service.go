package link

import (
	"errors"
	"log/slog"
	"time"

	"github.com/piheta/seq.re/internal/shared"
)

type LinkService struct {
	linkRepo *LinkRepo
}

func NewLinkService(linkRepo *LinkRepo) *LinkService {
	return &LinkService{linkRepo: linkRepo}
}

func (s *LinkService) CreateLink(url string, encrypted, onetime bool) (*Link, error) {
	link := Link{
		Short:     shared.CreateShort(),
		URL:       url,
		Encrypted: encrypted,
		OneTime:   onetime,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}

	err := s.linkRepo.Create(&link)
	if err != nil {
		return nil, err
	}

	return &link, nil
}

func (s *LinkService) GetLinkByShort(short string) (*Link, error) {
	link, err := s.linkRepo.GetByShort(short)
	if err != nil {
		return nil, err
	}

	if link.OneTime {
		if err := s.DeleteLink(short); err != nil {
			slog.With("error", err).With("short", short).Error("failed to delete onetime paste after retrieval")
			return nil, errors.New("failed to delete paste")
		}
	}

	return link, nil
}

func (s *LinkService) DeleteLink(short string) error {
	return s.linkRepo.Delete(short)
}

func (s *LinkService) CheckLinkExists(short string) (*Link, error) {
	return s.linkRepo.GetByShort(short)
}
