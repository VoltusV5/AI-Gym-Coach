package publications_transport_http

import (
	"net/http"
	core_auth "sport_app/internal/core/auth"
	core_logger "sport_app/internal/core/logger"
	core_http_middleware "sport_app/internal/core/transport/http/middleware"
	core_http_responce "sport_app/internal/core/transport/http/responce"
	core_http_server "sport_app/internal/core/transport/http/server"
	"sport_app/internal/core/transport/http/v2request"
	publications_postgres_repository "sport_app/internal/features/publications/repository/postgres"
	publications_service "sport_app/internal/features/publications/service"
	"strings"
)

type Handler struct {
	service *publications_service.Service
	jwt     *core_auth.JWT
}

func NewHandler(service *publications_service.Service, jwt *core_auth.JWT) *Handler {
	return &Handler{service: service, jwt: jwt}
}
func (h *Handler) Routes() []core_http_server.Route {
	protect := core_http_middleware.ProtectV2(h.jwt)
	return []core_http_server.Route{
		core_http_server.NewRoute("GET", "/trainer/programs", h.list, protect), core_http_server.NewRoute("POST", "/trainer/programs", h.create, protect),
		core_http_server.NewRoute("GET", "/trainer/programs/{program_id}", h.get, protect), core_http_server.NewRoute("PATCH", "/trainer/programs/{program_id}", h.update, protect),
		core_http_server.NewRoute("POST", "/trainer/programs/{program_id}/publish", h.publish, protect), core_http_server.NewRoute("POST", "/trainer/programs/{program_id}/versions", h.newVersion, protect),
		core_http_server.NewRoute("GET", "/admin/programs", h.list, protect), core_http_server.NewRoute("PATCH", "/admin/programs/{program_id}/moderation", h.moderate, protect),
	}
}
func actor(r *http.Request) string {
	id, _ := core_http_middleware.UserIDFromContext(r.Context())
	return id
}
func finish(w http.ResponseWriter, r *http.Request, value any, err error, status int) {
	resp := core_http_responce.NewHTTPResponceV2(core_logger.FromContext(r.Context()), w, r.Header.Get("X-Request-ID"))
	if err != nil {
		resp.ErrorResponse(err, "publication operation")
		return
	}
	resp.JSONResponse(value, status)
}
func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	p, err := h.service.Get(r.Context(), actor(r), r.PathValue("program_id"))
	finish(w, r, p, err, 200)
}
func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var d publications_postgres_repository.Draft
	if _, err := v2request.Decode(w, r, &d, "title", "description", "goal", "level", "duration_weeks", "sessions_per_week", "equipment_ids", "price_minor", "currency", "days"); err != nil {
		finish(w, r, nil, err, 0)
		return
	}
	p, err := h.service.Create(r.Context(), actor(r), d)
	finish(w, r, p, err, 201)
}
func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	var d publications_postgres_repository.Draft
	fields, err := v2request.Decode(w, r, &d, "version")
	if err != nil {
		finish(w, r, nil, err, 0)
		return
	}
	p, err := h.service.Update(r.Context(), actor(r), r.PathValue("program_id"), fields)
	finish(w, r, p, err, 200)
}
func (h *Handler) newVersion(w http.ResponseWriter, r *http.Request) {
	p, err := h.service.NewVersion(r.Context(), actor(r), r.PathValue("program_id"))
	finish(w, r, p, err, 201)
}
func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	f, err := v2request.Pagination(r, r.URL.Path+":"+actor(r))
	if err != nil {
		finish(w, r, nil, err, 0)
		return
	}
	p, err := h.service.List(r.Context(), actor(r), strings.HasSuffix(r.URL.Path, "/admin/programs"), f)
	finish(w, r, p, err, 200)
}
func (h *Handler) publish(w http.ResponseWriter, r *http.Request) {
	var b struct {
		Version int64 `json:"version"`
	}
	if _, err := v2request.Decode(w, r, &b, "version"); err != nil {
		finish(w, r, nil, err, 0)
		return
	}
	p, err := h.service.Publish(r.Context(), actor(r), r.PathValue("program_id"), b.Version)
	finish(w, r, p, err, 202)
}
func (h *Handler) moderate(w http.ResponseWriter, r *http.Request) {
	var m publications_postgres_repository.Moderation
	if _, err := v2request.Decode(w, r, &m, "status", "reason"); err != nil {
		finish(w, r, nil, err, 0)
		return
	}
	p, err := h.service.Moderate(r.Context(), actor(r), r.PathValue("program_id"), m, r.Header.Get("X-Request-ID"))
	finish(w, r, p, err, 200)
}
