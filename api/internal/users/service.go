package users

import (
	"auth/internal/emails"
	"auth/internal/repositories"
	"crypto/ed25519"
	"database/sql"
)

type UsersService struct {
	db                  *sql.DB
	jwtSigningKey       ed25519.PrivateKey
	issuer              string
	emailService        *emails.EmailService
	blockedEmailDomains []string

	userRepo                    repositories.UserRepository
	clientRepo                  repositories.ClientRepository
	userClientAuthorizationRepo repositories.UserClientAuthorizationRepository
	roleRepo                    repositories.RoleRepository
	refreshTokenRepo            repositories.RefreshTokenRepository
	emailVerificationTokenRepo  repositories.EmailVerificationTokenRepository
	eventRepo                   repositories.EventRepository
}

func NewUsersService(db *sql.DB, jwtSigningKey ed25519.PrivateKey, issuer string, emailService *emails.EmailService, blockedEmailDomains []string, onClientChange func()) *UsersService {
	return &UsersService{
		db:                          db,
		jwtSigningKey:               jwtSigningKey,
		issuer:                      issuer,
		emailService:                emailService,
		blockedEmailDomains:         blockedEmailDomains,
		userRepo:                    repositories.NewUserRepository(db),
		clientRepo:                  repositories.NewClientRepositoryWithChangeHook(db, onClientChange),
		userClientAuthorizationRepo: repositories.NewUserClientAuthorizationRepository(db),
		roleRepo:                    repositories.NewRoleRepository(db),
		refreshTokenRepo:            repositories.NewRefreshTokenRepository(db),
		emailVerificationTokenRepo:  repositories.NewEmailVerificationTokenRepository(db),
		eventRepo:                   repositories.NewEventRepository(db),
	}
}
