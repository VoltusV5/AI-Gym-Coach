package core_http_middleware

import (
	"github.com/stretchr/testify/require"
	"net/url"
	"testing"
)

func TestShareBearerTokenIsRedactedFromLoggedURL(t *testing.T) {
	for _, path := range []string{"/api/v2/shares/synthetic-bearer-value", "/api/v2/shares/synthetic-bearer-value/fork", "/api/v2/shares/synthetic-bearer-value/save"} {
		u, err := url.Parse(path + "?token=query-secret&limit=20")
		require.NoError(t, err)
		logged := sanitizeURL(u)
		require.NotContains(t, logged, "synthetic-bearer-value")
		require.NotContains(t, logged, "query-secret")
		require.Contains(t, logged, "[REDACTED]")
		require.Contains(t, logged, "limit=20")
	}
}
