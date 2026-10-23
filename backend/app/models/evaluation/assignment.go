package evaluation

import "time"

type EvaluationAssignment struct {
	ID                   int64      `json:"id" db:"id"`
	EvaluationPeriodID   int64      `json:"evaluation_period_id" db:"evaluation_period_id"`
	EvaluationTemplateID int64      `json:"evaluation_template_id" db:"evaluation_template_id"`
	EvaluatorID          int64      `json:"evaluator_id" db:"evaluator_id"`
	EvaluateeID          int64      `json:"evaluatee_id" db:"evaluatee_id"`
	CandidateID          *int64     `json:"candidate_id" db:"candidate_id"`
	DueDate              *time.Time `json:"due_date" db:"due_date"`
	Status               string     `json:"status" db:"status"`
	StartedAt            *time.Time `json:"started_at" db:"started_at"`
	SubmittedAt          *time.Time `json:"submitted_at" db:"submitted_at"`
	CompletedAt          *time.Time `json:"completed_at" db:"completed_at"`
	CreatedAt            time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at" db:"updated_at"`
}

type EvaluationResponse struct {
	ID                     int64     `json:"id" db:"id"`
	EvaluationAssignmentID int64     `json:"evaluation_assignment_id" db:"evaluation_assignment_id"`
	EvaluationQuestionID   int64     `json:"evaluation_question_id" db:"evaluation_question_id"`
	RatingValue            *float64  `json:"rating_value" db:"rating_value"`
	TextValue              string    `json:"text_value" db:"text_value"`
	NumericValue           *float64  `json:"numeric_value" db:"numeric_value"`
	BooleanValue           *bool     `json:"boolean_value" db:"boolean_value"`
	SelectedOptionID       *int64    `json:"selected_option_id" db:"selected_option_id"`
	CreatedAt              time.Time `json:"created_at" db:"created_at"`
	UpdatedAt              time.Time `json:"updated_at" db:"updated_at"`
}

type EvaluationResult struct {
	ID                     int64     `json:"id" db:"id"`
	EvaluationAssignmentID int64     `json:"evaluation_assignment_id" db:"evaluation_assignment_id"`
	TotalScore             float64   `json:"total_score" db:"total_score"`
	Percentage             float64   `json:"percentage" db:"percentage"`
	Grade                  string    `json:"grade" db:"grade"`
	CalculatedAt           time.Time `json:"calculated_at" db:"calculated_at"`
	CreatedAt              time.Time `json:"created_at" db:"created_at"`
	UpdatedAt              time.Time `json:"updated_at" db:"updated_at"`
}

// EvaluationPair is one evaluator/candidate pairing inside a bulk launch. The
// wizard only ever evaluates applicants, so a pair is somebody doing the
// assessing plus the candidate they will assess.
type EvaluationPair struct {
	EvaluatorID int64
	CandidateID *int64
}
