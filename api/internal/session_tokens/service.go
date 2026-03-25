package sessiontokens

import (
	"auth/internal/repositories"
	"crypto/ed25519"
	"time"
)

type SessionTokenService struct {
	jwtSigningKey      ed25519.PrivateKey
	jwtSigningKeyID    string
	issuer             string
	accessTokenExpiry  time.Duration
	refreshTokenExpiry time.Duration

	roleRepo         repositories.RoleRepository
	refreshTokenRepo repositories.RefreshTokenRepository
	eventRepo        repositories.EventRepository
}

func NewSessionTokenService(
	signingKey ed25519.PrivateKey,
	signingKeyID string,
	issuer string,
	accessTokenExpiry time.Duration,
	refreshTokenExpiry time.Duration,
	roleRepo repositories.RoleRepository,
	refreshTokenRepo repositories.RefreshTokenRepository,
	eventRepo repositories.EventRepository,
) SessionTokenService {
	return SessionTokenService{
		jwtSigningKey:      signingKey,
		jwtSigningKeyID:    signingKeyID,
		issuer:             issuer,
		accessTokenExpiry:  accessTokenExpiry,
		refreshTokenExpiry: refreshTokenExpiry,
		roleRepo:           roleRepo,
		refreshTokenRepo:   refreshTokenRepo,
		eventRepo:          eventRepo,
	}
}
