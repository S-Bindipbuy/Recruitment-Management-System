package cycles

import (
	evalmodel "ems/app/models/evaluation"
	candidatesvc "ems/app/services/candidate"
	organizationsvc "ems/app/services/organization"
	usersvc "ems/app/services/user"
	"fmt"
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

// PreviewCandidatePairs is the pairing table for step 3: the evaluatee is an
// applicant, so the only things that can go wrong are a missing (or no longer
// active) evaluator and a candidate this period already has this template for.
func PreviewCandidatePairs(periodID, templateID, evaluatorID int64, ids []int64) ([]CyclePairPreview, int) {
	if len(ids) == 0 {
		return nil, 0
	}
	names, _ := candidatesvc.CandidateServiceInstance.FindNames(ids)
	assigned, _ := CycleCandidateState(periodID, templateID)
	usable := evaluatorID > 0 && organizationsvc.IsActiveWorker(evaluatorID)
	evaluatorName := ""
	if evaluatorID > 0 {
		evaluatorName = userName(evaluatorID)
	}

	previews := make([]CyclePairPreview, 0, len(ids))
	skipped := 0
	for _, id := range ids {
		name := names[id]
		if name == "" {
			name = fmt.Sprintf("Candidate #%d", id)
		}
		p := CyclePairPreview{EvaluateeName: name}
		switch {
		case !usable:
			p.EvaluatorName = "-"
			p.Skipped = true
			if evaluatorID > 0 {
				p.SkipReason = CycleSkipEvaluatorInactive
			} else {
				p.SkipReason = CycleSkipNoEvaluator
			}
		default:
			p.EvaluatorName = evaluatorName
			if assigned[id] {
				p.Skipped = true
				p.SkipReason = CycleSkipAlreadyAssigned
			}
		}
		if p.Skipped {
			skipped++
		}
		previews = append(previews, p)
	}
	return previews, skipped
}

// CycleCandidatePairs rebuilds the id-level pairing for a candidate run at
// launch time, using the same rules as PreviewCandidatePairs.
func CycleCandidatePairs(periodID, templateID, evaluatorID int64, ids []int64) []evalmodel.EvaluationPair {
	if evaluatorID <= 0 || !organizationsvc.IsActiveWorker(evaluatorID) {
		return nil
	}
	assigned, _ := CycleCandidateState(periodID, templateID)

	var pairs []evalmodel.EvaluationPair
	for _, id := range ids {
		if assigned[id] {
			continue
		}
		candidateID := id
		pairs = append(pairs, evalmodel.EvaluationPair{EvaluatorID: evaluatorID, CandidateID: &candidateID})
	}
	return pairs
}

// CycleSkipMessage is the "why can't I continue" copy for the candidate scope,
// built from a tally of the skip reasons so it names the one real blocker
// rather than answering every dead end with the same sentence.
func CycleSkipMessage(pairs []CyclePairPreview) string {
	base := "No evaluations could be prepared. "
	if len(pairs) == 0 {
		return base + "Tick at least one candidate above."
	}

	var duplicate, noEvaluator, evaluatorDead int
	for _, p := range pairs {
		switch p.SkipReason {
		case CycleSkipAlreadyAssigned:
			duplicate++
		case CycleSkipNoEvaluator:
			noEvaluator++
		case CycleSkipEvaluatorInactive:
			evaluatorDead++
		}
	}

	switch {
	case noEvaluator == len(pairs):
		return base + "Choose who will evaluate these candidates above."
	case evaluatorDead == len(pairs):
		return base + "The chosen evaluator is not an active worker - pick somebody else above."
	case duplicate == len(pairs):
		return base + "Every candidate here is already assigned for this period - pick another period, or change the evaluator."
	default:
		return base + "The table above says why each candidate was skipped."
	}
}

// userName is one user's display name, falling back to the id when the account
// is gone - the wizard would rather show "User #7" than an empty cell.
func userName(id int64) string {
	name, _ := usersvc.UserServiceInstance.UserName(id)
	if name == "" {
		return fmt.Sprintf("User #%d", id)
	}
	return name
}
