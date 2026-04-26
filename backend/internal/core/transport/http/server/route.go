package core_http_server

import (
	"net/http"

	core_http_middleware "sport_app/internal/core/transport/http/middleware"
)

type Route struct {
	Method      string
	Path        string
	Handler     http.HandlerFunc
	Middlewares []core_http_middleware.Middleware
}

func NewRoute(
	method string,
	path string,
	handler http.HandlerFunc,
	middlewares ...core_http_middleware.Middleware,
) Route {
	return Route{
		Method:      method,
		Path:        path,
		Handler:     handler,
		Middlewares: middlewares,
	}
}
