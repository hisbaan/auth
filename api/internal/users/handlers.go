package users

import (
	"auth/internal/apperror"
	"auth/internal/auth"
	"auth/internal/jet/postgres/public/model"
	"auth/internal/ulidutil"
	"time"

	"github.com/oklog/ulid/v2"
)

type GetUserParams struct {
	userID ulid.ULID
}

type GetUserResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	UpdatedAt time.Time `json:"updated_at"`
	CreatedAt time.Time `json:"created_at"`
}

func (s *UsersService) GetUser(userID ulid.ULID) (GetUserResponse, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return GetUserResponse{}, err
	}

	return GetUserResponse{
		ID:        ulidutil.ToPrefixed("user", userID),
		Email:     user.Email,
		Username:  user.Username,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

type UpdateUserParams struct {
	Email    string `json:"email"`
	Username string `json:"username"`
}

func (s *UsersService) UpdateUser(userID ulid.ULID, params UpdateUserParams) error {
	user := model.Users{
		ID:       userID.Bytes(),
		Email:    params.Email,
		Username: params.Username,
	}

	willConflict, err := s.userRepo.WillConflict(user)
	if err != nil {
		return err
	}
	if willConflict {
		return apperror.NewConflict("Username or email already in use")
	}

	return s.userRepo.Update(user)
}

type UpdatePasswordParams struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

func (s *UsersService) UpdatePassword(userID ulid.ULID, params UpdatePasswordParams) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return err
	}

	if !auth.ComparePasswordAndHash(params.CurrentPassword, user.PasswordHash) {
		return apperror.NewUnauthorized("Unauthorized")
	}

	passwordHash, err := auth.HashPassword(params.NewPassword)
	if err != nil {
		return err
	}

	err = s.userRepo.SetPassword(userID, passwordHash)
	if err != nil {
		return err
	}

	err = s.refreshRokenRepo.RevokeByUserID(userID)
	if err != nil {
		return err
	}

	return nil
}

func (s *UsersService) DeleteUser(userID ulid.ULID) error {
	return s.userRepo.Delete(userID)
}
