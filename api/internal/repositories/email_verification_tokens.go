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

func (r *EmailVerificationTokenRepository) GetLatestActiveByUserID(userID ulid.ULID) (*model.EmailVerificationTokens, error) {
	query := EmailVerificationTokens.SELECT(EmailVerificationTokens.AllColumns).
		WHERE(
			AND(
				EmailVerificationTokens.UserID.EQ(Bytea(userID.Bytes())),
				EmailVerificationTokens.RevokedAt.IS_NULL(),
				EmailVerificationTokens.ExpiresAt.GT(TimestampzT(time.Now())),
			),
		).
		ORDER_BY(EmailVerificationTokens.CreatedAt.DESC()).
		LIMIT(1)

	var tokens []model.EmailVerificationTokens
	err := query.Query(r.db, &tokens)
	if err != nil {
		log.Printf("[ERROR] GetLatestActiveByUserID query failed: %v", err)
		return nil, apperror.NewInternalServerError("Database query error")
	}

	if len(tokens) == 0 {
		return nil, nil
	}

	return &tokens[0], nil
}

func (r *EmailVerificationTokenRepository) ActiveEmailVerificationWillConflict(userID ulid.ULID, email string) (bool, error) {
	query := EmailVerificationTokens.SELECT(EmailVerificationTokens.ID).
		WHERE(
			AND(
				EmailVerificationTokens.Email.EQ(String(email)),
				EmailVerificationTokens.UserID.NOT_EQ(Bytea(userID.Bytes())),
				EmailVerificationTokens.RevokedAt.IS_NULL(),
				EmailVerificationTokens.ExpiresAt.GT(TimestampzT(time.Now())),
			),
		).
		LIMIT(1)

	var tokens []model.EmailVerificationTokens
	err := query.Query(r.db, &tokens)
	if err != nil {
		log.Printf("[ERROR] Active email verification conflict query failed: %v", err)
		return false, apperror.NewInternalServerError("Database query error")
	}

	return len(tokens) > 0, nil
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
