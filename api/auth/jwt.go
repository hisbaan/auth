package auth

import (
	"auth/internal/apperror"
	"crypto/ed25519"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/oklog/ulid/v2"
)

type GenerateAccessTokenParams struct {
	privateKey ed25519.PrivateKey
	issuer     string
	userID     ulid.ULID
	expiry     time.Duration
}

func GenerateAccessToken(params GenerateAccessTokenParams) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   params.userID.String(),
		Issuer:    params.issuer,
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(params.expiry)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
	return token.SignedString(params.privateKey)
}

type GenerateRefreshTokenParams struct {
	privateKey ed25519.PrivateKey
	issuer     string
	userID     ulid.ULID
	tokenID    ulid.ULID
	expiry     time.Duration
}

func GenerateRefreshToken(params GenerateRefreshTokenParams) (string, error) {
	claims := jwt.RegisteredClaims{
		ID:        params.tokenID.String(),
		Subject:   params.userID.String(),
		Issuer:    params.issuer,
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(params.expiry)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
	return token.SignedString(params.privateKey)
}

func ValidateToken(publicKey ed25519.PublicKey, token string) (*jwt.Token, *jwt.RegisteredClaims, error) {
	claims := &jwt.RegisteredClaims{}
	verifiedToken, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodEd25519); !ok {
			return nil, apperror.NewUnauthorized("Invalid token")
		}
		return publicKey, nil
	})
	if err != nil || !verifiedToken.Valid {
		return nil, nil, apperror.NewUnauthorized("Invalid token")
	}
	return verifiedToken, claims, nil
}
