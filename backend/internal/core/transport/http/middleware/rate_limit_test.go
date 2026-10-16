package core_http_middleware

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func TestSanitizeURL(t *testing.T) {
	tests := []struct {
		name     string
		rawURL   string
		expected string
	}{
		{"nil url", "", ""},
		{"path only", "/api/v1/notes", "/api/v1/notes"},
		{"safe query params", "/api/v1/notes?limit=10&offset=0", "/api/v1/notes?limit=10&offset=0"},
		{"sensitive token redacted", "/api/v1/notes?token=secret123&limit=5", "/api/v1/notes?limit=5&token=%5BREDACTED%5D"},
		{"sensitive password redacted", "/login?password=mysecret&user=john", "/login?password=%5BREDACTED%5D&user=john"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var u *url.URL
			if tt.rawURL != "" {
				var err error
				u, err = url.Parse(tt.rawURL)
				if err != nil {
					t.Fatalf("Failed to parse URL: %v", err)
				}
			}
			got := sanitizeURL(u)
			if got != tt.expected {
				t.Errorf("sanitizeURL(%q) = %q; want %q", tt.rawURL, got, tt.expected)
			}
		})
	}
}

func TestSlidingWindowLimiter(t *testing.T) {
	lim := newSlidingWindowLimiter()

	// Allow up to 3 requests per 100ms
	window := 100 * time.Millisecond
	key := "client-1"

	if !lim.allow(key, 3, window) {
		t.Errorf("1st request should be allowed")
	}
	if !lim.allow(key, 3, window) {
		t.Errorf("2nd request should be allowed")
	}
	if !lim.allow(key, 3, window) {
		t.Errorf("3rd request should be allowed")
	}
	if lim.allow(key, 3, window) {
		t.Errorf("4th request should be blocked")
	}

	// Wait for window to expire
	time.Sleep(110 * time.Millisecond)
	if !lim.allow(key, 3, window) {
		t.Errorf("Request after window expiration should be allowed")
	}
}

func TestRateLimitMiddleware(t *testing.T) {
	handler := RateLimit(2, 50*time.Millisecond)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req1 := httptest.NewRequest("GET", "/test", nil)
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)
	if rec1.Code != http.StatusOK {
		t.Errorf("req1 status = %d; want 200", rec1.Code)
	}

	req2 := httptest.NewRequest("GET", "/test", nil)
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Errorf("req2 status = %d; want 200", rec2.Code)
	}

	req3 := httptest.NewRequest("GET", "/test", nil)
	rec3 := httptest.NewRecorder()
	handler.ServeHTTP(rec3, req3)
	if rec3.Code != http.StatusTooManyRequests {
		t.Errorf("req3 status = %d; want 429", rec3.Code)
	}
}
