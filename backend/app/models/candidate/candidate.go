package candidate

import "time"

type Candidate struct {
	ID           int64     `json:"id" db:"id"`
	FullName     string    `json:"full_name" db:"full_name"`
	Email        string    `json:"email" db:"email"`
	Phone        string    `json:"phone" db:"phone"`
	JobPostingID *int64    `json:"job_posting_id" db:"job_posting_id"`
	Skills       string    `json:"skills" db:"skills"`
	CVPath       string    `json:"-" db:"cv_path"`
	CVName       string    `json:"cv_name" db:"cv_name"`
	CVMimeType   string    `json:"cv_mime_type" db:"cv_mime_type"`
	CVSize       int64     `json:"cv_size" db:"cv_size"`
	Status       string    `json:"status" db:"status"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

type CandidateListItem struct {
	Candidate
	PostingTitle string `json:"posting_title" db:"posting_title"`
}
