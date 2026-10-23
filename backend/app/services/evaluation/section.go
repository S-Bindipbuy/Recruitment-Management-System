package evaluation

import (
	evalmodel "ems/app/models/evaluation"
	evalrepo "ems/app/repositories/evaluation"
)

type EvaluationSectionService struct{}

func (s *EvaluationSectionService) Create(templateID int64, name, description string, sortOrder int) (*evalmodel.EvaluationSection, error) {
	sec := &evalmodel.EvaluationSection{
		EvaluationTemplateID: templateID,
		Name:                 name,
		Description:          description,
		SortOrder:            sortOrder,
	}
	if err := evalrepo.SectionRepo.Create(sec); err != nil {
		return nil, err
	}
	return sec, nil
}

func (s *EvaluationSectionService) FindByTemplateID(templateID int64) ([]evalmodel.EvaluationSection, error) {
	return evalrepo.SectionRepo.FindByTemplateID(templateID)
}

func (s *EvaluationSectionService) Update(id int64, name, description string, sortOrder int) (*evalmodel.EvaluationSection, error) {
	sec, err := evalrepo.SectionRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	sec.Name = name
	sec.Description = description
	sec.SortOrder = sortOrder
	if err := evalrepo.SectionRepo.Update(sec); err != nil {
		return nil, err
	}
	return sec, nil
}

func (s *EvaluationSectionService) Delete(id int64) error {
	return evalrepo.SectionRepo.Delete(id)
}

func (s *EvaluationSectionService) FindByID(id int64) (*evalmodel.EvaluationSection, error) {
	return evalrepo.SectionRepo.FindByID(id)
}
