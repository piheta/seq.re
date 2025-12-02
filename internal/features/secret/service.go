package secret

import (
	"errors"
	"log/slog"
	"time"

	"github.com/piheta/seq.re/internal/shared"
)

type SecretService struct {
	secretRepo *SecretRepo
}

func NewSecretService(secretRepo *SecretRepo) *SecretService {
	return &SecretService{secretRepo: secretRepo}
}

func (s *SecretService) CreateSecret(encryptedSecret string) (*Secret, error) {
	secret := Secret{
		Short:     shared.CreateShort(),
		Data:      encryptedSecret,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}

	err := s.secretRepo.Create(&secret)
	if err != nil {
		return nil, err
	}

	return &secret, nil
}

func (s *SecretService) GetSecret(short string) (*Secret, error) {
	secret, err := s.secretRepo.GetByShort(short)
	if err != nil {
		return nil, err
	}

	if err := s.DeleteSecret(short); err != nil {
		slog.With("error", err).With("short", short).Error("failed to delete secret after retrieval")
		return nil, errors.New("failed to delete secret")
	}

	return secret, nil
}

func (s *SecretService) DeleteSecret(short string) error {
	return s.secretRepo.Delete(short)
}
