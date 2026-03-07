package admin

import (
	"database/sql"

	adminevents "auth/internal/admin/events"
	adminrefreshtokens "auth/internal/admin/refresh_tokens"
	adminroles "auth/internal/admin/roles"
	adminusers "auth/internal/admin/users"
)

type AdminService struct {
	Users         *adminusers.AdminUsersService
	Roles         *adminroles.AdminRolesService
	Events        *adminevents.AdminEventsService
	RefreshTokens *adminrefreshtokens.AdminRefreshTokensService
}

func NewAdminService(db *sql.DB) *AdminService {
	return &AdminService{
		Users:         adminusers.NewAdminUsersService(db),
		Roles:         adminroles.NewAdminRolesService(db),
		Events:        adminevents.NewAdminEventsService(db),
		RefreshTokens: adminrefreshtokens.NewAdminRefreshTokensService(db),
	}
}
