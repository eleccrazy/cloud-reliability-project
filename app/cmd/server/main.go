package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/eleccrazy/cloud-reliability-project/app/internal/handlers"
	"github.com/eleccrazy/cloud-reliability-project/app/internal/service"
)

func main() {
	mux := http.NewServeMux()

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

	mux.HandleFunc("GET /api/v1/items", itemHandler.List)
	mux.HandleFunc("GET /api/v1/items/{id}", itemHandler.GetByID)

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
