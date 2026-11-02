package evaluation

import (
	"database/sql"
	"ems/app/enums"
	evalmodel "ems/app/models/evaluation"
	evalrepo "ems/app/repositories/evaluation"
	candidatesvc "ems/app/services/candidate"
	organizationsvc "ems/app/services/organization"
	recsvc "ems/app/services/recruitment"
	"ems/database"
	"errors"
	"log"
	"strings"
	"time"
)

type EvaluationAssignmentService struct{}

// The rules for creating a candidate assignment, in the order they are
// checked. They live here rather than in each handler because the dialog, the
// candidates page and the JSON API all have to apply the same ones; each
// surface turns a sentinel into its own wording and keeps every check it does
// not need from being re-typed.
var (
	ErrAssignmentIncomplete = errors.New("period, template, evaluator and candidate are required")
	ErrPeriodNotFound       = errors.New("The selected period was not found")
	ErrTemplateNotFound     = errors.New("The selected template was not found")
	ErrTemplateInactive     = errors.New("The selected template is inactive")
	ErrEvaluatorNotActive   = errors.New("The evaluator must be an active worker")
	ErrCandidateNotFound    = errors.New("The selected candidate was not found")
	ErrCandidateHired       = errors.New("This candidate has already been accepted as a worker, so they can no longer be evaluated")
	ErrInterviewsIncomplete = errors.New("Complete all interview rounds for this candidate before assigning or starting an evaluation")
	ErrAlreadyAssigned      = errors.New("This candidate already has this template assigned in this period")
)

// assignmentRuleErrors holds every rule a person at the screen can fix. A
// failure outside this list (a rejected insert, say) is not their doing.
var assignmentRuleErrors = []error{
	ErrAssignmentIncomplete,
	ErrPeriodNotFound,
	ErrTemplateNotFound,
	ErrTemplateInactive,
	ErrEvaluatorNotActive,
	ErrCandidateNotFound,
	ErrCandidateHired,
	ErrInterviewsIncomplete,
	ErrAlreadyAssigned,
}

// AssignmentRuleError reports whether err is one of the rules above, so the
// callers can pick between a 400 and a 500 (or between scolding and apologising)
// without repeating the list.
func AssignmentRuleError(err error) bool {
	for _, rule := range assignmentRuleErrors {
		if errors.Is(err, rule) {
			return true
		}
	}
	return false
}

// CreateForCandidate makes the one assignment every entry point is allowed to
// make: an active worker assessing an applicant who has not been hired yet, on
// an active template, once per period. A rejected run is closed, not a ban: HR
// may decide to evaluate the same applicant again (for example after another
// job submission), so only an active assignment blocks a new one.
func (s *EvaluationAssignmentService) CreateForCandidate(periodID, templateID, evaluatorID, candidateID int64, dueDate *time.Time, assignedBy int64) (*evalmodel.EvaluationAssignment, error) {
	if periodID <= 0 || templateID <= 0 || evaluatorID <= 0 || candidateID <= 0 {
		return nil, ErrAssignmentIncomplete
	}
	if _, err := PeriodServiceInstance.FindByID(periodID); err != nil {
		return nil, ErrPeriodNotFound
	}
	template, err := TemplateServiceInstance.FindByID(templateID)
	if err != nil {
		return nil, ErrTemplateNotFound
	}
	if template.Status != string(enums.TemplateStatusActive) {
		return nil, ErrTemplateInactive
	}
	if !organizationsvc.IsActiveWorker(evaluatorID) {
		return nil, ErrEvaluatorNotActive
	}
	candidate, err := candidatesvc.CandidateServiceInstance.FindByID(candidateID)
	if err != nil {
		return nil, ErrCandidateNotFound
	}
	if strings.EqualFold(candidate.Status, "hired") {
		return nil, ErrCandidateHired
	}
	ready, err := recsvc.InterviewServiceInst.CandidateInterviewReady(candidateID)
	if err != nil {
		return nil, err
	}
	if !ready {
		return nil, ErrInterviewsIncomplete
	}
	assigned, err := evalrepo.AssignmentRepo.CandidateAssignments(periodID, templateID)
	if err != nil {
		return nil, err
	}
	if assigned[candidateID] {
		return nil, ErrAlreadyAssigned
	}
	assignment, err := s.createAssignment(periodID, templateID, evaluatorID, &candidateID, dueDate)
	if err != nil {
		return nil, err
	}
	reopenRejectedCandidate(candidateID, assignedBy)
	return assignment, nil
}

// CreateForInterview creates the per-round evaluations that belong to a newly
// scheduled interview: one pending assignment per panel interviewer, using the
// round's template and its selected sections. A templateOverride lets the
// scheduler pick the evaluation when the round itself has none configured; the
// rest of the pipeline still treats it as that round's evaluation.
//
// It is deliberately best-effort and idempotent - scheduling must not fail
// because a round has no template or no active period is configured.
func (s *EvaluationAssignmentService) CreateForInterview(interviewID int64, templateOverride ...int64) ([]evalmodel.EvaluationAssignment, error) {
	interview, err := recsvc.InterviewServiceInst.FindByID(interviewID)
	if err != nil {
		return nil, err
	}
	if interview.CandidateID <= 0 {
		return nil, nil
	}

	round, err := recsvc.InterviewRoundServiceInst.FindByID(interview.InterviewRoundID)
	if err != nil {
		return nil, err
	}
	templateID := round.EvaluationTemplateID
	if len(templateOverride) > 0 && templateOverride[0] > 0 {
		templateID = templateOverride[0]
	}
	if templateID <= 0 {
		log.Printf("[RoundEvaluation] round %d has no evaluation template; interview %d has no round evaluation", round.ID, interviewID)
		return nil, nil
	}
	template, err := TemplateServiceInstance.FindByID(templateID)
	if err != nil || template.Status != string(enums.TemplateStatusActive) {
		log.Printf("[RoundEvaluation] template %d for round %d is not active; interview %d has no round evaluation", templateID, round.ID, interviewID)
		return nil, nil
	}
	if candidate, err := candidatesvc.CandidateServiceInstance.FindByID(interview.CandidateID); err != nil {
		return nil, err
	} else if strings.EqualFold(candidate.Status, "hired") {
		return nil, nil
	}

	on := interview.ScheduledAt
	if on.IsZero() {
		on = time.Now()
	}
	period, err := PeriodServiceInstance.FindActive(on)
	if err != nil || period == nil {
		log.Printf("[RoundEvaluation] no active evaluation period for interview %d: %v", interviewID, err)
		return nil, nil
	}

	panel, _ := recsvc.InterviewServiceInst.FindPanelMembers(interviewID)
	panelIDs := make([]int64, 0, len(panel))
	for _, member := range panel {
		panelIDs = append(panelIDs, member.UserID)
	}
	if len(panelIDs) == 0 && interview.InterviewerID > 0 {
		panelIDs = []int64{interview.InterviewerID}
	}

	candidateID := interview.CandidateID
	roundID := round.ID
	created := make([]evalmodel.EvaluationAssignment, 0, len(panelIDs))
	for _, evaluatorID := range panelIDs {
		if evaluatorID <= 0 || !organizationsvc.IsActiveWorker(evaluatorID) {
			continue
		}
		if existing, err := evalrepo.AssignmentRepo.FindByInterviewAndEvaluator(interviewID, evaluatorID); err == nil && existing != nil {
			continue
		}
		a := &evalmodel.EvaluationAssignment{
			EvaluationPeriodID:   period.ID,
			EvaluationTemplateID: template.ID,
			EvaluatorID:          evaluatorID,
			CandidateID:          &candidateID,
			InterviewID:          &interviewID,
			InterviewRoundID:     &roundID,
			Status:               string(enums.AssignmentStatusPending),
		}
		if err := evalrepo.AssignmentRepo.Create(a); err != nil {
			log.Printf("[RoundEvaluation] create interview %d evaluator %d: %v", interviewID, evaluatorID, err)
			continue
		}
		created = append(created, *a)
	}
	return created, nil
}

// StartForInterview moves every pending round evaluation of an interview to
// started. It runs when the interview is completed or feedback is saved, so
// the evaluation form is ready as soon as the interview work is done.
func (s *EvaluationAssignmentService) StartForInterview(interviewID int64) error {
	assignments, err := evalrepo.AssignmentRepo.FindByInterviewID(interviewID)
	if err != nil {
		return err
	}
	now := time.Now()
	for _, a := range assignments {
		if a.Status != string(enums.AssignmentStatusPending) {
			continue
		}
		a.Status = string(enums.AssignmentStatusStarted)
		a.StartedAt = &now
		if err := evalrepo.AssignmentRepo.Update(&a); err != nil {
			return err
		}
	}
	return nil
}

// CompleteRoundFromFeedback is the whole per-round evaluation in one step:
// copy the panel member's saved interview answers into their evaluation, score
// it, and mark it complete. Saving interview feedback is the evaluation, so
// there is no separate "start evaluation" / "submit evaluation" step to go
// through for each candidate.
func (s *EvaluationAssignmentService) CompleteRoundFromFeedback(interviewID, interviewerID int64) error {
	assignment, err := evalrepo.AssignmentRepo.FindByInterviewAndEvaluator(interviewID, interviewerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}
	if assignment == nil {
		return nil
	}
	if err := s.SyncRoundFromFeedback(interviewID, interviewerID); err != nil {
		return err
	}
	return s.CompleteRound(assignment.ID)
}

// SyncRoundFromFeedback copies a panel member's saved feedback answers into
// their round evaluation. The evaluation form then starts from what was
// already recorded during the interview instead of an empty page.
func (s *EvaluationAssignmentService) SyncRoundFromFeedback(interviewID, interviewerID int64) error {
	assignment, err := evalrepo.AssignmentRepo.FindByInterviewAndEvaluator(interviewID, interviewerID)
	if err != nil || assignment == nil {
		return err
	}
	feedback, err := recsvc.InterviewFeedbackServiceInst.FindByInterviewAndInterviewer(interviewID, interviewerID)
	if err != nil || feedback == nil {
		return err
	}
	answers, err := recsvc.InterviewFeedbackServiceInst.FindAnswersByFeedbackID(feedback.ID)
	if err != nil {
		return err
	}
	for _, answer := range answers {
		if err := ResponseServiceInstance.SaveResponse(assignment.ID, answer.EvaluationQuestionID,
			answer.RatingValue, answer.TextValue, answer.NumericValue, answer.BooleanValue, nil); err != nil {
			return err
		}
	}
	return nil
}

// reopenRejectedCandidate puts an applicant back into the pipeline when HR
// assigns a fresh evaluation. Without this a re-evaluated candidate would stay
// "rejected", so the approval could complete but they still could not be hired.
func reopenRejectedCandidate(candidateID, changedBy int64) {
	candidate, err := candidatesvc.CandidateServiceInstance.FindByID(candidateID)
	if err != nil || !strings.EqualFold(candidate.Status, "rejected") {
		return
	}
	if err := candidatesvc.CandidateServiceInstance.SetStatus(candidateID, "interviewing", "Evaluation assigned", changedBy); err != nil {
		log.Printf("[AssignEval] could not reopen candidate %d: %v", candidateID, err)
	}
}

// createAssignment is the insert behind the single create and the bulk launch.
func (s *EvaluationAssignmentService) createAssignment(periodID, templateID, evaluatorID int64, candidateID *int64, dueDate *time.Time) (*evalmodel.EvaluationAssignment, error) {
	a := &evalmodel.EvaluationAssignment{
		EvaluationPeriodID:   periodID,
		EvaluationTemplateID: templateID,
		EvaluatorID:          evaluatorID,
		CandidateID:          candidateID,
		DueDate:              dueDate,
		Status:               string(enums.AssignmentStatusPending),
	}
	if err := evalrepo.AssignmentRepo.Create(a); err != nil {
		return nil, err
	}
	return a, nil
}

// CandidateAssignments hands the shared "already assigned" rule to the wizard,
// so step 3 flags the same rows a create would refuse.
func (s *EvaluationAssignmentService) CandidateAssignments(periodID, templateID int64) (map[int64]bool, error) {
	return evalrepo.AssignmentRepo.CandidateAssignments(periodID, templateID)
}

func (s *EvaluationAssignmentService) FindAll() ([]evalmodel.EvaluationAssignment, error) {
	return evalrepo.AssignmentRepo.FindAll()
}

func (s *EvaluationAssignmentService) FindByCandidate(candidateID int64) ([]evalmodel.EvaluationAssignment, error) {
	return evalrepo.AssignmentRepo.FindByCandidate(candidateID)
}

// FindByInterviewID lists the per-round evaluations created for one interview.
func (s *EvaluationAssignmentService) FindByInterviewID(interviewID int64) ([]evalmodel.EvaluationAssignment, error) {
	return evalrepo.AssignmentRepo.FindByInterviewID(interviewID)
}

func (s *EvaluationAssignmentService) FindByID(id int64) (*evalmodel.EvaluationAssignment, error) {
	return evalrepo.AssignmentRepo.FindByID(id)
}

// Delete drops an assignment and everything recorded against it.
func (s *EvaluationAssignmentService) Delete(id int64) error {
	return evalrepo.AssignmentRepo.Delete(id)
}

func (s *EvaluationAssignmentService) FindByEvaluator(evaluatorID int64) ([]evalmodel.EvaluationAssignment, error) {
	return evalrepo.AssignmentRepo.FindByEvaluator(evaluatorID)
}

func (s *EvaluationAssignmentService) FindByEvaluatee(evaluateeID int64) ([]evalmodel.EvaluationAssignment, error) {
	return evalrepo.AssignmentRepo.FindByEvaluatee(evaluateeID)
}

func (s *EvaluationAssignmentService) FindByStatus(status string) ([]evalmodel.EvaluationAssignment, error) {
	return evalrepo.AssignmentRepo.FindByStatus(status)
}

func (s *EvaluationAssignmentService) FindByPeriodID(periodID int64) ([]evalmodel.EvaluationAssignment, error) {
	return evalrepo.AssignmentRepo.FindByPeriodID(periodID)
}

// CreateMany launches a whole evaluation cycle's worth of assignments in one
// transaction. Every row is all-or-nothing: a partial run would leave a
// period half-evaluated with no record of what was skipped. Pairs the shared
// "already assigned" rule covers are dropped here too, so a second submit
// cannot put the same candidate in twice.
func (s *EvaluationAssignmentService) CreateMany(periodID, templateID int64, pairs []evalmodel.EvaluationPair, dueDate *time.Time, assignedBy int64) error {
	if len(pairs) == 0 {
		return nil
	}
	assigned, err := evalrepo.AssignmentRepo.CandidateAssignments(periodID, templateID)
	if err != nil {
		return err
	}

	tx, err := database.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Candidate assignments leave evaluatee_id NULL; the column is nullable so
	// the employees(id) foreign key is not asked to accept a placeholder zero.
	stmt, err := tx.Prepare(
		`INSERT INTO evaluation_assignments (evaluation_period_id, evaluation_template_id, evaluator_id, evaluatee_id, candidate_id, due_date, status) 
		 VALUES (?, ?, ?, NULL, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	created := make([]int64, 0, len(pairs))
	for _, p := range pairs {
		if p.CandidateID == nil || *p.CandidateID <= 0 || assigned[*p.CandidateID] {
			continue
		}
		if _, err := stmt.Exec(periodID, templateID, p.EvaluatorID, p.CandidateID, dueDate, string(enums.AssignmentStatusPending)); err != nil {
			return err
		}
		assigned[*p.CandidateID] = true
		created = append(created, *p.CandidateID)
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	// The assignments exist now, so reopening a rejected applicant cannot be
	// rolled back. It only happens for candidates HR explicitly picked again.
	for _, candidateID := range created {
		reopenRejectedCandidate(candidateID, assignedBy)
	}
	return nil
}

func (s *EvaluationAssignmentService) Start(id int64) error {
	a, err := evalrepo.AssignmentRepo.FindByID(id)
	if err != nil {
		return err
	}
	if a.Status == string(enums.AssignmentStatusCompleted) || a.Status == "rejected" || a.Status == string(enums.AssignmentStatusSubmitted) {
		return errors.New("this evaluation cannot be started again")
	}
	// A round evaluation is about one interview, so it does not have to wait
	// for every other round of the candidate to be scored. Candidate-level
	// evaluations still do.
	if a.InterviewID == nil && a.CandidateID != nil && *a.CandidateID > 0 {
		ready, err := recsvc.InterviewServiceInst.CandidateInterviewReady(*a.CandidateID)
		if err != nil {
			return err
		}
		if !ready {
			return ErrInterviewsIncomplete
		}
	}
	now := time.Now()
	a.Status = string(enums.AssignmentStatusStarted)
	a.StartedAt = &now
	return evalrepo.AssignmentRepo.Update(a)
}

// CompleteRound marks a per-round evaluation as submitted and completed
// without creating a candidate-level approval. Round scores feed back into the
// interview feedback; the final hiring decision still goes through interview
// scoring, not this evaluation's approval list.
func (s *EvaluationAssignmentService) CompleteRound(id int64) error {
	a, err := evalrepo.AssignmentRepo.FindByID(id)
	if err != nil {
		return err
	}
	if a.InterviewID == nil {
		return errors.New("not a round evaluation")
	}
	now := time.Now()
	a.Status = string(enums.AssignmentStatusCompleted)
	if a.StartedAt == nil {
		a.StartedAt = &now
	}
	a.SubmittedAt = &now
	a.CompletedAt = &now
	if err := evalrepo.AssignmentRepo.Update(a); err != nil {
		return err
	}
	if _, err := ResultServiceInstance.CalculateAndSave(a.ID); err != nil {
		log.Printf("[CompleteRound] calculate result for %d: %v", a.ID, err)
	}
	return nil
}

func (s *EvaluationAssignmentService) Submit(id int64) error {
	a, err := evalrepo.AssignmentRepo.FindByID(id)
	if err != nil {
		return err
	}
	if a.InterviewID != nil {
		return s.CompleteRound(id)
	}
	now := time.Now()
	a.Status = string(enums.AssignmentStatusCompleted)
	a.SubmittedAt = &now
	a.CompletedAt = &now
	if err := evalrepo.AssignmentRepo.Update(a); err != nil {
		return err
	}
	_, err = ResultServiceInstance.CalculateAndSave(id)
	return err
}
