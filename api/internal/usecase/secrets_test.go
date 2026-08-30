package usecase

import (
	"context"
	"encoding/base64"
	"errors"
	"testing"
	"time"

	"github.com/keepitsecret/api/internal/domain"
	"github.com/keepitsecret/api/internal/repository"
)

func TestSecretServiceCreateAndClaim(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 8, 30, 12, 0, 0, 0, time.UTC)
	service := NewSecretService(repository.NewMemorySecretRepository(), func() time.Time { return now })
	id := base64.RawURLEncoding.EncodeToString(make([]byte, domain.SecretIDMinBytes))

	created, err := service.Create(context.Background(), CreateSecretInput{
		ID:         id,
		Version:    domain.CurrentSecretVersion,
		Ciphertext: []byte("ciphertext"),
		Nonce:      make([]byte, domain.NonceSize),
		ExpiresAt:  now.Add(time.Hour),
	})
	if err != nil {
		t.Fatalf("Create(valid input) returned error: %v", err)
	}
	if created.ID != id {
		t.Errorf("Create(valid input).ID = %q, want %q", created.ID, id)
	}

	claimed, err := service.Claim(context.Background(), id)
	if err != nil {
		t.Fatalf("Claim(%q) returned error: %v", id, err)
	}
	if string(claimed.Ciphertext) != "ciphertext" {
		t.Errorf("Claim(%q).Ciphertext = %q, want %q", id, claimed.Ciphertext, "ciphertext")
	}
	if _, err := service.Claim(context.Background(), id); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Claim(%q) after consumption error = %v, want ErrNotFound", id, err)
	}
}

func TestSecretServiceCreateValidation(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 8, 30, 12, 0, 0, 0, time.UTC)
	service := NewSecretService(repository.NewMemorySecretRepository(), func() time.Time { return now })
	validID := base64.RawURLEncoding.EncodeToString(make([]byte, domain.SecretIDMinBytes))

	tests := []struct {
		name  string
		input CreateSecretInput
	}{
		{name: "short id", input: CreateSecretInput{ID: "AA", Version: 1, Ciphertext: []byte("x"), Nonce: make([]byte, 12), ExpiresAt: now.Add(time.Hour)}},
		{name: "wrong version", input: CreateSecretInput{ID: validID, Version: 2, Ciphertext: []byte("x"), Nonce: make([]byte, 12), ExpiresAt: now.Add(time.Hour)}},
		{name: "empty ciphertext", input: CreateSecretInput{ID: validID, Version: 1, Nonce: make([]byte, 12), ExpiresAt: now.Add(time.Hour)}},
		{name: "wrong nonce", input: CreateSecretInput{ID: validID, Version: 1, Ciphertext: []byte("x"), Nonce: make([]byte, 11), ExpiresAt: now.Add(time.Hour)}},
		{name: "too soon", input: CreateSecretInput{ID: validID, Version: 1, Ciphertext: []byte("x"), Nonce: make([]byte, 12), ExpiresAt: now.Add(30 * time.Second)}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if _, err := service.Create(context.Background(), tt.input); !errors.Is(err, domain.ErrValidation) {
				t.Errorf("Create(%s) error = %v, want ErrValidation", tt.name, err)
			}
		})
	}
}
