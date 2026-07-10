package oidc

import (
	"auth/internal/apperror"
	"auth/internal/utils/jwtutil"
	"auth/internal/utils/ulidutil"
	"crypto/ed25519"
	"slices"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/oklog/ulid/v2"
)

type AccessTokenClaims struct {
	ClientID string `json:"client_id"`
	Scope    string `json:"scope"`
	jwt.RegisteredClaims
}

func (c *AccessTokenClaims) Scopes() []string {
	return strings.Fields(c.Scope)
}

type GenerateAccessTokenParams struct {
	privateKey ed25519.PrivateKey
	keyID      string
	issuer     string
	userID     ulid.ULID
	clientID   ulid.ULID
	scopes     []string
	expiry     time.Duration
}

func GenerateAccessToken(params GenerateAccessTokenParams) (string, error) {
	clientID := ulidutil.ToPrefixed("client", params.clientID)
	claims := AccessTokenClaims{
		ClientID: clientID,
		Scope:    strings.Join(params.scopes, " "),
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        ulidutil.ToPrefixed("token", ulid.Make()),
			Subject:   ulidutil.ToPrefixed("user", params.userID),
			Issuer:    params.issuer,
			Audience:  jwt.ClaimStrings{clientID},
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(params.expiry)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
	token.Header["kid"] = params.keyID
	token.Header["typ"] = jwtutil.AccessTokenJWTType
	return token.SignedString(params.privateKey)
}

func ValidateAccessToken(publicKey ed25519.PublicKey, issuer string, token string) (*AccessTokenClaims, error) {
	claims, err := jwtutil.ValidateToken(publicKey, issuer, jwtutil.AccessTokenJWTType, token, &AccessTokenClaims{})
	if err != nil {
		return nil, err
	}
	if claims.ClientID == "" || !slices.Contains(claims.Audience, claims.ClientID) {
		return nil, apperror.NewUnauthorized("Invalid token")
	}
	return claims, nil
}
