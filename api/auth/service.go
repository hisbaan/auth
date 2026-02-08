package auth

import (
	"auth/repositories"
	"crypto/ed25519"
	"database/sql"
	"time"
)

type AuthService struct {
	db                 *sql.DB
	jwtAccessKey       ed25519.PrivateKey
	jwtRefreshKey      ed25519.PrivateKey
	issuer             string
	accessTokenExpiry  time.Duration
	refreshTokenExpiry time.Duration

	userRepo         repositories.UserRepository
	refreshTokenRepo repositories.RefreshTokenRepository
}

func NewAuthService(db *sql.DB, accessKey ed25519.PrivateKey, refreshKey ed25519.PrivateKey, issuer string) (*AuthService, error) {
	return &AuthService{
		db:                 db,
		jwtAccessKey:       accessKey,
		jwtRefreshKey:      refreshKey,
		issuer:             issuer,
		accessTokenExpiry:  15 * time.Minute,
		refreshTokenExpiry: 168 * time.Hour, // 7 days
		userRepo:           repositories.NewUserRepository(db),
		refreshTokenRepo:   repositories.NewRefreshTokenRepository(db),
	}, nil
}
