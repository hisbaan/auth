package auth

import (
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
