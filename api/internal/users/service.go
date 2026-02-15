package users

import (
	"auth/internal/repositories"
	"crypto/ed25519"
	"database/sql"
)

type UsersService struct {
	db            *sql.DB
	jwtAccessKey  ed25519.PrivateKey
	jwtRefreshKey ed25519.PrivateKey
	issuer        string

	userRepo         repositories.UserRepository
	refreshRokenRepo repositories.RefreshTokenRepository
}

func NewUsersService(db *sql.DB, jwtAccessKey ed25519.PrivateKey, jwtRefreshKey ed25519.PrivateKey, issuer string) (*UsersService, error) {
	return &UsersService{
		db:               db,
		jwtAccessKey:     jwtAccessKey,
		jwtRefreshKey:    jwtRefreshKey,
		issuer:           issuer,
		userRepo:         repositories.NewUserRepository(db),
		refreshRokenRepo: repositories.NewRefreshTokenRepository(db),
	}, nil
}
