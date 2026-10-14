package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/eleccrazy/cloud-reliability-project/app/internal/handlers"
	"github.com/eleccrazy/cloud-reliability-project/app/internal/metrics"
	"github.com/eleccrazy/cloud-reliability-project/app/internal/middleware"
	"github.com/eleccrazy/cloud-reliability-project/app/internal/service"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	mux := http.NewServeMux()

	registry := prometheus.NewRegistry()
	httpMetrics := metrics.NewHTTPMetrics(registry)

	itemService := service.NewItemService()
	itemHandler := handlers.NewItemHandler(itemService)

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
		middleware.Instrument(
			"/api/v1/items",
			httpMetrics,
			http.HandlerFunc(itemHandler.List),
		),
	)

	mux.Handle(
		"GET /api/v1/items/{id}",
		middleware.Instrument(
			"/api/v1/items/{id}",
			httpMetrics,
			http.HandlerFunc(itemHandler.GetByID),
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

	log.Printf("Starting API on port %s", port)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
