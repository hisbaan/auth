package users

import (
	"auth/internal/apperror"
	"auth/internal/auth"
	"auth/internal/events"
	"auth/internal/jet/postgres/public/model"
	"auth/internal/utils"
	"auth/internal/utils/ulidutil"
	"context"
	"time"

	"github.com/oklog/ulid/v2"
)

type GetUserResponse struct {
	ID            string    `json:"id"`
	Email         string    `json:"email"`
	Username      string    `json:"username"`
	EmailVerified bool      `json:"email_verified"`
	Roles         []string  `json:"roles"`
	UpdatedAt     time.Time `json:"updated_at"`
	CreatedAt     time.Time `json:"created_at"`
}

func (s *UsersService) GetUser(userID ulid.ULID) (GetUserResponse, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return GetUserResponse{}, err
	}

	roles, err := s.roleRepo.GetByUserID(userID)
	if err != nil {
		return GetUserResponse{}, err
	}

	return GetUserResponse{
		ID:            ulidutil.ToPrefixed("user", userID),
		Email:         user.Email,
		Username:      user.Username,
		EmailVerified: user.EmailVerified,
		Roles:         utils.Map(roles, func(role model.Roles) string { return role.Name }),
		CreatedAt:     user.CreatedAt,
		UpdatedAt:     user.UpdatedAt,
	}, nil
}

type UpdateUserParams struct {
	Email    string `json:"email"`
	Username string `json:"username"`
}

func (s *UsersService) UpdateUser(ctx context.Context, userID ulid.ULID, params UpdateUserParams) error {
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
		events.Log(ctx, &s.eventRepo, events.UserEmailVerificationRevoked, &userID, events.UserEmailVerificationRevokedData{
			Email: existing.Email,
		})
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
		events.Log(ctx, &s.eventRepo, events.UserEmailVerificationCreated, &userID, events.UserEmailVerificationCreatedData{
			Email: params.Email,
		})
		urlEncodedToken := auth.URLEncodeToken(token)

		s.emailService.SendVerifyEmail(params.Email, params.Username, urlEncodedToken)
	}

	if err := s.userRepo.Update(user); err != nil {
		return err
	}

	events.Log(ctx, &s.eventRepo, events.UserUpdated, &userID, events.UserUpdatedData{})

	return nil
}

type UpdatePasswordParams struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

func (s *UsersService) UpdatePassword(ctx context.Context, userID ulid.ULID, params UpdatePasswordParams) error {
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

	events.Log(ctx, &s.eventRepo, events.UserPasswordChanged, &userID, events.UserPasswordChangedData{})

	return nil
}

func (s *UsersService) DeleteUser(ctx context.Context, userID ulid.ULID) error {
	events.Log(ctx, &s.eventRepo, events.UserDeleted, &userID, events.UserDeletedData{})

	return s.userRepo.Delete(userID)
}
