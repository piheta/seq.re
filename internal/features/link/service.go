package link

import (
	"time"

	"github.com/piheta/seq.re/internal/shared"
)

type LinkService struct {
	linkRepo *LinkRepo
}

func NewLinkService(linkRepo *LinkRepo) *LinkService {
	return &LinkService{linkRepo: linkRepo}
}

func (s *LinkService) CreateLink(url string) (*Link, error) {
	link := Link{
		Short:     shared.CreateShort(),
		URL:       url,
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

	return link, nil
}
