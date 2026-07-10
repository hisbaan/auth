package oidc

import (
	"auth/internal/emails"
	"auth/internal/repositories"
	"auth/internal/sessions"
	"crypto/ed25519"
	"database/sql"
	"time"
)

type OIDCService struct {
	db                  *sql.DB
	jwtSigningKey       ed25519.PrivateKey
	jwtSigningKeyID     string
	issuer              string
	frontendURL         string
	cookieDomain        string
	accessTokenExpiry   time.Duration
	idTokenExpiry       time.Duration
	emailService        *emails.EmailService
	sessionTokenService sessions.SessionTokenService

	userRepo                    repositories.UserRepository
	clientRepo                  repositories.ClientRepository
	authorizationCodeRepo       repositories.AuthorizationCodeRepository
	userClientAuthorizationRepo repositories.UserClientAuthorizationRepository
	roleRepo                    repositories.RoleRepository
	refreshTokenRepo            repositories.RefreshTokenRepository
	passwordResetTokenRepo      repositories.PasswordResetTokenRepository
	emailVerificationTokenRepo  repositories.EmailVerificationTokenRepository
	eventRepo                   repositories.EventRepository
}

func NewOIDCService(
	db *sql.DB,
	signingKey ed25519.PrivateKey,
	signingKeyID string,
	issuer string,
	frontendURL string,
	emailService *emails.EmailService,
	cookieDomain string,
	sessionTokenService sessions.SessionTokenService,
	accessTokenExpiry time.Duration,
	idTokenExpiry time.Duration,
) *OIDCService {
	return &OIDCService{
		db:                          db,
		jwtSigningKey:               signingKey,
		jwtSigningKeyID:             signingKeyID,
		issuer:                      issuer,
		frontendURL:                 frontendURL,
		cookieDomain:                cookieDomain,
		accessTokenExpiry:           accessTokenExpiry,
		idTokenExpiry:               idTokenExpiry,
		emailService:                emailService,
		sessionTokenService:         sessionTokenService,
		userRepo:                    repositories.NewUserRepository(db),
		clientRepo:                  repositories.NewClientRepository(db),
		authorizationCodeRepo:       repositories.NewAuthorizationCodeRepository(db),
		userClientAuthorizationRepo: repositories.NewUserClientAuthorizationRepository(db),
		roleRepo:                    repositories.NewRoleRepository(db),
		refreshTokenRepo:            repositories.NewRefreshTokenRepository(db),
		passwordResetTokenRepo:      repositories.NewPasswordResetTokenRepository(db),
		emailVerificationTokenRepo:  repositories.NewEmailVerificationTokenRepository(db),
		eventRepo:                   repositories.NewEventRepository(db),
	}
}
