package main

import (
	"context"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

type scenario struct {
	name   string
	url    string
	weight int
}

func chooseScenario(scenarios []scenario) scenario {
	total := 0
	for _, s := range scenarios {
		total += s.weight
	}

	n := rand.Intn(total)

	for _, s := range scenarios {
		if n < s.weight {
			return s
		}
		n -= s.weight
	}

	return scenarios[len(scenarios)-1]
}

func envInt(name string, defaultValue int) int {
	value := os.Getenv(name)
	if value == "" {
		return defaultValue
	}

	n, err := strconv.Atoi(value)
	if err != nil || n <= 0 {
		return defaultValue
	}

	return n
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	rps := envInt("LOADGEN_RPS", 1)
	timeoutSeconds := envInt("LOADGEN_TIMEOUT_SECONDS", 3)

	apiURL := os.Getenv("LOADGEN_API_URL")
	if apiURL == "" {
		apiURL = "http://127.0.0.1:8080"
	}

	faultURL := os.Getenv("LOADGEN_FAULT_URL")
	if faultURL == "" {
		faultURL = "http://127.0.0.1:8081"
	}

	scenarios := []scenario{
		{"success", apiURL + "/api/v1/items/1", 90},
		{"client_error", apiURL + "/api/v1/items/999", 5},
		{"server_failure", faultURL + "/internal/faults/failure", 3},
		{"slow_response", faultURL + "/internal/faults/slow", 2},
	}

	client := &http.Client{
		Timeout: time.Duration(timeoutSeconds) * time.Second,
	}

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	interval := time.Second / time.Duration(rps)

	if interval <= 0 {
		logger.Error("invalid_request_rate", "rps", rps)
		os.Exit(1)
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	logger.Info("load_generator_started",
		"rps", rps,
		"timeout_seconds", timeoutSeconds,
	)

	for {
		select {
		case <-ctx.Done():
			logger.Info("load_generator_stopped")
			return

		case <-ticker.C:
			s := chooseScenario(scenarios)

			req, err := http.NewRequestWithContext(
				ctx,
				http.MethodGet,
				s.url,
				nil,
			)
			if err != nil {
				logger.Error("request_creation_failed",
					"error", err,
				)
				continue
			}

			start := time.Now()
			resp, err := client.Do(req)

			if err != nil {
				if ctx.Err() != nil {
					logger.Info("load_generator_stopped")
					return
				}

				logger.Error("request_failed",
					"scenario", s.name,
					"error", err,
				)
				continue
			}

			// Reuse connections without buffering response bodies.
			_ = resp.Body.Close()

			logger.Info("request_completed",
				"scenario", s.name,
				"status_code", resp.StatusCode,
				"duration_ms", time.Since(start).Milliseconds(),
			)
		}
	}
}
