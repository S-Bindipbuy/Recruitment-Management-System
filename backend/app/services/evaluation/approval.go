package evaluation

import (
	evalmodel "ems/app/models/evaluation"
	evalrepo "ems/app/repositories/evaluation"
)

type EvaluationApprovalService struct{}

// FindForAssignments reads historical approval records for the period history
// view. New evaluation submissions no longer create approval requests.
func (s *EvaluationApprovalService) FindForAssignments(assignmentIDs []int64) (map[int64][]evalmodel.EvaluationApproval, error) {
	grouped := map[int64][]evalmodel.EvaluationApproval{}
	approvals, err := evalrepo.ApprovalRepo.FindByAssignmentIDs(assignmentIDs)
	if err != nil {
		return nil, err
	}
	for _, approval := range approvals {
		grouped[approval.EvaluationAssignmentID] = append(grouped[approval.EvaluationAssignmentID], approval)
	}
	return grouped, nil
}
