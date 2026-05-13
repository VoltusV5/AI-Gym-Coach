package users_transport_http

import (
	"net/http"
	core_logger "sport_app/internal/core/logger"
	core_http_middleware "sport_app/internal/core/transport/http/middleware"
	core_http_request "sport_app/internal/core/transport/http/request"
	core_http_responce "sport_app/internal/core/transport/http/responce"
	"strconv"
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

	notes, err := h.usersService.GetListNotes(ctx, userID)
	if err != nil {
		responseHandler.ErrorResponse(err, "get list notes")
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	responseHandler.JSONResponse(map[string]any{"items": notes}, http.StatusOK)
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
		responseHandler.ErrorResponse(err, "failed to get note id")
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
		responseHandler.ErrorResponse(err, "failed to get note id")
		return
	}

	if err := h.usersService.DeleteNotesUser(ctx, userID, noteID); err != nil {
		responseHandler.ErrorResponse(err, "failed to delete user note")
		return
	}

	responseHandler.NoContentResponse()
}