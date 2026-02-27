package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var secretKey = []byte("JWT_SECRET")

func CreateToken(user_id string, is_anonymous bool) (string, error) {
	iat := time.Now()
	claims := jwt.MapClaims{
		"sub":          user_id,
		"is_anonymous": is_anonymous,
		"iat":          iat.Unix(),
		"exp":          iat.AddDate(0, 0, 90).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}
