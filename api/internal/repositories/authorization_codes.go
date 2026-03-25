package repositories

import (
	"database/sql"
	"log"
	"time"

	"auth/internal/apperror"
	"auth/internal/jet/postgres/public/model"
	. "auth/internal/jet/postgres/public/table"

	. "github.com/go-jet/jet/v2/postgres"
	"github.com/oklog/ulid/v2"
)

type AuthorizationCodeRepository struct {
	db *sql.DB
}

func NewAuthorizationCodeRepository(db *sql.DB) AuthorizationCodeRepository {
	return AuthorizationCodeRepository{db: db}
}

func (r *AuthorizationCodeRepository) GetByID(id ulid.ULID) (*model.AuthorizationCodes, error) {
	query := AuthorizationCodes.SELECT(AuthorizationCodes.AllColumns).
		WHERE(AuthorizationCodes.ID.EQ(Bytea(id.Bytes()))).
		LIMIT(1)

	var codes []model.AuthorizationCodes
	err := query.Query(r.db, &codes)
	if err != nil {
		log.Printf("[ERROR] GetByID authorization code query failed: %v", err)
		return nil, apperror.NewInternalServerError("Database query error")
	}

	if len(codes) == 0 {
		return nil, nil
	}

	return &codes[0], nil
}

func (r *AuthorizationCodeRepository) GetByCodeHash(codeHash []byte) (*model.AuthorizationCodes, error) {
	query := AuthorizationCodes.SELECT(AuthorizationCodes.AllColumns).
		WHERE(AuthorizationCodes.CodeHash.EQ(Bytea(codeHash))).
		LIMIT(1)

	var codes []model.AuthorizationCodes
	err := query.Query(r.db, &codes)
	if err != nil {
		log.Printf("[ERROR] GetByCodeHash authorization code query failed: %v", err)
		return nil, apperror.NewInternalServerError("Database query error")
	}

	if len(codes) == 0 {
		return nil, nil
	}

	return &codes[0], nil
}

func (r *AuthorizationCodeRepository) Create(code model.AuthorizationCodes) error {
	if code.CreatedAt.IsZero() {
		code.CreatedAt = time.Now()
	}

	_, err := AuthorizationCodes.INSERT().MODEL(code).Exec(r.db)
	if err != nil {
		log.Printf("[ERROR] Create authorization code failed: %v", err)
		return apperror.FromPGError(err)
	}

	return nil
}

func (r *AuthorizationCodeRepository) MarkUsed(id ulid.ULID) error {
	result, err := AuthorizationCodes.UPDATE().
		SET(AuthorizationCodes.UsedAt.SET(TimestampzT(time.Now()))).
		WHERE(AND(AuthorizationCodes.ID.EQ(Bytea(id.Bytes())), AuthorizationCodes.UsedAt.IS_NULL())).
		Exec(r.db)
	if err != nil {
		log.Printf("[ERROR] MarkUsed authorization code failed: %v", err)
		return apperror.NewInternalServerError("Database query error")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("[ERROR] MarkUsed authorization code rows affected failed: %v", err)
		return apperror.NewInternalServerError("Database query error")
	}
	if rowsAffected == 0 {
		return apperror.NewNotFound("Authorization code not found")
	}

	return nil
}
