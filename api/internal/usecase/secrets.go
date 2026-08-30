package usecase

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/keepitsecret/api/internal/domain"
)

const (
	DefaultMaxLifetime = 24 * time.Hour
	DefaultMinLifetime = time.Minute
)

type Clock func() time.Time

type CreateSecretInput struct {
	ID         string
	Version    int
	Ciphertext []byte
	Nonce      []byte
	ExpiresAt  time.Time
}

type SecretService interface {
	Create(ctx context.Context, input CreateSecretInput) (domain.Secret, error)
	Claim(ctx context.Context, id string) (domain.Secret, error)
}

type secretService struct {
	repository  domain.SecretRepository
	now         Clock
	minLifetime time.Duration
	maxLifetime time.Duration
}

func NewSecretService(repository domain.SecretRepository, now Clock) SecretService {
	if now == nil {
		now = time.Now
	}

	return &secretService{
		repository:  repository,
		now:         now,
		minLifetime: DefaultMinLifetime,
		maxLifetime: DefaultMaxLifetime,
	}
}

func (s *secretService) Create(ctx context.Context, input CreateSecretInput) (domain.Secret, error) {
	now := s.now().UTC()
	if err := validateCreateInput(input, now, s.minLifetime, s.maxLifetime); err != nil {
		return domain.Secret{}, err
	}

	secret := domain.Secret{
		ID:         input.ID,
		Version:    input.Version,
		Ciphertext: append([]byte(nil), input.Ciphertext...),
		Nonce:      append([]byte(nil), input.Nonce...),
		ExpiresAt:  input.ExpiresAt.UTC(),
		CreatedAt:  now,
	}
	if err := s.repository.Create(ctx, secret); err != nil {
		return domain.Secret{}, fmt.Errorf("create secret: %w", err)
	}

	return secret, nil
}

func (s *secretService) Claim(ctx context.Context, id string) (domain.Secret, error) {
	if id == "" {
		return domain.Secret{}, fmt.Errorf("%w: missing id", domain.ErrValidation)
	}

	secret, err := s.repository.Claim(ctx, id, s.now().UTC())
	if err != nil {
		return domain.Secret{}, fmt.Errorf("claim secret: %w", err)
	}

	return secret, nil
}

func validateCreateInput(input CreateSecretInput, now time.Time, minLifetime, maxLifetime time.Duration) error {
	decodedID, err := base64.RawURLEncoding.DecodeString(input.ID)
	if err != nil || len(decodedID) < domain.SecretIDMinBytes {
		return fmt.Errorf("%w: id must be base64url encoding of at least %d bytes", domain.ErrValidation, domain.SecretIDMinBytes)
	}
	if input.Version != domain.CurrentSecretVersion {
		return fmt.Errorf("%w: unsupported version", domain.ErrValidation)
	}
	if len(input.Ciphertext) == 0 || len(input.Ciphertext) > domain.MaxCiphertextBytes {
		return fmt.Errorf("%w: ciphertext size is invalid", domain.ErrValidation)
	}
	if len(input.Nonce) != domain.NonceSize {
		return fmt.Errorf("%w: nonce size is invalid", domain.ErrValidation)
	}
	if input.ExpiresAt.Before(now.Add(minLifetime)) || input.ExpiresAt.After(now.Add(maxLifetime)) {
		return fmt.Errorf("%w: expiration is outside the allowed window", domain.ErrValidation)
	}

	return nil
}
