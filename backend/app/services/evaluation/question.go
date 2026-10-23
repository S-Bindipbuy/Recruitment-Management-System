package evaluation

import (
	"ems/app/enums"
	evalmodel "ems/app/models/evaluation"
	evalrepo "ems/app/repositories/evaluation"
)

type EvaluationQuestionService struct{}

func (s *EvaluationQuestionService) Create(sectionID int64, question, description, questionType string, isRequired bool, weight float64, sortOrder int) (*evalmodel.EvaluationQuestion, error) {
	q := &evalmodel.EvaluationQuestion{
		EvaluationSectionID: sectionID,
		Question:            question,
		Description:         description,
		QuestionType:        questionType,
		IsRequired:          isRequired,
		Weight:              weight,
		SortOrder:           sortOrder,
		Status:              string(enums.QuestionStatusActive),
	}
	if err := evalrepo.QuestionRepo.Create(q); err != nil {
		return nil, err
	}
	return q, nil
}

func (s *EvaluationQuestionService) FindBySectionID(sectionID int64) ([]evalmodel.EvaluationQuestion, error) {
	return evalrepo.QuestionRepo.FindBySectionID(sectionID)
}

func (s *EvaluationQuestionService) Update(id int64, question, description, questionType string, isRequired bool, weight float64, sortOrder int, status string) (*evalmodel.EvaluationQuestion, error) {
	q, err := evalrepo.QuestionRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	q.Question = question
	q.Description = description
	q.QuestionType = questionType
	q.IsRequired = isRequired
	q.Weight = weight
	q.SortOrder = sortOrder
	q.Status = status
	if err := evalrepo.QuestionRepo.Update(q); err != nil {
		return nil, err
	}
	return q, nil
}

func (s *EvaluationQuestionService) Delete(id int64) error {
	return evalrepo.QuestionRepo.Delete(id)
}

func (s *EvaluationQuestionService) FindByID(id int64) (*evalmodel.EvaluationQuestion, error) {
	return evalrepo.QuestionRepo.FindByID(id)
}
