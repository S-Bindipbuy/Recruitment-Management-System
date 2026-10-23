package cycles

import (
	"ems/app/enums"
	candidatesvc "ems/app/services/candidate"
	evaluationsvc "ems/app/services/evaluation"
)

// ── Guided evaluation cycle: the model behind the wizard ──
//
// The wizard walks four steps and carries the choices forward as hidden form
// fields. What each step may do lives here rather than in the handler, so the
// rules can be tested without a request:
//
//	1. period   -> pick an existing period or create one inline
//	2. template -> pick an active template
//	3. scope    -> tick the individual candidates to evaluate and choose who
//	               assesses them
//	4. review   -> confirm and launch the whole batch

// IsDroppableStatus reports whether a run can still be pulled back out of a
// period: pending and started. Anything handed in, approved or overdue has
// left the wizard's hands.
func IsDroppableStatus(status string) bool {
	return status == string(enums.AssignmentStatusPending) ||
		status == string(enums.AssignmentStatusStarted)
}

// CycleCandidateState answers two questions about a period in one pass: which
// candidates already carry this template - the same rule a single create
// refuses - and which of those assignments this step is allowed to take back
// out again.
func CycleCandidateState(periodID, templateID int64) (map[int64]bool, map[int64][]DroppableAssignment) {
	assigned, _ := evaluationsvc.AssignmentServiceInstance.CandidateAssignments(periodID, templateID)
	droppable := map[int64][]DroppableAssignment{}
	if periodID <= 0 {
		return assigned, droppable
	}
	list, _ := evaluationsvc.AssignmentServiceInstance.FindByPeriodID(periodID)
	for _, a := range list {
		if a.CandidateID == nil || *a.CandidateID <= 0 {
			continue
		}
		if a.EvaluationTemplateID == templateID && IsDroppableStatus(a.Status) {
			droppable[*a.CandidateID] = append(droppable[*a.CandidateID],
				DroppableAssignment{ID: a.ID, Status: a.Status})
		}
	}
	return assigned, droppable
}

// LoadCycleCandidates builds the step-3 picker: everybody still in the
// pipeline, with the tick they already carry and a flag on anyone who already
// has this period's template - they are listed, but not silently re-created.
func LoadCycleCandidates(periodID, templateID int64, selected []int64) []CycleCandidateOption {
	assigned, droppable := CycleCandidateState(periodID, templateID)
	chosen := make(map[int64]bool, len(selected))
	for _, id := range selected {
		chosen[id] = true
	}

	all, _ := candidatesvc.CandidateServiceInstance.FindHireableOptions()
	options := make([]CycleCandidateOption, 0, len(all))
	for _, cand := range all {
		options = append(options, CycleCandidateOption{
			ID: cand.ID, Name: cand.Name, Email: cand.Email, Status: cand.Status,
			Assigned:  assigned[cand.ID],
			Checked:   chosen[cand.ID],
			Droppable: droppable[cand.ID],
		})
	}
	return options
}

// PruneCycleCandidateIDs drops ticks pointing at somebody who can no longer be
// evaluated (hired, deleted), so the preview never names a ghost.
func PruneCycleCandidateIDs(ids []int64, options []CycleCandidateOption) []int64 {
	if len(ids) == 0 {
		return ids
	}
	valid := make(map[int64]bool, len(options))
	for _, o := range options {
		valid[o.ID] = true
	}
	kept := make([]int64, 0, len(ids))
	for _, id := range ids {
		if valid[id] {
			kept = append(kept, id)
		}
	}
	return kept
}
