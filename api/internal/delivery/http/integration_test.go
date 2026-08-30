//go:build integration

package httpdelivery

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/keepitsecret/api/internal/repository"
	"github.com/keepitsecret/api/internal/usecase"
)

func TestIntegrationSecretLifecycle(t *testing.T) {
	now := time.Date(2026, 8, 30, 12, 0, 0, 0, time.UTC)
	service := usecase.NewSecretService(repository.NewMemorySecretRepository(), func() time.Time { return now })
	handler := NewHandler(service, nil)

	id := base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{7}, 16))
	payload := map[string]any{
		"secretId":   id,
		"version":    1,
		"ciphertext": base64.RawURLEncoding.EncodeToString([]byte("integration ciphertext")),
		"nonce":      base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{8}, 12)),
		"expiresAt":  now.Add(time.Hour).Format(time.RFC3339),
	}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Marshal(payload) returned error: %v", err)
	}
	createRequest := httptest.NewRequest(http.MethodPost, "/api/v1/secrets", bytes.NewReader(body))
	createRequest.Header.Set("Content-Type", "application/json")
	createResponse := httptest.NewRecorder()
	handler.ServeHTTP(createResponse, createRequest)
	if createResponse.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want %d", createResponse.Code, http.StatusCreated)
	}

	claimRequest := httptest.NewRequest(http.MethodPost, "/api/v1/secrets/"+id+"/claim", nil)
	claimResponse := httptest.NewRecorder()
	handler.ServeHTTP(claimResponse, claimRequest)
	if claimResponse.Code != http.StatusOK {
		t.Errorf("claim status = %d, want %d", claimResponse.Code, http.StatusOK)
	}
}
