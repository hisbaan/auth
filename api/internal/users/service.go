package users

import (
	"auth/internal/emails"
	"auth/internal/repositories"
	"crypto/ed25519"
	"database/sql"
)

type UsersService struct {
	db            *sql.DB
	jwtSigningKey ed25519.PrivateKey
	issuer        string
	emailService  *emails.EmailService

	userRepo                   repositories.UserRepository
	roleRepo                   repositories.RoleRepository
	refreshTokenRepo           repositories.RefreshTokenRepository
	emailVerificationTokenRepo repositories.EmailVerificationTokenRepository
}

func NewUsersService(db *sql.DB, jwtSigningKey ed25519.PrivateKey, issuer string, emailService *emails.EmailService) *UsersService {
	return &UsersService{
		db:                         db,
		jwtSigningKey:              jwtSigningKey,
		issuer:                     issuer,
		emailService:               emailService,
		userRepo:                   repositories.NewUserRepository(db),
		roleRepo:                   repositories.NewRoleRepository(db),
		refreshTokenRepo:           repositories.NewRefreshTokenRepository(db),
		emailVerificationTokenRepo: repositories.NewEmailVerificationTokenRepository(db),
	}
}
