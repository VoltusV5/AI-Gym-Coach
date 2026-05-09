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

type slidingWindowLimiter struct {
	mu      sync.Mutex
	entries map[string]rateLimitEntry
}

func newSlidingWindowLimiter() *slidingWindowLimiter {
	return &slidingWindowLimiter{entries: make(map[string]rateLimitEntry)}
}

func (l *slidingWindowLimiter) allow(key string, maxRequests int, window time.Duration) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	entry := l.entries[key]
	if entry.windowStart.IsZero() || now.Sub(entry.windowStart) >= window {
		l.entries[key] = rateLimitEntry{count: 1, windowStart: now}
		return true
	}
	if entry.count >= maxRequests {
		return false
	}
	entry.count++
	l.entries[key] = entry
	return true
}

func RateLimit(maxRequests int, window time.Duration) Middleware {
	lim := newSlidingWindowLimiter()
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			log := core_logger.FromContext(ctx)
			responseHandler := core_http_responce.NewHTTPResponce(log, w)

			key := clientIPFromRequest(r)
			if !lim.allow(key, maxRequests, window) {
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
			next.ServeHTTP(w, r)
		})
	}
}

func UserRateLimit(maxRequests int, window time.Duration) Middleware {
	lim := newSlidingWindowLimiter()
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			log := core_logger.FromContext(ctx)
			responseHandler := core_http_responce.NewHTTPResponce(log, w)

			userID, err := UserIDFromContext(ctx)
			if err != nil {
				responseHandler.ErrorResponse(err, "rate limit: user id")
				return
			}

			key := "user_rate:" + userID
			if !lim.allow(key, maxRequests, window) {
				responseHandler.ErrorResponse(
					fmt.Errorf(
						"user rate limit exceeded for user_id='%s': %w",
						userID,
						core_errors.ErrTooManyRequests,
					),
					"rate limit exceeded",
				)
				return
			}
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
