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

type PasswordResetTokenRepository struct {
	db *sql.DB
}

func NewPasswordResetTokenRepository(db *sql.DB) PasswordResetTokenRepository {
	return PasswordResetTokenRepository{db: db}
}

func (r *PasswordResetTokenRepository) Create(token model.PasswordResetTokens) error {
	_, err := PasswordResetTokens.INSERT().MODEL(token).Exec(r.db)
	if err != nil {
		log.Printf("[ERROR] Create password reset token failed: %v", err)
		return apperror.NewInternalServerError("Database query error")
	}
	return nil
}

func (r *PasswordResetTokenRepository) GetByHash(hash []byte) (*model.PasswordResetTokens, error) {
	query := PasswordResetTokens.SELECT(PasswordResetTokens.AllColumns).
		WHERE(PasswordResetTokens.TokenHash.EQ(Bytea(hash))).
		LIMIT(1)

	var tokens []model.PasswordResetTokens
	err := query.Query(r.db, &tokens)
	if err != nil {
		log.Printf("[ERROR] GetByHash query failed: %v", err)
		return nil, apperror.NewInternalServerError("Database query error")
	}

	if len(tokens) == 0 {
		return nil, apperror.NewNotFound("Token not found")
	}

	return &tokens[0], nil
}

func (r *PasswordResetTokenRepository) Revoke(id ulid.ULID) error {
	_, err := PasswordResetTokens.UPDATE().
		SET(PasswordResetTokens.RevokedAt.SET(TimestampzT(time.Now()))).
		WHERE(AND(PasswordResetTokens.ID.EQ(Bytea(id.Bytes())), PasswordResetTokens.RevokedAt.IS_NULL())).
		Exec(r.db)
	if err != nil {
		log.Printf("[ERROR] Revoke password reset token failed: %v", err)
		return apperror.NewInternalServerError("Database query error")
	}
	return nil
}

func (r *PasswordResetTokenRepository) RevokeByUserID(userID ulid.ULID) error {
	_, err := PasswordResetTokens.UPDATE().
		SET(PasswordResetTokens.RevokedAt.SET(TimestampzT(time.Now()))).
		WHERE(AND(PasswordResetTokens.UserID.EQ(Bytea(userID.Bytes())), PasswordResetTokens.RevokedAt.IS_NULL())).
		Exec(r.db)
	if err != nil {
		log.Printf("[ERROR] Revoke password reset tokens by userID failed: %v", err)
		return apperror.NewInternalServerError("Database query error")
	}
	return nil
}
