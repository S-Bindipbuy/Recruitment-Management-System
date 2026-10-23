package evaluation

import (
	evalmodel "ems/app/models/evaluation"
	"ems/database"
	"time"
)

type EvaluationAssignmentRepository struct{}

const assignmentColumns = `id, evaluation_period_id, evaluation_template_id, evaluator_id, evaluatee_id, candidate_id,
	due_date, status, started_at, submitted_at, completed_at, created_at, updated_at`

func (r *EvaluationAssignmentRepository) FindAll() ([]evalmodel.EvaluationAssignment, error) {
	var assignments []evalmodel.EvaluationAssignment
	err := database.DB.Select(&assignments,
		`SELECT `+assignmentColumns+` FROM evaluation_assignments ORDER BY id DESC`,
	)
	if err != nil {
		return nil, err
	}
	return assignments, nil
}

func (r *EvaluationAssignmentRepository) FindByPeriodID(periodID int64) ([]evalmodel.EvaluationAssignment, error) {
	var assignments []evalmodel.EvaluationAssignment
	err := database.DB.Select(&assignments,
		`SELECT `+assignmentColumns+` FROM evaluation_assignments WHERE evaluation_period_id = ? ORDER BY id DESC`,
		periodID,
	)
	if err != nil {
		return nil, err
	}
	return assignments, nil
}

// CandidateAssignments reports which candidates already carry this template in
// this period. It is the one "already assigned" rule: the wizard preview, the
// wizard launch and a single create all read it, so they cannot drift apart.
func (r *EvaluationAssignmentRepository) CandidateAssignments(periodID, templateID int64) (map[int64]bool, error) {
	assigned := make(map[int64]bool)
	if periodID <= 0 || templateID <= 0 {
		return assigned, nil
	}
	var ids []int64
	err := database.DB.Select(&ids,
		`SELECT candidate_id FROM evaluation_assignments
		 WHERE evaluation_period_id = ? AND evaluation_template_id = ? AND COALESCE(candidate_id, 0) > 0`,
		periodID, templateID,
	)
	if err != nil {
		return nil, err
	}
	for _, id := range ids {
		assigned[id] = true
	}
	return assigned, nil
}

func (r *EvaluationAssignmentRepository) Create(a *evalmodel.EvaluationAssignment) error {
	result, err := database.DB.Exec(
		`INSERT INTO evaluation_assignments (evaluation_period_id, evaluation_template_id, evaluator_id, evaluatee_id, candidate_id, due_date, status) 
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		a.EvaluationPeriodID, a.EvaluationTemplateID, a.EvaluatorID, a.EvaluateeID, a.CandidateID, a.DueDate, a.Status,
	)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	a.ID = id
	return nil
}

func (r *EvaluationAssignmentRepository) FindByID(id int64) (*evalmodel.EvaluationAssignment, error) {
	a := &evalmodel.EvaluationAssignment{}
	err := database.DB.Get(a,
		`SELECT `+assignmentColumns+` FROM evaluation_assignments WHERE id = ?`, id,
	)
	if err != nil {
		return nil, err
	}
	return a, nil
}

// Delete removes an assignment together with everything hanging off it - the
// answers, the score, the comments and the approval - so dropping a run from
// the wizard leaves no orphaned rows behind.
func (r *EvaluationAssignmentRepository) Delete(id int64) error {
	tx, err := database.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, stmt := range []string{
		`DELETE FROM evaluation_responses WHERE evaluation_assignment_id = ?`,
		`DELETE FROM evaluation_results WHERE evaluation_assignment_id = ?`,
		`DELETE FROM evaluation_comments WHERE evaluation_assignment_id = ?`,
		`DELETE FROM evaluation_approvals WHERE evaluation_assignment_id = ?`,
		`DELETE FROM evaluation_assignments WHERE id = ?`,
	} {
		if _, err := tx.Exec(stmt, id); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *EvaluationAssignmentRepository) FindByCandidate(candidateID int64) ([]evalmodel.EvaluationAssignment, error) {
	var assignments []evalmodel.EvaluationAssignment
	err := database.DB.Select(&assignments,
		`SELECT `+assignmentColumns+` FROM evaluation_assignments WHERE candidate_id = ?`,
		candidateID,
	)
	if err != nil {
		return nil, err
	}
	return assignments, nil
}

func (r *EvaluationAssignmentRepository) FindByEvaluator(evaluatorID int64) ([]evalmodel.EvaluationAssignment, error) {
	var assignments []evalmodel.EvaluationAssignment
	err := database.DB.Select(&assignments,
		`SELECT `+assignmentColumns+` FROM evaluation_assignments WHERE evaluator_id = ?`,
		evaluatorID,
	)
	if err != nil {
		return nil, err
	}
	return assignments, nil
}

func (r *EvaluationAssignmentRepository) FindByEvaluatee(evaluateeID int64) ([]evalmodel.EvaluationAssignment, error) {
	var assignments []evalmodel.EvaluationAssignment
	err := database.DB.Select(&assignments,
		`SELECT `+assignmentColumns+` FROM evaluation_assignments WHERE evaluatee_id = ?`,
		evaluateeID,
	)
	if err != nil {
		return nil, err
	}
	return assignments, nil
}

func (r *EvaluationAssignmentRepository) Update(a *evalmodel.EvaluationAssignment) error {
	_, err := database.DB.Exec(
		`UPDATE evaluation_assignments SET status = ?, started_at = ?, submitted_at = ?, completed_at = ?, updated_at = ? WHERE id = ?`,
		a.Status, a.StartedAt, a.SubmittedAt, a.CompletedAt, time.Now(), a.ID,
	)
	return err
}

func (r *EvaluationAssignmentRepository) FindByStatus(status string) ([]evalmodel.EvaluationAssignment, error) {
	var assignments []evalmodel.EvaluationAssignment
	err := database.DB.Select(&assignments,
		`SELECT `+assignmentColumns+` FROM evaluation_assignments WHERE status = ?`,
		status,
	)
	if err != nil {
		return nil, err
	}
	return assignments, nil
}
