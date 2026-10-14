package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/eleccrazy/cloud-reliability-project/app/internal/metrics"
)

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (r *statusRecorder) WriteHeader(code int) {
	if r.statusCode != 0 {
		return
	}

	if code >= 100 && code < 200 {
		r.ResponseWriter.WriteHeader(code)
		return
	}

	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *statusRecorder) Write(data []byte) (int, error) {
	if r.statusCode == 0 {
		r.WriteHeader(http.StatusOK)
	}

	return r.ResponseWriter.Write(data)
}

func Instrument(
	route string,
	m *metrics.HTTPMetrics,
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		m.RequestsInFlight.Inc()
		defer m.RequestsInFlight.Dec()

		recorder := &statusRecorder{
			ResponseWriter: w,
		}

		defer func() {
			status := recorder.statusCode
			if status == 0 {
				status = http.StatusOK
			}

			labels := []string{
				r.Method,
				route,
				strconv.Itoa(status),
			}

			m.RequestsTotal.WithLabelValues(labels...).Inc()
			m.RequestDuration.WithLabelValues(labels...).Observe(
				time.Since(start).Seconds(),
			)
		}()

		next.ServeHTTP(recorder, r)
	})
}
