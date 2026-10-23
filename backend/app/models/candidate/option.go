package candidate

// CandidateOption is an applicant, identified the way the wizard and the
// assignment dialog show them.
type CandidateOption struct {
	ID     int64  `json:"id" db:"id"`
	Name   string `json:"name" db:"full_name"`
	Email  string `json:"email" db:"email"`
	Status string `json:"status" db:"status"`
}
