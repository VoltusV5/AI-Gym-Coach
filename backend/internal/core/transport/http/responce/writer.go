package core_http_responce

import "net/http"

type ResponceWriter struct {
	http.ResponseWriter
	statusCode int
}

var (
	StatusCodeUninitialized = -1
)

func NewResponceWriter(w http.ResponseWriter) *ResponceWriter {
	return &ResponceWriter{
		ResponseWriter: w,
		statusCode:     StatusCodeUninitialized,
	}
}

func (rw *ResponceWriter) WriteHeader(statusCode int) {
	rw.ResponseWriter.WriteHeader(statusCode)
	rw.statusCode = statusCode
}

func (rw *ResponceWriter) GetStatusCode() int {
	if rw.statusCode == StatusCodeUninitialized {
		return http.StatusOK
	}

	return rw.statusCode
}

func (rw *ResponceWriter) Flush() {
	if fl, ok := rw.ResponseWriter.(http.Flusher); ok {
		fl.Flush()
	}
}

