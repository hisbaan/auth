package auth

import (
	"auth/internal/apperror"
	"auth/internal/utils/ulidutil"
	"crypto/ed25519"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/oklog/ulid/v2"
)

type GenerateAccessTokenParams struct {
	privateKey ed25519.PrivateKey
	keyID      string
	issuer     string
	userID     ulid.ULID
	roles      []string
	expiry     time.Duration
}

type AccessClaims struct {
	TokenType string   `json:"token_type"`
	Roles     []string `json:"roles"`
	jwt.RegisteredClaims
}

type RefreshClaims struct {
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

func GenerateAccessToken(params GenerateAccessTokenParams) (string, error) {
	claims := AccessClaims{
		TokenType: "access",
		Roles:     params.roles,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   ulidutil.ToPrefixed("user", params.userID),
			Issuer:    params.issuer,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(params.expiry)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
	token.Header["kid"] = params.keyID
	return token.SignedString(params.privateKey)
}

type GenerateRefreshTokenParams struct {
	privateKey ed25519.PrivateKey
	keyID      string
	issuer     string
	userID     ulid.ULID
	tokenID    ulid.ULID
	expiry     time.Duration
}

func GenerateRefreshToken(params GenerateRefreshTokenParams) (string, error) {
	claims := RefreshClaims{
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        ulidutil.ToPrefixed("token", params.tokenID),
			Subject:   ulidutil.ToPrefixed("user", params.userID),
			Issuer:    params.issuer,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(params.expiry)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
	token.Header["kid"] = params.keyID
	return token.SignedString(params.privateKey)
}

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
