package auth

import (
	"auth/internal/emails"
	"auth/internal/repositories"
	"crypto/ed25519"
	"database/sql"
	"time"
)

type AuthService struct {
	db                 *sql.DB
	jwtSigningKey      ed25519.PrivateKey
	jwtSigningKeyID    string
	issuer             string
	cookieDomain       string
	accessTokenExpiry  time.Duration
	refreshTokenExpiry time.Duration
	emailService       *emails.EmailService

	userRepo                   repositories.UserRepository
	roleRepo                   repositories.RoleRepository
	refreshTokenRepo           repositories.RefreshTokenRepository
	passwordResetTokenRepo     repositories.PasswordResetTokenRepository
	emailVerificationTokenRepo repositories.EmailVerificationTokenRepository
}

func NewAuthService(db *sql.DB, signingKey ed25519.PrivateKey, signingKeyID string, issuer string, emailService *emails.EmailService, cookieDomain string) *AuthService {
	return &AuthService{
		db:                         db,
		jwtSigningKey:              signingKey,
		jwtSigningKeyID:            signingKeyID,
		issuer:                     issuer,
		cookieDomain:               cookieDomain,
		accessTokenExpiry:          15 * time.Minute,
		refreshTokenExpiry:         168 * time.Hour, // 7 days
		emailService:               emailService,
		userRepo:                   repositories.NewUserRepository(db),
		roleRepo:                   repositories.NewRoleRepository(db),
		refreshTokenRepo:           repositories.NewRefreshTokenRepository(db),
		passwordResetTokenRepo:     repositories.NewPasswordResetTokenRepository(db),
		emailVerificationTokenRepo: repositories.NewEmailVerificationTokenRepository(db),
	}
}
