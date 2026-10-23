package evaluation

import "time"

type EvaluationPeriod struct {
	ID          int64     `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	StartDate   string    `json:"start_date" db:"start_date"`
	EndDate     string    `json:"end_date" db:"end_date"`
	Status      string    `json:"status" db:"status"`
	CreatedBy   int64     `json:"created_by" db:"created_by"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}
