package recruitment

import "time"

type InterviewRound struct {
	ID           int64     `json:"id" db:"id"`
	JobPostingID int64     `json:"job_posting_id" db:"job_posting_id"`
	Name         string    `json:"name" db:"name"`
	RoundNumber  int       `json:"round_number" db:"round_number"`
	RoundType    string    `json:"round_type" db:"round_type"`
	Description  string    `json:"description" db:"description"`
	MaxScore     float64   `json:"max_score" db:"max_score"`
	PassScore    float64   `json:"pass_score" db:"pass_score"`
	Status       string    `json:"status" db:"status"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

type Interview struct {
	ID               int64     `json:"id" db:"id"`
	CandidateID      int64     `json:"candidate_id" db:"candidate_id"`
	InterviewRoundID int64     `json:"interview_round_id" db:"interview_round_id"`
	InterviewerID    int64     `json:"interviewer_id" db:"interviewer_id"`
	ScheduledAt      time.Time `json:"scheduled_at" db:"scheduled_at"`
	DurationMinutes  int       `json:"duration_minutes" db:"duration_minutes"`
	Location         string    `json:"location" db:"location"`
	MeetingLink      string    `json:"meeting_link" db:"meeting_link"`
	Status           string    `json:"status" db:"status"`
	Notes            string    `json:"notes" db:"notes"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}

type InterviewListItem struct {
	Interview
	CandidateName   string `json:"candidate_name" db:"candidate_name"`
	RoundName       string `json:"round_name" db:"round_name"`
	InterviewerName string `json:"interviewer_name" db:"interviewer_name"`
}

type InterviewFeedback struct {
	ID             int64      `json:"id" db:"id"`
	InterviewID    int64      `json:"interview_id" db:"interview_id"`
	InterviewerID  int64      `json:"interviewer_id" db:"interviewer_id"`
	Score          *float64   `json:"score" db:"score"`
	Recommendation string     `json:"recommendation" db:"recommendation"`
	Strengths      string     `json:"strengths" db:"strengths"`
	Weaknesses     string     `json:"weaknesses" db:"weaknesses"`
	Comments       string     `json:"comments" db:"comments"`
	SubmittedAt    *time.Time `json:"submitted_at" db:"submitted_at"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}
