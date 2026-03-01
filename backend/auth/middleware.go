package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const (
	UserIDKey      contextKey = "userID"
	IsAnonymousKey contextKey = "isAnonymous"
)

func Protect(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Error HTTP request body", http.StatusBadRequest)
			return
		}

		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		var data map[string]any
		if err := json.Unmarshal(bodyBytes, &data); err != nil {
			http.Error(w, "Invalid format JSON", http.StatusBadRequest)
			return
		}

		tokenInterface, ok := data["token"]
		if !ok {
			http.Error(w, "Missing field token", http.StatusBadRequest)
			return
		}
		tokenString, ok := tokenInterface.(string)
		if !ok {
			http.Error(w, "Missing field token", http.StatusBadRequest)
			return
		}
		tokenString, ok = tokenInterface.(string)
		if !ok {
			http.Error(w, "The token field must be a string.", http.StatusBadRequest)
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return SecretKey, nil
		})

		if err != nil {
			http.Error(w, "Token is invalid or expired: "+err.Error(), http.StatusUnauthorized)
			return
		}
		if !token.Valid {
			http.Error(w, "Token invalid", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "Invalid token structure", http.StatusUnauthorized)
			return
		}

		sub, ok := claims["sub"].(string)
		if !ok || sub == "" {
			http.Error(w, "The token is missing a user ID.", http.StatusUnauthorized)
			return
		}

		isAnonymous, _ := claims["is_anonymous"].(bool)
		ctx := context.WithValue(r.Context(), UserIDKey, sub)
		ctx = context.WithValue(ctx, IsAnonymousKey, isAnonymous)
		r = r.WithContext(ctx)

		next(w, r)
	}
}
