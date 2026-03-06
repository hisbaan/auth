package roles

import "auth/internal/jet/postgres/public/model"

type CreateRoleParams struct {
	Name string `json:"name"`
}

func (s *AdminRolesService) CreateRole(params CreateRoleParams) error {
	role := model.Roles{
		Name: params.Name,
	}
	return s.roleRepo.Create(role)
}

type DeleteRoleParams struct {
	Name string `json:"name"`
}

func (s *AdminRolesService) DeleteRole(params DeleteRoleParams) error {
	return s.roleRepo.Delete(params.Name)
}
