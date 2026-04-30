package stringutil

import (
	"auth/internal/apperror"
	"net/mail"
	"strings"
)

const (
	MaxEmailLength    = 254
	MaxUsernameLength = 64
	MaxPasswordLength = 1024
)

func NonEmpty(input string) (string, error) {
	output := strings.TrimSpace(input)
	if output == "" {
		return "", apperror.NewBadRequest("Invalid request")
	}
	return output, nil
}

func NormalizeEmail(input string) (string, error) {
	email := strings.ToLower(strings.TrimSpace(input))
	if email == "" || len(email) > MaxEmailLength {
		return "", apperror.NewBadRequest("Invalid email")
	}

	address, err := mail.ParseAddress(email)
	if err != nil || address.Address != email || strings.Count(email, "@") != 1 {
		return "", apperror.NewBadRequest("Invalid email")
	}

	return email, nil
}

func ValidateUsername(input string) (string, error) {
	username := strings.TrimSpace(input)
	if username == "" || len(username) > MaxUsernameLength {
		return "", apperror.NewBadRequest("Invalid username")
	}

	return username, nil
}

func ValidatePassword(input string) error {
	if input == "" || len(input) > MaxPasswordLength {
		return apperror.NewBadRequest("Invalid password")
	}

	return nil
}
