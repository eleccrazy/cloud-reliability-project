package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestFailureHandler(t *testing.T) {
	handler := NewFaultHandler()

	req := httptest.NewRequest(http.MethodGet, "/internal/faults/failure", nil)
	rec := httptest.NewRecorder()

	handler.Failure(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}

	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}

	if body["error"] != "simulated internal server failure" {
		t.Errorf("unexpected response: %v", body)
	}
}

func TestSlowHandler(t *testing.T) {
	handler := NewFaultHandler()

	req := httptest.NewRequest(http.MethodGet, "/internal/faults/slow", nil)
	rec := httptest.NewRecorder()

	start := time.Now()
	handler.Slow(rec, req)
	elapsed := time.Since(start)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	if elapsed < 600*time.Millisecond {
		t.Errorf("expected at least 600ms, got %v", elapsed)
	}
}
