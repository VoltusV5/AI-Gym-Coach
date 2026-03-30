package handlers

import (
	"errors"
	"net/http"
	"sport_app/auth"
	simpleconnection "sport_app/core/models/simple_connection"

	"github.com/gorilla/mux"
)

var dbpool *simpleconnection.ConnectionPool

func StartHTTPServer(pool *simpleconnection.ConnectionPool) error {
	dbpool = pool
	router := mux.NewRouter()

	router.HandleFunc("/auth/guest", GuestHandler).Methods("POST")
	router.HandleFunc("/profile", auth.Protect(ProfileGetHandler)).Methods("GET")
	router.HandleFunc("/profile", auth.Protect(ProfileHandler)).Methods("POST")
	router.HandleFunc("/api/v1/plans/generate", auth.Protect(ResponceGenerateHandler)).Methods("POST")

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
