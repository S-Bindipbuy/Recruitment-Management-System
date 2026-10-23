package evaluation

import "time"

type EvaluationComment struct {
	ID                     int64     `json:"id" db:"id"`
	EvaluationAssignmentID int64     `json:"evaluation_assignment_id" db:"evaluation_assignment_id"`
	EvaluationQuestionID   *int64    `json:"evaluation_question_id" db:"evaluation_question_id"`
	UserID                 int64     `json:"user_id" db:"user_id"`
	Comment                string    `json:"comment" db:"comment"`
	CreatedAt              time.Time `json:"created_at" db:"created_at"`
	UpdatedAt              time.Time `json:"updated_at" db:"updated_at"`
}

type EvaluationApproval struct {
	ID                     int64      `json:"id" db:"id"`
	EvaluationAssignmentID int64      `json:"evaluation_assignment_id" db:"evaluation_assignment_id"`
	ApproverID             int64      `json:"approver_id" db:"approver_id"`
	Level                  int        `json:"level" db:"level"`
	Status                 string     `json:"status" db:"status"`
	Comment                string     `json:"comment" db:"comment"`
	ApprovedAt             *time.Time `json:"approved_at" db:"approved_at"`
	RejectedAt             *time.Time `json:"rejected_at" db:"rejected_at"`
	CreatedAt              time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at" db:"updated_at"`
}
