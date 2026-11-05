package workouts_transport_http

import (
	"net/http"
	core_auth "sport_app/internal/core/auth"
	core_logger "sport_app/internal/core/logger"
	core_http_middleware "sport_app/internal/core/transport/http/middleware"
	core_http_responce "sport_app/internal/core/transport/http/responce"
	core_http_server "sport_app/internal/core/transport/http/server"
	"sport_app/internal/core/transport/http/v2request"
	workouts_domain "sport_app/internal/features/workouts/domain"
	workouts_service "sport_app/internal/features/workouts/service"
	"strconv"
)

type Handler struct {
	service *workouts_service.Service
	jwt     *core_auth.JWT
}

func NewHandler(service *workouts_service.Service, jwt *core_auth.JWT) *Handler {
	return &Handler{service: service, jwt: jwt}
}
func (h *Handler) Routes() []core_http_server.Route {
	protect := core_http_middleware.ProtectV2(h.jwt)
	return []core_http_server.Route{
		core_http_server.NewRoute("POST", "/workouts/complete", h.complete, protect), core_http_server.NewRoute("GET", "/workouts", h.list, protect),
		core_http_server.NewRoute("GET", "/workouts/{session_id}", h.get, protect), core_http_server.NewRoute("POST", "/progress/weight", h.weight, protect),
		core_http_server.NewRoute("GET", "/progress", h.progress, protect), core_http_server.NewRoute("GET", "/progress/exercises/{exercise_id}", h.exercise, protect),
		core_http_server.NewRoute("GET", "/progress/summaries", h.summaries, protect),
	}
}
func actor(r *http.Request) string {
	id, _ := core_http_middleware.UserIDFromContext(r.Context())
	return id
}
func finish(w http.ResponseWriter, r *http.Request, value any, err error, status int) {
	resp := core_http_responce.NewHTTPResponceV2(core_logger.FromContext(r.Context()), w, r.Header.Get("X-Request-ID"))
	if err != nil {
		resp.ErrorResponse(err, "workout operation")
		return
	}
	resp.JSONResponse(value, status)
}
func (h *Handler) complete(w http.ResponseWriter, r *http.Request) {
	var b workouts_domain.Input
	if _, err := v2request.Decode(w, r, &b, "client_session_id", "plan_id", "plan_version", "plan_day_id", "started_at", "finished_at", "slots"); err != nil {
		finish(w, r, nil, err, 0)
		return
	}
	p, err := h.service.Complete(r.Context(), actor(r), r.Header.Get("Idempotency-Key"), b)
	finish(w, r, p, err, 201)
}
func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	p, err := h.service.Get(r.Context(), actor(r), r.PathValue("session_id"))
	finish(w, r, p, err, 200)
}
func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	limit, err := queryLimit(r)
	if err != nil {
		finish(w, r, nil, err, 0)
		return
	}
	q := r.URL.Query()
	p, err := h.service.List(r.Context(), actor(r), workouts_service.Query{Limit: limit, Cursor: q.Get("cursor"), From: q.Get("from"), To: q.Get("to"), PlanID: q.Get("plan_id")})
	finish(w, r, p, err, 200)
}
func (h *Handler) weight(w http.ResponseWriter, r *http.Request) {
	var b workouts_domain.Weight
	if _, err := v2request.Decode(w, r, &b, "day", "weight_kg"); err != nil {
		finish(w, r, nil, err, 0)
		return
	}
	err := h.service.Weight(r.Context(), actor(r), b)
	finish(w, r, map[string]bool{"ok": true}, err, 201)
}
func (h *Handler) progress(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	p, err := h.service.Progress(r.Context(), actor(r), q.Get("period"), q.Get("anchor_date"), q.Get("time_zone"))
	finish(w, r, p, err, 200)
}
func (h *Handler) exercise(w http.ResponseWriter, r *http.Request) {
	p, err := h.service.Exercise(r.Context(), actor(r), r.PathValue("exercise_id"))
	finish(w, r, p, err, 200)
}
func (h *Handler) summaries(w http.ResponseWriter, r *http.Request) {
	limit, err := queryLimit(r)
	if err != nil {
		finish(w, r, nil, err, 0)
		return
	}
	q := r.URL.Query()
	p, err := h.service.Summaries(r.Context(), actor(r), workouts_service.SummaryQuery{Limit: limit, Cursor: q.Get("cursor"), Period: q.Get("period"), TimeZone: q.Get("time_zone")})
	finish(w, r, p, err, 200)
}

func queryLimit(r *http.Request) (int, error) {
	if !r.URL.Query().Has("limit") {
		return 20, nil
	}
	n, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil || n < 1 || n > 100 {
		return 0, v2request.Invalid()
	}
	return n, nil
}
