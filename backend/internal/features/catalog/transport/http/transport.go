package catalog_transport_http

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"strconv"

	core_auth "sport_app/internal/core/auth"
	core_errors "sport_app/internal/core/errors"
	core_logger "sport_app/internal/core/logger"
	core_http_middleware "sport_app/internal/core/transport/http/middleware"
	core_http_responce "sport_app/internal/core/transport/http/responce"
	core_http_server "sport_app/internal/core/transport/http/server"
	catalog_postgres_repository "sport_app/internal/features/catalog/repository/postgres"
)

type Service interface {
	ListEquipment(context.Context, catalog_postgres_repository.Filter, string) (catalog_postgres_repository.Page[catalog_postgres_repository.Equipment], error)
	ListExercises(context.Context, catalog_postgres_repository.Filter, string) (catalog_postgres_repository.Page[catalog_postgres_repository.Exercise], error)
	ListPrograms(context.Context, catalog_postgres_repository.Filter, string) (catalog_postgres_repository.Page[catalog_postgres_repository.ProgramCard], error)
	GetProgram(context.Context, string, int) (catalog_postgres_repository.ProgramDetail, error)
	GrantFreeAccess(context.Context, string, int) (catalog_postgres_repository.Entitlement, error)
}

type CatalogHTTPHandler struct {
	service Service
	jwt     *core_auth.JWT
}

func NewCatalogHTTPHandler(service Service, jwt *core_auth.JWT) *CatalogHTTPHandler {
	return &CatalogHTTPHandler{service: service, jwt: jwt}
}

func (h *CatalogHTTPHandler) Routes() []core_http_server.Route {
	protect := core_http_middleware.ProtectV2(h.jwt)
	optional := core_http_middleware.OptionalProtectV2(h.jwt)
	return []core_http_server.Route{
		core_http_server.NewRoute(http.MethodGet, "/equipment", h.ListEquipment, protect),
		core_http_server.NewRoute(http.MethodGet, "/exercises", h.ListExercises, protect),
		core_http_server.NewRoute(http.MethodGet, "/catalog/programs", h.ListPrograms, optional),
		core_http_server.NewRoute(http.MethodGet, "/catalog/programs/{program_id}", h.GetProgram, optional),
		core_http_server.NewRoute(http.MethodPost, "/programs/{program_id}/access", h.GrantFreeAccess, protect),
	}
}

func response(w http.ResponseWriter, r *http.Request) *core_http_responce.HTTPResponceHandler {
	return core_http_responce.NewHTTPResponceV2(core_logger.FromContext(r.Context()), w, r.Header.Get("X-Request-ID"))
}

func parseFilter(r *http.Request) (catalog_postgres_repository.Filter, error) {
	q := r.URL.Query()
	filter := catalog_postgres_repository.Filter{
		Query: q.Get("query"), MuscleGroup: q.Get("muscle_group"),
		Goal: q.Get("goal"), Level: q.Get("level"),
		PriceType: q.Get("price_type"), Sort: q.Get("sort"),
		Equipment: q["equipment"], EquipmentSet: q.Has("equipment"),
	}
	for key, dest := range map[string]*int{
		"limit": &filter.Limit, "duration_weeks": &filter.DurationWeeks, "trainer_id": &filter.TrainerID,
	} {
		if q.Has(key) {
			n, err := strconv.Atoi(q.Get(key))
			if err != nil || n <= 0 {
				return filter, fmt.Errorf("invalid %s: %w", key, core_errors.ErrInvalidArgument)
			}
			*dest = n
		}
	}
	return filter, nil
}

func (h *CatalogHTTPHandler) ListEquipment(w http.ResponseWriter, r *http.Request) {
	resp := response(w, r)
	filter, err := parseFilter(r)
	if err != nil {
		resp.ErrorResponse(err, "invalid equipment filter")
		return
	}
	page, err := h.service.ListEquipment(r.Context(), filter, r.URL.Query().Get("cursor"))
	if err != nil {
		resp.ErrorResponse(err, "list equipment")
		return
	}
	resp.JSONResponse(page, http.StatusOK)
}

func (h *CatalogHTTPHandler) ListExercises(w http.ResponseWriter, r *http.Request) {
	resp := response(w, r)
	filter, err := parseFilter(r)
	if err != nil {
		resp.ErrorResponse(err, "invalid exercise filter")
		return
	}
	page, err := h.service.ListExercises(r.Context(), filter, r.URL.Query().Get("cursor"))
	if err != nil {
		resp.ErrorResponse(err, "list exercises")
		return
	}
	resp.JSONResponse(page, http.StatusOK)
}

func (h *CatalogHTTPHandler) ListPrograms(w http.ResponseWriter, r *http.Request) {
	resp := response(w, r)
	filter, err := parseFilter(r)
	if err != nil {
		resp.ErrorResponse(err, "invalid catalog filter")
		return
	}
	page, err := h.service.ListPrograms(r.Context(), filter, r.URL.Query().Get("cursor"))
	if err != nil {
		resp.ErrorResponse(err, "list programs")
		return
	}
	resp.JSONResponse(page, http.StatusOK)
}

func requestUserID(r *http.Request) (int, error) {
	raw, err := core_http_middleware.UserIDFromContext(r.Context())
	if err != nil {
		return 0, err
	}
	id, err := strconv.Atoi(raw)
	if err != nil || id <= 0 || id > math.MaxInt32 {
		return 0, fmt.Errorf("invalid user id: %w", core_errors.ErrUnauthorized)
	}
	return id, nil
}

func (h *CatalogHTTPHandler) GetProgram(w http.ResponseWriter, r *http.Request) {
	resp := response(w, r)
	userID := 0
	if r.Header.Get("Authorization") != "" {
		var err error
		userID, err = requestUserID(r)
		if err != nil {
			resp.ErrorResponse(err, "invalid identity")
			return
		}
	}
	item, err := h.service.GetProgram(r.Context(), r.PathValue("program_id"), userID)
	if err != nil {
		resp.ErrorResponse(err, "get program")
		return
	}
	resp.JSONResponse(item, http.StatusOK)
}

func (h *CatalogHTTPHandler) GrantFreeAccess(w http.ResponseWriter, r *http.Request) {
	resp := response(w, r)
	userID, err := requestUserID(r)
	if err != nil {
		resp.ErrorResponse(err, "invalid identity")
		return
	}
	item, err := h.service.GrantFreeAccess(r.Context(), r.PathValue("program_id"), userID)
	if err != nil {
		resp.ErrorResponse(err, "grant free access")
		return
	}
	resp.JSONResponse(item, http.StatusCreated)
}
