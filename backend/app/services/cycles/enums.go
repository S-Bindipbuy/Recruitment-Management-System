package cycles

// Cycle wizard steps. Step decides which block of the template renders.
const (
	CycleStepPeriod   = 1
	CycleStepTemplate = 2
	CycleStepScope    = 3
	CycleStepReview   = 4
)

// Cycle skip reasons, kept as constants because the "why can't I continue"
// copy is built from a tally of them.
const (
	CycleSkipNoEvaluator       = "No evaluator chosen"
	CycleSkipEvaluatorInactive = "Selected evaluator is not an active worker"
	CycleSkipAlreadyAssigned   = "Already assigned in this period"
)
