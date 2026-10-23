package evaluation

import (
	evalmodel "ems/app/models/evaluation"
	"ems/database"
	"strings"
	"time"
)

type EvaluationApprovalRepository struct{}

func (r *EvaluationApprovalRepository) Create(a *evalmodel.EvaluationApproval) error {
	result, err := database.DB.Exec(
		`INSERT INTO evaluation_approvals (evaluation_assignment_id, approver_id, level, status, comment) 
		 VALUES (?, ?, ?, ?, ?)`,
		a.EvaluationAssignmentID, a.ApproverID, a.Level, a.Status, a.Comment,
	)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	a.ID = id
	return nil
}

func (r *EvaluationApprovalRepository) FindByID(id int64) (*evalmodel.EvaluationApproval, error) {
	a := &evalmodel.EvaluationApproval{}
	err := database.DB.Get(a,
		`SELECT id, evaluation_assignment_id, approver_id, level, status, comment, approved_at, rejected_at, created_at, updated_at 
		 FROM evaluation_approvals WHERE id = ?`, id,
	)
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (r *EvaluationApprovalRepository) FindByAssignmentID(assignmentID int64) ([]evalmodel.EvaluationApproval, error) {
	var approvals []evalmodel.EvaluationApproval
	err := database.DB.Select(&approvals,
		`SELECT id, evaluation_assignment_id, approver_id, level, status, COALESCE(comment, '') AS comment, approved_at, rejected_at, created_at, updated_at 
		 FROM evaluation_approvals WHERE evaluation_assignment_id = ? ORDER BY level`,
		assignmentID,
	)
	if err != nil {
		return nil, err
	}
	return approvals, nil
}

// FindByAssignmentIDs fetches approvals for many assignments in one query.
// Used by the period history page so it does not fan out one query per row.
func (r *EvaluationApprovalRepository) FindByAssignmentIDs(assignmentIDs []int64) ([]evalmodel.EvaluationApproval, error) {
	ids := make([]int64, 0, len(assignmentIDs))
	seen := make(map[int64]bool, len(assignmentIDs))
	for _, id := range assignmentIDs {
		if id > 0 && !seen[id] {
			seen[id] = true
			ids = append(ids, id)
		}
	}
	if len(ids) == 0 {
		return nil, nil
	}

	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(ids)), ",")
	args := make([]any, len(ids))
	for i, id := range ids {
		args[i] = id
	}

	var approvals []evalmodel.EvaluationApproval
	err := database.DB.Select(&approvals,
		`SELECT id, evaluation_assignment_id, approver_id, level, status, COALESCE(comment, '') AS comment, approved_at, rejected_at, created_at, updated_at 
		 FROM evaluation_approvals WHERE evaluation_assignment_id IN (`+placeholders+`) ORDER BY level`, args...,
	)
	if err != nil {
		return nil, err
	}
	return approvals, nil
}

func (r *EvaluationApprovalRepository) FindPendingForApprover(approverID int64) ([]evalmodel.EvaluationApproval, error) {
	var approvals []evalmodel.EvaluationApproval
	err := database.DB.Select(&approvals,
		`SELECT id, evaluation_assignment_id, approver_id, level, status, COALESCE(comment, '') AS comment, approved_at, rejected_at, created_at, updated_at 
		 FROM evaluation_approvals WHERE status = 'pending' ORDER BY id DESC`,
	)
	if err != nil {
		return nil, err
	}
	return approvals, nil
}

func (r *EvaluationApprovalRepository) Update(a *evalmodel.EvaluationApproval) error {
	_, err := database.DB.Exec(
		`UPDATE evaluation_approvals SET status = ?, comment = ?, approved_at = ?, rejected_at = ?, updated_at = ? WHERE id = ?`,
		a.Status, a.Comment, a.ApprovedAt, a.RejectedAt, time.Now(), a.ID,
	)
	return err
}
