package roles

import (
	"database/sql"

	"auth/internal/repositories"
)

type RolesService struct {
	db       *sql.DB
	roleRepo repositories.RoleRepository
}

func NewRolesService(db *sql.DB) *RolesService {
	return &RolesService{
		db:       db,
		roleRepo: repositories.NewRoleRepository(db),
	}
}
