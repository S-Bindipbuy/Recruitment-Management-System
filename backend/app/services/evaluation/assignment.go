package evaluation

import (
	"ems/app/enums"
	evalmodel "ems/app/models/evaluation"
	evalrepo "ems/app/repositories/evaluation"
	userrepo "ems/app/repositories/user"
	candidatesvc "ems/app/services/candidate"
	organizationsvc "ems/app/services/organization"
	"ems/database"
	"errors"
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
// an active template, once per period.
func (s *EvaluationAssignmentService) CreateForCandidate(periodID, templateID, evaluatorID, candidateID int64, dueDate *time.Time) (*evalmodel.EvaluationAssignment, error) {
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
	if strings.EqualFold(candidate.Status, "rejected") {
		return nil, errors.New("This candidate has been rejected")
	}
	assigned, err := evalrepo.AssignmentRepo.CandidateAssignments(periodID, templateID)
	if err != nil {
		return nil, err
	}
	if assigned[candidateID] {
		return nil, ErrAlreadyAssigned
	}
	return s.createAssignment(periodID, templateID, evaluatorID, &candidateID, dueDate)
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
func (s *EvaluationAssignmentService) CreateMany(periodID, templateID int64, pairs []evalmodel.EvaluationPair, dueDate *time.Time) error {
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

	// evaluatee_id is NOT NULL and stays 0: the wizard only ever evaluates
	// applicants, and the column exists for the worker rows this code no
	// longer writes.
	stmt, err := tx.Prepare(
		`INSERT INTO evaluation_assignments (evaluation_period_id, evaluation_template_id, evaluator_id, evaluatee_id, candidate_id, due_date, status) 
		 VALUES (?, ?, ?, 0, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, p := range pairs {
		if p.CandidateID == nil || *p.CandidateID <= 0 || assigned[*p.CandidateID] {
			continue
		}
		if _, err := stmt.Exec(periodID, templateID, p.EvaluatorID, p.CandidateID, dueDate, string(enums.AssignmentStatusPending)); err != nil {
			return err
		}
		assigned[*p.CandidateID] = true
	}
	return tx.Commit()
}

func (s *EvaluationAssignmentService) Start(id int64) error {
	a, err := evalrepo.AssignmentRepo.FindByID(id)
	if err != nil {
		return err
	}
	if a.Status == string(enums.AssignmentStatusCompleted) || a.Status == "rejected" || a.Status == string(enums.AssignmentStatusSubmitted) {
		return errors.New("this evaluation cannot be started again")
	}
	now := time.Now()
	a.Status = string(enums.AssignmentStatusStarted)
	a.StartedAt = &now
	return evalrepo.AssignmentRepo.Update(a)
}

func (s *EvaluationAssignmentService) Submit(id int64) error {
	a, err := evalrepo.AssignmentRepo.FindByID(id)
	if err != nil {
		return err
	}
	now := time.Now()
	a.Status = string(enums.AssignmentStatusSubmitted)
	a.SubmittedAt = &now
	if err := evalrepo.AssignmentRepo.Update(a); err != nil {
		return err
	}

	// Ask somebody OTHER than the evaluator to approve. The old code used
	// a.EvaluatorID as the approver, so every submitted evaluation became a
	// request for the author to approve their own work.
	exclude := []int64{a.EvaluatorID}
	if a.EvaluateeID > 0 {
		exclude = append(exclude, a.EvaluateeID)
	}
	if approverID, err := userrepo.UserRepo.FindActiveUserWithPermission(
		exclude, string(enums.PermissionEvaluationApprove),
	); err == nil && approverID > 0 {
		_, _ = ApprovalServiceInstance.Create(a.ID, approverID, 1)
	}
	// If nobody else can approve (e.g. only the evaluator holds the
	// permission), no request is created; an admin can still request one
	// explicitly from the Approvals page.

	return nil
}
