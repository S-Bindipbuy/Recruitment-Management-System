package report

import (
	core "ems/app/handlers/pages/core"
	recmodel "ems/app/models/recruitment"
	recsvc "ems/app/services/recruitment"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type ReportHandler struct{}

type ReportsPageData struct {
	Stats                core.DashboardStats
	Summaries            []recmodel.CandidateEvaluationSummary
	Activities           []recmodel.InterviewerActivity
	Decisions            []recmodel.HiringDecision
	Postings             []recmodel.JobPostingListItem
	SelectedJobPostingID int64
	SearchQuery          string
	SelectedDecision     string
	HasFilters           bool
}

func (h *ReportHandler) Reports(c *fiber.Ctx) error {
	// Filters are simple GET parameters so the filtered report stays
	// bookmarkable and the Refresh control can re-use the current URL.
	postingID, _ := strconv.ParseInt(c.Query("job_posting_id"), 10, 64)
	search := strings.TrimSpace(c.Query("q"))
	decision := strings.TrimSpace(c.Query("decision"))

	summaries, err1 := recsvc.InterviewServiceInst.GetAllCandidateSummaries()
	activities, err2 := recsvc.InterviewServiceInst.GetInterviewerActivity(postingID)
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

	// Rank before filtering so a candidate keeps the same rank it has on the
	// other pages even when the report is narrowed down.
	summaries = core.PipelineCandidates(summaries)
	summaries = core.AssignCandidateRanks(summaries)

	filteredSummaries := make([]recmodel.CandidateEvaluationSummary, 0, len(summaries))
	for _, summary := range summaries {
		if postingID > 0 && summary.JobPostingID != postingID {
			continue
		}
		if !matchesSearch(search, summary.CandidateName, summary.PositionName, summary.PostingTitle) {
			continue
		}
		if !matchesDecision(decision, summary.FinalDecision) {
			continue
		}
		filteredSummaries = append(filteredSummaries, summary)
	}
	summaries = filteredSummaries

	filteredDecisions := make([]recmodel.HiringDecision, 0, len(decisions))
	for _, decisionRow := range decisions {
		if postingID > 0 && decisionRow.JobPostingID != postingID {
			continue
		}
		if !matchesSearch(search, decisionRow.CandidateName, decisionRow.PositionName, decisionRow.DepartmentName) {
			continue
		}
		if !matchesHiringDecision(decision, decisionRow.Decision) {
			continue
		}
		filteredDecisions = append(filteredDecisions, decisionRow)
	}
	decisions = filteredDecisions

	postings, _ := recsvc.JobPostingServiceInst.FindAll()
	data := ReportsPageData{
		Stats:                core.LoadDashboardStats(core.CurrentUserID(c)),
		Summaries:            summaries,
		Activities:           activities,
		Decisions:            decisions,
		Postings:             postings,
		SelectedJobPostingID: postingID,
		SearchQuery:          search,
		SelectedDecision:     decision,
		HasFilters:           postingID > 0 || search != "" || decision != "",
	}

	return core.RenderPage(c, "page-reports", data)
}

// matchesSearch reports whether any of the supplied fields contains the
// case-insensitive search term. An empty term matches everything.
func matchesSearch(term string, fields ...string) bool {
	term = strings.ToLower(strings.TrimSpace(term))
	if term == "" {
		return true
	}
	for _, field := range fields {
		if strings.Contains(strings.ToLower(field), term) {
			return true
		}
	}
	return false
}

// matchesDecision compares the evaluation recommendation (hire, maybe, reject,
// pending, in_progress) with the selected report filter.
func matchesDecision(filter, recommendation string) bool {
	filter = strings.ToLower(strings.TrimSpace(filter))
	if filter == "" {
		return true
	}
	return strings.EqualFold(filter, recommendation)
}

// matchesHiringDecision maps the evaluation filter onto the candidate status
// stored on a hiring-decision row: a "hire" recommendation corresponds to a
// "hired" status and a "reject" recommendation to a "rejected" status.
func matchesHiringDecision(filter, status string) bool {
	filter = strings.ToLower(strings.TrimSpace(filter))
	if filter == "" {
		return true
	}
	switch filter {
	case "hire":
		return strings.EqualFold(status, "hired") || strings.EqualFold(status, "hire")
	case "reject":
		return strings.EqualFold(status, "rejected") || strings.EqualFold(status, "reject")
	default:
		return strings.EqualFold(filter, status)
	}
}
