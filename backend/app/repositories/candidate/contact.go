package candidate

import (
	candmodel "ems/app/models/candidate"
	"ems/database"
)

// CreateContact records one contact attempt with an applicant.
func (r *CandidateRepository) CreateContact(contact *candmodel.ContactLog) error {
	result, err := database.DB.Exec(
		`INSERT INTO candidate_contact_log
		 (candidate_id, contacted_at, method, direction, summary, outcome, next_follow_up_at, created_by_user_id)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		contact.CandidateID, contact.ContactedAt, contact.Method, contact.Direction,
		contact.Summary, contact.Outcome, contact.NextFollowUpAt, nullID(contact.CreatedByID),
	)
	if err != nil {
		return err
	}
	contact.ID, _ = result.LastInsertId()
	return nil
}

// FindContactsByCandidate returns a candidate's contact history, newest first.
func (r *CandidateRepository) FindContactsByCandidate(candidateID int64) ([]candmodel.ContactLog, error) {
	var contacts []candmodel.ContactLog
	err := database.DB.Select(&contacts, contactSelect+`
		 WHERE cl.candidate_id = ?
		 ORDER BY cl.contacted_at DESC, cl.id DESC`, candidateID)
	if err != nil {
		return nil, err
	}
	return contacts, nil
}

// FindContactsForActiveCandidates returns contact rows for every candidate
// still in the pipeline, newest first. The dashboard walks this list and keeps
// the first row per candidate, which is their latest contact.
func (r *CandidateRepository) FindContactsForActiveCandidates() ([]candmodel.ContactLog, error) {
	var contacts []candmodel.ContactLog
	err := database.DB.Select(&contacts, contactSelect+`
		 WHERE c.status NOT IN ('hired', 'rejected')
		 ORDER BY cl.contacted_at DESC, cl.id DESC`)
	if err != nil {
		return nil, err
	}
	return contacts, nil
}

const contactSelect = `SELECT cl.id, cl.candidate_id, c.full_name AS candidate_name,
	 cl.contacted_at, cl.method, cl.direction,
	 COALESCE(cl.summary, '') AS summary,
	 COALESCE(cl.outcome, '') AS outcome,
	 cl.next_follow_up_at,
	 COALESCE(cl.created_by_user_id, 0) AS created_by_user_id,
	 COALESCE(u.name, '') AS created_by_name,
	 cl.created_at
	 FROM candidate_contact_log cl
	 JOIN candidates c ON c.id = cl.candidate_id
	 LEFT JOIN users u ON u.id = cl.created_by_user_id`

func nullID(id int64) any {
	if id > 0 {
		return id
	}
	return nil
}
