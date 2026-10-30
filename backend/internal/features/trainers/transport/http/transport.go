package trainers_transport_http

import (
	"net/http"
	core_auth "sport_app/internal/core/auth"
	core_logger "sport_app/internal/core/logger"
	core_http_middleware "sport_app/internal/core/transport/http/middleware"
	core_http_responce "sport_app/internal/core/transport/http/responce"
	core_http_server "sport_app/internal/core/transport/http/server"
	"sport_app/internal/core/transport/http/v2request"
	trainers_postgres_repository "sport_app/internal/features/trainers/repository/postgres"
	trainers_service "sport_app/internal/features/trainers/service"
)

type Handler struct {
	service *trainers_service.Service
	jwt     *core_auth.JWT
}

func NewHandler(service *trainers_service.Service, jwt *core_auth.JWT) *Handler {
	return &Handler{service: service, jwt: jwt}
}
func (h *Handler) Routes() []core_http_server.Route {
	protect := core_http_middleware.ProtectV2(h.jwt)
	return []core_http_server.Route{
		core_http_server.NewRoute("GET", "/trainer/profile", h.profile, protect),
		core_http_server.NewRoute("PATCH", "/trainer/profile", h.update, protect),
		core_http_server.NewRoute("GET", "/trainers/{trainer_id}", h.profile),
		core_http_server.NewRoute("GET", "/trainer/verification", h.verification, protect),
		core_http_server.NewRoute("POST", "/trainer/verification", h.submit, protect),
		core_http_server.NewRoute("GET", "/admin/verifications", h.verifications, protect),
		core_http_server.NewRoute("PATCH", "/admin/verifications/{verification_id}", h.decide, protect),
		core_http_server.NewRoute("GET", "/admin/trainers", h.trainers, protect),
	}
}
func response(w http.ResponseWriter, r *http.Request) *core_http_responce.HTTPResponceHandler {
	return core_http_responce.NewHTTPResponceV2(core_logger.FromContext(r.Context()), w, r.Header.Get("X-Request-ID"))
}
func actor(r *http.Request) string {
	id, _ := core_http_middleware.UserIDFromContext(r.Context())
	return id
}
func finish(w http.ResponseWriter, r *http.Request, value any, err error, status int) {
	resp := response(w, r)
	if err != nil {
		resp.ErrorResponse(err, "trainer operation")
		return
	}
	resp.JSONResponse(value, status)
}
func (h *Handler) profile(w http.ResponseWriter, r *http.Request) {
	value, err := h.service.Profile(r.Context(), actor(r), r.PathValue("trainer_id"))
	finish(w, r, value, err, 200)
}
func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	var p trainers_postgres_repository.Patch
	fields, err := v2request.Decode(w, r, &p, "display_name")
	if err != nil {
		finish(w, r, nil, err, 0)
		return
	}
	_, p.PhotoSet = fields["photo_ref"]
	value, err := h.service.Update(r.Context(), actor(r), p)
	finish(w, r, value, err, 200)
}
func (h *Handler) verification(w http.ResponseWriter, r *http.Request) {
	value, err := h.service.Verification(r.Context(), actor(r))
	finish(w, r, value, err, 200)
}
func (h *Handler) submit(w http.ResponseWriter, r *http.Request) {
	var p trainers_postgres_repository.VerificationInput
	if _, err := v2request.Decode(w, r, &p, "qualification", "document_refs"); err != nil {
		finish(w, r, nil, err, 0)
		return
	}
	value, err := h.service.Submit(r.Context(), actor(r), p)
	finish(w, r, value, err, 201)
}
func (h *Handler) decide(w http.ResponseWriter, r *http.Request) {
	var p trainers_postgres_repository.Decision
	if _, err := v2request.Decode(w, r, &p, "status", "reason"); err != nil {
		finish(w, r, nil, err, 0)
		return
	}
	value, err := h.service.Decide(r.Context(), actor(r), r.PathValue("verification_id"), p, r.Header.Get("X-Request-ID"))
	finish(w, r, value, err, 200)
}
func (h *Handler) verifications(w http.ResponseWriter, r *http.Request) {
	f, err := v2request.Pagination(r, "verifications:"+actor(r))
	if err != nil {
		finish(w, r, nil, err, 0)
		return
	}
	value, err := h.service.Verifications(r.Context(), actor(r), f)
	finish(w, r, value, err, 200)
}
func (h *Handler) trainers(w http.ResponseWriter, r *http.Request) {
	f, err := v2request.Pagination(r, "trainers:"+actor(r))
	if err != nil {
		finish(w, r, nil, err, 0)
		return
	}
	value, err := h.service.Trainers(r.Context(), actor(r), f)
	finish(w, r, value, err, 200)
}
