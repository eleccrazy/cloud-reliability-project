package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/eleccrazy/cloud-reliability-project/app/internal/service"
)

type ItemHandler struct {
	service *service.ItemService
}

func NewItemHandler(s *service.ItemService) *ItemHandler {
	return &ItemHandler{service: s}
}

func (h *ItemHandler) List(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, h.service.List())
}

func (h *ItemHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))

	if err != nil || id <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid item ID",
		})
		return
	}

	item, err := h.service.GetByID(id)

	if errors.Is(err, service.ErrItemNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{
			"error": "item not found",
		})
		return
	}

	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "internal server error",
		})
		return
	}

	writeJSON(w, http.StatusOK, item)
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(data)
}
