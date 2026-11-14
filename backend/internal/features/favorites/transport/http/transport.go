package favorites_transport_http

import (
	"net/http"
	core_auth "sport_app/internal/core/auth"
	core_logger "sport_app/internal/core/logger"
	core_http_middleware "sport_app/internal/core/transport/http/middleware"
	core_http_responce "sport_app/internal/core/transport/http/responce"
	core_http_server "sport_app/internal/core/transport/http/server"
	"sport_app/internal/core/transport/http/v2request"
	favorites_service "sport_app/internal/features/favorites/service"
)

type Handler struct {
	service *favorites_service.Service
	jwt     *core_auth.JWT
}

func NewHandler(s *favorites_service.Service, jwt *core_auth.JWT) *Handler { return &Handler{s, jwt} }
func (h *Handler) Routes() []core_http_server.Route {
	p := core_http_middleware.ProtectV2(h.jwt)
	return []core_http_server.Route{core_http_server.NewRoute("GET", "/me/favorites", h.list, p), core_http_server.NewRoute("PUT", "/me/favorites/{program_id}", h.set, p), core_http_server.NewRoute("DELETE", "/me/favorites/{program_id}", h.set, p)}
}
func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	id, _ := core_http_middleware.UserIDFromContext(r.Context())
	resp := core_http_responce.NewHTTPResponceV2(core_logger.FromContext(r.Context()), w, r.Header.Get("X-Request-ID"))
	f, err := v2request.Pagination(r, r.URL.Path+":"+id)
	if err != nil {
		resp.ErrorResponse(err, "favorites pagination")
		return
	}
	v, err := h.service.List(r.Context(), id, f)
	if err != nil {
		resp.ErrorResponse(err, "favorites list")
		return
	}
	resp.JSONResponse(v, 200)
}
func (h *Handler) set(w http.ResponseWriter, r *http.Request) {
	id, _ := core_http_middleware.UserIDFromContext(r.Context())
	resp := core_http_responce.NewHTTPResponceV2(core_logger.FromContext(r.Context()), w, r.Header.Get("X-Request-ID"))
	if err := h.service.Set(r.Context(), id, r.PathValue("program_id"), r.Method == "PUT"); err != nil {
		resp.ErrorResponse(err, "favorite change")
		return
	}
	if r.Method == "DELETE" {
		resp.NoContentResponse()
		return
	}
	resp.JSONResponse(map[string]bool{"ok": true}, 200)
}
