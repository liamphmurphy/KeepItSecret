package httpdelivery

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/keepitsecret/api/internal/domain"
	"github.com/keepitsecret/api/internal/usecase"
)

const maxRequestBytes = domain.MaxCiphertextBytes + 16*1024

// Server implements the generated OpenAPI server interface.
type Server struct {
	service usecase.SecretService
	logger  *slog.Logger
}

// NewServer creates an HTTP API server backed by service.
func NewServer(service usecase.SecretService, logger *slog.Logger) *Server {
	if logger == nil {
		logger = slog.Default()
	}

	return &Server{service: service, logger: logger}
}

// CreateSecret creates a one-time secret.
func (s *Server) CreateSecret(w http.ResponseWriter, r *http.Request) {
	var request CreateSecretJSONRequestBody
	if err := decodeJSON(w, r, &request); err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid request")
		return
	}
	ciphertext, err := base64.RawURLEncoding.DecodeString(request.Ciphertext)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid ciphertext")
		return
	}
	nonce, err := base64.RawURLEncoding.DecodeString(request.Nonce)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid nonce")
		return
	}

	secret, err := s.service.Create(r.Context(), usecase.CreateSecretInput{
		ID:         request.SecretId,
		Version:    int(request.Version),
		Ciphertext: ciphertext,
		Nonce:      nonce,
		ExpiresAt:  request.ExpiresAt,
	})
	if err != nil {
		s.handleServiceError(r, w, "create secret", err)
		return
	}

	s.writeJSON(w, http.StatusCreated, CreateSecretResponse{SecretId: secret.ID})
}

// ClaimSecret claims a one-time secret.
func (s *Server) ClaimSecret(w http.ResponseWriter, r *http.Request, secretID string) {
	secret, err := s.service.Claim(r.Context(), secretID)
	if err != nil {
		s.handleServiceError(r, w, "claim secret", err)
		return
	}

	s.writeJSON(w, http.StatusOK, toResponse(secret))
}

// LiveHealth reports whether the process is alive.
func (s *Server) LiveHealth(w http.ResponseWriter, _ *http.Request) {
	s.writeJSON(w, http.StatusOK, HealthResponse{Status: Ok})
}

// ReadyHealth reports whether the process is ready.
func (s *Server) ReadyHealth(w http.ResponseWriter, _ *http.Request) {
	s.writeJSON(w, http.StatusOK, HealthResponse{Status: Ok})
}

// NoStore prevents intermediaries from caching API responses.
func NoStore(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}

func (s *Server) handleServiceError(r *http.Request, w http.ResponseWriter, operation string, err error) {
	switch {
	case errors.Is(err, domain.ErrValidation):
		s.writeError(w, http.StatusUnprocessableEntity, "invalid secret")
	case errors.Is(err, domain.ErrNotFound), errors.Is(err, domain.ErrExpired), errors.Is(err, domain.ErrAlreadyConsumed):
		s.writeError(w, http.StatusNotFound, "secret unavailable")
	default:
		s.logger.ErrorContext(r.Context(), operation+" failed", "err", err)
		s.writeError(w, http.StatusInternalServerError, "internal server error")
	}
}

func (s *Server) writeError(w http.ResponseWriter, status int, message string) {
	s.writeJSON(w, status, ErrorResponse{Error: message})
}

func (s *Server) writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		s.logger.Error("write response failed", "err", err)
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

func toResponse(secret domain.Secret) SecretResponse {
	return SecretResponse{
		SecretId:   secret.ID,
		Version:    secret.Version,
		Ciphertext: base64.RawURLEncoding.EncodeToString(secret.Ciphertext),
		Nonce:      base64.RawURLEncoding.EncodeToString(secret.Nonce),
		ExpiresAt:  secret.ExpiresAt,
	}
}

var _ ServerInterface = (*Server)(nil)
