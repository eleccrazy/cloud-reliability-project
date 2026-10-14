package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"time"
)

func requestID() string {
	var b [16]byte

	if _, err := rand.Read(b[:]); err != nil {
		panic(err)
	}

	return hex.EncodeToString(b[:])
}

func RequestLogging(
	route string,
	logger *slog.Logger,
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		id := requestID()

		w.Header().Set("X-Request-ID", id)

		recorder := &statusRecorder{
			ResponseWriter: w,
		}

		next.ServeHTTP(recorder, r)

		status := recorder.statusCode
		if status == 0 {
			status = http.StatusOK
		}

		logger.Info("http_request",
			"request_id", id,
			"method", r.Method,
			"route", route,
			"status_code", status,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	})
}
