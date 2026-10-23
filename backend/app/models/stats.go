package models

// DashboardStats is the handful of numbers both dashboards show: one struct
// built by one query, so the page and the API cannot disagree.
type DashboardStats struct {
	TotalEmployees   int `json:"total_employees" db:"total_employees"`
	TotalAssignments int `json:"total_assignments" db:"total_assignments"`
	PendingApprovals int `json:"pending_approvals" db:"pending_approvals"`
	CompletedEvals   int `json:"completed_evals" db:"completed_evals"`
}
