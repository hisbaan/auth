package jwtutil

import (
	"auth/internal/apperror"
	"crypto/ed25519"

	"github.com/golang-jwt/jwt/v5"
)

const (
	// issued to OIDC clients.
	AccessTokenJWTType = "at+jwt"
	// marks first-party session access tokens.
	SessionTokenJWTType = "session+jwt"
	// marks refresh tokens (first-party and client).
	RefreshTokenJWTType = "refresh+jwt"
)

// verifies signature, algorithm, JOSE typ header, issuer
func ValidateToken[T jwt.Claims](publicKey ed25519.PublicKey, issuer string, typ string, tokenString string, claims T) (T, error) {
	var zero T
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodEd25519); !ok {
			return nil, apperror.NewUnauthorized("Invalid token")
		}
		return publicKey, nil
	}, jwt.WithIssuer(issuer), jwt.WithExpirationRequired())
	if err != nil || !token.Valid {
		return zero, apperror.NewUnauthorized("Invalid token")
	}
	if headerTyp, _ := token.Header["typ"].(string); headerTyp != typ {
		return zero, apperror.NewUnauthorized("Invalid token")
	}
	return token.Claims.(T), nil
}
