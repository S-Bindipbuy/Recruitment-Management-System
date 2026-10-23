package candidate

import (
	core "ems/app/handlers/pages/core"
	candmodel "ems/app/models/candidate"
	evalmodel "ems/app/models/evaluation"
	recmodel "ems/app/models/recruitment"
	candsvc "ems/app/services/candidate"
	evalsvc "ems/app/services/evaluation"
	recsvc "ems/app/services/recruitment"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

type CandidateHandler struct{}

type CandidatesPageData struct {
	Candidates           []core.CandidateWithEvaluation
	Postings             []recmodel.JobPostingListItem
	Periods              []evalmodel.EvaluationPeriod
	Templates            []evalmodel.EvaluationTemplate
	Workers              []core.WorkerOption
	SelectedJobPostingID int64
	SelectedStatus       string
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

	data := CandidatesPageData{
		ErrorMessage:         errorMessage,
		Notice:               notice,
		Postings:             postings,
		Periods:              periods,
		Templates:            templates,
		Workers:              workers,
		SelectedJobPostingID: jobPostingID,
		SelectedStatus:       status,
	}

	for _, candidate := range candidates {
		cwe := core.CandidateWithEvaluation{CandidateListItem: candidate}
		assignments, _ := evalsvc.AssignmentServiceInstance.FindByCandidate(candidate.ID)
		if len(assignments) > 0 {
			cwe.HasEvaluation = true
			for _, a := range assignments {
				if a.Status == "completed" {
					cwe.EvaluationCompleted = true
				} else if a.Status == "pending" || a.Status == "started" {
					cwe.EvaluationPending = true
				}
			}
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
	if _, err := evalsvc.AssignmentServiceInstance.CreateForCandidate(periodID, templateID, evaluatorID, candidateID, dueDate); err != nil {
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

// hireErrorMessage turns the repository's raw errors into something the person
// hiring can act on. The underlying causes were only ever written to the log.
func hireErrorMessage(err error) string {
	msg := err.Error()
	switch {
	case strings.Contains(strings.ToLower(msg), "already exists"):
		return "No worker was created: a user with this email already exists. Use a different email address for the candidate."
	case strings.Contains(strings.ToLower(msg), "already been accepted"):
		return "This candidate has already been accepted, so there is nothing left to create."
	case strings.Contains(strings.ToLower(msg), "not found"):
		return "That candidate no longer exists."
	default:
		return "No worker was created: " + msg
	}
}

func (h *CandidateHandler) HireCandidate(c *fiber.Ctx) error {
	candidateID, _ := strconv.ParseInt(c.Params("id"), 10, 64)
	rating, ratingErr := strconv.ParseFloat(c.FormValue("rating"), 64)
	temporaryPassword := c.FormValue("temporary_password")

	fail := func(msg string) error {
		core.TriggerToast(c, msg, "error")
		return h.candidatesPage(c, msg, "")
	}

	if candidateID <= 0 {
		return fail("That candidate no longer exists.")
	}
	if ratingErr != nil || rating < 1 || rating > 5 {
		return fail("The interview rating has to be a number between 1 and 5.")
	}
	if len(temporaryPassword) < 8 {
		return fail("The temporary password needs at least 8 characters.")
	}

	uid := core.CurrentUserID(c)
	if uid <= 0 {
		log.Printf("[Hire] core.CurrentUserID returned %d", uid)
		return fail("Your session has expired. Sign in again and retry.")
	}

	// Check if candidate has a completed evaluation
	assignments, _ := evalsvc.AssignmentServiceInstance.FindByCandidate(candidateID)
	hasCompleted := false
	for _, a := range assignments {
		if a.Status == "completed" || a.Status == "submitted" {
			hasCompleted = true
			break
		}
	}
	if !hasCompleted {
		log.Printf("[Hire] candidate %d has no completed evaluation", candidateID)
		return fail("This candidate needs a completed evaluation before they can be accepted.")
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(temporaryPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("[Hire] bcrypt error: %v", err)
		return fail("That password could not be prepared. Try a different one.")
	}

	name := ""
	if cand, cerr := candsvc.CandidateServiceInstance.FindByID(candidateID); cerr == nil {
		name = cand.FullName
	}
	if err := candsvc.CandidateServiceInstance.Hire(candidateID, uid, 0, rating, c.FormValue("notes"), string(passwordHash)); err != nil {
		log.Printf("[Hire] error hiring candidate %d: %v", candidateID, err)
		return fail(hireErrorMessage(err))
	}

	notice := "Worker account created."
	if name != "" {
		notice = name + " is now an active worker — they appear on the Employees page."
	}
	core.TriggerToast(c, notice, "success")
	return h.candidatesPage(c, "", notice)
}
