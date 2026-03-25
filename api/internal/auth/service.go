package auth

import (
	"auth/internal/emails"
	"auth/internal/repositories"
	sessiontokens "auth/internal/session_tokens"
	"crypto/ed25519"
	"database/sql"
	"time"
)

type AuthService struct {
	db                  *sql.DB
	jwtSigningKey       ed25519.PrivateKey
	jwtSigningKeyID     string
	issuer              string
	cookieDomain        string
	accessTokenExpiry   time.Duration
	refreshTokenExpiry  time.Duration
	emailService        *emails.EmailService
	sessionTokenService sessiontokens.SessionTokenService

	userRepo                   repositories.UserRepository
	roleRepo                   repositories.RoleRepository
	refreshTokenRepo           repositories.RefreshTokenRepository
	passwordResetTokenRepo     repositories.PasswordResetTokenRepository
	emailVerificationTokenRepo repositories.EmailVerificationTokenRepository
	eventRepo                  repositories.EventRepository
}

func NewAuthService(db *sql.DB, signingKey ed25519.PrivateKey, signingKeyID string, issuer string, emailService *emails.EmailService, cookieDomain string) *AuthService {
	return &AuthService{
		db:                         db,
		jwtSigningKey:              signingKey,
		jwtSigningKeyID:            signingKeyID,
		issuer:                     issuer,
		cookieDomain:               cookieDomain,
		accessTokenExpiry:          time.Duration(15) * time.Minute,
		refreshTokenExpiry:         time.Duration(168) * time.Hour, // 7 days
		emailService:               emailService,
		sessionTokenService:        sessiontokens.NewSessionTokenService(signingKey, signingKeyID, issuer, time.Duration(15)*time.Minute, time.Duration(168)*time.Hour, repositories.NewRoleRepository(db), repositories.NewRefreshTokenRepository(db), repositories.NewEventRepository(db)),
		userRepo:                   repositories.NewUserRepository(db),
		roleRepo:                   repositories.NewRoleRepository(db),
		refreshTokenRepo:           repositories.NewRefreshTokenRepository(db),
		passwordResetTokenRepo:     repositories.NewPasswordResetTokenRepository(db),
		emailVerificationTokenRepo: repositories.NewEmailVerificationTokenRepository(db),
		eventRepo:                  repositories.NewEventRepository(db),
	}
}
