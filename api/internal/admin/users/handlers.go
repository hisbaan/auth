package users

import (
	"auth/internal/apperror"
	"auth/internal/utils/ulidutil"
	"time"

	"github.com/oklog/ulid/v2"
)

type ListUsersParams struct {
	Cursor string
	Limit  int
}

type ListUsersResponse struct {
	Users      []UserWithMetadata `json:"users"`
	NextCursor string             `json:"next_cursor"`
}

type UserWithMetadata struct {
	ID            string    `json:"id"`
	Email         string    `json:"email"`
	Username      string    `json:"username"`
	EmailVerified bool      `json:"email_verified"`
	Roles         []string  `json:"roles"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (s *AdminUsersService) ListUsers(params ListUsersParams) (ListUsersResponse, error) {
	if params.Limit == 0 {
		params.Limit = 20
	}
	if params.Limit > 100 {
		params.Limit = 100
	}

	var cursor *ulid.ULID
	if params.Cursor != "" {
		c, err := ulidutil.FromPrefixed("user", params.Cursor)
		if err != nil {
			return ListUsersResponse{}, apperror.NewBadRequest("Invalid cursor")
		}
		cursor = &c
	}

	users, err := s.userRepo.ListWithRoles(params.Limit+1, cursor)
	if err != nil {
		return ListUsersResponse{}, err
	}

	hasMore := len(users) > params.Limit
	if hasMore {
		users = users[:params.Limit]
	}

	result := make([]UserWithMetadata, len(users))
	for i, user := range users {
		userID := ulidutil.MustFromBytes(user.ID)

		result[i] = UserWithMetadata{
			ID:            ulidutil.ToPrefixed("user", userID),
			Email:         user.Email,
			Username:      user.Username,
			EmailVerified: user.EmailVerified,
			Roles:         user.Roles,
			CreatedAt:     user.CreatedAt,
			UpdatedAt:     user.UpdatedAt,
		}
	}

	var nextCursor string
	if hasMore && len(users) > 0 {
		lastUser := users[len(users)-1]
		nextCursor = ulidutil.ToPrefixed("user", ulidutil.MustFromBytes(lastUser.ID))
	}

	return ListUsersResponse{
		Users:      result,
		NextCursor: nextCursor,
	}, nil
}

type GetUserResponse struct {
	ID            string    `json:"id"`
	Email         string    `json:"email"`
	Username      string    `json:"username"`
	EmailVerified bool      `json:"email_verified"`
	Roles         []string  `json:"roles"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (s *AdminUsersService) GetUser(userID ulid.ULID) (GetUserResponse, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return GetUserResponse{}, err
	}

	roles, err := s.roleRepo.GetByUserID(userID)
	if err != nil {
		return GetUserResponse{}, err
	}

	roleNames := make([]string, len(roles))
	for i, role := range roles {
		roleNames[i] = role.Name
	}

	return GetUserResponse{
		ID:            ulidutil.ToPrefixed("user", userID),
		Email:         user.Email,
		Username:      user.Username,
		EmailVerified: user.EmailVerified,
		Roles:         roleNames,
		CreatedAt:     user.CreatedAt,
		UpdatedAt:     user.UpdatedAt,
	}, nil
}

type UpdateUserRoleParams struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
}

type UpdateUserRoleBody struct {
	Role string `json:"role"`
}

func (s *AdminUsersService) AddUserRole(params UpdateUserRoleParams) error {
	userID, err := ulidutil.FromPrefixed("user", params.UserID)
	if err != nil {
		return apperror.NewBadRequest("Invalid user ID")
	}
	return s.roleRepo.CreateUserRole(userID, params.Role)
}

func (s *AdminUsersService) RemoveUserRole(params UpdateUserRoleParams) error {
	userID, err := ulidutil.FromPrefixed("user", params.UserID)
	if err != nil {
		return apperror.NewBadRequest("Invalid user ID")
	}
	return s.roleRepo.DeleteUserRole(userID, params.Role)
}
