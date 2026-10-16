// Package core_auth provides token generation and validation mechanisms for user authentication.
package core_auth

import (
	"fmt"
	"time"

	core_errors "sport_app/internal/core/errors"

	"github.com/golang-jwt/jwt/v5"
)

// Claims represents authenticated identity claims embedded in the JWT.
// Parameters:
//   - UserID: unique identifier of the user (guest or registered)
//   - IsAnonymous: boolean indicator whether the session belongs to a guest user
type Claims struct {
	UserID      string
	IsAnonymous bool
}

// JWT manages cryptographic signing and verification of user access tokens.
type JWT struct {
	secret   []byte
	tokenTTL time.Duration
}

// NewJWT initializes a new JWT service with configured secret and token expiration time.
// Parameters:
//   - config: Config struct containing Secret key string and TokenTTL duration
//
// Returns:
//   - *JWT: instantiated token manager ready to sign and verify tokens
func NewJWT(config Config) *JWT {
	return &JWT{
		secret:   []byte(config.Secret),
		tokenTTL: config.TokenTTL,
	}
}

// CreateToken creates and signs a new HMAC-SHA256 JWT containing subject ID and anonymous state.
// Parameters:
//   - userID: unique identifier string for the subject
//   - isAnonymous: true if the user is a guest, false if registered
//
// Returns:
//   - string: signed compact JWT string
//   - error: non-nil if token signing fails
func (j *JWT) CreateToken(userID string, isAnonymous bool) (string, error) {
	iat := time.Now()
	claims := jwt.MapClaims{
		"sub":          userID,
		"is_anonymous": isAnonymous,
		"iat":          iat.Unix(),
		"exp":          iat.Add(j.tokenTTL).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(j.secret)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}

	return tokenString, nil
}

// ParseToken decodes, cryptographically verifies HMAC signature, and extracts claims from a raw JWT string.
// Parameters:
//   - tokenString: raw compact JWT token string
//
// Returns:
//   - Claims: extracted user claims on success
//   - error: core_errors.ErrUnauthorized if the token is invalid, expired, or malformed
func (j *JWT) ParseToken(tokenString string) (Claims, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return j.secret, nil
	})
	if err != nil {
		return Claims{}, fmt.Errorf("parse token: %v: %w", err, core_errors.ErrUnauthorized)
	}

	if !token.Valid {
		return Claims{}, fmt.Errorf("token is invalid: %w", core_errors.ErrUnauthorized)
	}

	mapClaims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return Claims{}, fmt.Errorf("invalid token structure: %w", core_errors.ErrUnauthorized)
	}

	sub, ok := mapClaims["sub"].(string)
	if !ok || sub == "" {
		return Claims{}, fmt.Errorf("token is missing 'sub' claim: %w", core_errors.ErrUnauthorized)
	}

	isAnonymous, _ := mapClaims["is_anonymous"].(bool)

	return Claims{
		UserID:      sub,
		IsAnonymous: isAnonymous,
	}, nil
}
