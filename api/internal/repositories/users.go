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

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return UserRepository{db: db}
}

func (r *UserRepository) GetByID(id ulid.ULID) (*model.Users, error) {
	query := Users.SELECT(Users.AllColumns).
		WHERE(Users.ID.EQ(Bytea(id.Bytes()))).
		LIMIT(1)

	var users []model.Users
	err := query.Query(r.db, &users)
	if err != nil {
		log.Printf("[ERROR] GetByID query failed: %v", err)
		return nil, apperror.NewInternalServerError("Database query error")
	}

	if len(users) == 0 {
		return nil, apperror.NewNotFound("User not found")
	}

	return &users[0], nil
}

func (r *UserRepository) GetByEmail(email string) (*model.Users, error) {
	query := Users.SELECT(Users.AllColumns).
		WHERE(Users.Email.EQ(String(email))).
		LIMIT(1)

	var users []model.Users
	err := query.Query(r.db, &users)
	if err != nil {
		log.Printf("[ERROR] GetByEmail query failed: %v", err)
		return nil, apperror.NewInternalServerError("Database query error")
	}

	if len(users) == 0 {
		return nil, nil
	}

	return &users[0], nil
}

func (r *UserRepository) Create(user model.Users) error {
	_, err := Users.INSERT().MODEL(user).ON_CONFLICT().DO_NOTHING().Exec(r.db)
	if err != nil {
		log.Printf("[ERROR] Create user failed: %v", err)
		return apperror.NewInternalServerError("Database query error")
	}
	return nil
}

func (r *UserRepository) Update(user model.Users) error {
	_, err := Users.UPDATE(Users.Email, Users.Username, Users.EmailVerified, Users.UpdatedAt).MODEL(user).WHERE(Users.ID.EQ(Bytea(user.ID))).Exec(r.db)
	if err != nil {
		log.Printf("[ERROR] Update user failed: %v", err)
		return apperror.NewInternalServerError("Database query error")
	}
	return nil
}

func (r *UserRepository) SetPassword(id ulid.ULID, passwordHash string) error {
	_, err := Users.UPDATE(Users.PasswordHash).SET(Users.PasswordHash.SET(String(passwordHash))).WHERE(Users.ID.EQ(Bytea(id.Bytes()))).Exec(r.db)
	if err != nil {
		log.Printf("[ERROR] Update password failed: %v", err)
		return apperror.NewInternalServerError("Database query error")
	}
	return nil
}

func (r *UserRepository) SetEmailVerified(id ulid.ULID) error {
	_, err := Users.UPDATE(Users.EmailVerified).
		SET(Users.EmailVerified.SET(Bool(true)), Users.UpdatedAt.SET(TimestampzT(time.Now()))).
		WHERE(Users.ID.EQ(Bytea(id.Bytes()))).
		Exec(r.db)
	if err != nil {
		log.Printf("[ERROR] Set email verified failed: %v", err)
		return apperror.NewInternalServerError("Database query error")
	}
	return nil
}

func (r *UserRepository) Delete(id ulid.ULID) error {
	_, err := Users.DELETE().WHERE(Users.ID.EQ(Bytea(id.Bytes()))).Exec(r.db)
	if err != nil {
		log.Printf("[ERROR] Delete user failed: %v", err)
		return apperror.NewInternalServerError("Database query error")
	}
	return nil
}

func (r *UserRepository) WillConflict(user model.Users) (bool, error) {
	query := Users.SELECT(Users.ID).
		WHERE(
			AND(
				OR(
					Users.Email.EQ(String(user.Email)),
					Users.Username.EQ(String(user.Username)),
				),
				Users.ID.NOT_EQ(Bytea(user.ID)),
			),
		).
		LIMIT(1)

	var users []model.Users
	err := query.Query(r.db, &users)
	if err != nil {
		log.Printf("[ERROR] Will conflict query failed: %v", err)
		return false, apperror.NewInternalServerError("Database query error")
	}

	return len(users) > 0, nil
}
