package auth

import (
	"auth/internal/apperror"

	"github.com/alexedwards/argon2id"
)

func HashPassword(password string) (string, error) {
	hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return "", apperror.NewInternalServerError("Internal server error")
	}
	return string(hash), nil
}

func ComparePasswordAndHash(password string, hash string) bool {
	match, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		return false
	}
	return match
}
