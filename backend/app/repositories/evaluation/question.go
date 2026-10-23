package evaluation

import (
	evalmodel "ems/app/models/evaluation"
	"ems/database"
	"time"
)

type EvaluationQuestionRepository struct{}

func (r *EvaluationQuestionRepository) Create(q *evalmodel.EvaluationQuestion) error {
	result, err := database.DB.Exec(
		`INSERT INTO evaluation_questions (evaluation_section_id, question, description, question_type, is_required, weight, sort_order, status) 
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		q.EvaluationSectionID, q.Question, q.Description, q.QuestionType, q.IsRequired, q.Weight, q.SortOrder, q.Status,
	)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	q.ID = id
	return nil
}

func (r *EvaluationQuestionRepository) FindByID(id int64) (*evalmodel.EvaluationQuestion, error) {
	q := &evalmodel.EvaluationQuestion{}
	err := database.DB.Get(q,
		`SELECT id, evaluation_section_id, question, description, question_type, is_required, weight, sort_order, status, created_at, updated_at 
		 FROM evaluation_questions WHERE id = ?`, id,
	)
	if err != nil {
		return nil, err
	}
	return q, nil
}

func (r *EvaluationQuestionRepository) FindBySectionID(sectionID int64) ([]evalmodel.EvaluationQuestion, error) {
	var questions []evalmodel.EvaluationQuestion
	err := database.DB.Select(&questions,
		`SELECT id, evaluation_section_id, question, COALESCE(description, '') AS description, question_type, is_required, weight, sort_order, status, created_at, updated_at 
		 FROM evaluation_questions WHERE evaluation_section_id = ? ORDER BY sort_order`,
		sectionID,
	)
	if err != nil {
		return nil, err
	}
	return questions, nil
}

func (r *EvaluationQuestionRepository) Update(q *evalmodel.EvaluationQuestion) error {
	_, err := database.DB.Exec(
		`UPDATE evaluation_questions SET question = ?, description = ?, question_type = ?, is_required = ?, 
		 weight = ?, sort_order = ?, status = ?, updated_at = ? WHERE id = ?`,
		q.Question, q.Description, q.QuestionType, q.IsRequired, q.Weight, q.SortOrder, q.Status, time.Now(), q.ID,
	)
	return err
}

func (r *EvaluationQuestionRepository) Delete(id int64) error {
	_, err := database.DB.Exec("DELETE FROM evaluation_questions WHERE id = ?", id)
	return err
}
