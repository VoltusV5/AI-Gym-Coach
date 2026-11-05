package generation_transport_http

import (
	"net/http"
	core_auth "sport_app/internal/core/auth"
	core_logger "sport_app/internal/core/logger"
	core_http_middleware "sport_app/internal/core/transport/http/middleware"
	core_http_responce "sport_app/internal/core/transport/http/responce"
	core_http_server "sport_app/internal/core/transport/http/server"
	"sport_app/internal/core/transport/http/v2request"
	generation_service "sport_app/internal/features/generation/service"
)

type Handler struct {
	service *generation_service.Service
	jwt     *core_auth.JWT
}

func NewHandler(service *generation_service.Service, jwt *core_auth.JWT) *Handler {
	return &Handler{service: service, jwt: jwt}
}
func (h *Handler) Routes() []core_http_server.Route {
	return []core_http_server.Route{core_http_server.NewRoute("POST", "/plans/generate", h.generate, core_http_middleware.ProtectV2(h.jwt))}
}
func (h *Handler) generate(w http.ResponseWriter, r *http.Request) {
	resp := core_http_responce.NewHTTPResponceV2(core_logger.FromContext(r.Context()), w, r.Header.Get("X-Request-ID"))
	var b struct {
		ProfileVersion int64 `json:"profile_version"`
	}
	if _, err := v2request.Decode(w, r, &b, "profile_version"); err != nil {
		resp.ErrorResponse(err, "generation input")
		return
	}
	id, err := core_http_middleware.UserIDFromContext(r.Context())
	if err != nil {
		resp.ErrorResponse(err, "generation authentication")
		return
	}
	p, err := h.service.Generate(r.Context(), id, r.Header.Get("Idempotency-Key"), r.Header.Get("X-Request-ID"), b.ProfileVersion)
	if err != nil {
		resp.ErrorResponse(err, "generation")
		return
	}
	resp.JSONResponse(p, 201)
}
