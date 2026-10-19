// Package core_auth реализует создание и проверку JWT и загрузку настроек авторизации.
package core_auth

import (
	"fmt"
	"time"

	core_errors "sport_app/internal/core/errors"

	"github.com/golang-jwt/jwt/v5"
)

// Claims содержит идентификатор пользователя UserID и признак гостевой сессии IsAnonymous в JWT.
type Claims struct {
	UserID      string
	IsAnonymous bool
}

// JWT подписывает и проверяет токены доступа с заданным секретом и сроком действия.
type JWT struct {
	secret   []byte
	tokenTTL time.Duration
}

// NewJWT создаёт менеджер токенов с секретом Secret и сроком действия TokenTTL из config.
func NewJWT(config Config) *JWT {
	return &JWT{
		secret:   []byte(config.Secret),
		tokenTTL: config.TokenTTL,
	}
}

// CreateToken подписывает JWT алгоритмом HMAC-SHA256 для userID и признака гостя isAnonymous.
// Возвращает строку токена либо ошибку подписи.
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

// ParseToken проверяет подпись и срок действия tokenString и возвращает Claims.
// При повреждённом, просроченном или неверно подписанном токене возвращает ErrUnauthorized.
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
