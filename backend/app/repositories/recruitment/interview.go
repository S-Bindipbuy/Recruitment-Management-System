package recruitment

import (
	recmodel "ems/app/models/recruitment"
	"ems/database"
)

type InterviewRepository struct{}

const interviewListItemColumns = `i.id, i.candidate_id, i.interview_round_id, i.interviewer_id,
	i.scheduled_at, i.duration_minutes, i.location, i.meeting_link,
	i.status, i.notes, i.created_at, i.updated_at,
	c.full_name AS candidate_name, ir.name AS round_name, u.name AS interviewer_name`

func (r *InterviewRepository) Create(interview *recmodel.Interview) error {
	result, err := database.DB.Exec(
		`INSERT INTO interviews (candidate_id, interview_round_id, interviewer_id, scheduled_at,
		 duration_minutes, location, meeting_link, status, notes)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		interview.CandidateID, interview.InterviewRoundID, interview.InterviewerID,
		interview.ScheduledAt, interview.DurationMinutes, interview.Location,
		interview.MeetingLink, interview.Status, interview.Notes,
	)
	if err != nil {
		return err
	}
	interview.ID, _ = result.LastInsertId()
	return nil
}

const interviewListItemJoins = `FROM interviews i
	 JOIN candidates c ON c.id = i.candidate_id
	 JOIN interview_rounds ir ON ir.id = i.interview_round_id
	 JOIN users u ON u.id = i.interviewer_id`

func (r *InterviewRepository) FindAll() ([]recmodel.InterviewListItem, error) {
	var interviews []recmodel.InterviewListItem
	err := database.DB.Select(&interviews,
		`SELECT `+interviewListItemColumns+` `+interviewListItemJoins+` ORDER BY i.scheduled_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	return interviews, nil
}

func (r *InterviewRepository) FindByID(id int64) (*recmodel.InterviewListItem, error) {
	iv := &recmodel.InterviewListItem{}
	err := database.DB.Get(iv,
		`SELECT `+interviewListItemColumns+` `+interviewListItemJoins+` WHERE i.id = ?`, id,
	)
	if err != nil {
		return nil, err
	}
	return iv, nil
}

func (r *InterviewRepository) FindByCandidateID(candidateID int64) ([]recmodel.InterviewListItem, error) {
	var interviews []recmodel.InterviewListItem
	err := database.DB.Select(&interviews,
		`SELECT `+interviewListItemColumns+` `+interviewListItemJoins+`
		 WHERE i.candidate_id = ? ORDER BY i.scheduled_at`, candidateID,
	)
	if err != nil {
		return nil, err
	}
	return interviews, nil
}

func (r *InterviewRepository) UpdateStatus(id int64, status string) error {
	_, err := database.DB.Exec("UPDATE interviews SET status=?, updated_at=CURRENT_TIMESTAMP WHERE id=?", status, id)
	return err
}

func (r *InterviewRepository) Delete(id int64) error {
	_, err := database.DB.Exec("DELETE FROM interviews WHERE id = ?", id)
	return err
}

// Interview Feedback Repository
