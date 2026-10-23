package cycles

import (
	evalmodel "ems/app/models/evaluation"
	usermodel "ems/app/models/user"
)

// CycleData drives every step of the wizard. Step decides which block of the
// template renders; the rest is carried along.
type CycleData struct {
	Step         int
	Periods      []evalmodel.EvaluationPeriod
	Templates    []CycleTemplateOption
	Workers      []usermodel.WorkerOption
	PeriodName   string
	TemplateName string
	DueDate      string
	// Chosen so far
	PeriodID      int64
	TemplateID    int64
	EvaluatorID   int64
	EvaluatorName string
	// Step 3: exactly who was ticked. Carried forward as one hidden input per
	// candidate so the picker survives a trip to the review step and back.
	SelectedCandidateIDs []int64
	AllSelected          bool
	// Preview / results
	Pairs      []CyclePairPreview
	Created    int
	Skipped    int
	WillCreate int
	// BlockedReason names the one thing standing between the choices on step 3
	// and a creatable run, so the Next button can state it instead of simply
	// sitting there disabled with no visible explanation.
	BlockedReason  string
	ReadyTemplates int
	Launched       bool
	ErrMessage     string
	PreviewReady   bool
	Candidates     []CycleCandidateOption
}

// CycleTemplateOption is an active template offered on step 2, with enough
// counted for the step to tell a usable one from a placeholder.
type CycleTemplateOption struct {
	evalmodel.EvaluationTemplate
	QuestionCount int
	SectionCount  int
}

// CyclePairPreview is one row of the step-3 pairing table.
type CyclePairPreview struct {
	EvaluatorName string
	EvaluateeName string
	Skipped       bool
	SkipReason    string
}

// CycleCandidateOption is one applicant offered by the step-3 picker.
type CycleCandidateOption struct {
	ID       int64
	Name     string
	Email    string
	Status   string
	Assigned bool
	Checked  bool
	// Droppable lists the assignments behind "already assigned" that this step
	// is allowed to take back out of the period - a completed evaluation is
	// left alone, so the warning can show with nothing to press beside it.
	Droppable []DroppableAssignment
}

// DroppableAssignment is a pending or started evaluation the wizard may drop.
type DroppableAssignment struct {
	ID     int64
	Status string
}
