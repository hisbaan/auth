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

type RoleRepository struct {
	db *sql.DB
}

func NewRoleRepository(db *sql.DB) RoleRepository {
	return RoleRepository{db: db}
}

func (r *RoleRepository) GetByName(name string) (*model.Roles, error) {
	query := Roles.SELECT(Roles.AllColumns).
		WHERE(Roles.Name.EQ(String(name))).
		LIMIT(1)

	var roles []model.Roles
	err := query.Query(r.db, &roles)
	if err != nil {
		log.Printf("[ERROR] GetByName query failed: %v", err)
		return nil, apperror.NewInternalServerError("Database query error")
	}

	if len(roles) == 0 {
		return nil, nil
	}

	return &roles[0], nil
}

func (r *RoleRepository) GetByUserID(id ulid.ULID) ([]model.Roles, error) {
	query := Roles.SELECT(Roles.AllColumns).
		FROM(Roles.INNER_JOIN(UserRoles, Roles.Name.EQ(UserRoles.Role))).
		WHERE(UserRoles.UserID.EQ(Bytea(id.Bytes())))

	var roles []model.Roles
	err := query.Query(r.db, &roles)
	if err != nil {
		log.Printf("[ERROR] GetByUserID query failed: %v", err)
		return nil, apperror.NewInternalServerError("Database query error")
	}

	return roles, nil
}

func (r *RoleRepository) GetAll() ([]model.Roles, error) {
	query := Roles.SELECT(Roles.AllColumns)

	var roles []model.Roles
	err := query.Query(r.db, &roles)
	if err != nil {
		log.Printf("[ERROR] GetAll query failed: %v", err)
		return nil, apperror.NewInternalServerError("Database query error")
	}

	return roles, nil
}

func (r *RoleRepository) Create(role model.Roles) error {
	role.CreatedAt = time.Now()
	role.UpdatedAt = time.Now()
	_, err := Roles.INSERT().MODEL(role).ON_CONFLICT().DO_NOTHING().Exec(r.db)
	if err != nil {
		log.Printf("[ERROR] Create role failed: %v", err)
		return apperror.FromPGError(err)
	}
	return nil
}

func (r *RoleRepository) Update(role model.Roles) error {
	role.UpdatedAt = time.Now()
	_, err := Roles.UPDATE(Roles.Name).MODEL(role).WHERE(Roles.Name.EQ(String(role.Name))).Exec(r.db)
	if err != nil {
		log.Printf("[ERROR] Update role failed: %v", err)
		return apperror.NewInternalServerError("Database query error")
	}
	return nil
}

func (r *RoleRepository) Delete(name string) error {
	_, err := Roles.DELETE().WHERE(Roles.Name.EQ(String(name))).Exec(r.db)
	if err != nil {
		log.Printf("[ERROR] Delete role failed: %v", err)
		return apperror.FromPGError(err)
	}
	return nil
}

func (r *RoleRepository) CreateUserRole(userID ulid.ULID, roleName string) error {
	userRole := model.UserRoles{
		UserID:    userID.Bytes(),
		Role:      roleName,
		CreatedAt: time.Now(),
	}
	_, err := UserRoles.INSERT().MODEL(userRole).ON_CONFLICT().DO_NOTHING().Exec(r.db)
	if err != nil {
		log.Printf("[ERROR] CreateUserRole failed: %v", err)
		return apperror.FromPGError(err)
	}
	return nil
}

func (r *RoleRepository) DeleteUserRole(userID ulid.ULID, roleName string) error {
	_, err := UserRoles.DELETE().
		WHERE(
			AND(
				UserRoles.UserID.EQ(Bytea(userID.Bytes())),
				UserRoles.Role.EQ(String(roleName)),
			),
		).
		Exec(r.db)
	if err != nil {
		log.Printf("[ERROR] DeleteUserRole failed: %v", err)
		return apperror.NewInternalServerError("Database query error")
	}
	return nil
}
