package recruitment

import (
	core "ems/app/handlers/pages/core"
	candmodel "ems/app/models/candidate"
	evalmodel "ems/app/models/evaluation"
	orgmodel "ems/app/models/organization"
	recmodel "ems/app/models/recruitment"
	"ems/app/services"
	candsvc "ems/app/services/candidate"
	evalsvc "ems/app/services/evaluation"
	orgsvc "ems/app/services/organization"
	recsvc "ems/app/services/recruitment"
	"os"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type RecruitmentHandler struct{}

func (h *RecruitmentHandler) JobPostings(c *fiber.Ctx) error {
	postings, _ := recsvc.JobPostingServiceInst.FindAll()
	positions, _ := orgsvc.PositionServiceInst.FindAll()
	return core.RenderPage(c, "page-job-postings", struct {
		Postings  []recmodel.JobPostingListItem
		Positions []orgmodel.Position
	}{Postings: postings, Positions: positions})
}

func (h *RecruitmentHandler) CreateJobPosting(c *fiber.Ctx) error {
	positionID := core.FormInt64(c, "position_id")
	title := strings.TrimSpace(c.FormValue("title"))
	description := strings.TrimSpace(c.FormValue("description"))
	requirements := strings.TrimSpace(c.FormValue("requirements"))
	skillsRequired := strings.TrimSpace(c.FormValue("skills_required"))
	employmentType := c.FormValue("employment_type")
	if employmentType == "" {
		employmentType = "full-time"
	}

	var salaryMin, salaryMax *float64
	if v := c.FormValue("salary_min"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			salaryMin = &f
		}
	}
	if v := c.FormValue("salary_max"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			salaryMax = &f
		}
	}

	userID := core.CurrentUserID(c)
	posting := &recmodel.JobPosting{
		PositionID:     positionID,
		Title:          title,
		Description:    description,
		Requirements:   requirements,
		SkillsRequired: skillsRequired,
		SalaryMin:      salaryMin,
		SalaryMax:      salaryMax,
		EmploymentType: employmentType,
		Status:         "draft",
		CreatedBy:      userID,
	}

	if err := recsvc.JobPostingServiceInst.Create(posting); err != nil {
		return core.RenderError(c, fiber.StatusInternalServerError, "Failed to create job posting.")
	}
	return h.JobPostings(c)
}

func (h *RecruitmentHandler) PublishJobPosting(c *fiber.Ctx) error {
	id, _ := strconv.ParseInt(c.Params("id"), 10, 64)
	if err := recsvc.JobPostingServiceInst.Publish(id); err != nil {
		return core.RenderError(c, fiber.StatusInternalServerError, "Failed to publish.")
	}
	return h.JobPostings(c)
}

func (h *RecruitmentHandler) CloseJobPosting(c *fiber.Ctx) error {
	id, _ := strconv.ParseInt(c.Params("id"), 10, 64)
	if err := recsvc.JobPostingServiceInst.Close(id); err != nil {
		return core.RenderError(c, fiber.StatusInternalServerError, "Failed to close.")
	}
	return h.JobPostings(c)
}

func (h *RecruitmentHandler) ReopenJobPosting(c *fiber.Ctx) error {
	id, _ := strconv.ParseInt(c.Params("id"), 10, 64)
	if err := recsvc.JobPostingServiceInst.Reopen(id); err != nil {
		return core.RenderError(c, fiber.StatusInternalServerError, "Failed to reopen.")
	}
	return h.JobPostings(c)
}

func (h *RecruitmentHandler) DeleteJobPosting(c *fiber.Ctx) error {
	id, _ := strconv.ParseInt(c.Params("id"), 10, 64)
	if err := recsvc.JobPostingServiceInst.Delete(id); err != nil {
		return core.RenderError(c, fiber.StatusInternalServerError, "Failed to delete.")
	}
	return h.JobPostings(c)
}

func (h *RecruitmentHandler) JobPostingCandidates(c *fiber.Ctx) error {
	postingID, _ := strconv.ParseInt(c.Params("id"), 10, 64)
	posting, err := recsvc.JobPostingServiceInst.FindByID(postingID)
	if err != nil {
		return core.RenderError(c, fiber.StatusNotFound, "Job posting not found.")
	}
	candidates, _ := candsvc.CandidateServiceInstance.FindByJobPostingID(postingID)
	periods, _ := evalsvc.PeriodServiceInstance.FindAll()
	templates, _ := evalsvc.TemplateServiceInstance.FindAll()
	workers := core.LoadActiveWorkers()

	var candidatesWithEval []core.CandidateWithEvaluation
	for _, candidate := range candidates {
		cwe := core.CandidateWithEvaluation{CandidateListItem: candidate}
		assignments, _ := evalsvc.AssignmentServiceInstance.FindByCandidate(candidate.ID)
		if len(assignments) > 0 {
			cwe.HasEvaluation = true
			for _, a := range assignments {
				if a.Status == "completed" || a.Status == "submitted" {
					cwe.EvaluationCompleted = true
				} else if a.Status == "pending" {
					cwe.EvaluationPending = true
				}
			}
		}
		candidatesWithEval = append(candidatesWithEval, cwe)
	}

	return core.RenderPage(c, "page-job-posting-candidates", struct {
		Posting    *recmodel.JobPosting
		Candidates []core.CandidateWithEvaluation
		Periods    []evalmodel.EvaluationPeriod
		Templates  []evalmodel.EvaluationTemplate
		Workers    []core.WorkerOption
	}{Posting: posting, Candidates: candidatesWithEval, Periods: periods, Templates: templates, Workers: workers})
}

// Interview Scheduling Page
func (h *RecruitmentHandler) Interviews(c *fiber.Ctx) error {
	interviews, _ := recsvc.InterviewServiceInst.FindAll()
	candidates, _ := candsvc.CandidateServiceInstance.FindAll()
	workers := core.LoadActiveWorkers()
	postings, _ := recsvc.JobPostingServiceInst.FindAll()

	var allRounds []recmodel.InterviewRound
	for _, p := range postings {
		rounds, _ := recsvc.InterviewRoundServiceInst.FindByJobPostingID(p.ID)
		allRounds = append(allRounds, rounds...)
	}

	return core.RenderPage(c, "page-interviews", struct {
		Interviews []recmodel.InterviewListItem
		Candidates []candmodel.CandidateListItem
		Workers    []core.WorkerOption
		Rounds     []recmodel.InterviewRound
	}{Interviews: interviews, Candidates: candidates, Workers: workers, Rounds: allRounds})
}

func (h *RecruitmentHandler) CreateInterview(c *fiber.Ctx) error {
	candidateID := core.FormInt64(c, "candidate_id")
	roundID := core.FormInt64(c, "interview_round_id")
	interviewerID := core.FormInt64(c, "interviewer_id")
	scheduledAt := c.FormValue("scheduled_at")
	duration, _ := strconv.Atoi(c.FormValue("duration_minutes"))
	location := strings.TrimSpace(c.FormValue("location"))
	meetingLink := strings.TrimSpace(c.FormValue("meeting_link"))
	notes := strings.TrimSpace(c.FormValue("notes"))

	if candidateID <= 0 || roundID <= 0 || interviewerID <= 0 || scheduledAt == "" {
		return core.RenderError(c, fiber.StatusBadRequest, "Candidate, round, interviewer, and date are required.")
	}

	if duration <= 0 {
		duration = 60
	}

	interview := &recmodel.Interview{
		CandidateID:      candidateID,
		InterviewRoundID: roundID,
		InterviewerID:    interviewerID,
		DurationMinutes:  duration,
		Location:         location,
		MeetingLink:      meetingLink,
		Status:           "scheduled",
		Notes:            notes,
	}

	// Parse date
	for _, format := range []string{"2006-01-02T15:04", "2006-01-02 15:04"} {
		if t, err := services.ParseTime(scheduledAt, format); err == nil {
			interview.ScheduledAt = t
			break
		}
	}

	if err := recsvc.InterviewServiceInst.Create(interview); err != nil {
		return core.RenderError(c, fiber.StatusInternalServerError, "Failed to schedule interview.")
	}
	return h.Interviews(c)
}

func (h *RecruitmentHandler) UpdateInterviewStatus(c *fiber.Ctx) error {
	id, _ := strconv.ParseInt(c.Params("id"), 10, 64)
	status := c.FormValue("status")
	if err := recsvc.InterviewServiceInst.UpdateStatus(id, status); err != nil {
		return core.RenderError(c, fiber.StatusInternalServerError, "Failed to update status.")
	}
	return h.Interviews(c)
}

func (h *RecruitmentHandler) DeleteInterview(c *fiber.Ctx) error {
	id, _ := strconv.ParseInt(c.Params("id"), 10, 64)
	if err := recsvc.InterviewServiceInst.Delete(id); err != nil {
		return core.RenderError(c, fiber.StatusInternalServerError, "Failed to delete interview.")
	}
	return h.Interviews(c)
}

// Interview Feedback Page
func (h *RecruitmentHandler) InterviewFeedback(c *fiber.Ctx) error {
	interviewID, _ := strconv.ParseInt(c.Params("id"), 10, 64)
	interview, err := recsvc.InterviewServiceInst.FindByID(interviewID)
	if err != nil {
		return core.RenderError(c, fiber.StatusNotFound, "Interview not found.")
	}

	fb, _ := recsvc.InterviewFeedbackServiceInst.FindByInterviewID(interviewID)

	return core.RenderPage(c, "page-interview-feedback", struct {
		Interview *recmodel.InterviewListItem
		Feedback  *recmodel.InterviewFeedback
	}{Interview: interview, Feedback: fb})
}

func (h *RecruitmentHandler) SubmitInterviewFeedback(c *fiber.Ctx) error {
	interviewID, _ := strconv.ParseInt(c.Params("id"), 10, 64)
	score, _ := strconv.ParseFloat(c.FormValue("score"), 64)
	recommendation := c.FormValue("recommendation")
	strengths := strings.TrimSpace(c.FormValue("strengths"))
	weaknesses := strings.TrimSpace(c.FormValue("weaknesses"))
	comments := strings.TrimSpace(c.FormValue("comments"))

	userID := core.CurrentUserID(c)

	fb := &recmodel.InterviewFeedback{
		InterviewID:    interviewID,
		InterviewerID:  userID,
		Score:          &score,
		Recommendation: recommendation,
		Strengths:      strengths,
		Weaknesses:     weaknesses,
		Comments:       comments,
	}

	existing, err := recsvc.InterviewFeedbackServiceInst.FindByInterviewID(interviewID)
	if err == nil && existing.ID > 0 {
		fb.ID = existing.ID
		fb.SubmittedAt = existing.SubmittedAt
		recsvc.InterviewFeedbackServiceInst.Update(fb)
	} else {
		recsvc.InterviewFeedbackServiceInst.Create(fb)
	}

	return h.InterviewFeedback(c)
}

// Reports Page
func (h *RecruitmentHandler) RecruitmentReports(c *fiber.Ctx) error {
	summaries, _ := recsvc.InterviewServiceInst.GetAllCandidateSummaries()
	activities, _ := recsvc.InterviewServiceInst.GetInterviewerActivity()
	decisions, _ := recsvc.InterviewServiceInst.GetHiringDecisions()

	return core.RenderPage(c, "page-recruitment-reports", struct {
		Summaries  []recmodel.CandidateEvaluationSummary
		Activities []recmodel.InterviewerActivity
		Decisions  []recmodel.HiringDecision
	}{Summaries: summaries, Activities: activities, Decisions: decisions})
}

// Suppress unused import
var _ = os.Remove
