package core_auth_test

import (
	"testing"
	"time"

	core_auth "sport_app/internal/core/auth"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWT_CreateAndParseToken_Success(t *testing.T) {
	jwt := core_auth.NewJWT(core_auth.Config{Secret: "super-secret-key-12345", TokenTTL: 2 * time.Hour})

	// Registered user
	token, err := jwt.CreateToken("user-42", false)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	claims, err := jwt.ParseToken(token)
	require.NoError(t, err)
	assert.Equal(t, "user-42", claims.UserID)
	assert.False(t, claims.IsAnonymous)

	// Anonymous guest user
	guestToken, err := jwt.CreateToken("guest-99", true)
	require.NoError(t, err)
	assert.NotEmpty(t, guestToken)

	guestClaims, err := jwt.ParseToken(guestToken)
	require.NoError(t, err)
	assert.Equal(t, "guest-99", guestClaims.UserID)
	assert.True(t, guestClaims.IsAnonymous)
}

func TestJWT_ParseToken_Expired(t *testing.T) {
	// Negative TTL produces an already-expired token
	jwt := core_auth.NewJWT(core_auth.Config{Secret: "super-secret-key-12345", TokenTTL: -1 * time.Minute})

	token, err := jwt.CreateToken("user-expired", false)
	require.NoError(t, err)

	claims, err := jwt.ParseToken(token)
	assert.Error(t, err)
	assert.Empty(t, claims.UserID)
	assert.Contains(t, err.Error(), "token is expired")
}

func TestJWT_ParseToken_InvalidSignature(t *testing.T) {
	jwt1 := core_auth.NewJWT(core_auth.Config{Secret: "key-alpha", TokenTTL: 1 * time.Hour})
	jwt2 := core_auth.NewJWT(core_auth.Config{Secret: "key-beta", TokenTTL: 1 * time.Hour})

	token, err := jwt1.CreateToken("user-1", false)
	require.NoError(t, err)

	// Validating with a different secret
	claims, err := jwt2.ParseToken(token)
	assert.Error(t, err)
	assert.Empty(t, claims.UserID)
	assert.Contains(t, err.Error(), "signature is invalid")
}

func TestJWT_ParseToken_MalformedToken(t *testing.T) {
	jwt := core_auth.NewJWT(core_auth.Config{Secret: "secret", TokenTTL: 1 * time.Hour})

	malformedTokens := []string{
		"",
		"not-a-jwt",
		"header.payload",
		"a.b.c.d",
	}

	for _, badToken := range malformedTokens {
		claims, err := jwt.ParseToken(badToken)
		assert.Error(t, err, "expected error for token: %s", badToken)
		assert.Empty(t, claims.UserID)
	}
}
