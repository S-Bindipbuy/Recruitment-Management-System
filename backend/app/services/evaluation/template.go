package evaluation

import (
	"ems/app/enums"
	evalmodel "ems/app/models/evaluation"
	evalrepo "ems/app/repositories/evaluation"
)

type EvaluationTemplateService struct{}

func (s *EvaluationTemplateService) Create(name, description string, createdBy int64) (*evalmodel.EvaluationTemplate, error) {
	tmpl := &evalmodel.EvaluationTemplate{
		Name:        name,
		Description: description,
		Status:      string(enums.TemplateStatusActive),
		CreatedBy:   createdBy,
	}
	if err := evalrepo.TemplateRepo.Create(tmpl); err != nil {
		return nil, err
	}
	return tmpl, nil
}

func (s *EvaluationTemplateService) FindAll() ([]evalmodel.EvaluationTemplate, error) {
	return evalrepo.TemplateRepo.FindAll()
}

func (s *EvaluationTemplateService) FindByID(id int64) (*evalmodel.EvaluationTemplate, error) {
	return evalrepo.TemplateRepo.FindByID(id)
}

func (s *EvaluationTemplateService) Update(id int64, name, description string, status string) (*evalmodel.EvaluationTemplate, error) {
	tmpl, err := evalrepo.TemplateRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	tmpl.Name = name
	tmpl.Description = description
	tmpl.Status = status
	if err := evalrepo.TemplateRepo.Update(tmpl); err != nil {
		return nil, err
	}
	return tmpl, nil
}

func (s *EvaluationTemplateService) Delete(id int64) error {
	return evalrepo.TemplateRepo.Delete(id)
}

// StructureCounts reports sections/questions per template so the templates
// list and the cycle wizard can warn about templates that would score 0%.
func (s *EvaluationTemplateService) StructureCounts() (map[int64]evalmodel.TemplateStructure, error) {
	return evalrepo.TemplateRepo.StructureCounts()
}
