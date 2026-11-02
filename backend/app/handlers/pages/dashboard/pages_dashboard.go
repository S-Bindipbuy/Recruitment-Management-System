package dashboard

import (
	core "ems/app/handlers/pages/core"
	candmodel "ems/app/models/candidate"
	recmodel "ems/app/models/recruitment"
	candsvc "ems/app/services/candidate"
	recsvc "ems/app/services/recruitment"
	"sort"
	"time"

	"github.com/gofiber/fiber/v2"
)

type DashboardHandler struct{}

type dashboardPageData struct {
	Stats          core.DashboardStats
	FollowUps      []candmodel.ContactLog
	AwaitingScores []recmodel.InterviewListItem
	Rankings       []recmodel.CandidateEvaluationSummary
}

// dashboardRankingLimit keeps the dashboard card a shortlist. The full, filtered
// ranking lives on the Interviews and Reports pages.
const dashboardRankingLimit = 5

func (h *DashboardHandler) Dashboard(c *fiber.Ctx) error {
	data := dashboardPageData{
		Stats:          core.LoadDashboardStats(core.CurrentUserID(c)),
		FollowUps:      loadFollowUps(),
		AwaitingScores: loadAwaitingScores(),
	}

	// The combined ranking is decision material, so only the roles that may
	// review or approve interviews see it on their dashboard.
	if core.CanApproveInterviews(c) {
		rankings := core.LoadCandidateRankings()
		sort.SliceStable(rankings, func(i, j int) bool {
			if rankings[i].AverageScore != rankings[j].AverageScore {
				return rankings[i].AverageScore > rankings[j].AverageScore
			}
			return rankings[i].CandidateName < rankings[j].CandidateName
		})
		if len(rankings) > dashboardRankingLimit {
			rankings = rankings[:dashboardRankingLimit]
		}
		data.Rankings = rankings
	}

	return core.RenderPage(c, "page-dashboard", data)
}

// loadFollowUps keeps only the latest contact per candidate and returns those
// whose follow-up falls within the next week (overdue ones included). An older
// follow-up is superseded by a newer contact that did not schedule one.
func loadFollowUps() []candmodel.ContactLog {
	contacts, err := candsvc.CandidateServiceInstance.FindActiveContacts()
	if err != nil {
		return nil
	}

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	seen := make(map[int64]bool, len(contacts))
	followUps := make([]candmodel.ContactLog, 0)
	for _, contact := range contacts {
		if seen[contact.CandidateID] {
			continue
		}
		seen[contact.CandidateID] = true
		if contact.NextFollowUpAt == nil {
			continue
		}
		due := *contact.NextFollowUpAt
		dueDay := time.Date(due.Year(), due.Month(), due.Day(), 0, 0, 0, 0, now.Location())
		days := int(dueDay.Sub(today).Hours() / 24)
		if days > 7 {
			continue
		}
		contact.DueInDays = days
		contact.Overdue = days < 0
		followUps = append(followUps, contact)
	}
	return followUps
}

// loadAwaitingScores lists interviews whose time has passed but that still have
// no score recorded, so HR does not lose track of the interviewer's feedback.
func loadAwaitingScores() []recmodel.InterviewListItem {
	interviews, err := recsvc.InterviewServiceInst.FindAwaitingFeedback()
	if err != nil {
		return nil
	}
	now := time.Now()
	awaiting := make([]recmodel.InterviewListItem, 0)
	for _, interview := range interviews {
		if interview.ScheduledAt.Before(now) {
			awaiting = append(awaiting, interview)
		}
	}
	return awaiting
}
