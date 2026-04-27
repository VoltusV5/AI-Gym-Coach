package core_http_middleware

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	core_errors "sport_app/internal/core/errors"
	core_logger "sport_app/internal/core/logger"
	core_http_responce "sport_app/internal/core/transport/http/responce"
)

type rateLimitEntry struct {
	count       int
	windowStart time.Time
}

func RateLimit(maxRequests int, window time.Duration) Middleware {
	var (
		mu      sync.Mutex
		entries = make(map[string]rateLimitEntry)
	)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			log := core_logger.FromContext(ctx)
			responseHandler := core_http_responce.NewHTTPResponce(log, w)

			key := clientIPFromRequest(r)
			now := time.Now()

			mu.Lock()
			entry := entries[key]
			if entry.windowStart.IsZero() || now.Sub(entry.windowStart) >= window {
				entry = rateLimitEntry{
					count:       1,
					windowStart: now,
				}
				entries[key] = entry
				mu.Unlock()

				next.ServeHTTP(w, r)
				return
			}

			if entry.count >= maxRequests {
				mu.Unlock()
				responseHandler.ErrorResponse(
					fmt.Errorf(
						"rate limit exceeded for ip='%s': %w",
						key,
						core_errors.ErrTooManyRequests,
					),
					"rate limit exceeded",
				)
				return
			}

			entry.count++
			entries[key] = entry
			mu.Unlock()

			next.ServeHTTP(w, r)
		})
	}
}

func clientIPFromRequest(r *http.Request) string {
	if ip := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); ip != "" {
		parts := strings.Split(ip, ",")
		if len(parts) > 0 {
			candidate := strings.TrimSpace(parts[0])
			if candidate != "" {
				return candidate
			}
		}
	}

	if ip := strings.TrimSpace(r.Header.Get("X-Real-IP")); ip != "" {
		return ip
	}

	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err == nil && host != "" {
		return host
	}

	return strings.TrimSpace(r.RemoteAddr)
}
