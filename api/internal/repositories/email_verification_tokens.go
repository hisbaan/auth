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

type EmailVerificationTokenRepository struct {
	db *sql.DB
}

func NewEmailVerificationTokenRepository(db *sql.DB) EmailVerificationTokenRepository {
	return EmailVerificationTokenRepository{db: db}
}

func (r *EmailVerificationTokenRepository) Create(token model.EmailVerificationTokens) error {
	_, err := EmailVerificationTokens.INSERT().MODEL(token).Exec(r.db)
	if err != nil {
		log.Printf("[ERROR] Create email verification token failed: %v", err)
		return apperror.NewInternalServerError("Database query error")
	}
	return nil
}

func (r *EmailVerificationTokenRepository) GetByHash(hash []byte) (*model.EmailVerificationTokens, error) {
	query := EmailVerificationTokens.SELECT(EmailVerificationTokens.AllColumns).
		WHERE(EmailVerificationTokens.TokenHash.EQ(Bytea(hash))).
		LIMIT(1)

	var tokens []model.EmailVerificationTokens
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

func (r *EmailVerificationTokenRepository) Revoke(id ulid.ULID) error {
	_, err := EmailVerificationTokens.UPDATE().
		SET(EmailVerificationTokens.RevokedAt.SET(TimestampzT(time.Now()))).
		WHERE(AND(EmailVerificationTokens.ID.EQ(Bytea(id.Bytes())), EmailVerificationTokens.RevokedAt.IS_NULL())).
		Exec(r.db)
	if err != nil {
		log.Printf("[ERROR] Revoke email verification token failed: %v", err)
		return apperror.NewInternalServerError("Database query error")
	}
	return nil
}

func (r *EmailVerificationTokenRepository) RevokeByUserID(userID ulid.ULID) error {
	_, err := EmailVerificationTokens.UPDATE().
		SET(EmailVerificationTokens.RevokedAt.SET(TimestampzT(time.Now()))).
		WHERE(AND(EmailVerificationTokens.UserID.EQ(Bytea(userID.Bytes())), EmailVerificationTokens.RevokedAt.IS_NULL())).
		Exec(r.db)
	if err != nil {
		log.Printf("[ERROR] Revoke email verification tokens by userID failed: %v", err)
		return apperror.NewInternalServerError("Database query error")
	}
	return nil
}
