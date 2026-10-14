package middleware

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestLogging(t *testing.T) {
	var logs bytes.Buffer

	logger := slog.New(slog.NewJSONHandler(&logs, nil))

	handler := RequestLogging(
		"/api/v1/items/{id}",
		logger,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/items/999",
		nil,
	)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}

	if rec.Header().Get("X-Request-ID") == "" {
		t.Fatal("missing X-Request-ID header")
	}

	var entry map[string]any

	if err := json.Unmarshal(logs.Bytes(), &entry); err != nil {
		t.Fatalf("invalid JSON log: %v", err)
	}

	if entry["msg"] != "http_request" {
		t.Errorf("unexpected log message: %v", entry["msg"])
	}

	if entry["route"] != "/api/v1/items/{id}" {
		t.Errorf("unexpected route: %v", entry["route"])
	}

	if entry["status_code"] != float64(404) {
		t.Errorf("expected status_code 404, got %v", entry["status_code"])
	}
}
