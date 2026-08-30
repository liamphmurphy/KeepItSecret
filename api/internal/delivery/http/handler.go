package httpdelivery

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/keepitsecret/api/internal/domain"
	"github.com/keepitsecret/api/internal/usecase"
)

const maxRequestBytes = domain.MaxCiphertextBytes + 16*1024

type Handler struct {
	service usecase.SecretService
	logger  *slog.Logger
}

type createSecretRequest struct {
	SecretID   string `json:"secretId"`
	Version    int    `json:"version"`
	Ciphertext string `json:"ciphertext"`
	Nonce      string `json:"nonce"`
	ExpiresAt  string `json:"expiresAt"`
}

type secretResponse struct {
	SecretID   string `json:"secretId"`
	Version    int    `json:"version"`
	Ciphertext string `json:"ciphertext"`
	Nonce      string `json:"nonce"`
	ExpiresAt  string `json:"expiresAt"`
}

func NewHandler(service usecase.SecretService, logger *slog.Logger) http.Handler {
	if logger == nil {
		logger = slog.Default()
	}

	h := &Handler{service: service, logger: logger}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/secrets", h.create)
	mux.HandleFunc("POST /api/v1/secrets/{secretID}/claim", h.claim)
	mux.HandleFunc("GET /api/v1/health/live", h.live)
	mux.HandleFunc("GET /api/v1/health/ready", h.ready)
	return h.noStore(mux)
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var request createSecretRequest
	if err := decodeJSON(w, r, &request); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request")
		return
	}
	ciphertext, err := base64.RawURLEncoding.DecodeString(request.Ciphertext)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid ciphertext")
		return
	}
	nonce, err := base64.RawURLEncoding.DecodeString(request.Nonce)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid nonce")
		return
	}
	expiresAt, err := time.Parse(time.RFC3339, request.ExpiresAt)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid expiration")
		return
	}

	secret, err := h.service.Create(r.Context(), usecase.CreateSecretInput{
		ID:         request.SecretID,
		Version:    request.Version,
		Ciphertext: ciphertext,
		Nonce:      nonce,
		ExpiresAt:  expiresAt,
	})
	if err != nil {
		h.handleServiceError(r, w, "create secret", err)
		return
	}

	h.writeJSON(w, http.StatusCreated, map[string]string{"secretId": secret.ID})
}

func (h *Handler) claim(w http.ResponseWriter, r *http.Request) {
	secret, err := h.service.Claim(r.Context(), r.PathValue("secretID"))
	if err != nil {
		h.handleServiceError(r, w, "claim secret", err)
		return
	}

	h.writeJSON(w, http.StatusOK, toResponse(secret))
}

func (h *Handler) live(w http.ResponseWriter, _ *http.Request) {
	h.writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) ready(w http.ResponseWriter, _ *http.Request) {
	h.writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) noStore(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}

func (h *Handler) handleServiceError(r *http.Request, w http.ResponseWriter, operation string, err error) {
	switch {
	case errors.Is(err, domain.ErrValidation):
		h.writeError(w, http.StatusUnprocessableEntity, "invalid secret")
	case errors.Is(err, domain.ErrNotFound), errors.Is(err, domain.ErrExpired), errors.Is(err, domain.ErrAlreadyConsumed):
		h.writeError(w, http.StatusNotFound, "secret unavailable")
	default:
		h.logger.ErrorContext(r.Context(), operation+" failed", "err", err)
		h.writeError(w, http.StatusInternalServerError, "internal server error")
	}
}

func (h *Handler) writeError(w http.ResponseWriter, status int, message string) {
	h.writeJSON(w, status, map[string]string{"error": message})
}

func (h *Handler) writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		h.logger.Error("write response failed", "err", err)
	}
}

func decodeJSON(w http.ResponseWriter, r *http.Request, destination any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBytes)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return errors.New("request contains multiple JSON values")
	}
	return nil
}

func toResponse(secret domain.Secret) secretResponse {
	return secretResponse{
		SecretID:   secret.ID,
		Version:    secret.Version,
		Ciphertext: base64.RawURLEncoding.EncodeToString(secret.Ciphertext),
		Nonce:      base64.RawURLEncoding.EncodeToString(secret.Nonce),
		ExpiresAt:  secret.ExpiresAt.UTC().Format(time.RFC3339),
	}
}
