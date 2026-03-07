package roles

import (
	"auth/internal/repositories"
	"database/sql"
)

type AdminRolesService struct {
	roleRepo  repositories.RoleRepository
	eventRepo repositories.EventRepository
}

func NewAdminRolesService(db *sql.DB) *AdminRolesService {
	return &AdminRolesService{
		roleRepo:  repositories.NewRoleRepository(db),
		eventRepo: repositories.NewEventRepository(db),
	}
}
