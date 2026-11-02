package candidate

import "time"

// ContactLog is one contact attempt with an applicant. Contact happens outside
// the system (phone, email, Telegram, in person); this row is HR's record of it.
type ContactLog struct {
	ID             int64      `json:"id" db:"id"`
	CandidateID    int64      `json:"candidate_id" db:"candidate_id"`
	CandidateName  string     `json:"candidate_name" db:"candidate_name"`
	ContactedAt    time.Time  `json:"contacted_at" db:"contacted_at"`
	Method         string     `json:"method" db:"method"`
	Direction      string     `json:"direction" db:"direction"`
	Summary        string     `json:"summary" db:"summary"`
	Outcome        string     `json:"outcome" db:"outcome"`
	NextFollowUpAt *time.Time `json:"next_follow_up_at" db:"next_follow_up_at"`
	CreatedByID    int64      `json:"created_by_user_id" db:"created_by_user_id"`
	CreatedByName  string     `json:"created_by_name" db:"created_by_name"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`

	// Derived for the dashboard task queue; not stored.
	Overdue   bool `json:"overdue" db:"-"`
	DueInDays int  `json:"due_in_days" db:"-"`
}

// ContactMethods are the channels HR can record. The system does not send
// anything itself - these are labels for what already happened.
var ContactMethods = []string{"phone", "email", "telegram", "in_person", "other"}

// ContactDirections distinguishes HR reaching out from the candidate replying.
var ContactDirections = []string{"outbound", "inbound"}

// ContactOutcomes are the short result labels shown in lists.
var ContactOutcomes = []string{"reached", "no_answer", "scheduled", "declined", "other"}
