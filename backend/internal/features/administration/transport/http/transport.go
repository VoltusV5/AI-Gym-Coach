package administration_transport_http

import (
	"net/http"
	core_auth "sport_app/internal/core/auth"
	core_logger "sport_app/internal/core/logger"
	core_http_middleware "sport_app/internal/core/transport/http/middleware"
	core_http_responce "sport_app/internal/core/transport/http/responce"
	core_http_server "sport_app/internal/core/transport/http/server"
	"sport_app/internal/core/transport/http/v2request"
	administration_postgres_repository "sport_app/internal/features/administration/repository/postgres"
	administration_service "sport_app/internal/features/administration/service"
	"strings"
)

type Handler struct {
	service *administration_service.Service
	jwt     *core_auth.JWT
}

func NewHandler(service *administration_service.Service, jwt *core_auth.JWT) *Handler {
	return &Handler{service, jwt}
}
func (h *Handler) Routes() []core_http_server.Route {
	protect := core_http_middleware.ProtectV2(h.jwt)
	return []core_http_server.Route{
		core_http_server.NewRoute("GET", "/catalog/categories", h.references),
		core_http_server.NewRoute("GET", "/admin/categories", h.references, protect), core_http_server.NewRoute("POST", "/admin/categories", h.saveReference, protect), core_http_server.NewRoute("PATCH", "/admin/categories/{category_id}", h.saveReference, protect),
		core_http_server.NewRoute("GET", "/admin/exercises", h.references, protect), core_http_server.NewRoute("POST", "/admin/exercises", h.saveReference, protect), core_http_server.NewRoute("PATCH", "/admin/exercises/{exercise_id}", h.saveReference, protect), core_http_server.NewRoute("GET", "/me/session", h.session, protect), core_http_server.NewRoute("GET", "/admin/users", h.list, protect), core_http_server.NewRoute("PATCH", "/admin/users/{user_id}", h.update, protect)}
}
func actor(r *http.Request) string {
	id, _ := core_http_middleware.UserIDFromContext(r.Context())
	return id
}
func finish(w http.ResponseWriter, r *http.Request, v any, err error) {
	resp := core_http_responce.NewHTTPResponceV2(core_logger.FromContext(r.Context()), w, r.Header.Get("X-Request-ID"))
	if err != nil {
		resp.ErrorResponse(err, "account operation")
		return
	}
	resp.JSONResponse(v, http.StatusOK)
}
func (h *Handler) session(w http.ResponseWriter, r *http.Request) {
	v, err := h.service.Session(r.Context(), actor(r))
	finish(w, r, v, err)
}
func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	f, err := v2request.Pagination(r, r.URL.Path+":"+actor(r)+":"+r.URL.Query().Get("query"))
	if err != nil {
		finish(w, r, nil, err)
		return
	}
	v, err := h.service.List(r.Context(), actor(r), f, r.URL.Query().Get("query"))
	finish(w, r, v, err)
}
func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	var p administration_postgres_repository.Change
	if _, err := v2request.Decode(w, r, &p, "reason"); err != nil {
		finish(w, r, nil, err)
		return
	}
	v, err := h.service.Update(r.Context(), actor(r), r.PathValue("user_id"), p, r.Header.Get("X-Request-ID"))
	finish(w, r, v, err)
}

func (h *Handler) references(w http.ResponseWriter, r *http.Request) {
	kind := "categories"
	if strings.Contains(r.URL.Path, "/exercises") {
		kind = "exercises"
	}
	f, err := v2request.Pagination(r, r.URL.Path+":"+actor(r)+":"+r.URL.Query().Get("query"))
	if err != nil {
		finish(w, r, nil, err)
		return
	}
	v, err := h.service.References(r.Context(), actor(r), kind, strings.HasPrefix(r.URL.Path, "/admin/"), f, r.URL.Query().Get("query"))
	finish(w, r, v, err)
}
func (h *Handler) saveReference(w http.ResponseWriter, r *http.Request) {
	kind, target := "categories", r.PathValue("category_id")
	required := []string{"slug", "name", "reason"}
	if strings.Contains(r.URL.Path, "/exercises") {
		kind, target = "exercises", r.PathValue("exercise_id")
		required = []string{"name", "equipment_ids", "reason"}
	}
	var in administration_postgres_repository.ReferenceInput
	fields, err := v2request.Decode(w, r, &in, required...)
	if err != nil {
		finish(w, r, nil, err)
		return
	}
	_, in.ImageSet = fields["image_url"]
	_, in.VideoSet = fields["video_url"]
	_, in.SubgroupSet = fields["muscle_subgroup"]
	v, err := h.service.SaveReference(r.Context(), actor(r), kind, target, in, r.Header.Get("X-Request-ID"))
	if err != nil {
		finish(w, r, nil, err)
		return
	}
	status := 200
	if r.Method == "POST" {
		status = 201
	}
	core_http_responce.NewHTTPResponceV2(core_logger.FromContext(r.Context()), w, r.Header.Get("X-Request-ID")).JSONResponse(v, status)
}
