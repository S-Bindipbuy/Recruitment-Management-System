package candidate

import (
	"ems/app/enums"
	core "ems/app/handlers/pages/core"
	candmodel "ems/app/models/candidate"
	evalmodel "ems/app/models/evaluation"
	recmodel "ems/app/models/recruitment"
	candsvc "ems/app/services/candidate"
	evalsvc "ems/app/services/evaluation"
	recsvc "ems/app/services/recruitment"
	"errors"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

type CandidateHandler struct{}

type CandidatesPageData struct {
	Candidates            []core.CandidateWithEvaluation
	Postings              []recmodel.JobPostingListItem
	Periods               []evalmodel.EvaluationPeriod
	Templates             []evalmodel.EvaluationTemplate
	Workers               []core.WorkerOption
	InterviewerGroups     []core.InterviewerGroup
	Rounds                []core.InterviewRoundView
	SelectedJobPostingID  int64
	SelectedStatus        string
	CanManageInterviews   bool
	CanScheduleInterviews bool
	// ErrorMessage and Notice are shown at the top of the page. They exist
	// because these handlers return HTTP 200 with a re-rendered page, which the
	// front end's error toast never sees — without them a failed action looks
	// exactly like a successful one.
	ErrorMessage string
	Notice       string
}

func (h *CandidateHandler) Candidates(c *fiber.Ctx) error {
	return h.candidatesPage(c, "", "")
}

func (h *CandidateHandler) candidatesPage(c *fiber.Ctx, errorMessage, notice string) error {
	jobPostingID, _ := strconv.ParseInt(c.Query("job_posting_id"), 10, 64)
	status := strings.TrimSpace(c.Query("status"))

	var candidates []candmodel.CandidateListItem
	var err error
	if jobPostingID > 0 {
		candidates, err = candsvc.CandidateServiceInstance.FindByJobPostingID(jobPostingID)
	} else {
		candidates, err = candsvc.CandidateServiceInstance.FindAll()
	}
	if err != nil {
		candidates = []candmodel.CandidateListItem{}
	}

	if status != "" {
		filtered := candidates[:0]
		for _, cand := range candidates {
			if strings.EqualFold(cand.Status, status) {
				filtered = append(filtered, cand)
			}
		}
		candidates = filtered
	}

	postings, _ := recsvc.JobPostingServiceInst.FindAll()
	periods, _ := evalsvc.PeriodServiceInstance.FindAll()
	templates, _ := evalsvc.TemplateServiceInstance.FindAll()
	workers := core.LoadActiveWorkers()
	interviewers := core.LoadInterviewers()
	if len(interviewers) == 0 {
		interviewers = workers
	}

	uid := core.CurrentUserID(c)
	canManage := core.CanManageInterviews(c)
	canViewAll := core.CanViewAllInterviews(c)
	assignedCandidateIDs, _ := recsvc.InterviewServiceInst.FindAssignedCandidateIDs(uid)
	canSchedule := canManage || (len(assignedCandidateIDs) > 0 && core.HasAnyPermission(c, enums.PermissionEvaluationEvaluate))
	if !canManage && !canViewAll {
		// For scoped users, only show postings they can see
		visible := map[int64]bool{}
		for _, cand := range candidates {
			if cand.JobPostingID != nil {
				visible[*cand.JobPostingID] = true
			}
		}
		filtered := make([]recmodel.JobPostingListItem, 0, len(postings))
		for _, p := range postings {
			if visible[p.ID] {
				filtered = append(filtered, p)
			}
		}
		postings = filtered
	}

	templateNames := map[int64]string{}
	for _, t := range templates {
		templateNames[t.ID] = t.Name
	}
	allRounds := make([]core.InterviewRoundView, 0)
	for _, p := range postings {
		rounds, _ := recsvc.InterviewRoundServiceInst.FindByJobPostingID(p.ID)
		for _, r := range rounds {
			allRounds = append(allRounds, core.InterviewRoundView{
				InterviewRound: r,
				PostingTitle:   p.Title,
				TemplateName:   templateNames[r.EvaluationTemplateID],
			})
		}
	}

	data := CandidatesPageData{
		ErrorMessage:          errorMessage,
		Notice:                notice,
		Postings:              postings,
		Periods:               periods,
		Templates:             templates,
		Workers:               workers,
		InterviewerGroups:     core.GroupInterviewers(interviewers),
		Rounds:                allRounds,
		SelectedJobPostingID:  jobPostingID,
		SelectedStatus:        status,
		CanManageInterviews:   canManage,
		CanScheduleInterviews: canSchedule,
	}

	summaries := core.LoadCandidateSummaryMap()
	for _, candidate := range candidates {
		cwe := core.CandidateWithEvaluation{CandidateListItem: candidate}
		if candidate.JobPostingID != nil {
			cwe.PostingID = *candidate.JobPostingID
		}
		cwe.Interview = summaries[candidate.ID]
		cwe.InterviewReady = core.InterviewReady(cwe.Interview)
		core.AttachCandidateContacts(&cwe)
		core.AttachCandidateEvaluation(&cwe)
		cwe.CanScheduleInterview = data.CanScheduleInterviews
		if cwe.CanScheduleInterview && !canManage && len(assignedCandidateIDs) > 0 {
			allowed := map[int64]bool{}
			for _, id := range assignedCandidateIDs {
				allowed[id] = true
			}
			if !allowed[candidate.ID] {
				cwe.CanScheduleInterview = false
			}
		}
		if cwe.CanScheduleInterview && strings.EqualFold(candidate.Status, "hired") {
			cwe.CanScheduleInterview = false
		}
		data.Candidates = append(data.Candidates, cwe)
	}

	return core.RenderPage(c, "page-candidates", data)
}

func (h *CandidateHandler) CreateCandidate(c *fiber.Ctx) error {
	candidate, err := core.CreateCandidateFromForm(c)
	if err != nil {
		core.TriggerToast(c, err.Error(), "error")
		return h.candidatesPage(c, err.Error(), "")
	}
	if err := candsvc.CandidateServiceInstance.Create(candidate); err != nil {
		_ = os.Remove(candidate.CVPath)
		msg := "Failed to save candidate."
		core.TriggerToast(c, msg, "error")
		return h.candidatesPage(c, msg, "")
	}
	return h.candidatesPage(c, "", "Candidate added.")
}

// LogContact records a contact attempt that happened outside the system
// (phone, email, Telegram, in person). The candidate must still exist, and the
// summary is required because it is the only explanation HR will have later.
func (h *CandidateHandler) LogContact(c *fiber.Ctx) error {
	candidateID, _ := strconv.ParseInt(c.Params("id"), 10, 64)

	fail := func(msg string) error {
		core.TriggerToast(c, msg, "error")
		return h.candidatesPage(c, msg, "")
	}

	if candidateID <= 0 {
		return fail("That candidate no longer exists.")
	}
	candidate, err := candsvc.CandidateServiceInstance.FindByID(candidateID)
	if err != nil {
		return fail("That candidate no longer exists.")
	}

	method := strings.ToLower(strings.TrimSpace(c.FormValue("method")))
	if !validContactChoice(method, candmodel.ContactMethods) {
		return fail("Choose how you contacted the candidate.")
	}
	direction := strings.ToLower(strings.TrimSpace(c.FormValue("direction")))
	if !validContactChoice(direction, candmodel.ContactDirections) {
		return fail("Choose whether you contacted them or they contacted you.")
	}
	outcome := strings.ToLower(strings.TrimSpace(c.FormValue("outcome")))
	if outcome != "" && !validContactChoice(outcome, candmodel.ContactOutcomes) {
		return fail("Choose a valid contact outcome.")
	}
	summary := strings.TrimSpace(c.FormValue("summary"))
	if summary == "" {
		return fail("Add a short summary of the conversation.")
	}

	contactedAt, ok := parseContactTime(c.FormValue("contacted_at"), true)
	if !ok {
		return fail("Choose a valid contact date.")
	}
	var nextFollowUp *time.Time
	if raw := strings.TrimSpace(c.FormValue("next_follow_up_at")); raw != "" {
		next, valid := parseContactTime(raw, false)
		if !valid {
			return fail("Choose a valid follow-up date.")
		}
		nextFollowUp = &next
	}

	contact := &candmodel.ContactLog{
		CandidateID:    candidateID,
		ContactedAt:    contactedAt,
		Method:         method,
		Direction:      direction,
		Summary:        summary,
		Outcome:        outcome,
		NextFollowUpAt: nextFollowUp,
		CreatedByID:    core.CurrentUserID(c),
	}
	if err := candsvc.CandidateServiceInstance.LogContact(contact); err != nil {
		core.TriggerToast(c, "Could not save the contact.", "error")
		return h.candidatesPage(c, "Could not save the contact.", "")
	}

	notice := "Contact logged for " + candidate.FullName + "."
	core.TriggerToast(c, notice, "success")
	return h.candidatesPage(c, "", notice)
}

func validContactChoice(value string, choices []string) bool {
	for _, choice := range choices {
		if choice == value {
			return true
		}
	}
	return false
}

// parseContactTime accepts the two shapes browsers submit: a datetime-local
// value (YYYY-MM-DDTHH:MM) and a date input (YYYY-MM-DD). A missing value means
// "now" for the contact itself; follow-up dates never default to anything.
func parseContactTime(raw string, defaultNow bool) (time.Time, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		if defaultNow {
			return time.Now(), true
		}
		return time.Time{}, false
	}
	for _, layout := range []string{"2006-01-02T15:04", "2006-01-02 15:04", "2006-01-02"} {
		if parsed, err := time.ParseInLocation(layout, raw, time.Local); err == nil {
			return parsed, true
		}
	}
	return time.Time{}, false
}

// UpdateCandidateStatus moves a candidate through the HR pipeline. Rejection
// requires a reason so the reports and history explain why someone was
// dropped. "hired" is not accepted here: it is written by the accept-worker
// flow, which creates the employee record at the same time.
func (h *CandidateHandler) UpdateCandidateStatus(c *fiber.Ctx) error {
	candidateID, _ := strconv.ParseInt(c.Params("id"), 10, 64)
	status := strings.ToLower(strings.TrimSpace(c.FormValue("status")))
	reason := strings.TrimSpace(c.FormValue("reason"))

	fail := func(msg string) error {
		core.TriggerToast(c, msg, "error")
		return h.candidatesPage(c, msg, "")
	}

	if candidateID <= 0 {
		return fail("That candidate no longer exists.")
	}

	valid := false
	for _, s := range candmodel.PipelineStatuses {
		if s == status {
			valid = true
			break
		}
	}
	if !valid {
		if status == "hired" {
			return fail("Use the Accept worker action to mark someone as hired.")
		}
		return fail("Choose a valid pipeline status.")
	}

	candidate, err := candsvc.CandidateServiceInstance.FindByID(candidateID)
	if err != nil {
		return fail("That candidate no longer exists.")
	}
	if candidate.Status == "hired" {
		return fail("This candidate has already been accepted as an employee.")
	}
	if status == "rejected" && reason == "" {
		return fail("Add a rejection reason before rejecting this candidate.")
	}

	if err := candsvc.CandidateServiceInstance.SetStatus(candidateID, status, reason, core.CurrentUserID(c)); err != nil {
		core.TriggerToast(c, "Could not update the candidate status.", "error")
		return h.candidatesPage(c, "Could not update the candidate status.", "")
	}

	notice := candidate.FullName + " is now " + status + "."
	core.TriggerToast(c, notice, "success")
	return h.candidatesPage(c, "", notice)
}

func (h *CandidateHandler) AssignCandidateEvaluation(c *fiber.Ctx) error {
	candidateID, _ := strconv.ParseInt(c.Params("id"), 10, 64)

	form := core.ReadAssignmentForm(c)
	form.CandidateID = candidateID
	periodID, templateID := form.PeriodID, form.TemplateID
	evaluatorID := form.EvaluatorID
	dueDate := core.DateFromForm(form.DueDate)

	if candidateID <= 0 || periodID <= 0 || templateID <= 0 || evaluatorID <= 0 {
		msg := "Please choose a period, a template and an evaluator."
		core.TriggerToast(c, msg, "error")
		return h.candidatesPage(c, msg, "")
	}

	// Same rules as the dialog and the JSON API: an active worker, an active
	// template, an applicant who has not been hired, and not twice.
	if _, err := evalsvc.AssignmentServiceInstance.CreateForCandidate(periodID, templateID, evaluatorID, candidateID, dueDate, core.CurrentUserID(c)); err != nil {
		log.Printf("[AssignEval] error: %v", err)
		msg := "Could not assign the evaluation. Check that the period and template still exist."
		if evalsvc.AssignmentRuleError(err) {
			msg = err.Error()
		}
		core.TriggerToast(c, msg, "error")
		return h.candidatesPage(c, msg, "")
	}
	return h.candidatesPage(c, "", "Evaluation assigned.")
}

// ScheduleCandidateInterview creates one interview for the candidate straight
// from their row. It is the same work the Interviews page does for a group:
// the chosen round supplies the evaluation template, the panel members become
// the evaluators, and CreateForInterview creates one pending evaluation per
// panel member for that round.
//
// The response is JSON so the caller can stay on whichever candidate list it
// used (Candidates or Job Posting Candidates) and refresh that page.
func (h *CandidateHandler) ScheduleCandidateInterview(c *fiber.Ctx) error {
	candidateID, _ := strconv.ParseInt(c.Params("id"), 10, 64)
	uid := core.CurrentUserID(c)

	if !core.CanManageInterviews(c) {
		if !core.HasAnyPermission(c, enums.PermissionEvaluationEvaluate) {
			return core.RenderError(c, fiber.StatusForbidden, "You do not have permission to schedule interviews.")
		}
		assigned, _ := recsvc.InterviewServiceInst.FindAssignedCandidateIDs(uid)
		allowed := false
		for _, id := range assigned {
			if id == candidateID {
				allowed = true
				break
			}
		}
		if !allowed {
			return core.RenderError(c, fiber.StatusForbidden, "You can only schedule interviews for candidates assigned to you.")
		}
	}

	candidate, err := candsvc.CandidateServiceInstance.FindByID(candidateID)
	if err != nil {
		return core.RenderError(c, fiber.StatusNotFound, "Candidate not found.")
	}
	if strings.EqualFold(candidate.Status, "hired") {
		return core.RenderError(c, fiber.StatusBadRequest, "This candidate has already been hired.")
	}

	roundID := core.FormInt64(c, "interview_round_id")
	interviewerIDs := core.FormInt64s(c, "interviewer_ids")
	if roundID <= 0 || len(interviewerIDs) == 0 {
		return core.RenderError(c, fiber.StatusBadRequest, "Choose an interview round and at least one interviewer.")
	}
	if err := recsvc.InterviewServiceInst.ValidateCandidateRound(candidateID, roundID); err != nil {
		if errors.Is(err, recsvc.ErrRoundNotForCandidate) {
			return core.RenderError(c, fiber.StatusBadRequest, "The selected round does not belong to this candidate's job posting.")
		}
		return core.RenderError(c, fiber.StatusBadRequest, "Candidate or interview round not found.")
	}

	var scheduled time.Time
	for _, format := range []string{"2006-01-02T15:04", "2006-01-02 15:04"} {
		if t, err := time.Parse(format, c.FormValue("scheduled_at")); err == nil {
			scheduled = t
			break
		}
	}
	if scheduled.IsZero() {
		return core.RenderError(c, fiber.StatusBadRequest, "The interview date is not valid.")
	}

	duration, _ := strconv.Atoi(c.FormValue("duration_minutes"))
	if duration <= 0 {
		duration = 60
	}

	interview := &recmodel.Interview{
		CandidateID:      candidateID,
		InterviewRoundID: roundID,
		InterviewerID:    interviewerIDs[0],
		InterviewerIDs:   interviewerIDs,
		ScheduledAt:      scheduled,
		DurationMinutes:  duration,
		Location:         strings.TrimSpace(c.FormValue("location")),
		MeetingLink:      strings.TrimSpace(c.FormValue("meeting_link")),
		Status:           "scheduled",
		Notes:            strings.TrimSpace(c.FormValue("notes")),
	}
	if err := recsvc.InterviewServiceInst.Create(interview); err != nil {
		log.Printf("[ScheduleCandidateInterview] create candidate %d: %v", candidateID, err)
		return core.RenderError(c, fiber.StatusInternalServerError, "Failed to schedule interview.")
	}
	// A failure to advance the pipeline or create an evaluation must not undo
	// the scheduled interview; those are best-effort and can be retried.
	if err := candsvc.CandidateServiceInstance.MarkInterviewing(candidateID, uid); err != nil {
		log.Printf("[ScheduleCandidateInterview] could not advance candidate %d: %v", candidateID, err)
	}
	if _, err := evalsvc.AssignmentServiceInstance.CreateForInterview(interview.ID, core.FormInt64(c, "evaluation_template_id")); err != nil {
		log.Printf("[ScheduleCandidateInterview] round evaluation for interview %d: %v", interview.ID, err)
	}

	return c.JSON(fiber.Map{"message": "Interview scheduled."})
}

