package auth

import "github.com/google/uuid"

func generateUserID() string {
	return uuid.New().String()
}
