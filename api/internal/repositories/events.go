package repositories

import (
	"auth/internal/apperror"
	"auth/internal/jet/postgres/public/model"
	. "auth/internal/jet/postgres/public/table"
	"database/sql"
	"log"
	"time"

	. "github.com/go-jet/jet/v2/postgres"
	"github.com/oklog/ulid/v2"
)

type EventRepository struct {
	db *sql.DB
}

func NewEventRepository(db *sql.DB) EventRepository {
	return EventRepository{db: db}
}

func (r *EventRepository) Create(event model.Events) error {
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now()
	}
	_, err := Events.INSERT().MODEL(event).Exec(r.db)
	if err != nil {
		log.Printf("[ERROR] Create event failed: %v", err)
		return apperror.NewInternalServerError("Database query error")
	}
	return nil
}

func (r *EventRepository) GetByID(id ulid.ULID) (*model.Events, error) {
	query := Events.SELECT(Events.AllColumns).
		WHERE(Events.ID.EQ(Bytea(id.Bytes()))).
		LIMIT(1)

	var events []model.Events
	err := query.Query(r.db, &events)
	if err != nil {
		log.Printf("[ERROR] GetByID query failed: %v", err)
		return nil, apperror.NewInternalServerError("Database query error")
	}

	if len(events) == 0 {
		return nil, nil
	}

	return &events[0], nil
}

func (r *EventRepository) List(limit int, cursor *ulid.ULID) ([]model.Events, error) {
	query := Events.SELECT(Events.AllColumns).
		ORDER_BY(Events.ID.DESC()).
		LIMIT(int64(limit))

	if cursor != nil {
		query = query.WHERE(Events.ID.LT(Bytea(cursor.Bytes())))
	}

	var events []model.Events
	err := query.Query(r.db, &events)
	if err != nil {
		log.Printf("[ERROR] List events failed: %v", err)
		return nil, apperror.NewInternalServerError("Database query error")
	}

	return events, nil
}

func (r *EventRepository) ListByUserID(userID ulid.ULID, limit int, cursor *ulid.ULID) ([]model.Events, error) {
	query := Events.SELECT(Events.AllColumns).
		ORDER_BY(Events.ID.DESC()).
		LIMIT(int64(limit))

	if cursor != nil {
		query = query.WHERE(AND(Events.UserID.EQ(Bytea(userID.Bytes())), Events.ID.LT(Bytea(cursor.Bytes()))))
	} else {
		query = query.WHERE(Events.UserID.EQ(Bytea(userID.Bytes())))
	}

	var events []model.Events
	err := query.Query(r.db, &events)
	if err != nil {
		log.Printf("[ERROR] ListByUserID query failed: %v", err)
		return nil, apperror.NewInternalServerError("Database query error")
	}

	return events, nil
}
