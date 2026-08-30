package repository

import (
	"context"
	"sync"
	"time"

	"github.com/keepitsecret/api/internal/domain"
)

type MemorySecretRepository struct {
	mu      sync.Mutex
	secrets map[string]domain.Secret
}

func NewMemorySecretRepository() *MemorySecretRepository {
	return &MemorySecretRepository{secrets: make(map[string]domain.Secret)}
}

func (r *MemorySecretRepository) Create(ctx context.Context, secret domain.Secret) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.secrets[secret.ID]; exists {
		return domain.ErrAlreadyConsumed
	}
	r.secrets[secret.ID] = domain.Secret{
		ID:         secret.ID,
		Version:    secret.Version,
		Ciphertext: append([]byte(nil), secret.Ciphertext...),
		Nonce:      append([]byte(nil), secret.Nonce...),
		ExpiresAt:  secret.ExpiresAt,
		CreatedAt:  secret.CreatedAt,
	}

	return nil
}

func (r *MemorySecretRepository) Claim(ctx context.Context, id string, now time.Time) (domain.Secret, error) {
	if err := ctx.Err(); err != nil {
		return domain.Secret{}, err
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	secret, exists := r.secrets[id]
	if !exists {
		return domain.Secret{}, domain.ErrNotFound
	}
	delete(r.secrets, id)
	if !now.Before(secret.ExpiresAt) {
		return domain.Secret{}, domain.ErrExpired
	}

	secret.Ciphertext = append([]byte(nil), secret.Ciphertext...)
	secret.Nonce = append([]byte(nil), secret.Nonce...)
	return secret, nil
}
