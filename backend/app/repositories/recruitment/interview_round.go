package recruitment

import (
	evalmodel "ems/app/models/evaluation"
	recmodel "ems/app/models/recruitment"
	"ems/database"
)

type InterviewRoundRepository struct{}

// nullableID turns the zero value into SQL NULL. Foreign keys must point at a
// real row, and 0 is not an evaluation template.
func nullableID(id int64) any {
	if id <= 0 {
		return nil
	}
	return id
}

func (r *InterviewRoundRepository) Create(round *recmodel.InterviewRound) error {
	result, err := database.DB.Exec(
		`INSERT INTO interview_rounds (job_posting_id, evaluation_template_id, name, round_number, round_type, description, max_score, pass_score, weight, status)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		round.JobPostingID, nullableID(round.EvaluationTemplateID), round.Name, round.RoundNumber, round.RoundType,
		round.Description, round.MaxScore, round.PassScore, round.Weight, round.Status,
	)
	if err != nil {
		return err
	}
	round.ID, _ = result.LastInsertId()
	return nil
}

func (r *InterviewRoundRepository) FindByID(id int64) (*recmodel.InterviewRound, error) {
	round := &recmodel.InterviewRound{}
	err := database.DB.Get(round,
		`SELECT id, job_posting_id, COALESCE(evaluation_template_id, 0) AS evaluation_template_id,
		 name, round_number, round_type, COALESCE(description, '') AS description,
		 max_score, pass_score, weight, status, created_at, updated_at
		 FROM interview_rounds WHERE id = ?`,
		id,
	)
	if err != nil {
		return nil, err
	}
	return round, nil
}

func (r *InterviewRoundRepository) FindByJobPostingID(postingID int64) ([]recmodel.InterviewRound, error) {
	var rounds []recmodel.InterviewRound
	err := database.DB.Select(&rounds,
		`SELECT id, job_posting_id, COALESCE(evaluation_template_id, 0) AS evaluation_template_id,
		 name, round_number, round_type, COALESCE(description, '') AS description,
		 max_score, pass_score, weight, status, created_at, updated_at
		 FROM interview_rounds WHERE job_posting_id = ? ORDER BY round_number`,
		postingID,
	)
	if err != nil {
		return nil, err
	}
	return rounds, nil
}

// ReplaceSections sets the round's question scope to exactly sectionIDs. The
// delete-and-insert runs in one transaction so a failed insert cannot leave a
// round with half of its sections.
func (r *InterviewRoundRepository) ReplaceSections(roundID int64, sectionIDs []int64) error {
	tx, err := database.DB.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM interview_round_sections WHERE interview_round_id = ?`, roundID); err != nil {
		return err
	}
	for _, sectionID := range sectionIDs {
		if sectionID <= 0 {
			continue
		}
		if _, err := tx.Exec(
			`INSERT OR IGNORE INTO interview_round_sections (interview_round_id, evaluation_section_id) VALUES (?, ?)`,
			roundID, sectionID,
		); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *InterviewRoundRepository) FindSectionsByRoundID(roundID int64) ([]evalmodel.EvaluationSection, error) {
	var sections []evalmodel.EvaluationSection
	err := database.DB.Select(&sections,
		`SELECT s.id, s.evaluation_template_id, s.name, COALESCE(s.description, '') AS description,
		 s.sort_order, s.created_at, s.updated_at
		 FROM evaluation_sections s
		 JOIN interview_round_sections rs ON rs.evaluation_section_id = s.id
		 WHERE rs.interview_round_id = ?
		 ORDER BY s.sort_order, s.id`,
		roundID,
	)
	if err != nil {
		return nil, err
	}
	return sections, nil
}

func (r *InterviewRoundRepository) Delete(id int64) error {
	_, err := database.DB.Exec("DELETE FROM interview_rounds WHERE id = ?", id)
	return err
}

func (r *InterviewRoundRepository) Update(round *recmodel.InterviewRound) error {
	_, err := database.DB.Exec(
		`UPDATE interview_rounds SET name = ?, round_number = ?, round_type = ?, description = ?,
		 max_score = ?, pass_score = ?, weight = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		round.Name, round.RoundNumber, round.RoundType, round.Description,
		round.MaxScore, round.PassScore, round.Weight, round.ID,
	)
	return err
}

func (r *InterviewRoundRepository) HasInterviews(id int64) (bool, error) {
	var hasInterviews bool
	err := database.DB.Get(&hasInterviews,
		`SELECT EXISTS(SELECT 1 FROM interviews WHERE interview_round_id = ?)`, id,
	)
	return hasInterviews, err
}

// Interview Repository
