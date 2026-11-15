package social_transport_http

import (
	"net/http"
	core_auth "sport_app/internal/core/auth"
	core_logger "sport_app/internal/core/logger"
	core_http_middleware "sport_app/internal/core/transport/http/middleware"
	core_http_responce "sport_app/internal/core/transport/http/responce"
	core_http_server "sport_app/internal/core/transport/http/server"
	"sport_app/internal/core/transport/http/v2request"
	social_postgres_repository "sport_app/internal/features/social/repository/postgres"
	social_service "sport_app/internal/features/social/service"
	"strings"
)

type Handler struct {
	service *social_service.Service
	jwt     *core_auth.JWT
}

func NewHandler(s *social_service.Service, jwt *core_auth.JWT) *Handler { return &Handler{s, jwt} }
func (h *Handler) Routes() []core_http_server.Route {
	p := core_http_middleware.ProtectV2(h.jwt)
	o := core_http_middleware.OptionalProtectV2(h.jwt)
	return []core_http_server.Route{
		core_http_server.NewRoute("GET", "/programs/{program_id}/review-eligibility", h.eligibility, p), core_http_server.NewRoute("GET", "/programs/{program_id}/reviews", h.reviews, o), core_http_server.NewRoute("POST", "/programs/{program_id}/reviews", h.saveReview, p), core_http_server.NewRoute("PATCH", "/reviews/{review_id}", h.saveReview, p), core_http_server.NewRoute("DELETE", "/reviews/{review_id}", h.deleteReview, p), core_http_server.NewRoute("GET", "/admin/reviews", h.reviews, p), core_http_server.NewRoute("PATCH", "/admin/reviews/{review_id}/moderation", h.moderateReview, p),
		core_http_server.NewRoute("POST", "/reports", h.report, p), core_http_server.NewRoute("GET", "/admin/reports", h.reports, p), core_http_server.NewRoute("PATCH", "/admin/reports/{report_id}", h.decideReport, p),
		core_http_server.NewRoute("PUT", "/trainers/{trainer_id}/follow", h.follow, p), core_http_server.NewRoute("DELETE", "/trainers/{trainer_id}/follow", h.follow, p), core_http_server.NewRoute("GET", "/me/trainers", h.trainers, p),
		core_http_server.NewRoute("GET", "/notifications", h.notifications, p), core_http_server.NewRoute("PATCH", "/notifications/{notification_id}", h.readNotification, p), core_http_server.NewRoute("POST", "/events", h.events, p),
	}
}
func actor(r *http.Request) string {
	id, _ := core_http_middleware.UserIDFromContext(r.Context())
	return id
}
func finish(w http.ResponseWriter, r *http.Request, v any, err error, status int) {
	p := core_http_responce.NewHTTPResponceV2(core_logger.FromContext(r.Context()), w, r.Header.Get("X-Request-ID"))
	if err != nil {
		p.ErrorResponse(err, "social operation")
		return
	}
	if status == 204 {
		p.NoContentResponse()
		return
	}
	p.JSONResponse(v, status)
}
func page(r *http.Request) (v2request.PageFilter, error) {
	return v2request.Pagination(r, r.URL.Path+":"+actor(r))
}
func (h *Handler) eligibility(w http.ResponseWriter, r *http.Request) {
	v, err := h.service.Eligibility(r.Context(), actor(r), r.PathValue("program_id"))
	finish(w, r, v, err, 200)
}
func (h *Handler) reviews(w http.ResponseWriter, r *http.Request) {
	f, err := page(r)
	if err != nil {
		finish(w, r, nil, err, 0)
		return
	}
	v, err := h.service.Reviews(r.Context(), actor(r), r.PathValue("program_id"), strings.HasPrefix(r.URL.Path, "/admin/"), f)
	finish(w, r, v, err, 200)
}
func (h *Handler) saveReview(w http.ResponseWriter, r *http.Request) {
	var in social_postgres_repository.ReviewInput
	if _, err := v2request.Decode(w, r, &in, "overall"); err != nil {
		finish(w, r, nil, err, 0)
		return
	}
	v, err := h.service.SaveReview(r.Context(), actor(r), r.PathValue("program_id"), r.PathValue("review_id"), in)
	status := 200
	if r.Method == "POST" {
		status = 201
	}
	finish(w, r, v, err, status)
}
func (h *Handler) deleteReview(w http.ResponseWriter, r *http.Request) {
	err := h.service.DeleteReview(r.Context(), actor(r), r.PathValue("review_id"))
	finish(w, r, nil, err, 204)
}
func (h *Handler) moderateReview(w http.ResponseWriter, r *http.Request) {
	var in social_postgres_repository.Decision
	if _, err := v2request.Decode(w, r, &in, "status", "reason"); err != nil {
		finish(w, r, nil, err, 0)
		return
	}
	v, err := h.service.ModerateReview(r.Context(), actor(r), r.PathValue("review_id"), in, r.Header.Get("X-Request-ID"))
	finish(w, r, v, err, 200)
}
func (h *Handler) report(w http.ResponseWriter, r *http.Request) {
	var in social_postgres_repository.ReportInput
	if _, err := v2request.Decode(w, r, &in, "target_type", "target_id", "reason"); err != nil {
		finish(w, r, nil, err, 0)
		return
	}
	v, err := h.service.Report(r.Context(), actor(r), in)
	finish(w, r, v, err, 201)
}
func (h *Handler) reports(w http.ResponseWriter, r *http.Request) {
	f, err := page(r)
	if err != nil {
		finish(w, r, nil, err, 0)
		return
	}
	v, err := h.service.Reports(r.Context(), actor(r), f)
	finish(w, r, v, err, 200)
}
func (h *Handler) decideReport(w http.ResponseWriter, r *http.Request) {
	var in social_postgres_repository.Decision
	if _, err := v2request.Decode(w, r, &in, "status", "reason"); err != nil {
		finish(w, r, nil, err, 0)
		return
	}
	v, err := h.service.DecideReport(r.Context(), actor(r), r.PathValue("report_id"), in, r.Header.Get("X-Request-ID"))
	finish(w, r, v, err, 200)
}
func (h *Handler) follow(w http.ResponseWriter, r *http.Request) {
	id, err := h.service.Follow(r.Context(), actor(r), r.PathValue("trainer_id"), r.Method == "PUT")
	status := 200
	if r.Method == "DELETE" {
		status = 204
	}
	finish(w, r, map[string]any{"trainer_id": id, "following": r.Method == "PUT"}, err, status)
}
func (h *Handler) trainers(w http.ResponseWriter, r *http.Request) {
	f, err := page(r)
	if err != nil {
		finish(w, r, nil, err, 0)
		return
	}
	v, err := h.service.Trainers(r.Context(), actor(r), f)
	finish(w, r, v, err, 200)
}
func (h *Handler) notifications(w http.ResponseWriter, r *http.Request) {
	f, err := page(r)
	if err != nil {
		finish(w, r, nil, err, 0)
		return
	}
	v, err := h.service.Notifications(r.Context(), actor(r), f)
	finish(w, r, v, err, 200)
}
func (h *Handler) readNotification(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Read bool `json:"read"`
	}
	if _, err := v2request.Decode(w, r, &in, "read"); err != nil {
		finish(w, r, nil, err, 0)
		return
	}
	v, err := h.service.ReadNotification(r.Context(), actor(r), r.PathValue("notification_id"), in.Read)
	finish(w, r, v, err, 200)
}
func (h *Handler) events(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Events []social_postgres_repository.Event `json:"events"`
	}
	if _, err := v2request.Decode(w, r, &in, "events"); err != nil {
		finish(w, r, nil, err, 0)
		return
	}
	v, err := h.service.Events(r.Context(), actor(r), in.Events)
	finish(w, r, v, err, 202)
}
