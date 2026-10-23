package evaluation

import (
	evalmodel "ems/app/models/evaluation"
	"ems/database"
)

type EvaluationResponseRepository struct{}

func (r *EvaluationResponseRepository) FindByAssignmentID(assignmentID int64) ([]evalmodel.EvaluationResponse, error) {
	var responses []evalmodel.EvaluationResponse
	err := database.DB.Select(&responses,
		`SELECT id, evaluation_assignment_id, evaluation_question_id, rating_value, COALESCE(text_value, '') AS text_value, 
		 numeric_value, boolean_value, selected_option_id, created_at, updated_at 
		 FROM evaluation_responses WHERE evaluation_assignment_id = ?`,
		assignmentID,
	)
	if err != nil {
		return nil, err
	}
	return responses, nil
}

func (r *EvaluationResponseRepository) Upsert(resp *evalmodel.EvaluationResponse) error {
	_, err := database.DB.Exec(
		`INSERT INTO evaluation_responses (evaluation_assignment_id, evaluation_question_id, rating_value, text_value, numeric_value, boolean_value, selected_option_id) 
		 VALUES (?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT(evaluation_assignment_id, evaluation_question_id) DO UPDATE SET 
		 rating_value = excluded.rating_value, text_value = excluded.text_value, 
		 numeric_value = excluded.numeric_value, boolean_value = excluded.boolean_value, 
		 selected_option_id = excluded.selected_option_id, updated_at = CURRENT_TIMESTAMP`,
		resp.EvaluationAssignmentID, resp.EvaluationQuestionID, resp.RatingValue, resp.TextValue,
		resp.NumericValue, resp.BooleanValue, resp.SelectedOptionID,
	)
	return err
}
