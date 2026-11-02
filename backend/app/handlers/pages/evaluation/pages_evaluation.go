package evaluation

import (
	"ems/app/enums"
	core "ems/app/handlers/pages/core"
	evalmodel "ems/app/models/evaluation"
	recmodel "ems/app/models/recruitment"
	candsvc "ems/app/services/candidate"
	evalsvc "ems/app/services/evaluation"
	recsvc "ems/app/services/recruitment"
	usersvc "ems/app/services/user"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

// roundSectionsForAssignment returns the question scope of an assignment. A
// per-round evaluation asks only the sections picked for its interview round;
// a candidate-level evaluation keeps the whole template.
func roundSectionsForAssignment(assignment *evalmodel.EvaluationAssignment) []evalmodel.EvaluationSection {
	if assignment.InterviewRoundID != nil && *assignment.InterviewRoundID > 0 {
		if sections, err := recsvc.InterviewRoundServiceInst.FindSectionsByRoundID(*assignment.InterviewRoundID); err == nil {
			// A scheduler may override the round's template for one interview.
			// In that case the round's sections belong to the old template and
			// must not leak into the evaluation.
			matching := make([]evalmodel.EvaluationSection, 0, len(sections))
			for _, section := range sections {
				if section.EvaluationTemplateID == assignment.EvaluationTemplateID {
					matching = append(matching, section)
				}
			}
			if len(matching) > 0 {
				return matching
			}
		}
	}
	sections, _ := evalsvc.SectionServiceInstance.FindByTemplateID(assignment.EvaluationTemplateID)
	return sections
}

// roundName returns the interview round's display name for a round evaluation.
func roundName(assignment *evalmodel.EvaluationAssignment) string {
	if assignment.InterviewRoundID == nil || *assignment.InterviewRoundID <= 0 {
		return ""
	}
	round, err := recsvc.InterviewRoundServiceInst.FindByID(*assignment.InterviewRoundID)
	if err != nil {
		return ""
	}
	return round.Name
}

// loadAssignableCandidates returns applicants that still need an evaluation.
// Hired candidates and anybody whose evaluation was already submitted or
// completed are filtered out, so they stop appearing in the picker. Rejected
// applicants remain available for a fresh evaluation. Candidates whose active
// interview rounds are not all scored are hidden too: the evaluation starts
// only from a complete interview record.
func loadAssignableCandidates() []CandidateOption {
	options, _ := candsvc.CandidateServiceInstance.FindEvaluationCandidateOptions()
	summaries := core.LoadCandidateSummaryMap()
	ready := make([]CandidateOption, 0, len(options))
	for _, option := range options {
		if core.InterviewReady(summaries[option.ID]) {
			ready = append(ready, option)
		}
	}
	return ready
}

// groupAssignableCandidates splits the assignment picker by job posting,
// preserving the posting order returned by the repository.
func groupAssignableCandidates(options []CandidateOption) []CandidateGroup {
	groups := make(map[int64]*CandidateGroup)
	order := make([]int64, 0)
	for _, option := range options {
		group, ok := groups[option.JobPostingID]
		if !ok {
			title := option.JobPostingTitle
			if title == "" {
				if option.JobPostingID == 0 {
					title = "No job posting"
				} else {
					title = "Job posting"
				}
			}
			group = &CandidateGroup{PostingID: option.JobPostingID, PostingTitle: title}
			groups[option.JobPostingID] = group
			order = append(order, option.JobPostingID)
		}
		group.Candidates = append(group.Candidates, option)
	}

	grouped := make([]CandidateGroup, 0, len(order))
	for _, id := range order {
		grouped = append(grouped, *groups[id])
	}
	return grouped
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
	data.Candidates = loadAssignableCandidates()
	data.CandidateGroups = groupAssignableCandidates(data.Candidates)

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
	summaries := core.LoadCandidateSummaryMap()

	for _, a := range assignments {
		evaluateeName := ""
		isCandidate := a.CandidateID != nil && *a.CandidateID > 0
		if isCandidate {
			evaluateeName = candidateNames[*a.CandidateID] + " (Candidate)"
		} else {
			evaluateeName = core.DisplayName(names, a.EvaluateeID)
		}
		view := AssignmentView{
			Assignment:     a,
			EvaluateeName:  evaluateeName,
			EvaluatorName:  core.DisplayName(evaluatorNames, a.EvaluatorID),
			IsRound:        a.InterviewID != nil,
			InterviewReady: true,
		}
		if view.IsRound {
			view.RoundName = roundName(&a)
			if a.InterviewID != nil {
				view.InterviewID = *a.InterviewID
			}
		} else if isCandidate {
			view.InterviewReady = core.InterviewReady(summaries[*a.CandidateID])
		}
		if isCandidate {
			if summary, ok := summaries[*a.CandidateID]; ok && summary.CompletedReviews > 0 {
				view.InterviewScore = summary.AverageScore
				view.InterviewScoreTotal = summary.TotalScore
				view.InterviewScoreAvailable = true
				view.CompletedReviews = summary.CompletedReviews
				view.TotalInterviews = summary.TotalInterviews
			}
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
	_, err := evalsvc.AssignmentServiceInstance.CreateForCandidate(periodID, templateID, evaluatorID, candidateID, dueDate, core.CurrentUserID(c))
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
	case errors.Is(err, evalsvc.ErrInterviewsIncomplete):
		return err.Error()
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
		if errors.Is(err, evalsvc.ErrInterviewsIncomplete) {
			return core.RenderError(c, fiber.StatusBadRequest, err.Error())
		}
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

	data := &FillPageData{
		Assignment:     *assignment,
		EvaluateeName:  evaluateeName,
		RoundName:      roundName(assignment),
		InterviewReady: true,
	}
	if assignment.CandidateID != nil && *assignment.CandidateID > 0 {
		summaries := core.LoadCandidateSummaryMap()
		summary := summaries[*assignment.CandidateID]
		if assignment.InterviewID == nil {
			data.InterviewReady = core.InterviewReady(summary)
		}
		if summary != nil && summary.CompletedReviews > 0 {
			data.InterviewScore = summary.AverageScore
			data.InterviewScoreTotal = summary.TotalScore
			data.InterviewScoreAvailable = true
			data.CompletedReviews = summary.CompletedReviews
			data.TotalInterviews = summary.TotalInterviews
		}
	}

	responseMap := map[int64]evalmodel.EvaluationResponse{}
	for _, response := range assignmentResponses(assignment.ID) {
		responseMap[response.EvaluationQuestionID] = response
	}

	if assignment.Status == string(enums.AssignmentStatusStarted) ||
		assignment.Status == string(enums.AssignmentStatusPending) {

		sections := roundSectionsForAssignment(assignment)
		for _, sec := range sections {
			view := SectionView{Name: sec.Name, Description: sec.Description}
			questions, _ := evalsvc.QuestionServiceInstance.FindBySectionID(sec.ID)
			for _, q := range questions {
				qv := QuestionView{
					Model:    q,
					Required: q.IsRequired && q.Status == "active",
				}
				if resp, ok := responseMap[q.ID]; ok {
					qv.TextValue = resp.TextValue
					if resp.NumericValue != nil {
						qv.NumericValue = strconv.FormatFloat(*resp.NumericValue, 'f', -1, 64)
					}
					if resp.RatingValue != nil {
						qv.RatingValue = strconv.FormatFloat(*resp.RatingValue, 'f', -1, 64)
					}
					if resp.BooleanValue != nil {
						qv.BooleanValue = strconv.FormatBool(*resp.BooleanValue)
					}
				}
				view.Questions = append(view.Questions, qv)
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

func assignmentResponses(assignmentID int64) []evalmodel.EvaluationResponse {
	responses, err := evalsvc.ResponseServiceInstance.GetResponses(assignmentID)
	if err != nil {
		return nil
	}
	return responses
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
		if errors.Is(err, evalsvc.ErrInterviewsIncomplete) {
			return core.RenderError(c, fiber.StatusBadRequest, err.Error())
		}
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

	if assignment.Status == string(enums.AssignmentStatusPending) {
		if err := evalsvc.AssignmentServiceInstance.Start(assignmentID); err != nil {
			if errors.Is(err, evalsvc.ErrInterviewsIncomplete) {
				return core.RenderError(c, fiber.StatusBadRequest, err.Error())
			}
			log.Printf("[SubmitEvaluation] auto-start assignment %d: %v", assignmentID, err)
		}
	}

	sections := roundSectionsForAssignment(assignment)
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
	result, _ := evalsvc.ResultServiceInstance.GetResult(assignmentID)

	// A round evaluation is the same interview evidence in a second view. Copy
	// the score and answers onto the interview's feedback so the panel list,
	// reports and candidate ranking see them too.
	if assignment.InterviewID != nil {
		if err := syncRoundFeedback(assignment, result, core.CurrentUserID(c)); err != nil {
			log.Printf("[SubmitEvaluation] sync feedback for assignment %d: %v", assignmentID, err)
			return core.RenderError(c, fiber.StatusInternalServerError,
				"The evaluation was saved, but the interview feedback could not be updated.")
		}
	}

	c.Set("HX-Push-Url", "/pages/evaluations")
	return h.Evaluations(c)
}

// syncRoundFeedback writes a completed round evaluation back onto the
// interview. It keeps one feedback row per interviewer (updating the existing
// row) and reuses the evaluation answers, so both screens agree.
func syncRoundFeedback(assignment *evalmodel.EvaluationAssignment, result *evalmodel.EvaluationResult, recordedBy int64) error {
	if assignment.InterviewID == nil || *assignment.InterviewID <= 0 {
		return nil
	}
	interviewID := *assignment.InterviewID
	now := time.Now()

	var score *float64
	if result != nil {
		value := result.TotalScore
		score = &value
	}

	feedback, err := recsvc.InterviewFeedbackServiceInst.FindByInterviewAndInterviewer(interviewID, assignment.EvaluatorID)
	if err == nil && feedback != nil && feedback.ID > 0 {
		// Do not blank an existing score just because this submission carried
		// no rating answers (for example, only text notes).
		if score != nil {
			feedback.Score = score
		}
		feedback.SubmittedAt = &now
		if err := recsvc.InterviewFeedbackServiceInst.Update(feedback); err != nil {
			return err
		}
	} else {
		feedback = &recmodel.InterviewFeedback{
			InterviewID:   interviewID,
			InterviewerID: assignment.EvaluatorID,
			RecordedByID:  recordedBy,
			Score:         score,
			SubmittedAt:   &now,
		}
		if err := recsvc.InterviewFeedbackServiceInst.Create(feedback); err != nil {
			return err
		}
	}

	responses, err := evalsvc.ResponseServiceInstance.GetResponses(assignment.ID)
	if err != nil {
		return err
	}
	for _, response := range responses {
		answer := &recmodel.InterviewFeedbackAnswer{
			FeedbackID:           feedback.ID,
			EvaluationQuestionID: response.EvaluationQuestionID,
			RatingValue:          response.RatingValue,
			TextValue:            response.TextValue,
			NumericValue:         response.NumericValue,
			BooleanValue:         response.BooleanValue,
		}
		if err := recsvc.InterviewFeedbackServiceInst.SaveAnswer(answer); err != nil {
			return err
		}
	}
	return nil
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
