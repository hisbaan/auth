package middleware

import (
	"auth/internal/sessions"
	"auth/internal/utils/jwtutil"
	"auth/internal/utils/ulidutil"
	"crypto/ed25519"
	"crypto/rand"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/oklog/ulid/v2"
)

const testIssuer = "https://auth.example.com"

func signSessionClaims(t *testing.T, privateKey ed25519.PrivateKey, typ string) string {
	t.Helper()
	claims := sessions.AccessClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   ulidutil.ToPrefixed("user", ulid.Make()),
			Issuer:    testIssuer,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
	if typ != "" {
		token.Header["typ"] = typ
	}
	signed, err := token.SignedString(privateKey)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return signed
}

func TestAccessClaimsFromRequestAcceptsSessionToken(t *testing.T) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("Authorization", "Bearer "+signSessionClaims(t, privateKey, jwtutil.SessionTokenJWTType))

	if _, err := AccessClaimsFromRequest(r, publicKey, testIssuer); err != nil {
		t.Fatalf("expected session token to be accepted, got %v", err)
	}
}

func TestAccessClaimsFromRequestRejectsOtherTyps(t *testing.T) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	// Valid session claims are not enough: only typ session+jwt is a session
	// token. at+jwt is an OIDC client token, refresh+jwt a refresh token, and
	// bare JWT could be anything (e.g. an ID token).
	for _, typ := range []string{jwtutil.AccessTokenJWTType, jwtutil.RefreshTokenJWTType, "", "JWT"} {
		r := httptest.NewRequest("GET", "/", nil)
		r.Header.Set("Authorization", "Bearer "+signSessionClaims(t, privateKey, typ))

		if _, err := AccessClaimsFromRequest(r, publicKey, testIssuer); err == nil {
			t.Errorf("expected token with typ %q to be rejected as a session token", typ)
		}
	}
}
