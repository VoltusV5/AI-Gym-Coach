package core_http_middleware

import (
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCORSAllowsIdempotentBrowserRequests(t *testing.T) {
	handler := CORS()(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { t.Error("preflight must not execute handler") }))
	r := httptest.NewRequest("OPTIONS", "/api/v2/workouts/complete", nil)
	r.Header.Set("Origin", "http://localhost:5173")
	r.Header.Set("Access-Control-Request-Method", "POST")
	r.Header.Set("Access-Control-Request-Headers", "authorization,content-type,idempotency-key")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	require.Equal(t, 204, w.Code)
	require.Contains(t, strings.ToLower(w.Header().Get("Access-Control-Allow-Headers")), "idempotency-key")
}
