package admin

import (
	"database/sql"

	"auth/internal/repositories"
)

type AdminService struct {
	db *sql.DB

	userRepo repositories.UserRepository
	roleRepo repositories.RoleRepository
}

func NewAdminService(db *sql.DB) *AdminService {
	return &AdminService{
		db:       db,
		userRepo: repositories.NewUserRepository(db),
		roleRepo: repositories.NewRoleRepository(db),
	}
}
