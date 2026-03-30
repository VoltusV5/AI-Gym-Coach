package handlers

import (
	"net/http"
	"net/url"
	"strings"
)

func isAllowedDevOrigin(origin string) bool {
	if origin == "" {
		return false
	}
	u, err := url.Parse(origin)
	if err != nil || u.Host == "" {
		return false
	}
	host := strings.ToLower(u.Hostname())
	switch host {
	case "localhost", "127.0.0.1", "::1":
		return true
	}
	if strings.HasPrefix(host, "192.168.") {
		return true
	}
	if strings.HasPrefix(host, "10.") {
		return true
	}
	// Частные сети (Vite --host)
	if strings.HasPrefix(host, "172.") {
		return true
	}
	return false
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if isAllowedDevOrigin(origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PATCH, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		w.Header().Set("Access-Control-Expose-Headers", "Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
