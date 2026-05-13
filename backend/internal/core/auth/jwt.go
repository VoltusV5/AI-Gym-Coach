package core_auth

import (
	"fmt"
	"time"

	core_errors "sport_app/internal/core/errors"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID      string
	IsAnonymous bool
}

type JWT struct {
	secret   []byte
	tokenTTL time.Duration
}

func NewJWT(config Config) *JWT {
	return &JWT{
		secret:   []byte(config.Secret),
		tokenTTL: config.TokenTTL,
	}
}

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