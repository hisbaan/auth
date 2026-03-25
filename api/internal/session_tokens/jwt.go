package sessiontokens

import (
	"auth/internal/utils/ulidutil"
	"crypto/ed25519"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/oklog/ulid/v2"
)

const (
	TokenSourceSelf   = "self"
	TokenSourceClient = "client"
)

type GenerateAccessTokenParams struct {
	privateKey  ed25519.PrivateKey
	keyID       string
	issuer      string
	userID      ulid.ULID
	clientID    *ulid.ULID
	tokenSource string
	roles       []string
	expiry      time.Duration
}

type AccessClaims struct {
	TokenType   string   `json:"token_type"`
	TokenSource string   `json:"token_source"`
	ClientID    *string  `json:"client_id,omitempty"`
	Roles       []string `json:"roles"`
	jwt.RegisteredClaims
}

type RefreshClaims struct {
	TokenType   string  `json:"token_type"`
	TokenSource string  `json:"token_source"`
	ClientID    *string `json:"client_id,omitempty"`
	jwt.RegisteredClaims
}

func GenerateAccessToken(params GenerateAccessTokenParams) (string, error) {
	var clientID *string
	if params.clientID != nil {
		value := ulidutil.ToPrefixed("client", *params.clientID)
		clientID = &value
	}

	claims := AccessClaims{
		TokenType:   "access",
		TokenSource: params.tokenSource,
		ClientID:    clientID,
		Roles:       params.roles,
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
	privateKey  ed25519.PrivateKey
	keyID       string
	issuer      string
	userID      ulid.ULID
	clientID    *ulid.ULID
	tokenSource string
	tokenID     ulid.ULID
	expiry      time.Duration
}

func GenerateRefreshToken(params GenerateRefreshTokenParams) (string, error) {
	var clientID *string
	if params.clientID != nil {
		value := ulidutil.ToPrefixed("client", *params.clientID)
		clientID = &value
	}

	claims := RefreshClaims{
		TokenType:   "refresh",
		TokenSource: params.tokenSource,
		ClientID:    clientID,
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
