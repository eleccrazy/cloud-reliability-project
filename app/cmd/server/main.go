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
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	registry := prometheus.NewRegistry()
	httpMetrics := metrics.NewHTTPMetrics(registry)

	itemService := service.NewItemService()
	itemHandler := handlers.NewItemHandler(itemService)

	// Main application router.
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

	apiServer := &http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	servers := []*http.Server{apiServer}

	// Fault injection is disabled unless explicitly enabled.
	if os.Getenv("ENABLE_FAULT_INJECTION") == "true" {
		faultHandler := handlers.NewFaultHandler()
		faultMux := http.NewServeMux()

		registerFault := func(route string, handler http.HandlerFunc) {
			faultMux.Handle(
				"GET "+route,
				middleware.RequestLogging(
					route,
					logger,
					middleware.Instrument(route, httpMetrics, handler),
				),
			)
		}

		registerFault("/internal/faults/failure", faultHandler.Failure)
		registerFault("/internal/faults/slow", faultHandler.Slow)

		faultServer := &http.Server{
			Addr:              "127.0.0.1:8081",
			Handler:           faultMux,
			ReadHeaderTimeout: 5 * time.Second,
			ReadTimeout:       10 * time.Second,
			WriteTimeout:      15 * time.Second,
			IdleTimeout:       60 * time.Second,
		}

		servers = append(servers, faultServer)
		logger.Info("fault_injection_enabled", "address", faultServer.Addr)
	}

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	serverErrors := make(chan error, len(servers))

	for _, server := range servers {
		srv := server

		go func() {
			logger.Info("server_starting", "address", srv.Addr)

			if err := srv.ListenAndServe(); err != nil &&
				!errors.Is(err, http.ErrServerClosed) {
				serverErrors <- err
			}
		}()
	}

	select {
	case <-ctx.Done():
		logger.Info("shutdown_signal_received")

	case err := <-serverErrors:
		logger.Error("server_failed", "error", err)
		stop()

		// Shut down every listener before returning failure.
		for _, server := range servers {
			_ = server.Close()
		}
		os.Exit(1)
	}

	stop()

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	logger.Info("graceful_shutdown_started")

	var shutdownFailed bool

	for _, server := range servers {
		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("graceful_shutdown_failed",
				"address", server.Addr,
				"error", err,
			)
			_ = server.Close()
			shutdownFailed = true
		}
	}

	if shutdownFailed {
		os.Exit(1)
	}

	logger.Info("graceful_shutdown_completed")
}
