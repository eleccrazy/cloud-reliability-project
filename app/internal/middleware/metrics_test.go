package middleware

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/eleccrazy/cloud-reliability-project/app/internal/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestInstrumentRecordsStatusCodes(t *testing.T) {
	tests := []struct {
		name   string
		status int
	}{
		{"success", http.StatusOK},
		{"bad request", http.StatusBadRequest},
		{"not found", http.StatusNotFound},
		{"server error", http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := prometheus.NewRegistry()
			m := metrics.NewHTTPMetrics(reg)

			handler := Instrument(
				"/api/v1/items/{id}",
				m,
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(tt.status)
				}),
			)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/items/1", nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != tt.status {
				t.Fatalf("expected %d, got %d", tt.status, rec.Code)
			}


count := testutil.ToFloat64(
	m.RequestsTotal.WithLabelValues(
		http.MethodGet,
		"/api/v1/items/{id}",
		strconv.Itoa(tt.status),
	),
)

if count != 1 {
	t.Errorf("expected request counter 1, got %v", count)
}

if inFlight := testutil.ToFloat64(m.RequestsInFlight); inFlight != 0 {
	t.Errorf("expected in-flight gauge 0, got %v", inFlight)
}


		})
	}
}
