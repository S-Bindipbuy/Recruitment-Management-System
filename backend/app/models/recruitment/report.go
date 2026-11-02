package recruitment

type CandidateEvaluationSummary struct {
	CandidateID   int64  `json:"candidate_id" db:"candidate_id"`
	CandidateName string `json:"candidate_name" db:"candidate_name"`
	// CandidateStatus feeds the pending/approval decision but is not shown
	// separately; the pipeline status is rendered from the candidate row.
	CandidateStatus string `json:"-" db:"candidate_status"`
	PositionName    string `json:"position_name" db:"position_name"`
	JobPostingID    int64  `json:"job_posting_id" db:"job_posting_id"`
	PostingTitle    string `json:"posting_title" db:"posting_title"`
	TotalInterviews int    `json:"total_interviews" db:"total_interviews"`
	// LatestInterviewID points at the most recent interview this candidate was
	// scored on, so ranking tables can link a name straight to its feedback.
	LatestInterviewID int64   `json:"latest_interview_id" db:"latest_interview_id"`
	CompletedReviews  int     `json:"completed_reviews" db:"completed_reviews"`
	InterviewerCount  int     `json:"interviewer_count" db:"interviewer_count"`
	CompletedRounds   int     `json:"completed_rounds" db:"completed_rounds"`
	TotalRounds       int     `json:"total_rounds" db:"total_rounds"`
	AverageScore      float64 `json:"average_score" db:"average_score"`
	TotalScore        float64 `json:"total_score" db:"total_score"`
	HighestScore      float64 `json:"highest_score" db:"highest_score"`
	LowestScore       float64 `json:"lowest_score" db:"lowest_score"`
	Rank              int     `json:"rank" db:"-"`
	FinalDecision     string  `json:"final_decision" db:"final_decision"`
}

type InterviewerActivity struct {
	InterviewerID    int64   `json:"interviewer_id" db:"interviewer_id"`
	InterviewerName  string  `json:"interviewer_name" db:"interviewer_name"`
	TotalInterviews  int     `json:"total_interviews" db:"total_interviews"`
	CompletedReviews int     `json:"completed_reviews" db:"completed_reviews"`
	PendingReviews   int     `json:"pending_reviews" db:"pending_reviews"`
	AverageScore     float64 `json:"average_score" db:"average_score"`
}

type HiringDecision struct {
	CandidateID    int64   `json:"candidate_id" db:"candidate_id"`
	CandidateName  string  `json:"candidate_name" db:"candidate_name"`
	JobPostingID   int64   `json:"job_posting_id" db:"job_posting_id"`
	PositionName   string  `json:"position_name" db:"position_name"`
	DepartmentName string  `json:"department_name" db:"department_name"`
	AverageScore   float64 `json:"average_score" db:"average_score"`
	InterviewCount int     `json:"interview_count" db:"interview_count"`
	// LatestInterviewID lets the reports link a candidate straight to feedback.
	LatestInterviewID int64  `json:"latest_interview_id" db:"latest_interview_id"`
	Decision          string `json:"decision" db:"decision"`
	DecisionDate      string `json:"decision_date" db:"decision_date"`
	DecidedBy         string `json:"decided_by" db:"decided_by"`
}
