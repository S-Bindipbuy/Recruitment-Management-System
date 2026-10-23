package organization

import (
	"ems/app/enums"
	orgmodel "ems/app/models/organization"
	orgrepo "ems/app/repositories/organization"
)

type PositionService struct{}

func (s *PositionService) Create(departmentID int64, name, description string) (*orgmodel.Position, error) {
	pos := &orgmodel.Position{
		DepartmentID: departmentID,
		Name:         name,
		Description:  description,
		Status:       string(enums.PositionStatusActive),
	}
	if err := orgrepo.PositionRepo.Create(pos); err != nil {
		return nil, err
	}
	return pos, nil
}

func (s *PositionService) FindAll() ([]orgmodel.Position, error) {
	return orgrepo.PositionRepo.FindAll()
}

func (s *PositionService) Update(id int64, name, description, status string) (*orgmodel.Position, error) {
	pos, err := orgrepo.PositionRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	pos.Name = name
	pos.Description = description
	pos.Status = status
	if err := orgrepo.PositionRepo.Update(pos); err != nil {
		return nil, err
	}
	return pos, nil
}

func (s *PositionService) Delete(id int64) error {
	return orgrepo.PositionRepo.Delete(id)
}
