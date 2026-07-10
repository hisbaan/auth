package oidc

import (
	"auth/internal/utils/jwtutil"
	"auth/internal/utils/ulidutil"
	"crypto/ed25519"
	"crypto/rand"
	"slices"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/oklog/ulid/v2"
)

const testIssuer = "https://auth.example.com"

func newTestKey(t *testing.T) (ed25519.PrivateKey, ed25519.PublicKey) {
	t.Helper()
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	return privateKey, publicKey
}

func generateTestAccessToken(t *testing.T, privateKey ed25519.PrivateKey, userID ulid.ULID, clientID ulid.ULID, scopes []string, expiry time.Duration) string {
	t.Helper()
	token, err := GenerateAccessToken(GenerateAccessTokenParams{
		privateKey: privateKey,
		keyID:      "test-key",
		issuer:     testIssuer,
		userID:     userID,
		clientID:   clientID,
		scopes:     scopes,
		expiry:     expiry,
	})
	if err != nil {
		t.Fatalf("generate access token: %v", err)
	}
	return token
}

func TestAccessTokenRoundTrip(t *testing.T) {
	privateKey, publicKey := newTestKey(t)
	userID := ulid.Make()
	clientID := ulid.Make()

	token := generateTestAccessToken(t, privateKey, userID, clientID, []string{"openid", "profile", "email"}, time.Minute)

	claims, err := ValidateAccessToken(publicKey, testIssuer, token)
	if err != nil {
		t.Fatalf("validate access token: %v", err)
	}

	if claims.Subject != ulidutil.ToPrefixed("user", userID) {
		t.Errorf("sub = %q, want %q", claims.Subject, ulidutil.ToPrefixed("user", userID))
	}
	wantClientID := ulidutil.ToPrefixed("client", clientID)
	if claims.ClientID != wantClientID {
		t.Errorf("client_id = %q, want %q", claims.ClientID, wantClientID)
	}
	if !slices.Contains(claims.Audience, wantClientID) {
		t.Errorf("aud = %v, want it to contain %q", claims.Audience, wantClientID)
	}
	if claims.ID == "" {
		t.Error("jti is empty")
	}
	if got := claims.Scopes(); !slices.Equal(got, []string{"openid", "profile", "email"}) {
		t.Errorf("scopes = %v, want [openid profile email]", got)
	}
}

func TestAccessTokenHasATJWTType(t *testing.T) {
	privateKey, publicKey := newTestKey(t)

	token := generateTestAccessToken(t, privateKey, ulid.Make(), ulid.Make(), []string{"openid"}, time.Minute)

	// jwtutil enforces the exact typ, so validating as at+jwt passing and any
	// other typ failing proves the header.
	if _, err := jwtutil.ValidateToken(publicKey, testIssuer, jwtutil.AccessTokenJWTType, token, &AccessTokenClaims{}); err != nil {
		t.Errorf("expected token to validate as %s: %v", jwtutil.AccessTokenJWTType, err)
	}
	if _, err := jwtutil.ValidateToken(publicKey, testIssuer, jwtutil.SessionTokenJWTType, token, &AccessTokenClaims{}); err == nil {
		t.Errorf("expected token not to validate as %s", jwtutil.SessionTokenJWTType)
	}
}

func TestValidateAccessTokenRejectsIDToken(t *testing.T) {
	privateKey, publicKey := newTestKey(t)
	userID := ulid.Make()
	clientID := ulid.Make()

	idToken, err := GenerateIDToken(GenerateIDTokenParams{
		privateKey: privateKey,
		keyID:      "test-key",
		issuer:     testIssuer,
		userID:     userID,
		clientID:   clientID,
		user:       nil,
		scopes:     []string{"openid"},
		nonce:      nil,
		expiry:     time.Minute,
	})
	if err != nil {
		t.Fatalf("generate id token: %v", err)
	}

	if _, err := ValidateAccessToken(publicKey, testIssuer, idToken); err == nil {
		t.Error("expected ID token to be rejected as an access token")
	}
}

func TestValidateAccessTokenRejectsWrongTyp(t *testing.T) {
	privateKey, publicKey := newTestKey(t)
	clientID := ulidutil.ToPrefixed("client", ulid.Make())

	// Correct claims shape but default "JWT" typ header.
	claims := AccessTokenClaims{
		ClientID: clientID,
		Scope:    "openid",
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        ulidutil.ToPrefixed("token", ulid.Make()),
			Subject:   ulidutil.ToPrefixed("user", ulid.Make()),
			Issuer:    testIssuer,
			Audience:  jwt.ClaimStrings{clientID},
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute)),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims).SignedString(privateKey)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	if _, err := ValidateAccessToken(publicKey, testIssuer, token); err == nil {
		t.Error("expected token without at+jwt typ to be rejected")
	}
}

func TestValidateAccessTokenRejectsWrongIssuer(t *testing.T) {
	privateKey, publicKey := newTestKey(t)

	token := generateTestAccessToken(t, privateKey, ulid.Make(), ulid.Make(), []string{"openid"}, time.Minute)

	if _, err := ValidateAccessToken(publicKey, "https://other.example.com", token); err == nil {
		t.Error("expected token with wrong issuer to be rejected")
	}
}

func TestValidateAccessTokenRejectsExpired(t *testing.T) {
	privateKey, publicKey := newTestKey(t)

	token := generateTestAccessToken(t, privateKey, ulid.Make(), ulid.Make(), []string{"openid"}, -time.Minute)

	if _, err := ValidateAccessToken(publicKey, testIssuer, token); err == nil {
		t.Error("expected expired token to be rejected")
	}
}

func TestValidateAccessTokenRejectsWrongKey(t *testing.T) {
	privateKey, _ := newTestKey(t)
	_, otherPublicKey := newTestKey(t)

	token := generateTestAccessToken(t, privateKey, ulid.Make(), ulid.Make(), []string{"openid"}, time.Minute)

	if _, err := ValidateAccessToken(otherPublicKey, testIssuer, token); err == nil {
		t.Error("expected token signed with a different key to be rejected")
	}
}
