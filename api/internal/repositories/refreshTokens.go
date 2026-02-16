package repositories

import (
	"auth/internal/apperror"
	"auth/internal/jet/postgres/public/model"
	. "auth/internal/jet/postgres/public/table"
	"database/sql"
	"log"
	"time"

	. "github.com/go-jet/jet/v2/postgres"
	"github.com/oklog/ulid/v2"
)

type RefreshTokenRepository struct {
	db *sql.DB
}

func NewRefreshTokenRepository(db *sql.DB) RefreshTokenRepository {
	return RefreshTokenRepository{db: db}
}

func (r *RefreshTokenRepository) GetByID(id ulid.ULID) (*model.RefreshTokens, error) {
	query := RefreshTokens.SELECT(RefreshTokens.AllColumns).
		WHERE(RefreshTokens.ID.EQ(Bytea(id.Bytes()))).
		LIMIT(1)

	var tokens []model.RefreshTokens
	err := query.Query(r.db, &tokens)
	if err != nil {
		log.Printf("[ERROR] GetByID query failed: %v", err)
		return nil, apperror.NewInternalServerError("Database query error")
	}

	if len(tokens) == 0 {
		return nil, nil
	}

	return &tokens[0], nil
}

func (r *RefreshTokenRepository) Revoke(id ulid.ULID) error {
	_, err := RefreshTokens.UPDATE().
		SET(RefreshTokens.RevokedAt.SET(TimestampzT(time.Now()))).
		WHERE(AND(RefreshTokens.ID.EQ(Bytea(id.Bytes())), RefreshTokens.RevokedAt.IS_NULL())).
		Exec(r.db)
	if err != nil {
		log.Printf("[ERROR] Revoke token failed: %v", err)
		return apperror.NewInternalServerError("Database query error")
	}
	return nil
}

func (r *RefreshTokenRepository) RevokeByUserID(userID ulid.ULID) error {
	_, err := RefreshTokens.UPDATE().
		SET(RefreshTokens.RevokedAt.SET(TimestampzT(time.Now()))).
		WHERE(AND(RefreshTokens.UserID.EQ(Bytea(userID.Bytes())), RefreshTokens.RevokedAt.IS_NULL())).
		Exec(r.db)
	if err != nil {
		log.Printf("[ERROR] Revoke tokens by userID failed: %v", err)
		return apperror.NewInternalServerError("Database query error")
	}
	return nil
}

func (r *RefreshTokenRepository) Create(token model.RefreshTokens) error {
	_, err := RefreshTokens.INSERT().MODEL(token).Exec(r.db)
	if err != nil {
		log.Printf("[ERROR] Create refresh token failed: %v", err)
		return apperror.NewInternalServerError("Database query error")
	}
	return nil
}
