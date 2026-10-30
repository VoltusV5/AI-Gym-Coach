package profile_transport_http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	core_auth "sport_app/internal/core/auth"
	core_errors "sport_app/internal/core/errors"
	core_logger "sport_app/internal/core/logger"
	core_http_middleware "sport_app/internal/core/transport/http/middleware"
	core_http_responce "sport_app/internal/core/transport/http/responce"
	core_http_server "sport_app/internal/core/transport/http/server"
	profile_postgres_repository "sport_app/internal/features/profile/repository/postgres"
)

// Service задаёт операции профиля, доступные транспорту v2.
type Service interface {
	GetProfile(context.Context, string) (profile_postgres_repository.Profile, error)
	UpdateProfile(context.Context, string, profile_postgres_repository.Patch) (profile_postgres_repository.Profile, error)
}

// ProfileHTTPHandler обрабатывает только анкету текущего пользователя.
type ProfileHTTPHandler struct {
	service Service
	jwt     *core_auth.JWT
}

// NewProfileHTTPHandler создаёт HTTP-обработчик с сервисом и проверкой JWT.
func NewProfileHTTPHandler(service Service, jwt *core_auth.JWT) *ProfileHTTPHandler {
	return &ProfileHTTPHandler{service: service, jwt: jwt}
}

// Routes возвращает GET/PATCH /me/profile, защищённые JWT, включая гостевые сессии.
func (h *ProfileHTTPHandler) Routes() []core_http_server.Route {
	protect := core_http_middleware.ProtectV2(h.jwt)
	return []core_http_server.Route{
		core_http_server.NewRoute(http.MethodGet, "/me/profile", h.GetProfile, protect),
		core_http_server.NewRoute(http.MethodPatch, "/me/profile", h.UpdateProfile, protect),
	}
}

// response создаёт обработчик безопасных ошибок v2 с request_id.
func response(w http.ResponseWriter, r *http.Request) *core_http_responce.HTTPResponceHandler {
	return core_http_responce.NewHTTPResponceV2(core_logger.FromContext(r.Context()), w, r.Header.Get("X-Request-ID"))
}

// GetProfile извлекает userID из JWT-контекста и возвращает анкету либо ошибку v2.
func (h *ProfileHTTPHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	resp := response(w, r)
	userID, err := core_http_middleware.UserIDFromContext(r.Context())
	if err != nil {
		resp.ErrorResponse(err, "profile authentication")
		return
	}
	profile, err := h.service.GetProfile(r.Context(), userID)
	if err != nil {
		resp.ErrorResponse(err, "get profile")
		return
	}
	resp.JSONResponse(profile, http.StatusOK)
}

// UpdateProfile декодирует единственный JSON-объект и возвращает сохранённую анкету с новой version.
// Не допускает чужой user_id, неизвестные поля, null и несколько JSON-значений в одном теле.
func (h *ProfileHTTPHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	resp := response(w, r)
	userID, err := core_http_middleware.UserIDFromContext(r.Context())
	if err != nil {
		resp.ErrorResponse(err, "profile authentication")
		return
	}
	patch, err := decodePatch(w, r)
	if err != nil {
		resp.ErrorResponse(err, "decode profile patch")
		return
	}
	profile, err := h.service.UpdateProfile(r.Context(), userID, patch)
	if err != nil {
		resp.ErrorResponse(err, "update profile")
		return
	}
	resp.JSONResponse(profile, http.StatusOK)
}

// decodePatch ограничивает тело 1 MiB и проверяет JSON-форму до вызова сервиса.
func decodePatch(w http.ResponseWriter, r *http.Request) (profile_postgres_repository.Patch, error) {
	var patch profile_postgres_repository.Patch
	invalid := fmt.Errorf("invalid profile JSON: %w", core_errors.ErrInvalidArgument)
	data, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 1<<20))
	if err != nil {
		return patch, invalid
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&patch); err != nil {
		return patch, invalid
	}
	var extra any
	if decoder.Decode(&extra) != io.EOF {
		return patch, invalid
	}
	var fields map[string]json.RawMessage
	if json.Unmarshal(data, &fields) != nil || fields == nil {
		return patch, invalid
	}
	for _, raw := range fields {
		if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
			return patch, invalid
		}
	}
	return patch, nil
}
