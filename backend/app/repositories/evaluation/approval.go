package evaluation

import (
	evalmodel "ems/app/models/evaluation"
	"ems/database"
	"strings"
)

type EvaluationApprovalRepository struct{}

// FindByAssignmentIDs is retained for historical approval entries shown in
// period history. The active evaluation workflow no longer creates or decides
// approval requests.
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
		`SELECT id, evaluation_assignment_id, approver_id, level, status,
		 COALESCE(comment, '') AS comment, approved_at, rejected_at, created_at, updated_at
		 FROM evaluation_approvals WHERE evaluation_assignment_id IN (`+placeholders+`) ORDER BY level`, args...,
	)
	if err != nil {
		return nil, err
	}
	return approvals, nil
}
