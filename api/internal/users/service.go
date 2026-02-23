package users

import (
	"auth/internal/emails"
	"auth/internal/repositories"
	"crypto/ed25519"
	"database/sql"
)

type UsersService struct {
	db            *sql.DB
	jwtAccessKey  ed25519.PrivateKey
	jwtRefreshKey ed25519.PrivateKey
	issuer        string
	emailService  *emails.EmailService

	userRepo                   repositories.UserRepository
	refreshTokenRepo           repositories.RefreshTokenRepository
	emailVerificationTokenRepo repositories.EmailVerificationTokenRepository
}

func NewUsersService(db *sql.DB, jwtAccessKey ed25519.PrivateKey, jwtRefreshKey ed25519.PrivateKey, issuer string, emailService *emails.EmailService) (*UsersService, error) {
	return &UsersService{
		db:                         db,
		jwtAccessKey:               jwtAccessKey,
		jwtRefreshKey:              jwtRefreshKey,
		issuer:                     issuer,
		emailService:               emailService,
		userRepo:                   repositories.NewUserRepository(db),
		refreshTokenRepo:           repositories.NewRefreshTokenRepository(db),
		emailVerificationTokenRepo: repositories.NewEmailVerificationTokenRepository(db),
	}, nil
}
