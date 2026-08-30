package repository

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/keepitsecret/api/internal/domain"
)

func TestMemorySecretRepositoryConcurrentClaim(t *testing.T) {
	t.Parallel()
	repo := NewMemorySecretRepository()
	now := time.Date(2026, 8, 30, 12, 0, 0, 0, time.UTC)
	secret := domain.Secret{ID: "id", Ciphertext: []byte("ciphertext"), Nonce: make([]byte, 12), ExpiresAt: now.Add(time.Hour)}
	if err := repo.Create(context.Background(), secret); err != nil {
		t.Fatalf("Create(%q) returned error: %v", secret.ID, err)
	}

	var group sync.WaitGroup
	results := make(chan error, 2)
	for range 2 {
		group.Add(1)
		go func() {
			defer group.Done()
			_, err := repo.Claim(context.Background(), secret.ID, now)
			results <- err
		}()
	}
	group.Wait()
	close(results)

	var successes, notFound int
	for err := range results {
		switch {
		case err == nil:
			successes++
		case errors.Is(err, domain.ErrNotFound):
			notFound++
		}
	}
	if successes != 1 || notFound != 1 {
		t.Errorf("concurrent Claim(%q) results = successes %d, not found %d, want 1 and 1", secret.ID, successes, notFound)
	}
}

func TestMemorySecretRepositoryExpiredClaim(t *testing.T) {
	t.Parallel()
	repo := NewMemorySecretRepository()
	now := time.Date(2026, 8, 30, 12, 0, 0, 0, time.UTC)
	secret := domain.Secret{ID: "id", Ciphertext: []byte("ciphertext"), ExpiresAt: now}
	if err := repo.Create(context.Background(), secret); err != nil {
		t.Fatalf("Create(%q) returned error: %v", secret.ID, err)
	}
	if _, err := repo.Claim(context.Background(), secret.ID, now); !errors.Is(err, domain.ErrExpired) {
		t.Errorf("Claim(%q) error = %v, want ErrExpired", secret.ID, err)
	}
}
