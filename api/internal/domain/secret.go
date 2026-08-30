package domain

import (
	"context"
	"errors"
	"time"
)

const (
	CurrentSecretVersion = 1
	SecretIDMinBytes     = 16
	NonceSize            = 12
	MaxCiphertextBytes   = 1 << 20
)

var (
	ErrAlreadyConsumed = errors.New("secret already consumed")
	ErrExpired         = errors.New("secret expired")
	ErrNotFound        = errors.New("secret not found")
	ErrValidation      = errors.New("secret validation failed")
)

type Secret struct {
	ID         string
	Version    int
	Ciphertext []byte
	Nonce      []byte
	ExpiresAt  time.Time
	CreatedAt  time.Time
}

type SecretRepository interface {
	Create(ctx context.Context, secret Secret) error
	Claim(ctx context.Context, id string, now time.Time) (Secret, error)
}
