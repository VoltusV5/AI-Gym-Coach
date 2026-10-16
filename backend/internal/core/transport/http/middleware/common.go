// Package core_http_middleware provides HTTP middleware functions for authentication,
// request identification, structured logging, CORS handling, panic recovery, and rate limiting.
package core_http_middleware

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	core_auth "sport_app/internal/core/auth"
	core_errors "sport_app/internal/core/errors"
	core_logger "sport_app/internal/core/logger"
	core_http_responce "sport_app/internal/core/transport/http/responce"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

const requestIDHeader = "X-Request-ID"

// RequestID generates a unique UUIDv4 for each HTTP request if not already present in the X-Request-ID header,
// and sets the header on both request and response.
//
// Side effects:
//   - Mutates request Header and response Header with "X-Request-ID".
func RequestID() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get(requestIDHeader)
			if requestID == "" {
				requestID = uuid.NewString()
			}

			r.Header.Set(requestIDHeader, requestID)
			w.Header().Set(requestIDHeader, requestID)

			next.ServeHTTP(w, r)
		})
	}
}

// Logger enriches the application logger with request-scoped fields (request_id, sanitized URL)
// and stores the enriched logger in the request context for downstream handlers.
//
// Parameters:
//   - log: Base application logger.
//
// Side effects:
//   - Attaches enriched *core_logger.Logger to the request Context.
func Logger(log *core_logger.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get(requestIDHeader)

			l := log.With(
				zap.String("request_id", requestID),
				zap.String("url", sanitizeURL(r.URL)),
			)

			ctx := core_logger.ToContext(r.Context(), l)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func sanitizeURL(u *url.URL) string {
	if u == nil {
		return ""
	}
	if u.RawQuery == "" {
		return u.Path
	}
	q := u.Query()
	for _, key := range []string{"token", "secret", "password", "key", "auth"} {
		if q.Has(key) {
			q.Set(key, "[REDACTED]")
		}
	}
	return u.Path + "?" + q.Encode()
}

type contextKey string

const (
	// UserIDKey identifies the authenticated user ID string stored in request context.
	UserIDKey contextKey = "userID"
	// IsAnonymousKey identifies the boolean flag distinguishing guest from registered users.
	IsAnonymousKey contextKey = "isAnonymous"
)

// UserIDFromContext extracts the authenticated user ID from context.
//
// Parameters:
//   - ctx: Request context.
//
// Returns:
//   - string: Authenticated user's UUID.
//   - error: core_errors.ErrUnauthorized if user ID is missing or empty.
func UserIDFromContext(ctx context.Context) (string, error) {
	userID, ok := ctx.Value(UserIDKey).(string)
	if !ok || userID == "" {
		return "", fmt.Errorf("user id not found in context: %w", core_errors.ErrUnauthorized)
	}

	return userID, nil
}

// Protect enforces Bearer JWT token authentication on secured routes.
// It parses claims, verifies token validity and expiration, and binds UserIDKey and IsAnonymousKey to context.
//
// Parameters:
//   - jwt: JWT parser and validator service.
//
// Returns:
//   - Middleware: HTTP middleware handler.
//
// Errors:
//   - Returns 401 Unauthorized with JSON error payload if header is missing, malformed, or token is invalid/expired.
func Protect(jwt *core_auth.JWT) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			log := core_logger.FromContext(ctx)
			responceHandler := core_http_responce.NewHTTPResponce(log, w)

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				responceHandler.ErrorResponse(
					fmt.Errorf("missing 'Authorization' header: %w", core_errors.ErrUnauthorized),
					"missing 'Authorization' header",
				)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				responceHandler.ErrorResponse(
					fmt.Errorf("invalid 'Authorization' header format, expected 'Bearer <token>': %w", core_errors.ErrUnauthorized),
					"invalid 'Authorization' header format",
				)
				return
			}

			claims, err := jwt.ParseToken(parts[1])
			if err != nil {
				responceHandler.ErrorResponse(err, "failed to parse JWT token")
				return
			}

			ctx = context.WithValue(ctx, UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, IsAnonymousKey, claims.IsAnonymous)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// CORS adds Cross-Origin Resource Sharing headers for local and private network development clients.
// Automatically responds with 204 No Content for preflight OPTIONS requests.
func CORS() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			if isAllowedDevOrigin(origin) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PATCH, PUT, DELETE")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Request-ID")
			w.Header().Set("Access-Control-Expose-Headers", "Authorization")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

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
	if strings.HasPrefix(host, "192.168.") ||
		strings.HasPrefix(host, "10.") ||
		strings.HasPrefix(host, "172.") {
		return true
	}
	return false
}

// Trace logs entry and exit timing for each HTTP request, recording method, start timestamp, status code, and latency.
func Trace() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			log := core_logger.FromContext(ctx)

			rw := core_http_responce.NewResponceWriter(w)

			before := time.Now()
			log.Debug(
				">>> incoming HTTP request",
				zap.String("http_method", r.Method),
				zap.Time("time", before.UTC()),
			)

			next.ServeHTTP(rw, r)

			log.Debug(
				">>> done HTTP request",
				zap.Int("status_code", rw.GetStatusCode()),
				zap.Duration("latency", time.Since(before)),
			)
		})
	}
}

// Panic recovers from uncaught panics inside HTTP handlers and formats a standardized 500 error response.
func Panic() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			log := core_logger.FromContext(ctx)
			responceHandler := core_http_responce.NewHTTPResponce(log, w)

			defer func() {
				if p := recover(); p != nil {
					responceHandler.PanicResponce(p, "during handle HTTP request got unexpected panic")
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
