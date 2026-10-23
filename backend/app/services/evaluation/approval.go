package evaluation

import (
	"ems/app/enums"
	evalmodel "ems/app/models/evaluation"
	evalrepo "ems/app/repositories/evaluation"
	candidatesvc "ems/app/services/candidate"
	usersvc "ems/app/services/user"
	"errors"
	"time"
)

type EvaluationApprovalService struct{}

func (s *EvaluationApprovalService) Create(assignmentID, approverID int64, level int) (*evalmodel.EvaluationApproval, error) {
	a := &evalmodel.EvaluationApproval{
		EvaluationAssignmentID: assignmentID,
		ApproverID:             approverID,
		Level:                  level,
		Status:                 string(enums.ApprovalStatusPending),
	}
	if err := evalrepo.ApprovalRepo.Create(a); err != nil {
		return nil, err
	}
	return a, nil
}

// FindForAssignments returns the approvals attached to many assignments,
// grouped by assignment ID, for the period history view.
func (s *EvaluationApprovalService) FindForAssignments(assignmentIDs []int64) (map[int64][]evalmodel.EvaluationApproval, error) {
	grouped := map[int64][]evalmodel.EvaluationApproval{}
	approvals, err := evalrepo.ApprovalRepo.FindByAssignmentIDs(assignmentIDs)
	if err != nil {
		return nil, err
	}
	for _, a := range approvals {
		grouped[a.EvaluationAssignmentID] = append(grouped[a.EvaluationAssignmentID], a)
	}
	return grouped, nil
}

func (s *EvaluationApprovalService) FindPendingForApprover(approverID int64) ([]evalmodel.EvaluationApproval, error) {
	// Sync any submitted assignments missing a pending approval record
	submittedAssignments, err := evalrepo.AssignmentRepo.FindByStatus(string(enums.AssignmentStatusSubmitted))
	if err == nil {
		for _, assign := range submittedAssignments {
			existing, _ := evalrepo.ApprovalRepo.FindByAssignmentID(assign.ID)
			hasPending := false
			for _, ap := range existing {
				if ap.Status == string(enums.ApprovalStatusPending) {
					hasPending = true
					break
				}
			}
			if hasPending {
				continue
			}
			approver := assign.EvaluatorID
			if approver == 0 {
				approver = assign.EvaluateeID
			}
			_, _ = s.Create(assign.ID, approver, 1)
		}
	}
	return evalrepo.ApprovalRepo.FindPendingForApprover(approverID)
}

func (s *EvaluationApprovalService) resolveForApprover(id, approverID int64) (*evalmodel.EvaluationApproval, error) {
	a, err := evalrepo.ApprovalRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if a.ApproverID != approverID && !usersvc.PolicyServiceInstance.CanAny(approverID, enums.PermissionEvaluationApprove, enums.PermissionEvaluationManage) {
		return nil, errors.New("only the assigned approver or a manager can decide this evaluation")
	}
	if a.Status != string(enums.ApprovalStatusPending) {
		return nil, errors.New("this approval has already been decided")
	}
	return a, nil
}

func (s *EvaluationApprovalService) Approve(id, approverID int64, comment string) error {
	a, err := s.resolveForApprover(id, approverID)
	if err != nil {
		return err
	}
	now := time.Now()
	a.Status = string(enums.ApprovalStatusApproved)
	a.Comment = comment
	a.ApprovedAt = &now
	if err := evalrepo.ApprovalRepo.Update(a); err != nil {
		return err
	}

	if assign, err := evalrepo.AssignmentRepo.FindByID(a.EvaluationAssignmentID); err == nil {
		assign.Status = string(enums.AssignmentStatusCompleted)
		// completed_at is what the history timeline keys off; without it an
		// approved evaluation never reaches a "completed" event.
		assign.CompletedAt = &now
		_ = evalrepo.AssignmentRepo.Update(assign)
	}

	return nil
}

func (s *EvaluationApprovalService) Reject(id, approverID int64, comment string) error {
	a, err := s.resolveForApprover(id, approverID)
	if err != nil {
		return err
	}
	now := time.Now()
	a.Status = string(enums.ApprovalStatusRejected)
	a.Comment = comment
	a.RejectedAt = &now
	if err := evalrepo.ApprovalRepo.Update(a); err != nil {
		return err
	}

	if assign, err := evalrepo.AssignmentRepo.FindByID(a.EvaluationAssignmentID); err == nil {
		// A rejected approval takes the applicant out of the pipeline: the
		// candidate list must say "rejected" instead of leaving them pending.
		if assign.CandidateID != nil {
			_ = candidatesvc.CandidateServiceInstance.UpdateStatus(*assign.CandidateID, "rejected")
		}
		now := time.Now()
		assign.Status = "rejected"
		assign.CompletedAt = &now
		_ = evalrepo.AssignmentRepo.Update(assign)
	}

	return nil
}
