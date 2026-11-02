package recruitment

import (
	recmodel "ems/app/models/recruitment"
	"ems/database"
	"strings"
)

type InterviewRepository struct{}

const interviewListItemColumns = `i.id, i.candidate_id, i.interview_round_id, i.interviewer_id,
	i.scheduled_at, i.duration_minutes,
	COALESCE(i.location, '') AS location, COALESCE(i.meeting_link, '') AS meeting_link,
	i.status, COALESCE(i.notes, '') AS notes, i.created_at, i.updated_at,
	c.full_name AS candidate_name, ir.name AS round_name,
	COALESCE((
		SELECT GROUP_CONCAT(name, ', ')
		FROM (
			SELECT u.name AS name
			FROM interview_interviewers ii
			JOIN users u ON u.id = ii.interviewer_id
			WHERE ii.interview_id = i.id
			ORDER BY u.name
		)
	), (SELECT u2.name FROM users u2 WHERE u2.id = i.interviewer_id), '') AS interviewer_name`

func (r *InterviewRepository) Create(interview *recmodel.Interview) error {
	panel := uniqueInterviewerIDs(interview.InterviewerIDs)
	if len(panel) == 0 && interview.InterviewerID > 0 {
		panel = []int64{interview.InterviewerID}
	}
	if len(panel) > 0 {
		interview.InterviewerID = panel[0]
	}
	interview.InterviewerIDs = panel

	tx, err := database.DB.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	result, err := tx.Exec(
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

	for _, interviewerID := range panel {
		if _, err := tx.Exec(
			`INSERT OR IGNORE INTO interview_interviewers (interview_id, interviewer_id) VALUES (?, ?)`,
			interview.ID, interviewerID,
		); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// uniqueInterviewerIDs drops empty and repeated ids, keeping the order HR
// picked so the first selection stays the lead interviewer.
func uniqueInterviewerIDs(ids []int64) []int64 {
	unique := make([]int64, 0, len(ids))
	seen := make(map[int64]bool, len(ids))
	for _, id := range ids {
		if id <= 0 || seen[id] {
			continue
		}
		seen[id] = true
		unique = append(unique, id)
	}
	return unique
}

const interviewListItemJoins = `FROM interviews i
	 JOIN candidates c ON c.id = i.candidate_id
	 JOIN interview_rounds ir ON ir.id = i.interview_round_id`

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

// InterviewFilter narrows the schedule table in SQL so large pipelines do not
// have to be filtered in the page handler.
//
// CandidateIDs and PanelUserID are the viewer's scope: they are OR-ed together
// because somebody can be both an assigned evaluator for a candidate and a
// panel member on another interview. RestrictToViewer marks the presence of a
// scope explicitly; without it an empty scope would accidentally mean "all".
type InterviewFilter struct {
	CandidateIDs     []int64
	PanelUserID      int64
	RestrictToViewer bool
	PostingID        int64
	RoundID          int64
	InterviewerID    int64
	Status           string
	Search           string
}

func (r *InterviewRepository) FindFiltered(filter InterviewFilter) ([]recmodel.InterviewListItem, error) {
	if filter.RestrictToViewer && len(filter.CandidateIDs) == 0 && filter.PanelUserID <= 0 {
		return []recmodel.InterviewListItem{}, nil
	}

	query := `SELECT ` + interviewListItemColumns + ` ` + interviewListItemJoins + ` WHERE 1=1`
	args := make([]any, 0)

	if filter.RestrictToViewer {
		switch {
		case len(filter.CandidateIDs) > 0 && filter.PanelUserID > 0:
			placeholders := make([]string, len(filter.CandidateIDs))
			for i, id := range filter.CandidateIDs {
				placeholders[i] = "?"
				args = append(args, id)
			}
			query += ` AND (i.candidate_id IN (` + strings.Join(placeholders, ",") + `)
			          OR EXISTS (SELECT 1 FROM interview_interviewers viewer
			                     WHERE viewer.interview_id = i.id AND viewer.interviewer_id = ?))`
			args = append(args, filter.PanelUserID)
		case len(filter.CandidateIDs) > 0:
			placeholders := make([]string, len(filter.CandidateIDs))
			for i, id := range filter.CandidateIDs {
				placeholders[i] = "?"
				args = append(args, id)
			}
			query += ` AND i.candidate_id IN (` + strings.Join(placeholders, ",") + `)`
		case filter.PanelUserID > 0:
			query += ` AND EXISTS (SELECT 1 FROM interview_interviewers viewer
			                       WHERE viewer.interview_id = i.id AND viewer.interviewer_id = ?)`
			args = append(args, filter.PanelUserID)
		}
	}

	if filter.PostingID > 0 {
		query += ` AND c.job_posting_id = ?`
		args = append(args, filter.PostingID)
	}
	if filter.RoundID > 0 {
		query += ` AND i.interview_round_id = ?`
		args = append(args, filter.RoundID)
	}
	if filter.InterviewerID > 0 {
		query += ` AND EXISTS (SELECT 1 FROM interview_interviewers chosen
		                       WHERE chosen.interview_id = i.id AND chosen.interviewer_id = ?)`
		args = append(args, filter.InterviewerID)
	}
	if filter.Status != "" {
		query += ` AND i.status = ?`
		args = append(args, filter.Status)
	}
	if search := strings.TrimSpace(filter.Search); search != "" {
		query += ` AND (LOWER(c.full_name) LIKE ? OR LOWER(ir.name) LIKE ? OR LOWER(COALESCE(i.location, '')) LIKE ?)`
		like := "%" + strings.ToLower(search) + "%"
		args = append(args, like, like, like)
	}

	query += ` ORDER BY i.scheduled_at DESC`
	var interviews []recmodel.InterviewListItem
	if err := database.DB.Select(&interviews, query, args...); err != nil {
		return nil, err
	}
	return interviews, nil
}

// FindPanelMembers lists everybody assigned to an interview panel.
func (r *InterviewRepository) FindPanelMembers(interviewID int64) ([]recmodel.InterviewPanelMember, error) {
	var members []recmodel.InterviewPanelMember
	err := database.DB.Select(&members,
		`SELECT u.id AS user_id, u.name AS name
		 FROM interview_interviewers ii
		 JOIN users u ON u.id = ii.interviewer_id
		 WHERE ii.interview_id = ?
		 ORDER BY u.name`,
		interviewID,
	)
	if err != nil {
		return nil, err
	}
	return members, nil
}

// IsPanelMember answers whether a user may score a particular interview.
func (r *InterviewRepository) IsPanelMember(interviewID, userID int64) (bool, error) {
	if interviewID <= 0 || userID <= 0 {
		return false, nil
	}
	var count int
	err := database.DB.Get(&count,
		`SELECT COUNT(*) FROM interview_interviewers WHERE interview_id = ? AND interviewer_id = ?`,
		interviewID, userID,
	)
	return count > 0, err
}

// FindAssignedCandidateIDs lists the candidates for whom this user is the
// assigned evaluator. Interviewers with only panel access do not appear here.
func (r *InterviewRepository) FindAssignedCandidateIDs(userID int64) ([]int64, error) {
	if userID <= 0 {
		return nil, nil
	}
	var ids []int64
	err := database.DB.Select(&ids,
		`SELECT DISTINCT candidate_id FROM evaluation_assignments
		 WHERE evaluator_id = ? AND COALESCE(candidate_id, 0) > 0
		   AND LOWER(status) NOT IN ('rejected')
		 ORDER BY candidate_id`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	return ids, nil
}

// FindAwaitingFeedback returns interviews that have no score yet. The caller
// decides which ones are actually due, because scheduled_at formatting varies
// with the driver and a string comparison in SQL would be fragile.
func (r *InterviewRepository) FindAwaitingFeedback() ([]recmodel.InterviewListItem, error) {
	var interviews []recmodel.InterviewListItem
	err := database.DB.Select(&interviews,
		`SELECT `+interviewListItemColumns+` `+interviewListItemJoins+`
		 WHERE EXISTS (
		     SELECT 1 FROM interview_interviewers ii
		     WHERE ii.interview_id = i.id
		       AND NOT EXISTS (
		           SELECT 1 FROM interview_feedback f
		           WHERE f.interview_id = i.id AND f.interviewer_id = ii.interviewer_id
		             AND f.score IS NOT NULL
		       )
		 )
		 AND i.status NOT IN ('cancelled', 'no_show')
		 ORDER BY i.scheduled_at ASC`,
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
