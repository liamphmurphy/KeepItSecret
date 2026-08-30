package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	httpdelivery "github.com/keepitsecret/api/internal/delivery/http"
	"github.com/keepitsecret/api/internal/repository"
	"github.com/keepitsecret/api/internal/usecase"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	port := envInt("PORT", 8080)
	address := fmt.Sprintf(":%d", port)

	repository := repository.NewMemorySecretRepository()
	service := usecase.NewSecretService(repository, time.Now)
	apiServer := httpdelivery.NewServer(service, logger)
	router := httpdelivery.Handler(apiServer)
	server := &http.Server{
		Addr:              address,
		Handler:           httpdelivery.NoStore(router),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	stopContext, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	serverError := make(chan error, 1)
	go func() {
		logger.Info("api server starting", "addr", address)
		serverError <- server.ListenAndServe()
	}()

	select {
	case err := <-serverError:
		if !errors.Is(err, http.ErrServerClosed) {
			logger.Error("api server stopped unexpectedly", "err", err)
			os.Exit(1)
		}
	case <-stopContext.Done():
		shutdownContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownContext); err != nil {
			logger.Error("api server shutdown failed", "err", err)
			os.Exit(1)
		}
	}
}

func envInt(name string, fallback int) int {
	value, ok := os.LookupEnv(name)
	if !ok {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 1 || parsed > 65535 {
		return fallback
	}
	return parsed
}
