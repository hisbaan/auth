package stringutil

import (
	"auth/internal/apperror"
	"strings"
)

func NonEmpty(input string) (string, error) {
	output := strings.TrimSpace(input)
	if output == "" {
		return "", apperror.NewBadRequest("Invalid request")
	}
	return output, nil
}
