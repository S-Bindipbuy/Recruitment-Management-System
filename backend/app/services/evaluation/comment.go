package evaluation

import (
	evalmodel "ems/app/models/evaluation"
	evalrepo "ems/app/repositories/evaluation"
)

type EvaluationCommentService struct{}

func (s *EvaluationCommentService) Create(assignmentID, userID int64, questionID *int64, comment string) (*evalmodel.EvaluationComment, error) {
	c := &evalmodel.EvaluationComment{
		EvaluationAssignmentID: assignmentID,
		EvaluationQuestionID:   questionID,
		UserID:                 userID,
		Comment:                comment,
	}
	if err := evalrepo.CommentRepo.Create(c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *EvaluationCommentService) FindByAssignmentID(assignmentID int64) ([]evalmodel.EvaluationComment, error) {
	return evalrepo.CommentRepo.FindByAssignmentID(assignmentID)
}
