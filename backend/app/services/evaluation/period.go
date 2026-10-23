package evaluation

import (
	"ems/app/enums"
	evalmodel "ems/app/models/evaluation"
	evalrepo "ems/app/repositories/evaluation"
)

type EvaluationPeriodService struct{}

func (s *EvaluationPeriodService) Create(name, description, startDate, endDate string, createdBy int64) (*evalmodel.EvaluationPeriod, error) {
	period := &evalmodel.EvaluationPeriod{
		Name:        name,
		Description: description,
		StartDate:   startDate,
		EndDate:     endDate,
		Status:      string(enums.PeriodStatusDraft),
		CreatedBy:   createdBy,
	}
	if err := evalrepo.PeriodRepo.Create(period); err != nil {
		return nil, err
	}
	return period, nil
}

func (s *EvaluationPeriodService) FindAll() ([]evalmodel.EvaluationPeriod, error) {
	return evalrepo.PeriodRepo.FindAll()
}

func (s *EvaluationPeriodService) FindByID(id int64) (*evalmodel.EvaluationPeriod, error) {
	return evalrepo.PeriodRepo.FindByID(id)
}

func (s *EvaluationPeriodService) Update(id int64, name, description, startDate, endDate, status string) (*evalmodel.EvaluationPeriod, error) {
	period, err := evalrepo.PeriodRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	period.Name = name
	period.Description = description
	period.StartDate = startDate
	period.EndDate = endDate
	period.Status = status
	if err := evalrepo.PeriodRepo.Update(period); err != nil {
		return nil, err
	}
	return period, nil
}

func (s *EvaluationPeriodService) Delete(id int64) error {
	return evalrepo.PeriodRepo.Delete(id)
}
