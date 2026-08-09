package auth

import (
	"auth/internal/emails"
	"auth/internal/repositories"
	"auth/internal/sessions"
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
	blockedEmailDomains []string
	emailService        *emails.EmailService
	sessionTokenService sessions.SessionTokenService

	emailVerificationExpiry         time.Duration
	emailVerificationResendCooldown time.Duration

	userRepo                   repositories.UserRepository
	roleRepo                   repositories.RoleRepository
	refreshTokenRepo           repositories.RefreshTokenRepository
	passwordResetTokenRepo     repositories.PasswordResetTokenRepository
	emailVerificationTokenRepo repositories.EmailVerificationTokenRepository
	eventRepo                  repositories.EventRepository
}

func NewAuthService(db *sql.DB, signingKey ed25519.PrivateKey, signingKeyID string, issuer string, emailService *emails.EmailService, cookieDomain string, blockedEmailDomains []string, sessionTokenService sessions.SessionTokenService) *AuthService {
	return &AuthService{
		db:                              db,
		jwtSigningKey:                   signingKey,
		jwtSigningKeyID:                 signingKeyID,
		issuer:                          issuer,
		cookieDomain:                    cookieDomain,
		blockedEmailDomains:             blockedEmailDomains,
		emailVerificationExpiry:         time.Duration(24) * time.Hour,
		emailVerificationResendCooldown: time.Duration(5) * time.Minute,
		emailService:                    emailService,
		sessionTokenService:             sessionTokenService,
		userRepo:                        repositories.NewUserRepository(db),
		roleRepo:                        repositories.NewRoleRepository(db),
		refreshTokenRepo:                repositories.NewRefreshTokenRepository(db),
		passwordResetTokenRepo:          repositories.NewPasswordResetTokenRepository(db),
		emailVerificationTokenRepo:      repositories.NewEmailVerificationTokenRepository(db),
		eventRepo:                       repositories.NewEventRepository(db),
	}
}
