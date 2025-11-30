package link

import (
	"crypto/rand"
	"time"
)

type LinkService struct {
	linkRepo *LinkRepo
}

func NewLinkService(linkRepo *LinkRepo) *LinkService {
	return &LinkService{linkRepo: linkRepo}
}

func (s *LinkService) CreateLink(url string) (*Link, error) {
	link := Link{
		Short:     s.createShort(),
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

func (s *LinkService) createShort() string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_"
	b := make([]byte, 6)
	rand.Read(b) // nolint
	for i := range b {
		b[i] = chars[b[i]%byte(len(chars))]
	}
	return string(b)
}
