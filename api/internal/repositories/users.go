package repositories

import (
	"auth/internal/apperror"
	"auth/internal/jet/postgres/public/model"
	. "auth/internal/jet/postgres/public/table"
	"database/sql"
	"log"

	. "github.com/go-jet/jet/v2/postgres"
	"github.com/oklog/ulid/v2"
)

type UserRepository interface {
	GetByID(id ulid.ULID) (*model.Users, error)
	GetByEmail(email string) (*model.Users, error)
	Create(user model.Users) error
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) GetByID(id ulid.ULID) (*model.Users, error) {
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
		return nil, nil
	}

	return &users[0], nil
}

func (r *userRepository) GetByEmail(email string) (*model.Users, error) {
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

func (r *userRepository) Create(user model.Users) error {
	_, err := Users.INSERT().MODEL(user).ON_CONFLICT().DO_NOTHING().Exec(r.db)
	if err != nil {
		log.Printf("[ERROR] Create user failed: %v", err)
		return apperror.NewInternalServerError("Database query error")
	}
	return nil
}
