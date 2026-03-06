package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const (
	UserIDKey      contextKey = "userID"
	IsAnonymousKey contextKey = "isAnonymous"
)

func Protect(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Header missing Authorization", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			http.Error(w, "Invalid header format. Expected 'Bearer <token>'", http.StatusUnauthorized)
			return
		}
		tokenString := parts[1]

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
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
			http.Error(w, "Token is invalid", http.StatusUnauthorized)
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
