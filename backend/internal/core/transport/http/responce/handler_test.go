package core_http_responce

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	core_errors "sport_app/internal/core/errors"
	core_logger "sport_app/internal/core/logger"
)

func TestJSONResponse(t *testing.T) {
	rec := httptest.NewRecorder()
	h := NewHTTPResponce(core_logger.Nop(), rec)

	data := map[string]string{"status": "ok"}
	h.JSONResponse(data, http.StatusOK)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", ct)
	}

	var parsed map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &parsed); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	if parsed["status"] != "ok" {
		t.Errorf("Expected status 'ok', got %s", parsed["status"])
	}
}

func TestErrorResponseStatusMapping(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
	}{
		{"invalid argument", core_errors.ErrInvalidArgument, http.StatusBadRequest},
		{"not found", core_errors.ErrNotFound, http.StatusNotFound},
		{"conflict", core_errors.ErrConflict, http.StatusConflict},
		{"unauthorized", core_errors.ErrUnauthorized, http.StatusUnauthorized},
		{"too many requests", core_errors.ErrTooManyRequests, http.StatusTooManyRequests},
		{"unknown internal", errors.New("something went wrong"), http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			h := NewHTTPResponce(core_logger.Nop(), rec)

			h.ErrorResponse(tt.err, "operation failed")

			if rec.Code != tt.wantStatus {
				t.Errorf("ErrorResponse(%v) status = %d; want %d", tt.err, rec.Code, tt.wantStatus)
			}
			if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
				t.Errorf("Expected Content-Type application/json, got %s", ct)
			}

			var body map[string]string
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatalf("Failed to parse error json: %v", err)
			}
			if body["message"] != "operation failed" {
				t.Errorf("Expected message 'operation failed', got %s", body["message"])
			}
		})
	}
}

func TestNoContentResponse(t *testing.T) {
	rec := httptest.NewRecorder()
	h := NewHTTPResponce(core_logger.Nop(), rec)

	h.NoContentResponse()

	if rec.Code != http.StatusNoContent {
		t.Errorf("Expected 204, got %d", rec.Code)
	}
}
