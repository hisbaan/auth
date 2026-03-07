package events

import (
	"auth/internal/apperror"
	"auth/internal/jet/postgres/public/model"
	"auth/internal/utils/ulidutil"
	"time"

	"github.com/oklog/ulid/v2"
)

type ListEventsParams struct {
	Cursor string
	Limit  int
}

type EventResponse struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Type      string    `json:"type"`
	Data      string    `json:"data"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
	CreatedAt time.Time `json:"created_at"`
}

type ListEventsResponse struct {
	Events     []EventResponse `json:"events"`
	NextCursor string          `json:"next_cursor"`
}

func (s *AdminEventsService) ListEvents(params ListEventsParams) (ListEventsResponse, error) {
	if params.Limit == 0 {
		params.Limit = 20
	}
	if params.Limit > 100 {
		params.Limit = 100
	}

	var cursor *ulid.ULID
	if params.Cursor != "" {
		c, err := ulidutil.FromPrefixed("event", params.Cursor)
		if err != nil {
			return ListEventsResponse{}, apperror.NewBadRequest("Invalid cursor")
		}
		cursor = &c
	}

	events, err := s.eventRepo.List(params.Limit+1, cursor)
	if err != nil {
		return ListEventsResponse{}, err
	}

	hasMore := len(events) > params.Limit
	if hasMore {
		events = events[:params.Limit]
	}

	result := make([]EventResponse, len(events))
	for i, event := range events {
		result[i] = mapEvent(event)
	}

	var nextCursor string
	if hasMore && len(events) > 0 {
		lastEvent := events[len(events)-1]
		lastID, err := ulidutil.FromBytes(lastEvent.ID)
		if err != nil {
			return ListEventsResponse{}, apperror.NewInternalServerError("Invalid event ID")
		}
		nextCursor = ulidutil.ToPrefixed("event", lastID)
	}

	return ListEventsResponse{Events: result, NextCursor: nextCursor}, nil
}

func (s *AdminEventsService) ListUserEvents(userID ulid.ULID, params ListEventsParams) (ListEventsResponse, error) {
	if params.Limit == 0 {
		params.Limit = 20
	}
	if params.Limit > 100 {
		params.Limit = 100
	}

	var cursor *ulid.ULID
	if params.Cursor != "" {
		c, err := ulidutil.FromPrefixed("event", params.Cursor)
		if err != nil {
			return ListEventsResponse{}, apperror.NewBadRequest("Invalid cursor")
		}
		cursor = &c
	}

	events, err := s.eventRepo.ListByUserID(userID, params.Limit+1, cursor)
	if err != nil {
		return ListEventsResponse{}, err
	}

	hasMore := len(events) > params.Limit
	if hasMore {
		events = events[:params.Limit]
	}

	result := make([]EventResponse, len(events))
	for i, event := range events {
		result[i] = mapEvent(event)
	}

	var nextCursor string
	if hasMore && len(events) > 0 {
		lastEvent := events[len(events)-1]
		lastID, err := ulidutil.FromBytes(lastEvent.ID)
		if err != nil {
			return ListEventsResponse{}, apperror.NewInternalServerError("Invalid event ID")
		}
		nextCursor = ulidutil.ToPrefixed("event", lastID)
	}

	return ListEventsResponse{Events: result, NextCursor: nextCursor}, nil
}

func (s *AdminEventsService) GetEvent(eventID ulid.ULID) (EventResponse, error) {
	event, err := s.eventRepo.GetByID(eventID)
	if err != nil {
		return EventResponse{}, err
	}
	if event == nil {
		return EventResponse{}, apperror.NewNotFound("Event not found")
	}

	return mapEvent(*event), nil
}

func mapEvent(event model.Events) EventResponse {
	return EventResponse{
		ID:        ulidutil.ToPrefixed("event", ulidutil.MustFromBytes(event.ID)),
		UserID:    ulidutil.ToPrefixed("user", ulidutil.MustFromBytes(*event.UserID)),
		Type:      event.Type,
		Data:      event.Data,
		IPAddress: event.IPAddress,
		UserAgent: event.UserAgent,
		CreatedAt: event.CreatedAt,
	}
}
