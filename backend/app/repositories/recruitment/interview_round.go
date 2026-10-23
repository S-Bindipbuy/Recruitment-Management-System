package recruitment

import (
	recmodel "ems/app/models/recruitment"
	"ems/database"
)

type InterviewRoundRepository struct{}

func (r *InterviewRoundRepository) Create(round *recmodel.InterviewRound) error {
	result, err := database.DB.Exec(
		`INSERT INTO interview_rounds (job_posting_id, name, round_number, round_type, description, max_score, pass_score, status)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		round.JobPostingID, round.Name, round.RoundNumber, round.RoundType,
		round.Description, round.MaxScore, round.PassScore, round.Status,
	)
	if err != nil {
		return err
	}
	round.ID, _ = result.LastInsertId()
	return nil
}

func (r *InterviewRoundRepository) FindByJobPostingID(postingID int64) ([]recmodel.InterviewRound, error) {
	var rounds []recmodel.InterviewRound
	err := database.DB.Select(&rounds,
		`SELECT id, job_posting_id, name, round_number, round_type, description,
		 max_score, pass_score, status, created_at, updated_at
		 FROM interview_rounds WHERE job_posting_id = ? ORDER BY round_number`,
		postingID,
	)
	if err != nil {
		return nil, err
	}
	return rounds, nil
}

func (r *InterviewRoundRepository) Delete(id int64) error {
	_, err := database.DB.Exec("DELETE FROM interview_rounds WHERE id = ?", id)
	return err
}

// Interview Repository
