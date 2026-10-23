package organization

import (
	"ems/app/enums"
	orgmodel "ems/app/models/organization"
	orgrepo "ems/app/repositories/organization"
)

type DepartmentService struct{}

func (s *DepartmentService) Create(organizationID int64, name, description string) (*orgmodel.Department, error) {
	dept := &orgmodel.Department{
		OrganizationID: organizationID,
		Name:           name,
		Description:    description,
		Status:         string(enums.DepartmentStatusActive),
	}
	if err := orgrepo.DepartmentRepo.Create(dept); err != nil {
		return nil, err
	}
	return dept, nil
}

func (s *DepartmentService) FindAll() ([]orgmodel.Department, error) {
	return orgrepo.DepartmentRepo.FindAll()
}

func (s *DepartmentService) Update(id int64, name, description, status string) (*orgmodel.Department, error) {
	dept, err := orgrepo.DepartmentRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	dept.Name = name
	dept.Description = description
	dept.Status = status
	if err := orgrepo.DepartmentRepo.Update(dept); err != nil {
		return nil, err
	}
	return dept, nil
}

func (s *DepartmentService) Delete(id int64) error {
	return orgrepo.DepartmentRepo.Delete(id)
}
