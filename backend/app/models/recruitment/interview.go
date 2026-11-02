package recruitment

import "time"

type InterviewRound struct {
	ID           int64 `json:"id" db:"id"`
	JobPostingID int64 `json:"job_posting_id" db:"job_posting_id"`
	// EvaluationTemplateID is the question bank the round draws from. Zero
	// means the round has no questions attached.
	EvaluationTemplateID int64     `json:"evaluation_template_id" db:"evaluation_template_id"`
	Name                 string    `json:"name" db:"name"`
	RoundNumber          int       `json:"round_number" db:"round_number"`
	RoundType            string    `json:"round_type" db:"round_type"`
	Description          string    `json:"description" db:"description"`
	MaxScore             float64   `json:"max_score" db:"max_score"`
	PassScore            float64   `json:"pass_score" db:"pass_score"`
	Weight               float64   `json:"weight" db:"weight"`
	Status               string    `json:"status" db:"status"`
	CreatedAt            time.Time `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time `json:"updated_at" db:"updated_at"`
}

// InterviewRoundSection joins a round to one section of its evaluation
// template. A large template is split across rounds, so each round only asks
// the questions that belong to the sections picked for it.
type InterviewRoundSection struct {
	ID                  int64     `json:"id" db:"id"`
	InterviewRoundID    int64     `json:"interview_round_id" db:"interview_round_id"`
	EvaluationSectionID int64     `json:"evaluation_section_id" db:"evaluation_section_id"`
	CreatedAt           time.Time `json:"created_at" db:"created_at"`
}

// InterviewFeedbackAnswer is one question answered during an interview. It
// mirrors EvaluationResponse so the same question types render the same way.
type InterviewFeedbackAnswer struct {
	ID                   int64     `json:"id" db:"id"`
	FeedbackID           int64     `json:"feedback_id" db:"feedback_id"`
	EvaluationQuestionID int64     `json:"evaluation_question_id" db:"evaluation_question_id"`
	RatingValue          *float64  `json:"rating_value" db:"rating_value"`
	TextValue            string    `json:"text_value" db:"text_value"`
	NumericValue         *float64  `json:"numeric_value" db:"numeric_value"`
	BooleanValue         *bool     `json:"boolean_value" db:"boolean_value"`
	CreatedAt            time.Time `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time `json:"updated_at" db:"updated_at"`
}

type Interview struct {
	ID               int64 `json:"id" db:"id"`
	CandidateID      int64 `json:"candidate_id" db:"candidate_id"`
	InterviewRoundID int64 `json:"interview_round_id" db:"interview_round_id"`
	InterviewerID    int64 `json:"interviewer_id" db:"interviewer_id"`
	// InterviewerIDs is the full panel. InterviewerID mirrors the first member
	// so existing feedback and report queries keep a stable attribution.
	InterviewerIDs  []int64   `json:"interviewer_ids" db:"-"`
	ScheduledAt     time.Time `json:"scheduled_at" db:"scheduled_at"`
	DurationMinutes int       `json:"duration_minutes" db:"duration_minutes"`
	Location        string    `json:"location" db:"location"`
	MeetingLink     string    `json:"meeting_link" db:"meeting_link"`
	Status          string    `json:"status" db:"status"`
	Notes           string    `json:"notes" db:"notes"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

type InterviewListItem struct {
	Interview
	CandidateName   string `json:"candidate_name" db:"candidate_name"`
	RoundName       string `json:"round_name" db:"round_name"`
	InterviewerName string `json:"interviewer_name" db:"interviewer_name"`
}

// InterviewPanelMember is one person on an interview panel. The feedback page
// uses it to show who has scored and to let a manager enter a score on behalf
// of a specific panel member.
type InterviewPanelMember struct {
	UserID int64  `json:"user_id" db:"user_id"`
	Name   string `json:"name" db:"name"`
}

type InterviewFeedback struct {
	ID              int64      `json:"id" db:"id"`
	InterviewID     int64      `json:"interview_id" db:"interview_id"`
	InterviewerID   int64      `json:"interviewer_id" db:"interviewer_id"`
	InterviewerName string     `json:"interviewer_name" db:"interviewer_name"`
	RecordedByID    int64      `json:"recorded_by_user_id" db:"recorded_by_user_id"`
	RecordedByName  string     `json:"recorded_by_name" db:"recorded_by_name"`
	Score           *float64   `json:"score" db:"score"`
	Recommendation  string     `json:"recommendation" db:"recommendation"`
	Strengths       string     `json:"strengths" db:"strengths"`
	Weaknesses      string     `json:"weaknesses" db:"weaknesses"`
	Comments        string     `json:"comments" db:"comments"`
	SubmittedAt     *time.Time `json:"submitted_at" db:"submitted_at"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
}
