// Package core_http_middleware реализует авторизацию, идентификацию запросов, журналирование, CORS, восстановление после паник и ограничение частоты запросов.
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

// RequestID создаёт UUID запроса, если заголовок X-Request-ID отсутствует, и записывает идентификатор в заголовки запроса и ответа.
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

// Logger добавляет к журналу log поля запроса и URL без секретов и помещает журнал в контекст запроса.
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
	// UserIDKey задаёт ключ идентификатора авторизованного пользователя в контексте.
	UserIDKey contextKey = "userID"
	// IsAnonymousKey задаёт ключ признака гостевой сессии в контексте.
	IsAnonymousKey contextKey = "isAnonymous"
)

// UserIDFromContext извлекает идентификатор пользователя из ctx.
// Возвращает ErrUnauthorized, если идентификатор отсутствует или пуст.
func UserIDFromContext(ctx context.Context) (string, error) {
	userID, ok := ctx.Value(UserIDKey).(string)
	if !ok || userID == "" {
		return "", fmt.Errorf("user id not found in context: %w", core_errors.ErrUnauthorized)
	}

	return userID, nil
}

// Protect проверяет Bearer JWT с помощью jwt и добавляет идентификатор пользователя и признак гостя в контекст.
// При отсутствии, повреждении или истечении токена отправляет HTTP 401 с JSON-ошибкой.
func Protect(jwt *core_auth.JWT) Middleware {
	return protect(jwt, false)
}

func ProtectV2(jwt *core_auth.JWT) Middleware {
	return protect(jwt, true)
}

func OptionalProtectV2(jwt *core_auth.JWT) Middleware {
	return func(next http.Handler) http.Handler {
		protected := ProtectV2(jwt)(next)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, present := r.Header["Authorization"]; !present {
				next.ServeHTTP(w, r)
				return
			}
			protected.ServeHTTP(w, r)
		})
	}
}

func protect(jwt *core_auth.JWT, v2 bool) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			log := core_logger.FromContext(ctx)
			responceHandler := core_http_responce.NewHTTPResponce(log, w)
			if v2 {
				responceHandler = core_http_responce.NewHTTPResponceV2(log, w, r.Header.Get(requestIDHeader))
			}

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

// CORS разрешает междоменные запросы от локальных клиентов и частных сетей разработки. На предварительный запрос OPTIONS отвечает HTTP 204.
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

// Trace записывает в журнал метод, время начала, HTTP-статус и длительность запроса.
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

// Panic перехватывает панику HTTP-обработчика и формирует стандартный ответ с HTTP 500.
func Panic() Middleware {
	return panicMiddleware(false)
}

func PanicV2() Middleware {
	return panicMiddleware(true)
}

func panicMiddleware(v2 bool) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			log := core_logger.FromContext(ctx)
			responceHandler := core_http_responce.NewHTTPResponce(log, w)
			if v2 {
				responceHandler = core_http_responce.NewHTTPResponceV2(log, w, r.Header.Get(requestIDHeader))
			}

			defer func() {
				if p := recover(); p != nil {
					responceHandler.PanicResponce(p, "during handle HTTP request got unexpected panic")
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
