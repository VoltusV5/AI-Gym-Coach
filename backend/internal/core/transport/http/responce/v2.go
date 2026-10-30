package core_http_responce

import (
	"context"
	"errors"
	"net/http"

	core_errors "sport_app/internal/core/errors"
	core_logger "sport_app/internal/core/logger"

	"go.uber.org/zap"
)

func NewHTTPResponceV2(log *core_logger.Logger, w http.ResponseWriter, requestID string) *HTTPResponceHandler {
	return &HTTPResponceHandler{log: log, rw: w, requestID: &requestID}
}

func (h *HTTPResponceHandler) errorResponseV2(err error, msg string) {
	status, code, message := http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error"
	switch {
	case errors.Is(err, core_errors.ErrInvalidArgument):
		status, code, message = http.StatusBadRequest, "INVALID_ARGUMENT", "Invalid request"
	case errors.Is(err, core_errors.ErrUnauthorized):
		status, code, message = http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required"
	case errors.Is(err, core_errors.ErrForbidden):
		status, code, message = http.StatusForbidden, "FORBIDDEN", "Registration required"
	case errors.Is(err, core_errors.ErrNotFound):
		status, code, message = http.StatusNotFound, "NOT_FOUND", "Object not found"
	case errors.Is(err, core_errors.ErrConflict):
		status, code, message = http.StatusConflict, "CONFLICT", "State conflict"
	case errors.Is(err, core_errors.ErrUnprocessable):
		status, code, message = http.StatusUnprocessableEntity, "FREE_ACCESS_UNAVAILABLE", "Free access unavailable"
	case errors.Is(err, context.DeadlineExceeded):
		status, code, message = http.StatusGatewayTimeout, "DEADLINE_EXCEEDED", "Request timed out"
	}
	var domain *core_errors.DomainError
	if errors.As(err, &domain) {
		status, code, message = domain.Status, domain.Code, domain.Message
	}
	h.log.Error(msg, zap.Error(err))
	h.JSONResponse(map[string]any{
		"error":      map[string]string{"code": code, "message": message},
		"request_id": *h.requestID,
	}, status)
}
