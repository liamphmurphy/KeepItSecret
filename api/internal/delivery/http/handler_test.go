package httpdelivery

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/keepitsecret/api/internal/repository"
	"github.com/keepitsecret/api/internal/usecase"
)

func TestHandlerCreateAndClaim(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 8, 30, 12, 0, 0, 0, time.UTC)
	service := usecase.NewSecretService(repository.NewMemorySecretRepository(), func() time.Time { return now })
	handler := Handler(NewServer(service, nil))

	id := base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{1}, 16))
	requestBody := map[string]any{
		"secretId":   id,
		"version":    1,
		"ciphertext": base64.RawURLEncoding.EncodeToString([]byte("encrypted")),
		"nonce":      base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{2}, 12)),
		"expiresAt":  now.Add(time.Hour).Format(time.RFC3339),
	}
	body, err := json.Marshal(requestBody)
	if err != nil {
		t.Fatalf("Marshal(create request) returned error: %v", err)
	}
	request := httptest.NewRequest(http.MethodPost, "/api/v1/secrets", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusCreated {
		t.Errorf("POST /api/v1/secrets status = %d, want %d", response.Code, http.StatusCreated)
	}

	request = httptest.NewRequest(http.MethodPost, "/api/v1/secrets/"+id+"/claim", nil)
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Errorf("POST /api/v1/secrets/%s/claim status = %d, want %d", id, response.Code, http.StatusOK)
	}
	var claim SecretResponse
	if err := json.NewDecoder(response.Body).Decode(&claim); err != nil {
		t.Fatalf("Decode(claim response) returned error: %v", err)
	}
	if claim.Ciphertext != requestBody["ciphertext"] {
		t.Errorf("claim ciphertext = %q, want %q", claim.Ciphertext, requestBody["ciphertext"])
	}

	request = httptest.NewRequest(http.MethodPost, "/api/v1/secrets/"+id+"/claim", nil)
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusNotFound {
		t.Errorf("second claim status = %d, want %d", response.Code, http.StatusNotFound)
	}
}

func TestHandlerRejectsMalformedJSON(t *testing.T) {
	t.Parallel()
	handler := Handler(NewServer(usecase.NewSecretService(repository.NewMemorySecretRepository(), time.Now), nil))
	request := httptest.NewRequest(http.MethodPost, "/api/v1/secrets", strings.NewReader("{"))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Errorf("POST /api/v1/secrets malformed JSON status = %d, want %d", response.Code, http.StatusBadRequest)
	}
	if _, err := io.ReadAll(response.Body); err != nil {
		t.Errorf("ReadAll(malformed JSON response) returned error: %v", err)
	}
}
