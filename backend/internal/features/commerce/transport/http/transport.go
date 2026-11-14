package commerce_transport_http

import (
	"bytes"
	"io"
	"net/http"
	core_auth "sport_app/internal/core/auth"
	core_errors "sport_app/internal/core/errors"
	core_logger "sport_app/internal/core/logger"
	core_http_middleware "sport_app/internal/core/transport/http/middleware"
	core_http_responce "sport_app/internal/core/transport/http/responce"
	core_http_server "sport_app/internal/core/transport/http/server"
	"sport_app/internal/core/transport/http/v2request"
	commerce_postgres_repository "sport_app/internal/features/commerce/repository/postgres"
	commerce_service "sport_app/internal/features/commerce/service"
	"strings"
)

type Handler struct {
	service *commerce_service.Service
	jwt     *core_auth.JWT
}

func NewHandler(s *commerce_service.Service, jwt *core_auth.JWT) *Handler { return &Handler{s, jwt} }
func (h *Handler) Routes() []core_http_server.Route {
	p := core_http_middleware.ProtectV2(h.jwt)
	return []core_http_server.Route{
		core_http_server.NewRoute("POST", "/orders", h.create, p), core_http_server.NewRoute("GET", "/orders/{order_id}", h.get, p),
		core_http_server.NewRoute("GET", "/me/orders", h.list, p), core_http_server.NewRoute("GET", "/admin/orders", h.list, p),
		core_http_server.NewRoute("GET", "/admin/settings/commission", h.commission, p), core_http_server.NewRoute("PATCH", "/admin/settings/commission", h.commission, p),
		core_http_server.NewRoute("GET", "/trainer/payouts", h.payouts, p), core_http_server.NewRoute("GET", "/admin/payouts", h.payouts, p),
		core_http_server.NewRoute("POST", "/admin/payouts", h.createPayout, p), core_http_server.NewRoute("PATCH", "/admin/payouts/{payout_id}", h.decidePayout, p),
		core_http_server.NewRoute("POST", "/payments/webhook", h.webhook), core_http_server.NewRoute("POST", "/payments/sandbox/{order_id}", h.sandbox, p),
	}
}
func actor(r *http.Request) string {
	id, _ := core_http_middleware.UserIDFromContext(r.Context())
	return id
}
func finish(w http.ResponseWriter, r *http.Request, v any, err error, status int) {
	p := core_http_responce.NewHTTPResponceV2(core_logger.FromContext(r.Context()), w, r.Header.Get("X-Request-ID"))
	if err != nil {
		p.ErrorResponse(err, "commerce operation")
		return
	}
	p.JSONResponse(v, status)
}
func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var in commerce_postgres_repository.OrderInput
	if _, err := v2request.Decode(w, r, &in, "program_id", "program_version_id"); err != nil {
		finish(w, r, nil, err, 0)
		return
	}
	v, err := h.service.Create(r.Context(), actor(r), r.Header.Get("Idempotency-Key"), in)
	finish(w, r, v, err, 201)
}
func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	v, err := h.service.Get(r.Context(), actor(r), r.PathValue("order_id"))
	finish(w, r, v, err, 200)
}
func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	f, err := v2request.Pagination(r, r.URL.Path+":"+actor(r))
	if err != nil {
		finish(w, r, nil, err, 0)
		return
	}
	v, err := h.service.List(r.Context(), actor(r), strings.HasPrefix(r.URL.Path, "/admin/"), f)
	finish(w, r, v, err, 200)
}
func (h *Handler) commission(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		v, err := h.service.Commission(r.Context(), actor(r))
		finish(w, r, v, err, 200)
		return
	}
	var in commerce_postgres_repository.Commission
	if _, err := v2request.Decode(w, r, &in, "commission_bps", "effective_at", "version"); err != nil {
		finish(w, r, nil, err, 0)
		return
	}
	v, err := h.service.SetCommission(r.Context(), actor(r), in, r.Header.Get("X-Request-ID"))
	finish(w, r, v, err, 200)
}
func (h *Handler) payouts(w http.ResponseWriter, r *http.Request) {
	f, err := v2request.Pagination(r, r.URL.Path+":"+actor(r))
	if err != nil {
		finish(w, r, nil, err, 0)
		return
	}
	v, err := h.service.Payouts(r.Context(), actor(r), strings.HasPrefix(r.URL.Path, "/admin/"), f)
	finish(w, r, v, err, 200)
}
func (h *Handler) createPayout(w http.ResponseWriter, r *http.Request) {
	var in commerce_postgres_repository.PayoutInput
	if _, err := v2request.Decode(w, r, &in, "trainer_id", "amount_minor", "currency"); err != nil {
		finish(w, r, nil, err, 0)
		return
	}
	v, err := h.service.CreatePayout(r.Context(), actor(r), r.Header.Get("Idempotency-Key"), in, r.Header.Get("X-Request-ID"))
	finish(w, r, v, err, 201)
}
func (h *Handler) decidePayout(w http.ResponseWriter, r *http.Request) {
	var in commerce_postgres_repository.Decision
	if _, err := v2request.Decode(w, r, &in, "action", "reason"); err != nil {
		finish(w, r, nil, err, 0)
		return
	}
	v, err := h.service.DecidePayout(r.Context(), actor(r), r.PathValue("payout_id"), in, r.Header.Get("X-Request-ID"))
	finish(w, r, v, err, 200)
}
func (h *Handler) webhook(w http.ResponseWriter, r *http.Request) {
	raw, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 1<<20))
	if err != nil {
		finish(w, r, nil, v2request.Invalid(), 0)
		return
	}
	if !h.service.Verify(raw, r.Header.Get("X-Payment-Signature")) {
		finish(w, r, nil, core_errors.ErrUnauthorized, 0)
		return
	}
	r.Body = io.NopCloser(bytes.NewReader(raw))
	var in commerce_postgres_repository.Event
	if _, err := v2request.Decode(w, r, &in, "provider_event_id", "provider_payment_id", "order_id", "event_type", "amount_minor", "currency"); err != nil {
		finish(w, r, nil, err, 0)
		return
	}
	err = h.service.Confirm(r.Context(), in, r.Header.Get("X-Request-ID"))
	finish(w, r, map[string]bool{"ok": true}, err, 200)
}
func (h *Handler) sandbox(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Action string `json:"action"`
	}
	if _, err := v2request.Decode(w, r, &in, "action"); err != nil {
		finish(w, r, nil, err, 0)
		return
	}
	v, err := h.service.Sandbox(r.Context(), actor(r), r.PathValue("order_id"), in.Action, r.Header.Get("X-Request-ID"))
	finish(w, r, v, err, 200)
}
