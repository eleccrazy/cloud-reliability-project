package main

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/eleccrazy/cloud-reliability-project/app/internal/handlers"
	"github.com/eleccrazy/cloud-reliability-project/app/internal/metrics"
	"github.com/eleccrazy/cloud-reliability-project/app/internal/middleware"
	"github.com/eleccrazy/cloud-reliability-project/app/internal/service"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	slog.SetDefault(logger)

	registry := prometheus.NewRegistry()
	httpMetrics := metrics.NewHTTPMetrics(registry)

	itemService := service.NewItemService()
	itemHandler := handlers.NewItemHandler(itemService)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		_ = json.NewEncoder(w).Encode(map[string]string{
			"status": "ok",
		})
	})

	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		_ = json.NewEncoder(w).Encode(map[string]string{
			"status": "ready",
		})
	})

	// Instrument and log business endpoints.
	mux.Handle(
		"GET /api/v1/items",
		middleware.RequestLogging(
			"/api/v1/items",
			logger,
			middleware.Instrument(
				"/api/v1/items",
				httpMetrics,
				http.HandlerFunc(itemHandler.List),
			),
		),
	)

	mux.Handle(
		"GET /api/v1/items/{id}",
		middleware.RequestLogging(
			"/api/v1/items/{id}",
			logger,
			middleware.Instrument(
				"/api/v1/items/{id}",
				httpMetrics,
				http.HandlerFunc(itemHandler.GetByID),
			),
		),
	)

	mux.Handle(
		"GET /metrics",
		promhttp.HandlerFor(registry, promhttp.HandlerOpts{}),
	)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// The context is cancelled when SIGINT or SIGTERM is received.
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	serverErrors := make(chan error, 1)

	go func() {
		logger.Info("server_starting", "port", port)

		if err := server.ListenAndServe(); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
		}
	}()

	select {
	case <-ctx.Done():
		logger.Info("shutdown_signal_received")

	case err := <-serverErrors:
		logger.Error("server_failed", "error", err)
		os.Exit(1)
	}

	// Restore default signal handling for a second signal.
	stop()

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	logger.Info("graceful_shutdown_started")

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful_shutdown_failed", "error", err)

		// Force remaining connections to close.
		_ = server.Close()
		os.Exit(1)
	}

	logger.Info("graceful_shutdown_completed")
}
