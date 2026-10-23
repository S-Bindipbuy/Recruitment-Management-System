package organization

import (
	"ems/app/enums"
	orgmodel "ems/app/models/organization"
	orgrepo "ems/app/repositories/organization"
)

type OrganizationService struct{}

func (s *OrganizationService) Create(name, description string) (*orgmodel.Organization, error) {
	org := &orgmodel.Organization{
		Name:        name,
		Description: description,
		Status:      string(enums.OrganizationStatusActive),
	}
	if err := orgrepo.OrganizationRepo.Create(org); err != nil {
		return nil, err
	}
	return org, nil
}

func (s *OrganizationService) FindAll() ([]orgmodel.Organization, error) {
	return orgrepo.OrganizationRepo.FindAll()
}

func (s *OrganizationService) Update(id int64, name, description, status string) (*orgmodel.Organization, error) {
	org, err := orgrepo.OrganizationRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	org.Name = name
	org.Description = description
	org.Status = status
	if err := orgrepo.OrganizationRepo.Update(org); err != nil {
		return nil, err
	}
	return org, nil
}

func (s *OrganizationService) Delete(id int64) error {
	return orgrepo.OrganizationRepo.Delete(id)
}
