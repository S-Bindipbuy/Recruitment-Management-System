package user

import (
	usermodel "ems/app/models/user"
	userrepo "ems/app/repositories/user"
)

type RoleService struct{}

func (s *RoleService) Create(name, description string) (*usermodel.Role, error) {
	role := &usermodel.Role{
		Name:        name,
		Description: description,
	}
	if err := userrepo.RoleRepo.Create(role); err != nil {
		return nil, err
	}
	return role, nil
}

func (s *RoleService) FindAll() ([]usermodel.Role, error) {
	return userrepo.RoleRepo.FindAll()
}

func (s *RoleService) FindByID(id int64) (*usermodel.Role, error) {
	return userrepo.RoleRepo.FindByID(id)
}

func (s *RoleService) Update(id int64, name, description string) (*usermodel.Role, error) {
	role, err := userrepo.RoleRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	role.Name = name
	role.Description = description
	if err := userrepo.RoleRepo.Update(role); err != nil {
		return nil, err
	}
	return role, nil
}

func (s *RoleService) Delete(id int64) error {
	return userrepo.RoleRepo.Delete(id)
}

// FindActiveNonEmployees lists active accounts with no employee record: the
// people who can be chosen as a manager before they are hired as one.
func (s *UserService) FindActiveNonEmployees() ([]usermodel.UserOption, error) {
	return userrepo.UserRepo.FindActiveNonEmployees()
}

// FindNames returns the display names behind a set of user ids in one query.
func (s *UserService) FindNames(ids []int64) (map[int64]string, error) {
	return userrepo.UserRepo.FindNames(ids)
}

// UserName returns one user's display name, empty when it is not known.
func (s *UserService) UserName(id int64) (string, error) {
	names, err := s.FindNames([]int64{id})
	if err != nil {
		return "", err
	}
	return names[id], nil
}
