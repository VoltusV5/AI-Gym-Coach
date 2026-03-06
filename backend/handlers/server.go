package handlers

import (
	"errors"
	"net/http"
	"sport_app/auth"

	"github.com/gorilla/mux"
)

func StartHTTPServer() error {
	router := mux.NewRouter()

	router.HandleFunc("/auth/guest", GuestHandler).Methods("POST")
	router.HandleFunc("/profile", auth.Protect(ProfileHandler)).Methods("POST")
	router.HandleFunc("/api/v1/plans/generate", auth.Protect(ResponceGenerateHandler)).Methods("POST")

	if err := http.ListenAndServe(":8080", router); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		} else {
			return err
		}
	}
	return nil
}
