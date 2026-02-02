package auth

import (
	"database/sql"

	"auth/jet/postgres/public/model"
	. "auth/jet/postgres/public/table"
	"auth/passwords"

	"github.com/oklog/ulid/v2"
)

type CreateUserParams struct {
	Username string
	Email    string
	Password string
}

func CreateUser(db *sql.DB, params CreateUserParams) error {
	hash, err := passwords.Hash(params.Password)
	if err != nil {
		return err
	}

	user := model.Users{
		ID:           ulid.Make().Bytes(),
		Username:     params.Username,
		Email:        params.Email,
		PasswordHash: hash,
	}

	_, err = Users.INSERT().MODEL(user).ON_CONFLICT().DO_NOTHING().Exec(db)
	if err != nil {
		return err
	}

	return nil
}
