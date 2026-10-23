package evaluation

import (
	evalmodel "ems/app/models/evaluation"
	evalrepo "ems/app/repositories/evaluation"
)

type EvaluationResponseService struct{}

func (s *EvaluationResponseService) SaveResponse(assignmentID, questionID int64, ratingValue *float64, textValue string, numericValue *float64, booleanValue *bool, selectedOptionID *int64) error {
	resp := &evalmodel.EvaluationResponse{
		EvaluationAssignmentID: assignmentID,
		EvaluationQuestionID:   questionID,
		RatingValue:            ratingValue,
		TextValue:              textValue,
		NumericValue:           numericValue,
		BooleanValue:           booleanValue,
		SelectedOptionID:       selectedOptionID,
	}
	return evalrepo.ResponseRepo.Upsert(resp)
}

func (s *EvaluationResponseService) GetResponses(assignmentID int64) ([]evalmodel.EvaluationResponse, error) {
	return evalrepo.ResponseRepo.FindByAssignmentID(assignmentID)
}
