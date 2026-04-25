package repositories

import (
	"auth/internal/apperror"
	"auth/internal/jet/postgres/public/model"
	. "auth/internal/jet/postgres/public/table"
	"database/sql"
	"log"
	"time"

	. "github.com/go-jet/jet/v2/postgres"
	"github.com/lib/pq"
	"github.com/oklog/ulid/v2"
)

type UserClientAuthorizationRepository struct {
	db *sql.DB
}

type UserClientAuthorizationWithClient struct {
	model.UserClientAuthorizations
	ClientName        string
	ClientRedirectURI string
}

func NewUserClientAuthorizationRepository(db *sql.DB) UserClientAuthorizationRepository {
	return UserClientAuthorizationRepository{db: db}
}

func (r *UserClientAuthorizationRepository) GetByID(id ulid.ULID) (*model.UserClientAuthorizations, error) {
	query := UserClientAuthorizations.SELECT(UserClientAuthorizations.AllColumns).
		WHERE(UserClientAuthorizations.ID.EQ(Bytea(id.Bytes()))).
		LIMIT(1)

	var authorizations []model.UserClientAuthorizations
	err := query.Query(r.db, &authorizations)
	if err != nil {
		log.Printf("[ERROR] GetByID user client authorization query failed: %v", err)
		return nil, apperror.NewInternalServerError("Database query error")
	}

	if len(authorizations) == 0 {
		return nil, nil
	}

	return &authorizations[0], nil
}

func (r *UserClientAuthorizationRepository) GetByUserIDAndClientID(userID ulid.ULID, clientID ulid.ULID) (*model.UserClientAuthorizations, error) {
	query := UserClientAuthorizations.SELECT(UserClientAuthorizations.AllColumns).
		WHERE(AND(
			UserClientAuthorizations.UserID.EQ(Bytea(userID.Bytes())),
			UserClientAuthorizations.ClientID.EQ(Bytea(clientID.Bytes())),
		)).
		ORDER_BY(UserClientAuthorizations.CreatedAt.DESC()).
		LIMIT(1)

	var authorizations []model.UserClientAuthorizations
	err := query.Query(r.db, &authorizations)
	if err != nil {
		log.Printf("[ERROR] GetByUserIDAndClientID user client authorization query failed: %v", err)
		return nil, apperror.NewInternalServerError("Database query error")
	}

	if len(authorizations) == 0 {
		return nil, nil
	}

	return &authorizations[0], nil
}

func (r *UserClientAuthorizationRepository) ListActiveByUserID(userID ulid.ULID) ([]model.UserClientAuthorizations, error) {
	query := UserClientAuthorizations.SELECT(UserClientAuthorizations.AllColumns).
		WHERE(AND(
			UserClientAuthorizations.UserID.EQ(Bytea(userID.Bytes())),
			UserClientAuthorizations.RevokedAt.IS_NULL(),
		)).
		ORDER_BY(UserClientAuthorizations.LastAuthorizedAt.DESC())

	var authorizations []model.UserClientAuthorizations
	err := query.Query(r.db, &authorizations)
	if err != nil {
		log.Printf("[ERROR] ListActiveByUserID user client authorizations query failed: %v", err)
		return nil, apperror.NewInternalServerError("Database query error")
	}

	return authorizations, nil
}

func (r *UserClientAuthorizationRepository) ListActiveWithClientByUserID(userID ulid.ULID) ([]UserClientAuthorizationWithClient, error) {
	query := SELECT(
		UserClientAuthorizations.AllColumns,
		Clients.Name.AS("userclientauthorizationwithclient.client_name"),
		Clients.RedirectURI.AS("userclientauthorizationwithclient.client_redirect_uri"),
	).
		FROM(UserClientAuthorizations.INNER_JOIN(Clients, UserClientAuthorizations.ClientID.EQ(Clients.ID))).
		WHERE(AND(
			UserClientAuthorizations.UserID.EQ(Bytea(userID.Bytes())),
			UserClientAuthorizations.RevokedAt.IS_NULL(),
			Clients.RevokedAt.IS_NULL(),
		)).
		ORDER_BY(UserClientAuthorizations.LastAuthorizedAt.DESC())

	var authorizations []UserClientAuthorizationWithClient
	err := query.Query(r.db, &authorizations)
	if err != nil {
		log.Printf("[ERROR] ListActiveWithClientByUserID user client authorizations query failed: %v", err)
		return nil, apperror.NewInternalServerError("Database query error")
	}

	return authorizations, nil
}

func (r *UserClientAuthorizationRepository) Create(authorization model.UserClientAuthorizations) error {
	now := time.Now()
	if authorization.CreatedAt.IsZero() {
		authorization.CreatedAt = now
	}
	if authorization.LastAuthorizedAt.IsZero() {
		authorization.LastAuthorizedAt = now
	}
	authorization.UpdatedAt = now

	_, err := UserClientAuthorizations.INSERT().MODEL(authorization).Exec(r.db)
	if err != nil {
		log.Printf("[ERROR] Create user client authorization failed: %v", err)
		return apperror.FromPGError(err)
	}

	return nil
}

func (r *UserClientAuthorizationRepository) Update(authorization model.UserClientAuthorizations) error {
	authorization.UpdatedAt = time.Now()

	_, err := UserClientAuthorizations.UPDATE(
		UserClientAuthorizations.GrantedScopes,
		UserClientAuthorizations.LastAuthorizedAt,
		UserClientAuthorizations.RevokedAt,
		UserClientAuthorizations.UpdatedAt,
	).MODEL(authorization).WHERE(UserClientAuthorizations.ID.EQ(Bytea(authorization.ID))).Exec(r.db)
	if err != nil {
		log.Printf("[ERROR] Update user client authorization failed: %v", err)
		return apperror.FromPGError(err)
	}

	return nil
}

func (r *UserClientAuthorizationRepository) UpsertActive(userID ulid.ULID, clientID ulid.ULID, grantedScopes []string) (*model.UserClientAuthorizations, error) {
	existing, err := r.GetByUserIDAndClientID(userID, clientID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	if existing == nil {
		authorization := model.UserClientAuthorizations{
			ID:               ulid.Make().Bytes(),
			UserID:           userID.Bytes(),
			ClientID:         clientID.Bytes(),
			GrantedScopes:    pq.StringArray(grantedScopes),
			LastAuthorizedAt: now,
			RevokedAt:        nil,
			CreatedAt:        now,
			UpdatedAt:        now,
		}
		if err := r.Create(authorization); err != nil {
			return nil, err
		}
		return &authorization, nil
	}

	existing.GrantedScopes = pq.StringArray(grantedScopes)
	existing.LastAuthorizedAt = now
	existing.RevokedAt = nil
	existing.UpdatedAt = now
	if err := r.Update(*existing); err != nil {
		return nil, err
	}

	return existing, nil
}

func (r *UserClientAuthorizationRepository) Revoke(id ulid.ULID) error {
	_, err := UserClientAuthorizations.UPDATE().
		SET(
			UserClientAuthorizations.RevokedAt.SET(TimestampzT(time.Now())),
			UserClientAuthorizations.UpdatedAt.SET(TimestampzT(time.Now())),
		).
		WHERE(AND(UserClientAuthorizations.ID.EQ(Bytea(id.Bytes())), UserClientAuthorizations.RevokedAt.IS_NULL())).
		Exec(r.db)
	if err != nil {
		log.Printf("[ERROR] Revoke user client authorization failed: %v", err)
		return apperror.NewInternalServerError("Database query error")
	}

	return nil
}

func (r *UserClientAuthorizationRepository) RevokeByClientID(clientID ulid.ULID) error {
	_, err := UserClientAuthorizations.UPDATE().
		SET(
			UserClientAuthorizations.RevokedAt.SET(TimestampzT(time.Now())),
			UserClientAuthorizations.UpdatedAt.SET(TimestampzT(time.Now())),
		).
		WHERE(AND(UserClientAuthorizations.ClientID.EQ(Bytea(clientID.Bytes())), UserClientAuthorizations.RevokedAt.IS_NULL())).
		Exec(r.db)
	if err != nil {
		log.Printf("[ERROR] Revoke user client authorizations by clientID failed: %v", err)
		return apperror.NewInternalServerError("Database query error")
	}

	return nil
}
