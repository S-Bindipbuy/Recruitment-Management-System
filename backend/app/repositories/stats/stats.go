package stats

import (
	"ems/app/models"
	"ems/database"
)

type StatsRepository struct{}

// Dashboard counts the four numbers the dashboards show, in one round trip.
// The approvals count is per user, because "waiting for me" is the only one
// that means anything to the person reading it.
func (r *StatsRepository) Dashboard(userID int64) (models.DashboardStats, error) {
	var stats models.DashboardStats
	err := database.DB.Get(&stats, `
		SELECT
			(SELECT COUNT(*) FROM employees) AS total_employees,
			(SELECT COUNT(*) FROM evaluation_assignments) AS total_assignments,
			(SELECT COUNT(*) FROM evaluation_approvals
			 WHERE approver_id = ? AND status = 'pending') AS pending_approvals,
			(SELECT COUNT(*) FROM evaluation_assignments
			 WHERE status = 'completed') AS completed_evals`,
		userID)
	if err != nil {
		return models.DashboardStats{}, err
	}
	return stats, nil
}
