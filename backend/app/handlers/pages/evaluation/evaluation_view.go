package evaluation

import (
	core "ems/app/handlers/pages/core"
	candmodel "ems/app/models/candidate"
	evalmodel "ems/app/models/evaluation"
)

// CandidateOption is one applicant offered in the "New Assignment" dialog.
// Only candidates that have not been hired yet can be evaluated.
type CandidateOption = candmodel.CandidateOption

type AssignmentView struct {
	Assignment    evalmodel.EvaluationAssignment
	Result        *evalmodel.EvaluationResult
	EvaluateeName string
	EvaluatorName string
}

type CommentView struct {
	Comment    evalmodel.EvaluationComment
	AuthorName string
}

type QuestionView struct {
	Model    evalmodel.EvaluationQuestion
	Required bool
}

type SectionView struct {
	Name        string
	Description string
	Questions   []QuestionView
}

type EvaluationsPageData struct {
	Periods     []evalmodel.EvaluationPeriod
	Templates   []evalmodel.EvaluationTemplate
	Assignments []AssignmentView
	Workers     []core.WorkerOption
	Candidates  []CandidateOption
}

type FillPageData struct {
	Assignment    evalmodel.EvaluationAssignment
	EvaluateeName string
	Sections      []SectionView
	Comments      []CommentView
	Result        *evalmodel.EvaluationResult
}
