package repositories

import (
	"auth/internal/apperror"
	"auth/internal/jet/postgres/public/model"
	. "auth/internal/jet/postgres/public/table"
	"database/sql"

	. "github.com/go-jet/jet/v2/postgres"
)

type UserRepository interface {
	GetByEmail(email string) (*model.Users, error)
	Create(user model.Users) error
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) GetByEmail(email string) (*model.Users, error) {
	query := Users.SELECT(Users.AllColumns).
		WHERE(Users.Email.EQ(String(email))).
		LIMIT(1)

	var users []model.Users
	err := query.Query(r.db, &users)
	if err != nil {
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
		return apperror.NewInternalServerError("Database query error")
	}
	return nil
}
