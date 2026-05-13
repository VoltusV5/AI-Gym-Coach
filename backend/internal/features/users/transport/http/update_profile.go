package users_transport_http

import (
	"encoding/json"
	"fmt"
	"net/http"

	core_errors "sport_app/internal/core/errors"
	core_logger "sport_app/internal/core/logger"
	core_http_middleware "sport_app/internal/core/transport/http/middleware"
	core_http_responce "sport_app/internal/core/transport/http/responce"
)

func (h *UsersHTTPHandler) UpdateProfile(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_responce.NewHTTPResponce(log, rw)

	userID, err := core_http_middleware.UserIDFromContext(ctx)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get user id from context")
		return
	}

	var request map[string]any
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		responseHandler.ErrorResponse(
			fmt.Errorf("decode profile update body: %v: %w", err, core_errors.ErrInvalidArgument),
			"failed to decode profile update body",
		)
		return
	}
	defer r.Body.Close()

	versionRaw, ok := request["version"]
	if !ok {
		responseHandler.ErrorResponse(
			fmt.Errorf("missing 'version' in profile update body: %w", core_errors.ErrInvalidArgument),
			"missing 'version' in profile update body",
		)
		return
	}
	versionFloat, ok := versionRaw.(float64)
	if !ok {
		responseHandler.ErrorResponse(
			fmt.Errorf("'version' must be a number: %w", core_errors.ErrInvalidArgument),
			"invalid 'version' in profile update body",
		)
		return
	}
	delete(request, "version")
	expectedVersion := int64(versionFloat)

	if len(request) == 0 {
		responseHandler.ErrorResponse(
			fmt.Errorf("empty profile update body: %w", core_errors.ErrInvalidArgument),
			"empty profile update body",
		)
		return
	}

	profile, err := h.usersService.UpdateProfile(ctx, userID, expectedVersion, request)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to update profile")
		return
	}

	responseHandler.JSONResponse(profile, http.StatusOK)
}