package refreshtokens

import (
	"auth/internal/repositories"
	"database/sql"
)

type AdminRefreshTokensService struct {
	refreshTokenRepo repositories.RefreshTokenRepository
}

func NewAdminRefreshTokensService(db *sql.DB) *AdminRefreshTokensService {
	return &AdminRefreshTokensService{
		refreshTokenRepo: repositories.NewRefreshTokenRepository(db),
	}
}
