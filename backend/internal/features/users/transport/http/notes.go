package users_transport_http

import (
	"fmt"
	"net/http"
	"strconv"

	core_errors "sport_app/internal/core/errors"
	core_logger "sport_app/internal/core/logger"
	core_http_middleware "sport_app/internal/core/transport/http/middleware"
	core_http_request "sport_app/internal/core/transport/http/request"
	core_http_responce "sport_app/internal/core/transport/http/responce"
)

func (h *UsersHTTPHandler) ListNotes(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_responce.NewHTTPResponce(log, rw)

	userID, err := core_http_middleware.UserIDFromContext(ctx)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get user id from context")
		return
	}

	limit := 50
	offset := 0
	q := r.URL.Query()
	if lStr := q.Get("limit"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil && l > 0 {
			if l > 100 {
				l = 100
			}
			limit = l
		}
	}
	if oStr := q.Get("offset"); oStr != "" {
		if o, err := strconv.Atoi(oStr); err == nil && o >= 0 {
			offset = o
		}
	}

	notes, total, err := h.usersService.GetListNotes(ctx, userID, limit, offset)
	if err != nil {
		responseHandler.ErrorResponse(err, "get list notes")
		return
	}

	res := map[string]any{
		"items":  notes,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	}

	rw.Header().Set("Content-Type", "application/json")
	responseHandler.JSONResponse(res, http.StatusOK)
}

type notePayload struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

func (h *UsersHTTPHandler) CreateNote(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_responce.NewHTTPResponce(log, rw)

	userID, err := core_http_middleware.UserIDFromContext(ctx)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get user id from context")
		return
	}

	var p notePayload
	if err := core_http_request.DecodeAndValidateRequest(r, &p); err != nil {
		responseHandler.ErrorResponse(err, "failed to decode change user request")
		return
	}

	note, err := h.usersService.CreateNotesUser(ctx, userID, p.Title, p.Body)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to create note")
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	responseHandler.JSONResponse(note, http.StatusOK)
}

func (h *UsersHTTPHandler) UpdateNote(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_responce.NewHTTPResponce(log, rw)

	userID, err := core_http_middleware.UserIDFromContext(ctx)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get user id from context")
		return
	}

	noteID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		responseHandler.ErrorResponse(
			fmt.Errorf("invalid note id '%s': %w", r.PathValue("id"), core_errors.ErrInvalidArgument),
			"invalid note id",
		)
		return
	}

	var p notePayload
	if err := core_http_request.DecodeAndValidateRequest(r, &p); err != nil {
		responseHandler.ErrorResponse(err, "failed to decode change user request")
		return
	}

	note, err := h.usersService.UpdateNotesUser(ctx, userID, noteID, p.Title, p.Body)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to update note")
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	responseHandler.JSONResponse(note, http.StatusOK)
}

func (h *UsersHTTPHandler) DeleteNote(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_responce.NewHTTPResponce(log, rw)

	userID, err := core_http_middleware.UserIDFromContext(ctx)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get user id from context")
		return
	}

	noteID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		responseHandler.ErrorResponse(
			fmt.Errorf("invalid note id '%s': %w", r.PathValue("id"), core_errors.ErrInvalidArgument),
			"invalid note id",
		)
		return
	}

	if err := h.usersService.DeleteNotesUser(ctx, userID, noteID); err != nil {
		responseHandler.ErrorResponse(err, "failed to delete user note")
		return
	}

	responseHandler.NoContentResponse()
}
