package handlers

import (
	"encoding/json"
	"net/http"
	"time"
)

type FaultHandler struct{}

func NewFaultHandler() *FaultHandler {
	return &FaultHandler{}
}

func (h *FaultHandler) Failure(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)

	_ = json.NewEncoder(w).Encode(map[string]string{
		"error": "simulated internal server failure",
	})
}

func (h *FaultHandler) Slow(w http.ResponseWriter, r *http.Request) {
	select {
	case <-time.After(600 * time.Millisecond):
	case <-r.Context().Done():
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}
