package report

import (
	core "ems/app/handlers/pages/core"
	recmodel "ems/app/models/recruitment"
	recsvc "ems/app/services/recruitment"

	"github.com/gofiber/fiber/v2"
)

type ReportHandler struct{}

type ReportsPageData struct {
	Stats      core.DashboardStats
	Summaries  []recmodel.CandidateEvaluationSummary
	Activities []recmodel.InterviewerActivity
	Decisions  []recmodel.HiringDecision
}

func (h *ReportHandler) Reports(c *fiber.Ctx) error {
	summaries, err1 := recsvc.InterviewServiceInst.GetAllCandidateSummaries()
	activities, err2 := recsvc.InterviewServiceInst.GetInterviewerActivity()
	decisions, err3 := recsvc.InterviewServiceInst.GetHiringDecisions()
	if err1 != nil {
		summaries = []recmodel.CandidateEvaluationSummary{}
	}
	if err2 != nil {
		activities = []recmodel.InterviewerActivity{}
	}
	if err3 != nil {
		decisions = []recmodel.HiringDecision{}
	}

	data := ReportsPageData{
		Stats:      core.LoadDashboardStats(core.CurrentUserID(c)),
		Summaries:  summaries,
		Activities: activities,
		Decisions:  decisions,
	}

	return core.RenderPage(c, "page-reports", data)
}
