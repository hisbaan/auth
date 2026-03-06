package events

import (
	"auth/internal/repositories"
	"database/sql"
)

type AdminEventsService struct {
	eventRepo repositories.EventRepository
}

func NewAdminEventsService(db *sql.DB) *AdminEventsService {
	return &AdminEventsService{
		eventRepo: repositories.NewEventRepository(db),
	}
}
