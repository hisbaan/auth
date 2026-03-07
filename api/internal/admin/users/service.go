package users

import (
	"auth/internal/repositories"
	"database/sql"
)

type AdminUsersService struct {
	userRepo  repositories.UserRepository
	roleRepo  repositories.RoleRepository
	eventRepo repositories.EventRepository
}

func NewAdminUsersService(db *sql.DB) *AdminUsersService {
	return &AdminUsersService{
		userRepo:  repositories.NewUserRepository(db),
		roleRepo:  repositories.NewRoleRepository(db),
		eventRepo: repositories.NewEventRepository(db),
	}
}
