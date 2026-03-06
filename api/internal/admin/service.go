package admin

import (
	adminevents "auth/internal/admin/events"
	adminroles "auth/internal/admin/roles"
	adminusers "auth/internal/admin/users"
	"database/sql"
)

type AdminService struct {
	Users  *adminusers.AdminUsersService
	Roles  *adminroles.AdminRolesService
	Events *adminevents.AdminEventsService
}

func NewAdminService(db *sql.DB) *AdminService {
	return &AdminService{
		Users:  adminusers.NewAdminUsersService(db),
		Roles:  adminroles.NewAdminRolesService(db),
		Events: adminevents.NewAdminEventsService(db),
	}
}
