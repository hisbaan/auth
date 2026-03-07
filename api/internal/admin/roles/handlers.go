package roles

import (
	"auth/internal/events"
	"auth/internal/jet/postgres/public/model"
	"context"
)

type CreateRoleParams struct {
	Name string `json:"name"`
}

func (s *AdminRolesService) CreateRole(ctx context.Context, params CreateRoleParams, actorID string) error {
	role := model.Roles{
		Name: params.Name,
	}
	if err := s.roleRepo.Create(role); err != nil {
		return err
	}

	events.Log(ctx, &s.eventRepo, events.RoleCreated, nil, events.RoleCreatedData{
		Role:    params.Name,
		ActorID: actorID,
	})

	return nil
}

type DeleteRoleParams struct {
	Name string `json:"name"`
}

func (s *AdminRolesService) DeleteRole(ctx context.Context, params DeleteRoleParams, actorID string) error {
	if err := s.roleRepo.Delete(params.Name); err != nil {
		return err
	}

	events.Log(ctx, &s.eventRepo, events.RoleDeleted, nil, events.RoleDeletedData{
		Role:    params.Name,
		ActorID: actorID,
	})

	return nil
}
