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

type ClientRepository struct {
	db *sql.DB
}

func NewClientRepository(db *sql.DB) ClientRepository {
	return ClientRepository{db: db}
}

func (r *ClientRepository) GetByID(id ulid.ULID) (*model.Clients, error) {
	query := Clients.SELECT(Clients.AllColumns).
		WHERE(Clients.ID.EQ(Bytea(id.Bytes()))).
		LIMIT(1)

	var clients []model.Clients
	err := query.Query(r.db, &clients)
	if err != nil {
		log.Printf("[ERROR] GetByID client query failed: %v", err)
		return nil, apperror.NewInternalServerError("Database query error")
	}

	if len(clients) == 0 {
		return nil, nil
	}

	return &clients[0], nil
}

func (r *ClientRepository) GetByIDAndUserID(id ulid.ULID, userID ulid.ULID) (*model.Clients, error) {
	query := Clients.SELECT(Clients.AllColumns).
		WHERE(
			AND(
				Clients.ID.EQ(Bytea(id.Bytes())),
				Clients.UserID.EQ(Bytea(userID.Bytes())),
			),
		).
		LIMIT(1)

	var clients []model.Clients
	err := query.Query(r.db, &clients)
	if err != nil {
		log.Printf("[ERROR] GetByIDAndUserID client query failed: %v", err)
		return nil, apperror.NewInternalServerError("Database query error")
	}

	if len(clients) == 0 {
		return nil, nil
	}

	return &clients[0], nil
}

func (r *ClientRepository) Create(client model.Clients) error {
	now := time.Now()
	if client.CreatedAt.IsZero() {
		client.CreatedAt = now
	}
	client.UpdatedAt = now

	_, err := Clients.INSERT().MODEL(client).Exec(r.db)
	if err != nil {
		log.Printf("[ERROR] Create client failed: %v", err)
		return apperror.FromPGError(err)
	}

	return nil
}

func (r *ClientRepository) Update(client model.Clients) error {
	client.UpdatedAt = time.Now()

	_, err := Clients.UPDATE(
		Clients.Name,
		Clients.RedirectURI,
		Clients.AllowedScopes,
		Clients.RevokedAt,
		Clients.UpdatedAt,
	).MODEL(client).WHERE(Clients.ID.EQ(Bytea(client.ID))).Exec(r.db)
	if err != nil {
		log.Printf("[ERROR] Update client failed: %v", err)
		return apperror.FromPGError(err)
	}

	return nil
}

func (r *ClientRepository) Revoke(id ulid.ULID) error {
	result, err := Clients.UPDATE().
		SET(
			Clients.RevokedAt.SET(TimestampzT(time.Now())),
			Clients.UpdatedAt.SET(TimestampzT(time.Now())),
		).
		WHERE(AND(Clients.ID.EQ(Bytea(id.Bytes())), Clients.RevokedAt.IS_NULL())).
		Exec(r.db)
	if err != nil {
		log.Printf("[ERROR] Revoke client failed: %v", err)
		return apperror.NewInternalServerError("Database query error")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("[ERROR] Revoke client rows affected failed: %v", err)
		return apperror.NewInternalServerError("Database query error")
	}
	if rowsAffected == 0 {
		return apperror.NewNotFound("Client not found")
	}

	return nil
}

func (r *ClientRepository) Delete(id ulid.ULID) error {
	result, err := Clients.DELETE().WHERE(Clients.ID.EQ(Bytea(id.Bytes()))).Exec(r.db)
	if err != nil {
		log.Printf("[ERROR] Delete client failed: %v", err)
		return apperror.FromPGError(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("[ERROR] Delete client rows affected failed: %v", err)
		return apperror.NewInternalServerError("Database query error")
	}
	if rowsAffected == 0 {
		return apperror.NewNotFound("Client not found")
	}

	return nil
}

func (r *ClientRepository) ListByUserID(userID ulid.ULID, limit int, cursor *ulid.ULID) ([]model.Clients, error) {
	query := Clients.SELECT(Clients.AllColumns).
		ORDER_BY(Clients.ID.DESC()).
		LIMIT(int64(limit))

	if cursor != nil {
		query = query.WHERE(AND(Clients.UserID.EQ(Bytea(userID.Bytes())), Clients.ID.LT(Bytea(cursor.Bytes()))))
	} else {
		query = query.WHERE(Clients.UserID.EQ(Bytea(userID.Bytes())))
	}

	var clients []model.Clients
	err := query.Query(r.db, &clients)
	if err != nil {
		log.Printf("[ERROR] ListByUserID clients failed: %v", err)
		return nil, apperror.NewInternalServerError("Database query error")
	}

	return clients, nil
}
