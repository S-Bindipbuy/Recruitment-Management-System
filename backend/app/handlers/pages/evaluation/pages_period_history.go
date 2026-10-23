package evaluation

import (
	"ems/app/enums"
	core "ems/app/handlers/pages/core"
	evalmodel "ems/app/models/evaluation"
	evalsvc "ems/app/services/evaluation"
	"fmt"
	"sort"
	"time"

	"github.com/gofiber/fiber/v2"
)

type EvaluationHandler struct{}

// ── Period history ──
//
// Once a period closes, the question becomes "what actually happened in
// there?". The assignments, results and approvals already carry the timestamps
// needed to answer it (created, started, submitted, approved, rejected), so
// this page reconstructs the timeline from those rather than from a separate
// log that nothing writes to.

type HistoryEvent struct {
	At      time.Time
	Kind    string
	Summary string
	Detail  string
	Badge   string
}

type HistoryRow struct {
	AssignmentID   int64
	EvaluateeName  string
	EvaluatorName  string
	Status         string
	Percentage     float64
	Grade          string
	HasResult      bool
	Overdue        bool
	ApprovalStatus string
	CreatedAt      time.Time
}

type PeriodHistoryData struct {
	Period     evalmodel.EvaluationPeriod
	Rows       []HistoryRow
	Events     []HistoryEvent
	Total      int
	Completed  int
	InProgress int
	Pending    int
	AvgScore   float64
	HasScore   bool
	// Progress percentages for the summary bar.
	PctCompleted int
	PctInProg    int
	PctPending   int
}

func (h *EvaluationHandler) PeriodHistory(c *fiber.Ctx) error {
	if err := core.RequirePermission(c,
		enums.PermissionEvaluationManage,
		enums.PermissionEvaluationAssign,
		enums.PermissionEvaluationEvaluate,
	); err != nil {
		return err
	}

	id, _ := c.ParamsInt("id")
	period, err := evalsvc.PeriodServiceInstance.FindByID(int64(id))
	if err != nil {
		return core.RenderError(c, fiber.StatusNotFound, "Evaluation period not found.")
	}

	data := PeriodHistoryData{Period: *period}
	assignments, _ := evalsvc.AssignmentServiceInstance.FindByPeriodID(period.ID)
	if len(assignments) == 0 {
		return core.RenderPage(c, "page-period-history", data)
	}

	assignmentIDs := make([]int64, 0, len(assignments))
	evaluateeIDs := make([]int64, 0, len(assignments))
	evaluatorIDs := make([]int64, 0, len(assignments))
	for _, a := range assignments {
		assignmentIDs = append(assignmentIDs, a.ID)
		if a.CandidateID != nil && *a.CandidateID > 0 {
			continue
		}
		evaluateeIDs = append(evaluateeIDs, a.EvaluateeID)
		evaluatorIDs = append(evaluatorIDs, a.EvaluatorID)
	}

	names := core.UserNames(append(append([]int64{}, evaluateeIDs...), evaluatorIDs...))
	candNames := core.CandidateNames(candidateIDsOf(assignments))
	results, _ := evalsvc.ResultServiceInstance.GetResultsForAssignments(assignmentIDs)
	approvals, _ := evalsvc.ApprovalServiceInstance.FindForAssignments(assignmentIDs)
	templateNames := templateNameMap(assignments)

	events := make([]HistoryEvent, 0, len(assignments)*3)
	totalScore, scoredCount := 0.0, 0

	for _, a := range assignments {
		subjectName := names[a.EvaluateeID]
		if a.CandidateID != nil && *a.CandidateID > 0 {
			subjectName = candNames[*a.CandidateID] + " (Candidate)"
		}
		evaluatorName := core.DisplayName(names, a.EvaluatorID)

		row := HistoryRow{
			AssignmentID:  a.ID,
			EvaluateeName: subjectName,
			EvaluatorName: evaluatorName,
			Status:        a.Status,
			CreatedAt:     a.CreatedAt,
		}
		// Overdue means a due date has passed and the assignment never reached
		// completed, whatever stage it is sitting at now.
		if a.DueDate != nil && a.DueDate.Before(time.Now()) && a.Status != string(enums.AssignmentStatusCompleted) {
			row.Overdue = true
		}
		if res, ok := results[a.ID]; ok && res != nil {
			row.HasResult = true
			row.Percentage = res.Percentage
			row.Grade = res.Grade
			totalScore += res.Percentage
			scoredCount++
		}
		if list := approvals[a.ID]; len(list) > 0 {
			row.ApprovalStatus = decisiveApprovalStatus(list)
		}

		switch a.Status {
		case string(enums.AssignmentStatusCompleted):
			data.Completed++
		case string(enums.AssignmentStatusPending):
			data.Pending++
		default:
			data.InProgress++
		}
		data.Rows = append(data.Rows, row)

		events = append(events, HistoryEvent{
			At:      a.CreatedAt,
			Kind:    "assigned",
			Summary: fmt.Sprintf("%s was assigned to %s", subjectName, evaluatorName),
			Detail:  fmt.Sprintf("Evaluation #%d using %s", a.ID, templateNames[a.EvaluationTemplateID]),
			Badge:   "badge-info",
		})
		if a.StartedAt != nil {
			events = append(events, HistoryEvent{
				At:      *a.StartedAt,
				Kind:    "started",
				Summary: fmt.Sprintf("%s started the evaluation of %s", evaluatorName, subjectName),
				Badge:   "badge-warning",
			})
		}
		if a.SubmittedAt != nil {
			events = append(events, HistoryEvent{
				At:      *a.SubmittedAt,
				Kind:    "submitted",
				Summary: fmt.Sprintf("%s submitted the evaluation of %s", evaluatorName, subjectName),
				Badge:   "badge-neutral",
			})
		}
		for _, ap := range approvals[a.ID] {
			events = append(events, approvalEvent(ap, subjectName, names, a.ID))
		}
		if at, ok := completionTime(a, approvals[a.ID]); ok {
			events = append(events, HistoryEvent{
				At:      at,
				Kind:    "completed",
				Summary: fmt.Sprintf("Evaluation of %s was completed", subjectName),
				Badge:   "badge-success",
			})
		}
	}

	sort.SliceStable(events, func(i, j int) bool { return events[i].At.After(events[j].At) })
	data.Events = events
	data.Total = len(assignments)
	if scoredCount > 0 {
		data.HasScore = true
		data.AvgScore = totalScore / float64(scoredCount)
	}
	if data.Total > 0 {
		data.PctCompleted = data.Completed * 100 / data.Total
		data.PctInProg = data.InProgress * 100 / data.Total
		data.PctPending = data.Pending * 100 / data.Total
	}
	return core.RenderPage(c, "page-period-history", data)
}

// approvalDecidedAt is when an approval record reached its final state.
func approvalDecidedAt(a evalmodel.EvaluationApproval) time.Time {
	if a.ApprovedAt != nil {
		return *a.ApprovedAt
	}
	if a.RejectedAt != nil {
		return *a.RejectedAt
	}
	return time.Time{}
}

// selectDecisiveApproval picks the record that reflects where the assignment
// actually stands: the highest approval level, and inside that level the one
// decided last. The query only orders by level, so two records sharing a level
// (rejected, then approved after a resubmit) have no guaranteed order and the
// stale one could otherwise win.
func selectDecisiveApproval(list []evalmodel.EvaluationApproval) (evalmodel.EvaluationApproval, bool) {
	if len(list) == 0 {
		return evalmodel.EvaluationApproval{}, false
	}
	best := list[0]
	for _, a := range list[1:] {
		switch {
		case a.Level > best.Level:
			best = a
		case a.Level < best.Level:
		case approvalDecidedAt(a).After(approvalDecidedAt(best)):
			best = a
		}
	}
	return best, true
}

func decisiveApprovalStatus(list []evalmodel.EvaluationApproval) string {
	best, ok := selectDecisiveApproval(list)
	if !ok {
		return ""
	}
	return best.Status
}

// completionTime is when the evaluation finished. Rows approved before
// completed_at was recorded carry no timestamp of their own, so fall back to
// the moment the approval was granted instead of dropping the event.
func completionTime(a evalmodel.EvaluationAssignment, list []evalmodel.EvaluationApproval) (time.Time, bool) {
	if a.CompletedAt != nil {
		return *a.CompletedAt, true
	}
	if a.Status != string(enums.AssignmentStatusCompleted) {
		return time.Time{}, false
	}
	best, ok := selectDecisiveApproval(list)
	if !ok || best.Status != string(enums.ApprovalStatusApproved) || best.ApprovedAt == nil {
		return time.Time{}, false
	}
	return *best.ApprovedAt, true
}

func approvalEvent(ap evalmodel.EvaluationApproval, subjectName string, names map[int64]string, assignmentID int64) HistoryEvent {
	approver := core.DisplayName(names, ap.ApproverID)
	switch ap.Status {
	case string(enums.ApprovalStatusApproved):
		at := ap.CreatedAt
		if ap.ApprovedAt != nil {
			at = *ap.ApprovedAt
		}
		return HistoryEvent{
			At:      at,
			Kind:    "approved",
			Summary: fmt.Sprintf("%s approved the evaluation of %s", approver, subjectName),
			Detail:  ap.Comment,
			Badge:   "badge-active",
		}
	case string(enums.ApprovalStatusRejected):
		at := ap.CreatedAt
		if ap.RejectedAt != nil {
			at = *ap.RejectedAt
		}
		return HistoryEvent{
			At:      at,
			Kind:    "rejected",
			Summary: fmt.Sprintf("%s sent back the evaluation of %s", approver, subjectName),
			Detail:  ap.Comment,
			Badge:   "badge-danger",
		}
	default:
		return HistoryEvent{
			At:      ap.CreatedAt,
			Kind:    "awaiting",
			Summary: fmt.Sprintf("Evaluation of %s is waiting for %s to approve", subjectName, approver),
			Detail:  fmt.Sprintf("Approval #%d for evaluation #%d", ap.ID, assignmentID),
			Badge:   "badge-inactive",
		}
	}
}

func candidateIDsOf(assignments []evalmodel.EvaluationAssignment) []int64 {
	ids := make([]int64, 0, len(assignments))
	for _, a := range assignments {
		if a.CandidateID != nil && *a.CandidateID > 0 {
			ids = append(ids, *a.CandidateID)
		}
	}
	return ids
}

// templateNameMap resolves the template names used by a set of assignments in
// one pass, so the timeline does not issue a lookup per row.
func templateNameMap(assignments []evalmodel.EvaluationAssignment) map[int64]string {
	ids := make([]int64, 0, len(assignments))
	seen := make(map[int64]bool, len(assignments))
	for _, a := range assignments {
		if !seen[a.EvaluationTemplateID] {
			seen[a.EvaluationTemplateID] = true
			ids = append(ids, a.EvaluationTemplateID)
		}
	}
	names := map[int64]string{}
	for _, id := range ids {
		if t, err := evalsvc.TemplateServiceInstance.FindByID(id); err == nil {
			names[id] = t.Name
		} else {
			names[id] = fmt.Sprintf("template #%d", id)
		}
	}
	return names
}
