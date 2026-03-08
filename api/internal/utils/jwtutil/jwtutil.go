package jwtutil

import (
	"auth/internal/apperror"
	"crypto/ed25519"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func ValidateToken[T jwt.Claims](publicKey ed25519.PublicKey, token string, claims T) (*jwt.Token, T, error) {
	var zero T
	verifiedToken, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodEd25519); !ok {
			return nil, apperror.NewUnauthorized("Invalid token")
		}
		return publicKey, nil
	})
	if err != nil || !verifiedToken.Valid {
		return nil, zero, apperror.NewUnauthorized("Invalid token")
	}
	return verifiedToken, verifiedToken.Claims.(T), nil
}

func ValidateClaims(claims jwt.RegisteredClaims, issuer string) error {
	if claims.Issuer != issuer {
		return apperror.NewUnauthorized("Invalid token")
	}
	if claims.ExpiresAt.Before(time.Now()) {
		return apperror.NewUnauthorized("Invalid token")
	}
	if claims.ExpiresAt.Before(time.Now()) {
		return apperror.NewUnauthorized("Invalid token")
	}

	return nil
}
