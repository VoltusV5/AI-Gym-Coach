package sharing_transport_http

import (
	"net/http"
	core_auth "sport_app/internal/core/auth"
	core_logger "sport_app/internal/core/logger"
	core_http_middleware "sport_app/internal/core/transport/http/middleware"
	core_http_responce "sport_app/internal/core/transport/http/responce"
	core_http_server "sport_app/internal/core/transport/http/server"
	"sport_app/internal/core/transport/http/v2request"
	sharing_postgres_repository "sport_app/internal/features/sharing/repository/postgres"
	sharing_service "sport_app/internal/features/sharing/service"
)

type Handler struct {
	service *sharing_service.Service
	jwt     *core_auth.JWT
}

func NewHandler(s *sharing_service.Service, jwt *core_auth.JWT) *Handler { return &Handler{s, jwt} }
func (h *Handler) Routes() []core_http_server.Route {
	p := core_http_middleware.ProtectV2(h.jwt)
	o := core_http_middleware.OptionalProtectV2(h.jwt)
	return []core_http_server.Route{
		core_http_server.NewRoute("GET", "/me/saved-shares", h.saved, p), core_http_server.NewRoute("GET", "/me/saved-shares/{item_id}", h.getSaved, p), core_http_server.NewRoute("DELETE", "/me/saved-shares/{item_id}", h.removeSaved, p), core_http_server.NewRoute("POST", "/me/saved-shares/{item_id}/fork", h.forkSaved, p),
		core_http_server.NewRoute("POST", "/programs/{program_id}/fork", h.forkProgram, p), core_http_server.NewRoute("POST", "/shares", h.create, p), core_http_server.NewRoute("GET", "/shares/{token}", h.get, o), core_http_server.NewRoute("DELETE", "/shares/{share_id}", h.revoke, p), core_http_server.NewRoute("POST", "/shares/{token}/save", h.save, p), core_http_server.NewRoute("POST", "/shares/{token}/fork", h.forkShare, p),
	}
}
func actor(r *http.Request) string {
	id, _ := core_http_middleware.UserIDFromContext(r.Context())
	return id
}
func finish(w http.ResponseWriter, r *http.Request, v any, err error, status int) {
	p := core_http_responce.NewHTTPResponceV2(core_logger.FromContext(r.Context()), w, r.Header.Get("X-Request-ID"))
	if err != nil {
		p.ErrorResponse(err, "sharing operation")
		return
	}
	if status == 204 {
		p.NoContentResponse()
		return
	}
	p.JSONResponse(v, status)
}
func (h *Handler) forkProgram(w http.ResponseWriter, r *http.Request) {
	var in sharing_postgres_repository.ForkInput
	if _, err := v2request.Decode(w, r, &in, "source_version_id", "title"); err != nil {
		finish(w, r, nil, err, 0)
		return
	}
	v, err := h.service.ForkProgram(r.Context(), actor(r), r.PathValue("program_id"), r.Header.Get("Idempotency-Key"), in)
	finish(w, r, v, err, 201)
}
func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var in sharing_postgres_repository.ShareInput
	if _, err := v2request.Decode(w, r, &in, "target_type", "target_id"); err != nil {
		finish(w, r, nil, err, 0)
		return
	}
	v, err := h.service.Create(r.Context(), actor(r), in)
	finish(w, r, v, err, 201)
}
func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	v, err := h.service.Get(r.Context(), actor(r), r.PathValue("token"))
	finish(w, r, v, err, 200)
}
func (h *Handler) revoke(w http.ResponseWriter, r *http.Request) {
	err := h.service.Revoke(r.Context(), actor(r), r.PathValue("share_id"))
	finish(w, r, nil, err, 204)
}
func (h *Handler) save(w http.ResponseWriter, r *http.Request) {
	v, err := h.service.Save(r.Context(), actor(r), r.PathValue("token"))
	finish(w, r, map[string]string{"library_item_id": v}, err, 201)
}
func (h *Handler) forkShare(w http.ResponseWriter, r *http.Request) {
	v, err := h.service.ForkShare(r.Context(), actor(r), r.PathValue("token"), r.Header.Get("Idempotency-Key"))
	finish(w, r, v, err, 201)
}

func (h *Handler) saved(w http.ResponseWriter, r *http.Request) {
	f, err := v2request.Pagination(r, "saved-shares:"+actor(r))
	if err != nil {
		finish(w, r, nil, err, 0)
		return
	}
	v, err := h.service.Saved(r.Context(), actor(r), f)
	finish(w, r, v, err, 200)
}
func (h *Handler) getSaved(w http.ResponseWriter, r *http.Request) {
	v, err := h.service.GetSaved(r.Context(), actor(r), r.PathValue("item_id"))
	finish(w, r, v, err, 200)
}
func (h *Handler) removeSaved(w http.ResponseWriter, r *http.Request) {
	err := h.service.RemoveSaved(r.Context(), actor(r), r.PathValue("item_id"))
	finish(w, r, nil, err, 204)
}
func (h *Handler) forkSaved(w http.ResponseWriter, r *http.Request) {
	v, err := h.service.ForkSaved(r.Context(), actor(r), r.PathValue("item_id"), r.Header.Get("Idempotency-Key"))
	finish(w, r, v, err, 201)
}
