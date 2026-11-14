package insights_transport_http

import (
	"net/http"
	core_auth "sport_app/internal/core/auth"
	core_logger "sport_app/internal/core/logger"
	middleware "sport_app/internal/core/transport/http/middleware"
	response "sport_app/internal/core/transport/http/responce"
	server "sport_app/internal/core/transport/http/server"
	"sport_app/internal/core/transport/http/v2request"
	service "sport_app/internal/features/insights/service"
	"strconv"
)

type Handler struct {
	service *service.Service
	jwt     *core_auth.JWT
}

func NewHandler(s *service.Service, jwt *core_auth.JWT) *Handler { return &Handler{s, jwt} }
func (h *Handler) Routes() []server.Route {
	p := middleware.ProtectV2(h.jwt)
	return []server.Route{server.NewRoute("GET", "/recommendations/programs", h.recommendations, p), server.NewRoute("GET", "/dashboard", h.dashboard, p), server.NewRoute("GET", "/trainer/analytics", h.analytics, p), server.NewRoute("GET", "/admin/analytics", h.analytics, p)}
}
func actor(r *http.Request) string { id, _ := middleware.UserIDFromContext(r.Context()); return id }
func finish(w http.ResponseWriter, r *http.Request, value any, err error) {
	resp := response.NewHTTPResponceV2(core_logger.FromContext(r.Context()), w, r.Header.Get("X-Request-ID"))
	if err != nil {
		resp.ErrorResponse(err, "insights operation")
		return
	}
	resp.JSONResponse(value, 200)
}
func (h *Handler) recommendations(w http.ResponseWriter, r *http.Request) {
	limit := 10
	var err error
	if raw := r.URL.Query().Get("limit"); raw != "" {
		limit, err = strconv.Atoi(raw)
	}
	if err != nil || r.URL.Query().Get("cursor") != "" {
		finish(w, r, nil, v2request.Invalid())
		return
	}
	value, err := h.service.Recommendations(r.Context(), actor(r), limit)
	finish(w, r, value, err)
}
func (h *Handler) dashboard(w http.ResponseWriter, r *http.Request) {
	value, err := h.service.Dashboard(r.Context(), actor(r))
	finish(w, r, value, err)
}
func (h *Handler) analytics(w http.ResponseWriter, r *http.Request) {
	value, err := h.service.Analytics(r.Context(), actor(r), r.URL.Path == "/admin/analytics", r.URL.Query().Get("from"), r.URL.Query().Get("to"))
	finish(w, r, value, err)
}
