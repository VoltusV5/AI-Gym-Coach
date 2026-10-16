// Package core_http_responce provides HTTP response helpers, JSON serialization,
// error mapping to status codes, and panic recovery handling.
package core_http_responce

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	core_errors "sport_app/internal/core/errors"
	core_logger "sport_app/internal/core/logger"

	"go.uber.org/zap"
)

// HTTPResponceHandler encapsulates structured response writing and error logging.
type HTTPResponceHandler struct {
	log *core_logger.Logger
	rw  http.ResponseWriter
}

// NewHTTPResponce constructs an HTTPResponceHandler with the provided logger and ResponseWriter.
//
// Parameters:
//   - log: Scoped zap logger for reporting response errors.
//   - rw: Underlying HTTP response writer.
//
// Returns:
//   - *HTTPResponceHandler: Initialized response helper.
func NewHTTPResponce(
	log *core_logger.Logger,
	rw http.ResponseWriter,
) *HTTPResponceHandler {
	return &HTTPResponceHandler{
		log: log,
		rw:  rw,
	}
}

// JSONResponse serializes responseBody into JSON, sets the Content-Type header to application/json,
// and writes the provided HTTP status code.
//
// Parameters:
//   - responseBody: Data structure to serialize as JSON.
//   - statusCode: HTTP status code to emit (e.g., 200, 201, 400).
//
// Side effects:
//   - Sets HTTP header "Content-Type: application/json".
//   - Emits HTTP status code and writes JSON bytes to response stream.
func (h *HTTPResponceHandler) JSONResponse(
	responseBody any,
	statusCode int,
) {
	h.rw.Header().Set("Content-Type", "application/json")
	h.rw.WriteHeader(statusCode)

	if err := json.NewEncoder(h.rw).Encode(responseBody); err != nil {
		h.log.Error("write HTTP response", zap.Error(err))
	}
}

// NoContentResponse writes HTTP 204 No Content status with an empty body.
//
// Side effects:
//   - Calls WriteHeader(http.StatusNoContent).
func (h *HTTPResponceHandler) NoContentResponse() {
	h.rw.WriteHeader(http.StatusNoContent)
}

// ErrorResponse maps domain errors (ErrInvalidArgument, ErrNotFound, ErrConflict, ErrUnauthorized, ErrTooManyRequests)
// to corresponding HTTP status codes (400, 404, 409, 401, 429, or 500), logs the error with appropriate level,
// and outputs a JSON body `{"message": "...", "error": "..."}`.
//
// Parameters:
//   - err: Source error to inspect and include in response.
//   - msg: Human-readable context message.
//
// Side effects:
//   - Logs warning/error to logger.
//   - Writes HTTP status header and JSON payload.
func (h *HTTPResponceHandler) ErrorResponse(err error, msg string) {
	var (
		statusCode int
		logFunc    func(string, ...zap.Field)
	)

	switch {
	case errors.Is(err, core_errors.ErrInvalidArgument):
		statusCode = http.StatusBadRequest
		logFunc = h.log.Warn

	case errors.Is(err, core_errors.ErrNotFound):
		statusCode = http.StatusNotFound
		logFunc = h.log.Debug

	case errors.Is(err, core_errors.ErrConflict):
		statusCode = http.StatusConflict
		logFunc = h.log.Warn

	case errors.Is(err, core_errors.ErrUnauthorized):
		statusCode = http.StatusUnauthorized
		logFunc = h.log.Warn

	case errors.Is(err, core_errors.ErrTooManyRequests):
		statusCode = http.StatusTooManyRequests
		logFunc = h.log.Warn

	default:
		statusCode = http.StatusInternalServerError
		logFunc = h.log.Error
	}

	logFunc(msg, zap.Error(err))

	h.errorResponse(
		statusCode,
		err,
		msg,
	)
}

// PanicResponce logs panic recovery information at Error level and returns HTTP 500 Internal Server Error.
//
// Parameters:
//   - p: Value recovered from panic (any).
//   - msg: Description of where the panic occurred.
func (h *HTTPResponceHandler) PanicResponce(p any, msg string) {
	statusCode := http.StatusInternalServerError
	err := fmt.Errorf("unexpected panic: %v", p)

	h.log.Error(msg, zap.Error(err))
	h.errorResponse(
		statusCode,
		err,
		msg,
	)
}

func (h *HTTPResponceHandler) errorResponse(
	statusCode int,
	err error,
	msg string,
) {
	responce := map[string]string{
		"message": msg,
		"error":   err.Error(),
	}

	h.JSONResponse(
		responce,
		statusCode,
	)
}
