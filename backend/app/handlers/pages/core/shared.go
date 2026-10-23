package core

import (
	"ems/app/models"
	candmodel "ems/app/models/candidate"
	usermodel "ems/app/models/user"
	statsrepo "ems/app/repositories/stats"
	"ems/database"
	"log"
)

// WorkerOption and UserOption are the pick-list shapes the templates are
// written against; the definitions live in the models.
type WorkerOption = usermodel.WorkerOption
type UserOption = usermodel.UserOption

// CandidateWithEvaluation is a candidate row with the status of the evaluation
// that may be assigned to them. Both the candidates page and the job-posting
// candidates page read it.
type CandidateWithEvaluation struct {
	candmodel.CandidateListItem
	HasEvaluation       bool
	EvaluationCompleted bool
	EvaluationPending   bool
}

// DashboardStats is the struct the dashboard template is written against.
type DashboardStats = models.DashboardStats

// LoadDashboardStats reads the four dashboard numbers for one user.
func LoadDashboardStats(userID int64) DashboardStats {
	if database.DB == nil {
		log.Println("dashboard stats: database not connected")
		return DashboardStats{}
	}
	stats, err := statsrepo.StatsRepo.Dashboard(userID)
	if err != nil {
		log.Println("dashboard stats:", err)
	}
	return stats
}
