package oidc

import (
	"auth/internal/jet/postgres/public/model"
	"auth/internal/utils/ulidutil"
	"crypto/ed25519"
	"slices"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/oklog/ulid/v2"
)

type GenerateIDTokenParams struct {
	privateKey ed25519.PrivateKey
	keyID      string
	issuer     string
	userID     ulid.ULID
	clientID   ulid.ULID
	user       *model.Users
	scopes     []string
	nonce      *string
	expiry     time.Duration
}

type IDTokenClaims struct {
	Nonce             *string `json:"nonce,omitempty"`
	Email             *string `json:"email,omitempty"`
	EmailVerified     *bool   `json:"email_verified,omitempty"`
	PreferredUsername *string `json:"preferred_username,omitempty"`
	jwt.RegisteredClaims
}

func GenerateIDToken(params GenerateIDTokenParams) (string, error) {
	claims := IDTokenClaims{
		Nonce: params.nonce,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   ulidutil.ToPrefixed("user", params.userID),
			Issuer:    params.issuer,
			Audience:  jwt.ClaimStrings{ulidutil.ToPrefixed("client", params.clientID)},
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(params.expiry)),
		},
	}

	if params.user != nil {
		if slices.Contains(params.scopes, "profile") {
			claims.PreferredUsername = &params.user.Username
		}
		if slices.Contains(params.scopes, "email") {
			claims.Email = &params.user.Email
			claims.EmailVerified = &params.user.EmailVerified
		}
	}

	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
	token.Header["kid"] = params.keyID
	return token.SignedString(params.privateKey)
}
