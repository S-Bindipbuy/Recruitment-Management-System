package user

import (
	"ems/app/enums"
	usermodel "ems/app/models/user"
	userrepo "ems/app/repositories/user"
)

type UserService struct{}

func (s *UserService) Register(name, email, password string) (*usermodel.User, error) {
	user := &usermodel.User{
		Name:     name,
		Email:    email,
		Password: password,
		Status:   string(enums.UserStatusActive),
	}
	if err := userrepo.UserRepo.Create(user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *UserService) FindByEmail(email string) (*usermodel.User, error) {
	return userrepo.UserRepo.FindByEmail(email)
}

func (s *UserService) FindByID(id int64) (*usermodel.User, error) {
	return userrepo.UserRepo.FindByID(id)
}

func (s *UserService) FindAll() ([]usermodel.User, error) {
	return userrepo.UserRepo.FindAll()
}

func (s *UserService) Update(id int64, name, email, status string) (*usermodel.User, error) {
	user, err := userrepo.UserRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	user.Name = name
	user.Email = email
	user.Status = status
	if err := userrepo.UserRepo.Update(user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *UserService) Delete(id int64) error {
	return userrepo.UserRepo.Delete(id)
}

func (s *UserService) GetUserRoles(userID int64) ([]string, error) {
	return userrepo.UserRepo.GetUserRoles(userID)
}

func (s *UserService) GetUserRoleIDs(userID int64) ([]int64, error) {
	return userrepo.UserRepo.GetUserRoleIDs(userID)
}

func (s *UserService) AssignRoles(userID int64, roleIDs []int64) error {
	return userrepo.UserRepo.AssignRoles(userID, roleIDs)
}
