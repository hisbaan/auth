package stringutil

import (
	"auth/internal/apperror"
	"net/mail"
	"strings"
	"unicode"
)

const (
	MaxEmailLength    = 254
	MaxUsernameLength = 64
	MaxPasswordLength = 1024
	MinPasswordLength = 8
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

func ValidateUserEmail(input string, senderAddress string, blockedDomains []string) (string, error) {
	email, err := NormalizeEmail(input)
	if err != nil {
		return "", err
	}
	if email == strings.ToLower(strings.TrimSpace(senderAddress)) || emailDomainBlocked(email, blockedDomains) {
		return "", apperror.NewBadRequest("Invalid email")
	}

	return email, nil
}

func emailDomainBlocked(email string, blockedDomains []string) bool {
	at := strings.LastIndex(email, "@")
	if at == -1 || at == len(email)-1 {
		return false
	}

	domain := email[at+1:]
	for _, blockedDomain := range blockedDomains {
		blockedDomain = strings.ToLower(strings.TrimSpace(blockedDomain))
		if blockedDomain == "" {
			continue
		}
		if strings.HasPrefix(blockedDomain, "@") {
			blockedDomain = strings.TrimPrefix(blockedDomain, "@")
		}
		if domain == blockedDomain {
			return true
		}
	}

	return false
}

func ValidateUsername(input string) (string, error) {
	username := strings.TrimSpace(input)
	if username == "" || len(username) > MaxUsernameLength {
		return "", apperror.NewBadRequest("Invalid username")
	}

	return username, nil
}

func ValidatePassword(input string) error {
	if input == "" || len(input) > MaxPasswordLength || len(input) < MinPasswordLength || !isStrongPassword(input) {
		return apperror.NewBadRequest("Invalid password")
	}

	return nil
}

func isStrongPassword(input string) bool {
	hasUpper := false
	hasLower := false
	hasNumber := false
	hasSpecial := false

	for _, char := range input {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	return hasUpper && hasLower && hasNumber && hasSpecial
}
