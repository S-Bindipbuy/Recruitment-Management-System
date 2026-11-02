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
	// RoundName names the interview round for an auto-created per-round
	// evaluation; it is empty for a candidate-level assignment.
	RoundName string
	IsRound   bool
	// InterviewID links a per-round evaluation back to the interview whose
	// feedback is the score. Zero for a candidate-level assignment.
	InterviewID int64
	// InterviewReady is false while the candidate still has unscored rounds.
	// Employee assignments and per-round evaluations are always ready.
	InterviewReady bool
	InterviewScore float64
	InterviewScoreTotal float64
	InterviewScoreAvailable bool
	CompletedReviews int
	TotalInterviews int
}

type CommentView struct {
	Comment    evalmodel.EvaluationComment
	AuthorName string
}

type QuestionView struct {
	Model        evalmodel.EvaluationQuestion
	Required     bool
	TextValue    string
	NumericValue string
	RatingValue  string
	BooleanValue string
}

type SectionView struct {
	Name        string
	Description string
	Questions   []QuestionView
}

// CandidateGroup keeps the assignment dialog's candidates under the job
// posting they applied for.
type CandidateGroup struct {
	PostingID    int64
	PostingTitle string
	Candidates   []CandidateOption
}

type EvaluationsPageData struct {
	Periods         []evalmodel.EvaluationPeriod
	Templates       []evalmodel.EvaluationTemplate
	Assignments     []AssignmentView
	Workers         []core.WorkerOption
	Candidates      []CandidateOption
	CandidateGroups []CandidateGroup
}

type FillPageData struct {
	Assignment     evalmodel.EvaluationAssignment
	EvaluateeName  string
	RoundName      string
	InterviewReady bool
	InterviewScore float64
	InterviewScoreTotal float64
	InterviewScoreAvailable bool
	CompletedReviews int
	TotalInterviews int
	Sections       []SectionView
	Comments       []CommentView
	Result         *evalmodel.EvaluationResult
}
