package evaluation

import "time"

type EvaluationTemplate struct {
	ID          int64     `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Status      string    `json:"status" db:"status"`
	CreatedBy   int64     `json:"created_by" db:"created_by"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// TemplateStructure counts how much content a template holds, keyed by
// template ID. The templates list uses it so a template with no questions is
// obvious before someone picks it in the cycle wizard.
type TemplateStructure struct {
	Sections  int
	Questions int
}
