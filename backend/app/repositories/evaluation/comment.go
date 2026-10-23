package evaluation

import (
	evalmodel "ems/app/models/evaluation"
	"ems/database"
)

type EvaluationCommentRepository struct{}

func (r *EvaluationCommentRepository) Create(c *evalmodel.EvaluationComment) error {
	result, err := database.DB.Exec(
		`INSERT INTO evaluation_comments (evaluation_assignment_id, evaluation_question_id, user_id, comment) 
		 VALUES (?, ?, ?, ?)`,
		c.EvaluationAssignmentID, c.EvaluationQuestionID, c.UserID, c.Comment,
	)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	c.ID = id
	return nil
}

func (r *EvaluationCommentRepository) FindByAssignmentID(assignmentID int64) ([]evalmodel.EvaluationComment, error) {
	var comments []evalmodel.EvaluationComment
	err := database.DB.Select(&comments,
		`SELECT id, evaluation_assignment_id, evaluation_question_id, user_id, comment, created_at, updated_at 
		 FROM evaluation_comments WHERE evaluation_assignment_id = ?`,
		assignmentID,
	)
	if err != nil {
		return nil, err
	}
	return comments, nil
}
