package roles

import (
	"auth/internal/jet/postgres/public/model"
	"auth/internal/utils"
)

type ListRolesResponse struct {
	Roles []string `json:"roles"`
}

func (s *RolesService) ListRoles() (ListRolesResponse, error) {
	roles, err := s.roleRepo.GetAll()
	if err != nil {
		return ListRolesResponse{}, err
	}

	return ListRolesResponse{Roles: utils.Map(roles, func(role model.Roles) string { return role.Name })}, nil
}

type CreateRoleParams struct {
	Name string `json:"name"`
}

func (s *RolesService) CreateRole(params CreateRoleParams) error {
	role := model.Roles{
		Name: params.Name,
	}
	return s.roleRepo.Create(role)
}

type DeleteRoleParams struct {
	Name string `json:"name"`
}

func (s *RolesService) DeleteRole(params DeleteRoleParams) error {
	return s.roleRepo.Delete(params.Name)
}
