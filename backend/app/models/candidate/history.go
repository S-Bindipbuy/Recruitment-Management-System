package candidate

import "time"

// StatusHistory is one transition in a candidate's recruitment pipeline. The
// current status lives on Candidate; these rows are the audit trail and carry
// the rejection reason.
type StatusHistory struct {
	ID            int64     `json:"id" db:"id"`
	CandidateID   int64     `json:"candidate_id" db:"candidate_id"`
	FromStatus    string    `json:"from_status" db:"from_status"`
	ToStatus      string    `json:"to_status" db:"to_status"`
	Reason        string    `json:"reason" db:"reason"`
	ChangedByID   int64     `json:"changed_by_user_id" db:"changed_by_user_id"`
	ChangedByName string    `json:"changed_by_name" db:"changed_by_name"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

// PipelineStatuses is the ordered set of statuses HR can move a candidate
// through. "hired" is intentionally not settable by hand: accepting a
// candidate creates the employee record in the same transaction.
var PipelineStatuses = []string{"new", "screening", "interviewing", "rejected"}
