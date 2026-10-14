package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/eleccrazy/cloud-reliability-project/app/internal/service"
)

func TestItemEndpoints(t *testing.T) {
	handler := NewItemHandler(service.NewItemService())

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/items", handler.List)
	mux.HandleFunc("GET /api/v1/items/{id}", handler.GetByID)

	tests := []struct {
		name       string
		path       string
		wantStatus int
	}{
		{
			name:       "list items",
			path:       "/api/v1/items",
			wantStatus: http.StatusOK,
		},
		{
			name:       "get existing item",
			path:       "/api/v1/items/1",
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid item ID",
			path:       "/api/v1/items/abc",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "item not found",
			path:       "/api/v1/items/999",
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()

			mux.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf(
					"expected status %d, got %d",
					tt.wantStatus,
					rec.Code,
				)
			}
		})
	}
}
