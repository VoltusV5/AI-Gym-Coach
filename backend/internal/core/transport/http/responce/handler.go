// Package core_http_responce формирует HTTP-ответы, сериализует JSON и сопоставляет ошибки со статусами.
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

// HTTPResponceHandler формирует HTTP-ответы и записывает ошибки в журнал.
type HTTPResponceHandler struct {
	log       *core_logger.Logger
	rw        http.ResponseWriter
	requestID *string
}

// NewHTTPResponce создаёт обработчик ответов с журналом log и потоком ответа rw.
func NewHTTPResponce(
	log *core_logger.Logger,
	rw http.ResponseWriter,
) *HTTPResponceHandler {
	return &HTTPResponceHandler{
		log: log,
		rw:  rw,
	}
}

// JSONResponse сериализует responseBody, устанавливает Content-Type: application/json и записывает statusCode и тело ответа. Ошибку записи сохраняет в журнал.
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

// NoContentResponse записывает HTTP 204 без тела ответа.
func (h *HTTPResponceHandler) NoContentResponse() {
	h.rw.WriteHeader(http.StatusNoContent)
}

// ErrorResponse сопоставляет err с HTTP-статусом, записывает ошибку и контекст msg в журнал и отправляет JSON с полями message и error. Неизвестной ошибке соответствует HTTP 500.
func (h *HTTPResponceHandler) ErrorResponse(err error, msg string) {
	if h.requestID != nil {
		h.errorResponseV2(err, msg)
		return
	}
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

// PanicResponce записывает восстановленное значение паники p и описание msg в журнал и отправляет HTTP 500.
func (h *HTTPResponceHandler) PanicResponce(p any, msg string) {
	statusCode := http.StatusInternalServerError
	err := fmt.Errorf("unexpected panic: %v", p)
	if h.requestID != nil {
		h.errorResponseV2(err, msg)
		return
	}

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
