package handlers

import (
	"errors"
	"net/http"

	"github.com/gorilla/mux"
)

func StartHTTPServer() error {
	router := mux.NewRouter()

	router.HandleFunc("/auth/guest", GuestHandler).Methods("POST")

	if err := http.ListenAndServe(":9091", router); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		} else {
			return err
		}
	}
	return nil
}
