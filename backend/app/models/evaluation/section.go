package evaluation

import "time"

type EvaluationSection struct {
	ID                   int64     `json:"id" db:"id"`
	EvaluationTemplateID int64     `json:"evaluation_template_id" db:"evaluation_template_id"`
	Name                 string    `json:"name" db:"name"`
	Description          string    `json:"description" db:"description"`
	SortOrder            int       `json:"sort_order" db:"sort_order"`
	CreatedAt            time.Time `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time `json:"updated_at" db:"updated_at"`
}

type EvaluationQuestion struct {
	ID                  int64     `json:"id" db:"id"`
	EvaluationSectionID int64     `json:"evaluation_section_id" db:"evaluation_section_id"`
	Question            string    `json:"question" db:"question"`
	Description         string    `json:"description" db:"description"`
	QuestionType        string    `json:"question_type" db:"question_type"`
	IsRequired          bool      `json:"is_required" db:"is_required"`
	Weight              float64   `json:"weight" db:"weight"`
	SortOrder           int       `json:"sort_order" db:"sort_order"`
	Status              string    `json:"status" db:"status"`
	CreatedAt           time.Time `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time `json:"updated_at" db:"updated_at"`
}
