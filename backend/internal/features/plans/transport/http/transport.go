package plans_transport_http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	core_auth "sport_app/internal/core/auth"
	core_errors "sport_app/internal/core/errors"
	core_logger "sport_app/internal/core/logger"
	core_http_middleware "sport_app/internal/core/transport/http/middleware"
	core_http_responce "sport_app/internal/core/transport/http/responce"
	core_http_server "sport_app/internal/core/transport/http/server"
	plans_postgres_repository "sport_app/internal/features/plans/repository/postgres"
)

type Service interface {
	ListPlans(context.Context, string, int, string) (plans_postgres_repository.Page, error)
	GetPlan(context.Context, string, string) (plans_postgres_repository.Plan, error)
	CreatePlan(context.Context, string, plans_postgres_repository.Input) (plans_postgres_repository.Plan, error)
	UpdatePlan(context.Context, string, string, plans_postgres_repository.Input) (plans_postgres_repository.Plan, error)
	ActivatePlan(context.Context, string, string, int64) (plans_postgres_repository.Plan, error)
	ArchivePlan(context.Context, string, string) error
}

type PlansHTTPHandler struct {
	service Service
	jwt     *core_auth.JWT
}

func NewPlansHTTPHandler(service Service, jwt *core_auth.JWT) *PlansHTTPHandler {
	return &PlansHTTPHandler{service: service, jwt: jwt}
}

func (h *PlansHTTPHandler) Routes() []core_http_server.Route {
	protect := core_http_middleware.ProtectV2(h.jwt)
	return []core_http_server.Route{
		core_http_server.NewRoute(http.MethodGet, "/me/plans", h.ListPlans, protect),
		core_http_server.NewRoute(http.MethodPost, "/me/plans", h.CreatePlan, protect),
		core_http_server.NewRoute(http.MethodGet, "/me/plans/{plan_id}", h.GetPlan, protect),
		core_http_server.NewRoute(http.MethodPatch, "/me/plans/{plan_id}", h.UpdatePlan, protect),
		core_http_server.NewRoute(http.MethodDelete, "/me/plans/{plan_id}", h.ArchivePlan, protect),
		core_http_server.NewRoute(http.MethodPost, "/me/plans/{plan_id}/activate", h.ActivatePlan, protect),
	}
}

func response(w http.ResponseWriter, r *http.Request) *core_http_responce.HTTPResponceHandler {
	return core_http_responce.NewHTTPResponceV2(core_logger.FromContext(r.Context()), w, r.Header.Get("X-Request-ID"))
}

func requiredObject(raw []byte, required ...string) (map[string]json.RawMessage, error) {
	invalid := fmt.Errorf("invalid plan JSON: %w", core_errors.ErrInvalidArgument)
	var fields map[string]json.RawMessage
	if json.Unmarshal(raw, &fields) != nil || fields == nil {
		return nil, invalid
	}
	for _, key := range required {
		if _, ok := fields[key]; !ok {
			return nil, invalid
		}
	}
	for key, value := range fields {
		if key != "weight_kg" && bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return nil, invalid
		}
	}
	return fields, nil
}

func decodeBody(w http.ResponseWriter, r *http.Request, dest any, required ...string) (map[string]json.RawMessage, error) {
	invalid := fmt.Errorf("invalid plan JSON: %w", core_errors.ErrInvalidArgument)
	data, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 1<<20))
	if err != nil {
		return nil, invalid
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dest); err != nil {
		return nil, invalid
	}
	var extra any
	if decoder.Decode(&extra) != io.EOF {
		return nil, invalid
	}
	return requiredObject(data, required...)
}

func decodeInput(w http.ResponseWriter, r *http.Request, update bool) (plans_postgres_repository.Input, error) {
	var input plans_postgres_repository.Input
	required := []string{"title", "days", "schedule"}
	if update {
		required = append(required, "version")
	}
	fields, err := decodeBody(w, r, &input, required...)
	if err != nil {
		return input, err
	}
	var days []json.RawMessage
	if err := json.Unmarshal(fields["days"], &days); err != nil {
		return input, fmt.Errorf("invalid plan days: %w", core_errors.ErrInvalidArgument)
	}
	for _, day := range days {
		object, err := requiredObject(day, "title", "position", "exercises")
		if err != nil {
			return input, err
		}
		var exercises []json.RawMessage
		if err := json.Unmarshal(object["exercises"], &exercises); err != nil {
			return input, fmt.Errorf("invalid prescriptions: %w", core_errors.ErrInvalidArgument)
		}
		for _, ex := range exercises {
			if _, err := requiredObject(ex, "exercise_id", "position", "sets", "reps_min", "reps_max", "rest_sec"); err != nil {
				return input, err
			}
		}
	}
	schedule, err := requiredObject(fields["schedule"], "time_zone", "start_date", "entries")
	if err != nil {
		return input, err
	}
	var entries []json.RawMessage
	if err := json.Unmarshal(schedule["entries"], &entries); err != nil {
		return input, fmt.Errorf("invalid schedule entries: %w", core_errors.ErrInvalidArgument)
	}
	for _, entry := range entries {
		if _, err := requiredObject(entry, "weekday", "plan_day_id"); err != nil {
			return input, err
		}
	}
	return input, nil
}

func (h *PlansHTTPHandler) ListPlans(w http.ResponseWriter, r *http.Request) {
	resp := response(w, r)
	id, err := core_http_middleware.UserIDFromContext(r.Context())
	if err != nil {
		resp.ErrorResponse(err, "plan authentication")
		return
	}
	limit := 20
	if r.URL.Query().Has("limit") {
		limit, err = strconv.Atoi(r.URL.Query().Get("limit"))
		if err != nil || limit < 1 {
			resp.ErrorResponse(fmt.Errorf("invalid plan limit: %w", core_errors.ErrInvalidArgument), "plan pagination")
			return
		}
	}
	page, err := h.service.ListPlans(r.Context(), id, limit, r.URL.Query().Get("cursor"))
	if err != nil {
		resp.ErrorResponse(err, "list plans")
		return
	}
	resp.JSONResponse(page, http.StatusOK)
}

func (h *PlansHTTPHandler) GetPlan(w http.ResponseWriter, r *http.Request) {
	resp := response(w, r)
	id, err := core_http_middleware.UserIDFromContext(r.Context())
	if err != nil {
		resp.ErrorResponse(err, "plan authentication")
		return
	}
	plan, err := h.service.GetPlan(r.Context(), id, r.PathValue("plan_id"))
	if err != nil {
		resp.ErrorResponse(err, "get plan")
		return
	}
	resp.JSONResponse(plan, http.StatusOK)
}

func (h *PlansHTTPHandler) CreatePlan(w http.ResponseWriter, r *http.Request) { h.save(w, r, false) }
func (h *PlansHTTPHandler) UpdatePlan(w http.ResponseWriter, r *http.Request) { h.save(w, r, true) }

func (h *PlansHTTPHandler) save(w http.ResponseWriter, r *http.Request, update bool) {
	resp := response(w, r)
	id, err := core_http_middleware.UserIDFromContext(r.Context())
	if err != nil {
		resp.ErrorResponse(err, "plan authentication")
		return
	}
	input, err := decodeInput(w, r, update)
	if err != nil {
		resp.ErrorResponse(err, "decode plan")
		return
	}
	var plan plans_postgres_repository.Plan
	status := http.StatusCreated
	if update {
		plan, err = h.service.UpdatePlan(r.Context(), id, r.PathValue("plan_id"), input)
		status = http.StatusOK
	} else {
		plan, err = h.service.CreatePlan(r.Context(), id, input)
	}
	if err != nil {
		resp.ErrorResponse(err, "save plan")
		return
	}
	resp.JSONResponse(plan, status)
}

func (h *PlansHTTPHandler) ActivatePlan(w http.ResponseWriter, r *http.Request) {
	resp := response(w, r)
	id, err := core_http_middleware.UserIDFromContext(r.Context())
	if err != nil {
		resp.ErrorResponse(err, "plan authentication")
		return
	}
	var body struct {
		Version int64 `json:"version"`
	}
	if _, err := decodeBody(w, r, &body, "version"); err != nil {
		resp.ErrorResponse(err, "decode activation")
		return
	}
	plan, err := h.service.ActivatePlan(r.Context(), id, r.PathValue("plan_id"), body.Version)
	if err != nil {
		resp.ErrorResponse(err, "activate plan")
		return
	}
	resp.JSONResponse(plan, http.StatusCreated)
}

func (h *PlansHTTPHandler) ArchivePlan(w http.ResponseWriter, r *http.Request) {
	resp := response(w, r)
	id, err := core_http_middleware.UserIDFromContext(r.Context())
	if err != nil {
		resp.ErrorResponse(err, "plan authentication")
		return
	}
	if err := h.service.ArchivePlan(r.Context(), id, r.PathValue("plan_id")); err != nil {
		resp.ErrorResponse(err, "archive plan")
		return
	}
	resp.NoContentResponse()
}
