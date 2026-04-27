package notes_transport_http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"go.uber.org/zap"

	core_auth "sport_app/internal/core/auth"
	core_logger "sport_app/internal/core/logger"
	core_http_middleware "sport_app/internal/core/transport/http/middleware"
	notes_postgres_repository "sport_app/internal/features/notes/repository/postgres"
)

type NotesService interface {
	ListNotes(ctx context.Context, userID int) ([]notes_postgres_repository.Note, error)
	CreateNote(ctx context.Context, userID int, title, body string) (notes_postgres_repository.Note, error)
	UpdateNote(ctx context.Context, userID int, noteID int, title, body string) (notes_postgres_repository.Note, error)
	DeleteNote(ctx context.Context, userID int, noteID int) error
}

type Handler struct {
	service NotesService
	jwt     *core_auth.JWT
}

func NewHandler(service NotesService, jwt *core_auth.JWT) *Handler {
	return &Handler{service: service, jwt: jwt}
}

func (h *Handler) RegisterRoutes(register func(method, path string, handler http.Handler)) {
	protect := core_http_middleware.Protect(h.jwt)

	register(http.MethodGet, "/notes", protect(http.HandlerFunc(h.listNotes)))
	register(http.MethodPost, "/notes", protect(http.HandlerFunc(h.createNote)))
	register(http.MethodPatch, "/notes/{id}", protect(http.HandlerFunc(h.updateNote)))
	register(http.MethodDelete, "/notes/{id}", protect(http.HandlerFunc(h.deleteNote)))
}

func userIDFromRequest(r *http.Request) (int, error) {
	raw, ok := r.Context().Value(core_http_middleware.UserIDKey).(string)
	if !ok || raw == "" {
		return 0, errors.New("user id missing")
	}
	id, err := strconv.Atoi(raw)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (h *Handler) listNotes(w http.ResponseWriter, r *http.Request) {
	userID, err := userIDFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	notes, err := h.service.ListNotes(r.Context(), userID)
	if err != nil {
		core_logger.FromContext(r.Context()).Error("Failed to list notes", zap.Error(err))
		http.Error(w, "Failed to list notes", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"items": notes})
}

type notePayload struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

func (h *Handler) createNote(w http.ResponseWriter, r *http.Request) {
	userID, err := userIDFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var p notePayload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	note, err := h.service.CreateNote(r.Context(), userID, p.Title, p.Body)
	if err != nil {
		http.Error(w, "Failed to create note", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(note)
}

func (h *Handler) updateNote(w http.ResponseWriter, r *http.Request) {
	userID, err := userIDFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	noteID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid note id", http.StatusBadRequest)
		return
	}

	var p notePayload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	note, err := h.service.UpdateNote(r.Context(), userID, noteID, p.Title, p.Body)
	if err != nil {
		http.Error(w, "Failed to update note", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(note)
}

func (h *Handler) deleteNote(w http.ResponseWriter, r *http.Request) {
	userID, err := userIDFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	noteID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid note id", http.StatusBadRequest)
		return
	}

	if err := h.service.DeleteNote(r.Context(), userID, noteID); err != nil {
		http.Error(w, "Failed to delete note", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
