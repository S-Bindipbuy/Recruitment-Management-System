package stats

import (
	"ems/app/models"
	"ems/database"
)

type StatsRepository struct{}

// Dashboard counts the numbers shown on the dashboard in one round trip.
func (r *StatsRepository) Dashboard(_ int64) (models.DashboardStats, error) {
	var stats models.DashboardStats
	err := database.DB.Get(&stats, `
		SELECT
			(SELECT COUNT(*) FROM employees) AS total_employees,
			(SELECT COUNT(*) FROM evaluation_assignments) AS total_assignments,
			(SELECT COUNT(*) FROM evaluation_assignments
			 WHERE status = 'completed') AS completed_evals`,
	)
	if err != nil {
		return models.DashboardStats{}, err
	}
	return stats, nil
}
