package sessions

import (
	"auth/internal/utils/jwtutil"
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

// AccessClaims is the first-party session access token (JOSE typ
// "session+jwt"). It is an internal format carried in cookies or bearer
// headers by this service's own frontend, never issued to OIDC clients.
type AccessClaims struct {
	Roles []string `json:"roles"`
	jwt.RegisteredClaims
}

// RefreshClaims is the refresh token (JOSE typ "refresh+jwt"). Both
// first-party sessions and OIDC client grants use it; token_source and
// client_id say which.
type RefreshClaims struct {
	TokenSource string  `json:"token_source"`
	ClientID    *string `json:"client_id,omitempty"`
	jwt.RegisteredClaims
}

type GenerateAccessTokenParams struct {
	privateKey ed25519.PrivateKey
	keyID      string
	issuer     string
	userID     ulid.ULID
	roles      []string
	expiry     time.Duration
}

func GenerateAccessToken(params GenerateAccessTokenParams) (string, error) {
	claims := AccessClaims{
		Roles: params.roles,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   ulidutil.ToPrefixed("user", params.userID),
			Issuer:    params.issuer,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(params.expiry)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
	token.Header["kid"] = params.keyID
	token.Header["typ"] = jwtutil.SessionTokenJWTType
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
	token.Header["typ"] = jwtutil.RefreshTokenJWTType
	return token.SignedString(params.privateKey)
}
