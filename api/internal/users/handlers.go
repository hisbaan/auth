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
	ID            string    `json:"id"`
	Email         string    `json:"email"`
	Username      string    `json:"username"`
	EmailVerified bool      `json:"email_verified"`
	UpdatedAt     time.Time `json:"updated_at"`
	CreatedAt     time.Time `json:"created_at"`
}

func (s *UsersService) GetUser(userID ulid.ULID) (GetUserResponse, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return GetUserResponse{}, err
	}

	return GetUserResponse{
		ID:            ulidutil.ToPrefixed("user", userID),
		Email:         user.Email,
		Username:      user.Username,
		EmailVerified: user.EmailVerified,
		CreatedAt:     user.CreatedAt,
		UpdatedAt:     user.UpdatedAt,
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

	existing, err := s.userRepo.GetByID(userID)
	if err != nil {
		return err
	}
	if existing.Email != params.Email {
		user.EmailVerified = false

		userID := ulidutil.MustFromBytes(user.ID)
		s.emailVerificationTokenRepo.RevokeByUserID(userID)

		token, hashedToken := auth.GenerateResetToken()
		emailVerificationTokenModel := model.EmailVerificationTokens{
			ID:        ulid.Make().Bytes(),
			UserID:    user.ID,
			TokenHash: hashedToken,
			ExpiresAt: time.Now().Add(time.Duration(24) * time.Hour),
			RevokedAt: nil,
			CreatedAt: time.Now(),
		}
		s.emailVerificationTokenRepo.Create(emailVerificationTokenModel)
		urlEncodedToken := auth.URLEncodeToken(token)

		s.emailService.SendVerifyEmail(params.Email, params.Username, urlEncodedToken)
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

	err = s.refreshTokenRepo.RevokeByUserID(userID)
	if err != nil {
		return err
	}

	return nil
}

func (s *UsersService) DeleteUser(userID ulid.ULID) error {
	return s.userRepo.Delete(userID)
}
