package core_http_middleware_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	core_auth "sport_app/internal/core/auth"
	core_http_middleware "sport_app/internal/core/transport/http/middleware"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProtectMiddleware(t *testing.T) {
	jwtService := core_auth.NewJWT(core_auth.Config{
		Secret:   "test-auth-secret-key-9999",
		TokenTTL: 1 * time.Hour,
	})

	handlerExecuted := false
	var capturedUserID string
	var capturedIsAnon bool

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerExecuted = true
		userID, err := core_http_middleware.UserIDFromContext(r.Context())
		if err == nil {
			capturedUserID = userID
		}
		if isAnon, ok := r.Context().Value(core_http_middleware.IsAnonymousKey).(bool); ok {
			capturedIsAnon = isAnon
		}
		w.WriteHeader(http.StatusOK)
	})

	protected := core_http_middleware.Protect(jwtService)(nextHandler)

	t.Run("Missing Authorization Header returns 401", func(t *testing.T) {
		handlerExecuted = false
		req := httptest.NewRequest(http.MethodGet, "/api/v1/protected", nil)
		rec := httptest.NewRecorder()

		protected.ServeHTTP(rec, req)

		assert.False(t, handlerExecuted)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)

		var body map[string]string
		err := json.NewDecoder(rec.Body).Decode(&body)
		require.NoError(t, err)
		assert.Contains(t, body["message"], "missing 'Authorization' header")
	})

	t.Run("Invalid Authorization format returns 401", func(t *testing.T) {
		invalidHeaders := []string{
			"Basic dXNlcjpwYXNz",
			"Bearer",
			"Bearer token1 token2",
			"Token 12345",
		}

		for _, header := range invalidHeaders {
			handlerExecuted = false
			req := httptest.NewRequest(http.MethodGet, "/api/v1/protected", nil)
			req.Header.Set("Authorization", header)
			rec := httptest.NewRecorder()

			protected.ServeHTTP(rec, req)

			assert.False(t, handlerExecuted)
			assert.Equal(t, http.StatusUnauthorized, rec.Code)
		}
	})

	t.Run("Expired Token returns 401", func(t *testing.T) {
		expiredJWT := core_auth.NewJWT(core_auth.Config{
			Secret:   "test-auth-secret-key-9999",
			TokenTTL: -1 * time.Minute,
		})
		expiredToken, err := expiredJWT.CreateToken("user-1", false)
		require.NoError(t, err)

		handlerExecuted = false
		req := httptest.NewRequest(http.MethodGet, "/api/v1/protected", nil)
		req.Header.Set("Authorization", "Bearer "+expiredToken)
		rec := httptest.NewRecorder()

		protected.ServeHTTP(rec, req)

		assert.False(t, handlerExecuted)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("Valid Registered User Token injects context and calls next", func(t *testing.T) {
		validToken, err := jwtService.CreateToken("usr-777", false)
		require.NoError(t, err)

		handlerExecuted = false
		capturedUserID = ""
		capturedIsAnon = true

		req := httptest.NewRequest(http.MethodGet, "/api/v1/protected", nil)
		req.Header.Set("Authorization", "Bearer "+validToken)
		rec := httptest.NewRecorder()

		protected.ServeHTTP(rec, req)

		assert.True(t, handlerExecuted)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "usr-777", capturedUserID)
		assert.False(t, capturedIsAnon)
	})

	t.Run("Valid Guest User Token injects anonymous flag", func(t *testing.T) {
		guestToken, err := jwtService.CreateToken("guest-555", true)
		require.NoError(t, err)

		handlerExecuted = false
		capturedUserID = ""

		req := httptest.NewRequest(http.MethodGet, "/api/v1/protected", nil)
		req.Header.Set("Authorization", "Bearer "+guestToken)
		rec := httptest.NewRecorder()

		protected.ServeHTTP(rec, req)

		assert.True(t, handlerExecuted)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "guest-555", capturedUserID)
		assert.True(t, capturedIsAnon)
	})
}
