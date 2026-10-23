package recruitment

type CandidateEvaluationSummary struct {
	CandidateID      int64   `json:"candidate_id" db:"candidate_id"`
	CandidateName    string  `json:"candidate_name" db:"candidate_name"`
	PositionName     string  `json:"position_name" db:"position_name"`
	TotalInterviews  int     `json:"total_interviews" db:"total_interviews"`
	CompletedReviews int     `json:"completed_reviews" db:"completed_reviews"`
	AverageScore     float64 `json:"average_score" db:"average_score"`
	HighestScore     float64 `json:"highest_score" db:"highest_score"`
	LowestScore      float64 `json:"lowest_score" db:"lowest_score"`
	FinalDecision    string  `json:"final_decision" db:"final_decision"`
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
	PositionName   string  `json:"position_name" db:"position_name"`
	DepartmentName string  `json:"department_name" db:"department_name"`
	AverageScore   float64 `json:"average_score" db:"average_score"`
	InterviewCount int     `json:"interview_count" db:"interview_count"`
	Decision       string  `json:"decision" db:"decision"`
	DecisionDate   string  `json:"decision_date" db:"decision_date"`
	DecidedBy      string  `json:"decided_by" db:"decided_by"`
}
