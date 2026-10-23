package evaluation

import (
	evalmodel "ems/app/models/evaluation"
	evalrepo "ems/app/repositories/evaluation"
	"time"
)

type EvaluationResultService struct{}

func (s *EvaluationResultService) CalculateAndSave(assignmentID int64) (*evalmodel.EvaluationResult, error) {
	responses, err := evalrepo.ResponseRepo.FindByAssignmentID(assignmentID)
	if err != nil {
		return nil, err
	}

	assignment, err := evalrepo.AssignmentRepo.FindByID(assignmentID)
	if err != nil {
		return nil, err
	}

	template, err := evalrepo.TemplateRepo.FindByID(assignment.EvaluationTemplateID)
	if err != nil {
		return nil, err
	}

	sections, err := evalrepo.SectionRepo.FindByTemplateID(template.ID)
	if err != nil {
		return nil, err
	}

	var totalWeightedScore, totalWeight float64

	for _, section := range sections {
		questions, err := evalrepo.QuestionRepo.FindBySectionID(section.ID)
		if err != nil {
			continue
		}

		for _, question := range questions {
			for _, resp := range responses {
				if resp.EvaluationQuestionID == question.ID && resp.RatingValue != nil {
					totalWeightedScore += *resp.RatingValue * question.Weight
					totalWeight += question.Weight
				}
			}
		}
	}

	if totalWeight == 0 {
		// No rating was captured, so there is no score to report. Writing
		// 0% / F here would read as a verdict rather than as "nothing was
		// ever scored", so drop any stale row instead.
		if existing, err := evalrepo.ResultRepo.FindByAssignmentID(assignmentID); err == nil && existing != nil {
			_ = evalrepo.ResultRepo.Delete(existing.ID)
		}
		return nil, nil
	}

	totalScore := totalWeightedScore / totalWeight
	percentage := (totalScore / 10.0) * 100

	grade := calculateGrade(percentage)

	result := &evalmodel.EvaluationResult{
		EvaluationAssignmentID: assignmentID,
		TotalScore:             totalScore,
		Percentage:             percentage,
		Grade:                  grade,
		CalculatedAt:           time.Now(),
	}

	existing, err := evalrepo.ResultRepo.FindByAssignmentID(assignmentID)
	if err == nil {
		result.ID = existing.ID
		err = evalrepo.ResultRepo.Update(result)
	} else {
		err = evalrepo.ResultRepo.Create(result)
	}

	return result, err
}

func (s *EvaluationResultService) GetResult(assignmentID int64) (*evalmodel.EvaluationResult, error) {
	return evalrepo.ResultRepo.FindByAssignmentID(assignmentID)
}

// GetResultsForAssignments returns results keyed by assignment ID.
func (s *EvaluationResultService) GetResultsForAssignments(assignmentIDs []int64) (map[int64]*evalmodel.EvaluationResult, error) {
	return evalrepo.ResultRepo.FindByAssignmentIDs(assignmentIDs)
}

func calculateGrade(percentage float64) string {
	switch {
	case percentage >= 90:
		return "A"
	case percentage >= 80:
		return "B"
	case percentage >= 70:
		return "C"
	case percentage >= 60:
		return "D"
	default:
		return "F"
	}
}
