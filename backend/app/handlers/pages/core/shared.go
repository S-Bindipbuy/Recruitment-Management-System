package core

import (
	"ems/app/models"
	candmodel "ems/app/models/candidate"
	evalmodel "ems/app/models/evaluation"
	recmodel "ems/app/models/recruitment"
	usermodel "ems/app/models/user"
	statsrepo "ems/app/repositories/stats"
	"ems/database"
	"log"
	"strings"
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
	Interview           *recmodel.CandidateEvaluationSummary
	InterviewReady      bool
	// PostingID is JobPostingID without the pointer, so a template can look up
	// the rounds that belong to this candidate's job posting.
	PostingID int64
	// CanScheduleInterview hides the schedule action from viewers who may not
	// create interviews for this particular candidate.
	CanScheduleInterview bool
	Contacts             []candmodel.ContactLog
	LastContact          *candmodel.ContactLog
	NextFollowUp         *candmodel.ContactLog
	FollowUpOverdue      bool
}

// InterviewRoundView carries the posting title alongside a round so a
// scheduling picker can tell rounds from different postings apart.
type InterviewRoundView struct {
	recmodel.InterviewRound
	PostingTitle string
	TemplateName string
	Sections     []evalmodel.EvaluationSection
}

// InterviewerGroup keeps a scheduling picker's panel under each department.
type InterviewerGroup struct {
	Department string
	Workers    []WorkerOption
}

// GroupInterviewers splits the interviewer pick-list by department while
// preserving the order the repository returned, so every scheduling surface
// (Interviews, Candidates, Job Posting Candidates) lists people the same way.
func GroupInterviewers(workers []WorkerOption) []InterviewerGroup {
	groups := make(map[string]*InterviewerGroup)
	order := make([]string, 0)
	for _, worker := range workers {
		department := strings.TrimSpace(worker.DepartmentName)
		if department == "" {
			department = "No department"
		}
		group, ok := groups[department]
		if !ok {
			group = &InterviewerGroup{Department: department}
			groups[department] = group
			order = append(order, department)
		}
		group.Workers = append(group.Workers, worker)
	}

	grouped := make([]InterviewerGroup, 0, len(order))
	for _, department := range order {
		grouped = append(grouped, *groups[department])
	}
	return grouped
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
