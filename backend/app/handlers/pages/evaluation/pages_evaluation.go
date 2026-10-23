package evaluation

import (
	"ems/app/enums"
	core "ems/app/handlers/pages/core"
	evalmodel "ems/app/models/evaluation"
	candsvc "ems/app/services/candidate"
	evalsvc "ems/app/services/evaluation"
	usersvc "ems/app/services/user"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// loadHireableCandidates returns applicants that are still in the pipeline -
// hired candidates have left it and can no longer be evaluated.
func loadHireableCandidates() []CandidateOption {
	options, _ := candsvc.CandidateServiceInstance.FindHireableOptions()
	return options
}

func (h *EvaluationHandler) Evaluations(c *fiber.Ctx) error {
	if err := core.RequirePermission(c,
		enums.PermissionEvaluationEvaluate,
		enums.PermissionEvaluationAssign,
		enums.PermissionEvaluationManage,
	); err != nil {
		return err
	}

	data := EvaluationsPageData{}
	data.Periods, _ = evalsvc.PeriodServiceInstance.FindAll()
	data.Templates, _ = evalsvc.TemplateServiceInstance.FindAll()
	data.Workers = core.LoadActiveWorkers()
	data.Candidates = loadHireableCandidates()

	uid := core.CurrentUserID(c)
	var assignments []evalmodel.EvaluationAssignment
	if usersvc.PolicyServiceInstance.CanAny(uid, enums.PermissionEvaluationManage, enums.PermissionEvaluationAssign) {
		assignments, _ = evalsvc.AssignmentServiceInstance.FindAll()
	} else {
		byEvaluator, _ := evalsvc.AssignmentServiceInstance.FindByEvaluator(uid)
		byEvaluatee, _ := evalsvc.AssignmentServiceInstance.FindByEvaluatee(uid)
		seen := make(map[int64]bool)
		for _, a := range byEvaluator {
			if !seen[a.ID] {
				seen[a.ID] = true
				assignments = append(assignments, a)
			}
		}
		for _, a := range byEvaluatee {
			if !seen[a.ID] {
				seen[a.ID] = true
				assignments = append(assignments, a)
			}
		}
	}

	evaluateeIDs := make([]int64, 0, len(assignments))
	candidateIDs := make([]int64, 0, len(assignments))
	evaluatorIDs := make([]int64, 0, len(assignments))
	for _, a := range assignments {
		if a.CandidateID != nil && *a.CandidateID > 0 {
			candidateIDs = append(candidateIDs, *a.CandidateID)
		} else {
			evaluateeIDs = append(evaluateeIDs, a.EvaluateeID)
		}
		if a.EvaluatorID > 0 {
			evaluatorIDs = append(evaluatorIDs, a.EvaluatorID)
		}
	}
	names := core.UserNames(evaluateeIDs)
	candidateNames := core.CandidateNames(candidateIDs)
	evaluatorNames := core.UserNames(evaluatorIDs)

	for _, a := range assignments {
		evaluateeName := ""
		if a.CandidateID != nil && *a.CandidateID > 0 {
			evaluateeName = candidateNames[*a.CandidateID] + " (Candidate)"
		} else {
			evaluateeName = core.DisplayName(names, a.EvaluateeID)
		}
		view := AssignmentView{
			Assignment:    a,
			EvaluateeName: evaluateeName,
			EvaluatorName: core.DisplayName(evaluatorNames, a.EvaluatorID),
		}
		if a.Status == string(enums.AssignmentStatusCompleted) || a.Status == string(enums.AssignmentStatusSubmitted) {
			view.Result, _ = evalsvc.ResultServiceInstance.GetResult(a.ID)
		}
		data.Assignments = append(data.Assignments, view)
	}
	return core.RenderPage(c, "page-evaluations", data)
}

func (h *EvaluationHandler) CreateAssignment(c *fiber.Ctx) error {
	if err := core.RequirePermission(c, enums.PermissionEvaluationAssign, enums.PermissionEvaluationManage); err != nil {
		return err
	}

	form := core.ReadAssignmentForm(c)
	periodID, templateID := form.PeriodID, form.TemplateID
	evaluatorID, candidateID := form.EvaluatorID, form.CandidateID
	dueDate := core.DateFromForm(form.DueDate)

	if periodID <= 0 || templateID <= 0 || evaluatorID <= 0 {
		return core.RenderError(c, fiber.StatusBadRequest, "Please fill in all required fields.")
	}
	if candidateID <= 0 {
		return core.RenderError(c, fiber.StatusBadRequest, "Choose the candidate to evaluate.")
	}

	// The rules themselves live in the service, so this dialog cannot accept
	// something the JSON API would turn away.
	_, err := evalsvc.AssignmentServiceInstance.CreateForCandidate(periodID, templateID, evaluatorID, candidateID, dueDate)
	if err != nil {
		status := fiber.StatusBadRequest
		if !evalsvc.AssignmentRuleError(err) {
			status = fiber.StatusInternalServerError
		}
		return core.RenderError(c, status, assignmentDialogMessage(err))
	}
	return h.Evaluations(c)
}

// assignmentDialogMessage keeps this dialog's wording. The rules are the
// service's; only the copy differs from one surface to the next.
func assignmentDialogMessage(err error) string {
	switch {
	case errors.Is(err, evalsvc.ErrAssignmentIncomplete):
		return "Please fill in all required fields."
	case errors.Is(err, evalsvc.ErrCandidateHired):
		return "This candidate has already been accepted as a worker, so they can no longer be evaluated from here."
	case errors.Is(err, evalsvc.ErrAlreadyAssigned):
		return "This candidate already has an evaluation for that period and template."
	default:
		return err.Error()
	}
}

func (h *EvaluationHandler) StartAssignment(c *fiber.Ctx) error {
	id, _ := c.ParamsInt("id")
	assignment, err := evalsvc.AssignmentServiceInstance.FindByID(int64(id))
	if err != nil {
		return core.RenderError(c, fiber.StatusNotFound, "Assignment not found.")
	}
	if err := core.RequireCanFillEvaluation(c, assignment); err != nil {
		return err
	}
	if err := evalsvc.AssignmentServiceInstance.Start(int64(id)); err != nil {
		return core.RenderError(c, fiber.StatusInternalServerError, "Failed to start assignment.")
	}
	return h.Evaluations(c)
}

// ── Evaluation fill (fill out / submit an evaluation) ──

func commentViews(comments []evalmodel.EvaluationComment) []CommentView {
	authorIDs := make([]int64, 0, len(comments))
	for _, cm := range comments {
		authorIDs = append(authorIDs, cm.UserID)
	}
	names := core.UserNames(authorIDs)

	views := make([]CommentView, 0, len(comments))
	for _, cm := range comments {
		views = append(views, CommentView{Comment: cm, AuthorName: core.DisplayName(names, cm.UserID)})
	}
	return views
}

func (h *EvaluationHandler) fillData(c *fiber.Ctx) (*FillPageData, error) {
	id, _ := c.ParamsInt("id")
	assignment, err := evalsvc.AssignmentServiceInstance.FindByID(int64(id))
	if err != nil {
		return nil, fiber.NewError(fiber.StatusNotFound, "Assignment not found")
	}

	evaluateeName := ""
	if assignment.CandidateID != nil && *assignment.CandidateID > 0 {
		candidate, err := candsvc.CandidateServiceInstance.FindByID(*assignment.CandidateID)
		if err == nil {
			evaluateeName = candidate.FullName + " (Candidate)"
		} else {
			evaluateeName = "Candidate"
		}
	} else {
		evaluateeName = core.UserName(assignment.EvaluateeID)
	}

	data := &FillPageData{Assignment: *assignment, EvaluateeName: evaluateeName}

	if assignment.Status == string(enums.AssignmentStatusStarted) ||
		assignment.Status == string(enums.AssignmentStatusPending) {

		sections, _ := evalsvc.SectionServiceInstance.FindByTemplateID(assignment.EvaluationTemplateID)
		for _, sec := range sections {
			view := SectionView{Name: sec.Name, Description: sec.Description}
			questions, _ := evalsvc.QuestionServiceInstance.FindBySectionID(sec.ID)
			for _, q := range questions {
				view.Questions = append(view.Questions, QuestionView{
					Model:    q,
					Required: q.IsRequired && q.Status == "active",
				})
			}
			data.Sections = append(data.Sections, view)
		}
	} else {
		data.Result, _ = evalsvc.ResultServiceInstance.GetResult(assignment.ID)
	}

	comments, _ := evalsvc.CommentServiceInstance.FindByAssignmentID(assignment.ID)
	data.Comments = commentViews(comments)
	return data, nil
}

func (h *EvaluationHandler) EvaluationFill(c *fiber.Ctx) error {
	data, err := h.fillData(c)
	if err != nil {
		return core.RenderError(c, fiber.StatusNotFound, "Evaluation not found.")
	}
	if err := core.RequireCanViewEvaluation(c, &data.Assignment); err != nil {
		return err
	}
	return core.RenderPage(c, "page-evaluation-fill", data)
}

func (h *EvaluationHandler) StartFromFill(c *fiber.Ctx) error {
	id, _ := c.ParamsInt("id")
	assignment, err := evalsvc.AssignmentServiceInstance.FindByID(int64(id))
	if err != nil {
		return core.RenderError(c, fiber.StatusNotFound, "Assignment not found.")
	}
	if err := core.RequireCanFillEvaluation(c, assignment); err != nil {
		return err
	}
	if err := evalsvc.AssignmentServiceInstance.Start(int64(id)); err != nil {
		return core.RenderError(c, fiber.StatusInternalServerError, "Failed to start evaluation.")
	}
	return h.EvaluationFill(c)
}

func (h *EvaluationHandler) SubmitEvaluation(c *fiber.Ctx) error {
	id, _ := c.ParamsInt("id")
	assignmentID := int64(id)

	assignment, err := evalsvc.AssignmentServiceInstance.FindByID(assignmentID)
	if err != nil {
		return core.RenderError(c, fiber.StatusNotFound, "Evaluation not found.")
	}
	if err := core.RequireCanFillEvaluation(c, assignment); err != nil {
		return err
	}

	sections, _ := evalsvc.SectionServiceInstance.FindByTemplateID(assignment.EvaluationTemplateID)
	for _, sec := range sections {
		questions, _ := evalsvc.QuestionServiceInstance.FindBySectionID(sec.ID)
		for _, q := range questions {
			if err := saveQuestionResponse(assignmentID, q, c.FormValue(fmt.Sprintf("q_%d", q.ID))); err != nil {
				// Stop before submitting: a half-saved evaluation would be
				// scored as though the missing answers were never given.
				log.Printf("[SubmitEvaluation] answer not saved for assignment %d question %d: %v", assignmentID, q.ID, err)
				return core.RenderError(c, fiber.StatusInternalServerError,
					"Your answers could not be saved, so nothing was submitted. Please try again.")
			}
		}
	}

	if err := evalsvc.AssignmentServiceInstance.Submit(assignmentID); err != nil {
		return core.RenderError(c, fiber.StatusInternalServerError, "Failed to submit evaluation.")
	}
	_, _ = evalsvc.ResultServiceInstance.CalculateAndSave(assignmentID)

	c.Set("HX-Push-Url", "/pages/evaluations")
	return h.Evaluations(c)
}

// saveQuestionResponse stores one answer. Failures are returned rather than
// discarded: an answer that did not save only shows up much later as a score
// of zero, with nothing to point at.
func saveQuestionResponse(assignmentID int64, q evalmodel.EvaluationQuestion, raw string) error {
	if raw == "" {
		return nil
	}
	switch enums.QuestionType(q.QuestionType) {
	case enums.QuestionTypeText:
		return evalsvc.ResponseServiceInstance.SaveResponse(assignmentID, q.ID, nil, raw, nil, nil, nil)
	case enums.QuestionTypeNumber:
		v, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return fmt.Errorf("%q is not a number", raw)
		}
		return evalsvc.ResponseServiceInstance.SaveResponse(assignmentID, q.ID, nil, "", &v, nil, nil)
	case enums.QuestionTypeBoolean:
		b := raw == "true"
		return evalsvc.ResponseServiceInstance.SaveResponse(assignmentID, q.ID, nil, "", nil, &b, nil)
	default:
		v, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return fmt.Errorf("%q is not a rating", raw)
		}
		return evalsvc.ResponseServiceInstance.SaveResponse(assignmentID, q.ID, &v, "", nil, nil, nil)
	}
}

func (h *EvaluationHandler) AddComment(c *fiber.Ctx) error {
	id, _ := c.ParamsInt("id")
	assignmentID := int64(id)
	comment := strings.TrimSpace(c.FormValue("comment"))

	assignment, err := evalsvc.AssignmentServiceInstance.FindByID(assignmentID)
	if err != nil {
		return core.RenderError(c, fiber.StatusNotFound, "Assignment not found.")
	}
	if err := core.RequireCanFillEvaluation(c, assignment); err != nil {
		return err
	}

	if comment != "" {
		_, _ = evalsvc.CommentServiceInstance.Create(assignmentID, core.CurrentUserID(c), nil, comment)
	}

	comments, _ := evalsvc.CommentServiceInstance.FindByAssignmentID(assignmentID)
	return core.RenderTemplate(c, "eval-comments", map[string]any{
		"AssignmentID": assignmentID,
		"Comments":     commentViews(comments),
	})
}
