package handlers

import (
	"errors"
	"net/http"
	simpleconnection "sport_app/internal/core/models/simple_connection"
	middleware "sport_app/internal/core/transport/http/middleware"
	"sport_app/internal/features/nutrition"

	"github.com/gorilla/mux"
)

var dbpool *simpleconnection.ConnectionPool

func StartHTTPServer(pool *simpleconnection.ConnectionPool) error {
	dbpool = pool
	router := mux.NewRouter()

	router.HandleFunc("/auth/guest", GuestHandler).Methods("POST")
	// Те же обработчики под префиксом /api/v1 — фронт ходит сюда из Docker/nginx (без лишних 404 в консоли).
	router.HandleFunc("/api/v1/auth/guest", GuestHandler).Methods("POST")

	router.Handle("/profile", middleware.Protect()(http.HandlerFunc(ProfileHandler))).Methods("POST", "PATCH")
	router.Handle("/profile", middleware.Protect()(http.HandlerFunc(ProfileGetHandler))).Methods("GET")
	router.Handle("/api/v1/profile", middleware.Protect()(http.HandlerFunc(ProfileHandler))).Methods("POST", "PATCH")
	router.Handle("/api/v1/profile", middleware.Protect()(http.HandlerFunc(ProfileGetHandler))).Methods("GET")
	router.Handle("/api/v1/plans/generate", middleware.Protect()(http.HandlerFunc(ResponceGenerateHandler))).Methods("POST")
	router.Handle("/api/v1/workouts/complete", middleware.Protect()(http.HandlerFunc(WorkoutCompleteHandler))).Methods("POST")

	nutrition.NewService(pool.Pool).RegisterRoutes(router)

	handler := corsMiddleware(router)

	if err := http.ListenAndServe(":5050", handler); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		} else {
			return err
		}
	}
	return nil
}
