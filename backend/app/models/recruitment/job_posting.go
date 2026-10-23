package recruitment

import "time"

type JobPosting struct {
	ID             int64      `json:"id" db:"id"`
	PositionID     int64      `json:"position_id" db:"position_id"`
	Title          string     `json:"title" db:"title"`
	Description    string     `json:"description" db:"description"`
	Requirements   string     `json:"requirements" db:"requirements"`
	SkillsRequired string     `json:"skills_required" db:"skills_required"`
	SalaryMin      *float64   `json:"salary_min" db:"salary_min"`
	SalaryMax      *float64   `json:"salary_max" db:"salary_max"`
	EmploymentType string     `json:"employment_type" db:"employment_type"`
	Status         string     `json:"status" db:"status"`
	PostedAt       *time.Time `json:"posted_at" db:"posted_at"`
	ClosedAt       *time.Time `json:"closed_at" db:"closed_at"`
	CreatedBy      int64      `json:"created_by" db:"created_by"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}

type JobPostingListItem struct {
	JobPosting
	PositionName   string `json:"position_name" db:"position_name"`
	DepartmentName string `json:"department_name" db:"department_name"`
}
