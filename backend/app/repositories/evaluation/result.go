package evaluation

import (
	evalmodel "ems/app/models/evaluation"
	"ems/database"
	"strings"
	"time"
)

type EvaluationResultRepository struct{}

func (r *EvaluationResultRepository) Create(result *evalmodel.EvaluationResult) error {
	res, err := database.DB.Exec(
		`INSERT INTO evaluation_results (evaluation_assignment_id, total_score, percentage, grade, calculated_at) 
		 VALUES (?, ?, ?, ?, ?)`,
		result.EvaluationAssignmentID, result.TotalScore, result.Percentage, result.Grade, result.CalculatedAt,
	)
	if err != nil {
		return err
	}
	id, _ := res.LastInsertId()
	result.ID = id
	return nil
}

func (r *EvaluationResultRepository) FindByAssignmentID(assignmentID int64) (*evalmodel.EvaluationResult, error) {
	result := &evalmodel.EvaluationResult{}
	err := database.DB.Get(result,
		`SELECT id, evaluation_assignment_id, total_score, percentage, grade, calculated_at, created_at, updated_at 
		 FROM evaluation_results WHERE evaluation_assignment_id = ?`, assignmentID,
	)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// FindByAssignmentIDs fetches results for many assignments in one query, keyed
// by assignment ID for the period history page.
func (r *EvaluationResultRepository) FindByAssignmentIDs(assignmentIDs []int64) (map[int64]*evalmodel.EvaluationResult, error) {
	ids := make([]int64, 0, len(assignmentIDs))
	seen := make(map[int64]bool, len(assignmentIDs))
	for _, id := range assignmentIDs {
		if id > 0 && !seen[id] {
			seen[id] = true
			ids = append(ids, id)
		}
	}
	results := map[int64]*evalmodel.EvaluationResult{}
	if len(ids) == 0 {
		return results, nil
	}

	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(ids)), ",")
	args := make([]any, len(ids))
	for i, id := range ids {
		args[i] = id
	}

	var rowsData []evalmodel.EvaluationResult
	err := database.DB.Select(&rowsData,
		`SELECT id, evaluation_assignment_id, total_score, percentage, COALESCE(grade, '') AS grade, calculated_at, created_at, updated_at 
		 FROM evaluation_results WHERE evaluation_assignment_id IN (`+placeholders+`)`, args...,
	)
	if err != nil {
		return nil, err
	}
	for i := range rowsData {
		result := rowsData[i]
		results[result.EvaluationAssignmentID] = &result
	}
	return results, nil
}

func (r *EvaluationResultRepository) Update(result *evalmodel.EvaluationResult) error {
	_, err := database.DB.Exec(
		`UPDATE evaluation_results SET total_score = ?, percentage = ?, grade = ?, calculated_at = ?, updated_at = ? WHERE id = ?`,
		result.TotalScore, result.Percentage, result.Grade, result.CalculatedAt, time.Now(), result.ID,
	)
	return err
}

func (r *EvaluationResultRepository) Delete(id int64) error {
	_, err := database.DB.Exec(`DELETE FROM evaluation_results WHERE id = ?`, id)
	return err
}
