package candidate

import (
	"database/sql"
	candmodel "ems/app/models/candidate"
	"ems/app/repositories"
	"ems/database"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

type CandidateRepository struct{}

func (r *CandidateRepository) Create(candidate *candmodel.Candidate) error {
	result, err := database.DB.Exec(
		`INSERT INTO candidates (full_name, email, phone, job_posting_id, skills, cv_path, cv_name, cv_mime_type, cv_size, status)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		candidate.FullName, candidate.Email, candidate.Phone, candidate.JobPostingID, candidate.Skills,
		candidate.CVPath, candidate.CVName, candidate.CVMimeType, candidate.CVSize, candidate.Status,
	)
	if err != nil {
		return err
	}
	candidate.ID, _ = result.LastInsertId()
	return nil
}

func (r *CandidateRepository) FindAll() ([]candmodel.CandidateListItem, error) {
	var candidates []candmodel.CandidateListItem
	err := database.DB.Select(&candidates,
		`SELECT c.id, c.full_name, c.email, c.phone, c.job_posting_id, c.skills, c.cv_path, c.cv_name,
		 c.cv_mime_type, c.cv_size, c.status, c.created_at, c.updated_at,
		 COALESCE(jp.title, '') AS posting_title,
		 COALESCE((SELECT h.reason FROM candidate_status_history h
		           WHERE h.candidate_id = c.id AND h.to_status = c.status
		           ORDER BY h.id DESC LIMIT 1), '') AS status_reason
		 FROM candidates c
		 LEFT JOIN job_postings jp ON jp.id = c.job_posting_id
		 ORDER BY c.created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	return candidates, nil
}

// FindByIDs lists only the candidates a scoped viewer is allowed to see. It
// keeps the scheduling picker small for evaluators instead of loading the
// whole pipeline and filtering it in the page.
func (r *CandidateRepository) FindByIDs(ids []int64) ([]candmodel.CandidateListItem, error) {
	if len(ids) == 0 {
		return []candmodel.CandidateListItem{}, nil
	}
	placeholders := make([]string, len(ids))
	args := make([]any, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}
	var candidates []candmodel.CandidateListItem
	err := database.DB.Select(&candidates,
		`SELECT c.id, c.full_name, c.email, c.phone, c.job_posting_id, c.skills, c.cv_path, c.cv_name,
		 c.cv_mime_type, c.cv_size, c.status, c.created_at, c.updated_at,
		 COALESCE(jp.title, '') AS posting_title,
		 COALESCE((SELECT h.reason FROM candidate_status_history h
		           WHERE h.candidate_id = c.id AND h.to_status = c.status
		           ORDER BY h.id DESC LIMIT 1), '') AS status_reason
		 FROM candidates c
		 LEFT JOIN job_postings jp ON jp.id = c.job_posting_id
		 WHERE c.id IN (`+strings.Join(placeholders, ",")+`)
		 ORDER BY c.created_at DESC`, args...,
	)
	if err != nil {
		return nil, err
	}
	return candidates, nil
}

func (r *CandidateRepository) FindByJobPostingID(postingID int64) ([]candmodel.CandidateListItem, error) {
	var candidates []candmodel.CandidateListItem
	err := database.DB.Select(&candidates,
		`SELECT c.id, c.full_name, c.email, c.phone, c.job_posting_id, c.skills, c.cv_path, c.cv_name,
		 c.cv_mime_type, c.cv_size, c.status, c.created_at, c.updated_at,
		 COALESCE(jp.title, '') AS posting_title,
		 COALESCE((SELECT h.reason FROM candidate_status_history h
		           WHERE h.candidate_id = c.id AND h.to_status = c.status
		           ORDER BY h.id DESC LIMIT 1), '') AS status_reason
		 FROM candidates c
		 LEFT JOIN job_postings jp ON jp.id = c.job_posting_id
		 WHERE c.job_posting_id = ?
		 ORDER BY c.created_at DESC`, postingID,
	)
	if err != nil {
		return nil, err
	}
	return candidates, nil
}

func (r *CandidateRepository) FindByID(id int64) (*candmodel.Candidate, error) {
	candidate := &candmodel.Candidate{}
	err := database.DB.Get(candidate,
		`SELECT id, full_name, email, phone, job_posting_id, skills, cv_path, cv_name, cv_mime_type, cv_size, status, created_at, updated_at
		 FROM candidates WHERE id = ?`, id,
	)
	if err != nil {
		return nil, err
	}
	return candidate, nil
}

func (r *CandidateRepository) UpdateStatus(id int64, status string) error {
	_, err := database.DB.Exec("UPDATE candidates SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?", status, id)
	return err
}

// SetStatus moves a candidate to a new pipeline status and records the
// transition in candidate_status_history. A no-op is not written when the
// status is unchanged, so refresh/replay does not create noise.
func (r *CandidateRepository) SetStatus(id int64, status, reason string, changedBy int64) error {
	tx, err := database.DB.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var current string
	if err := tx.Get(&current, `SELECT status FROM candidates WHERE id = ?`, id); err != nil {
		return err
	}
	if current == status {
		return nil
	}

	if _, err := tx.Exec(
		`UPDATE candidates SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		status, id,
	); err != nil {
		return err
	}
	if err := insertStatusHistory(tx, id, current, status, reason, changedBy); err != nil {
		return err
	}
	return tx.Commit()
}

// MarkInterviewing advances a fresh applicant to "interviewing" when HR
// schedules their first interview. Candidates already further along (hired or
// rejected) are left untouched.
func (r *CandidateRepository) MarkInterviewing(id, changedBy int64) error {
	var current string
	if err := database.DB.Get(&current, `SELECT status FROM candidates WHERE id = ?`, id); err != nil {
		return err
	}
	if current != "new" && current != "screening" {
		return nil
	}
	return r.SetStatus(id, "interviewing", "", changedBy)
}

// FindStatusHistory returns a candidate's pipeline transitions, newest first.
func (r *CandidateRepository) FindStatusHistory(candidateID int64) ([]candmodel.StatusHistory, error) {
	var history []candmodel.StatusHistory
	err := database.DB.Select(&history,
		`SELECT h.id, h.candidate_id, h.from_status, h.to_status,
		 COALESCE(h.reason, '') AS reason,
		 COALESCE(h.changed_by_user_id, 0) AS changed_by_user_id,
		 COALESCE(u.name, '') AS changed_by_name,
		 h.created_at
		 FROM candidate_status_history h
		 LEFT JOIN users u ON u.id = h.changed_by_user_id
		 WHERE h.candidate_id = ? ORDER BY h.id DESC`, candidateID,
	)
	if err != nil {
		return nil, err
	}
	return history, nil
}

func insertStatusHistory(tx *sqlx.Tx, candidateID int64, from, to, reason string, changedBy int64) error {
	var userID any
	if changedBy > 0 {
		userID = changedBy
	}
	_, err := tx.Exec(
		`INSERT INTO candidate_status_history
		 (candidate_id, from_status, to_status, reason, changed_by_user_id)
		 VALUES (?, ?, ?, ?, ?)`,
		candidateID, from, to, strings.TrimSpace(reason), userID,
	)
	return err
}

// Hire creates an employee record for the candidate and records the
// assessment atomically.
func (r *CandidateRepository) Hire(candidateID, evaluatorID, roleID int64, rating float64, notes, passwordHash string) error {
	tx, err := database.DB.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var fullName, email, status string
	var jobPostingID sql.NullInt64
	err = tx.QueryRow(`
		SELECT c.full_name, c.email, c.job_posting_id, c.status
		FROM candidates c
		WHERE c.id = ?`, candidateID,
	).Scan(&fullName, &email, &jobPostingID, &status)
	if err == sql.ErrNoRows {
		return fmt.Errorf("candidate not found")
	}
	if err != nil {
		return err
	}
	if strings.EqualFold(status, "hired") {
		return fmt.Errorf("candidate has already been accepted")
	}

	var positionID, departmentID sql.NullInt64
	if jobPostingID.Valid {
		var pid int64
		if err := tx.QueryRow(`SELECT position_id FROM job_postings WHERE id = ?`, jobPostingID.Int64).Scan(&pid); err == nil && pid > 0 {
			positionID = sql.NullInt64{Int64: pid, Valid: true}
			var did int64
			if err := tx.QueryRow(`SELECT department_id FROM positions WHERE id = ?`, pid).Scan(&did); err == nil && did > 0 {
				departmentID = sql.NullInt64{Int64: did, Valid: true}
			}
		}
	}

	employeeCode := fmt.Sprintf("EMP-%06d", candidateID)
	if _, err := tx.Exec(
		`INSERT INTO employees (user_id, candidate_id, employee_code, department_id, position_id, hire_date, status)
		 VALUES (?, ?, ?, ?, ?, ?, 'active')`,
		nil, candidateID, employeeCode, departmentID, positionID, time.Now(),
	); err != nil {
		return err
	}

	actualEvaluator := evaluatorID
	var evalExists int
	_ = tx.QueryRow("SELECT COUNT(*) FROM users WHERE id = ?", actualEvaluator).Scan(&evalExists)
	if evalExists == 0 {
		// Fall back to a safe evaluator if session is invalid
		_ = tx.QueryRow("SELECT id FROM users ORDER BY id LIMIT 1").Scan(&actualEvaluator)
	}
	if actualEvaluator > 0 {
		if _, err := tx.Exec(
			`INSERT INTO candidate_assessments (candidate_id, evaluator_id, skill_score, notes, recommendation)
			 VALUES (?, ?, ?, ?, 'accepted')`,
			candidateID, actualEvaluator, rating, strings.TrimSpace(notes),
		); err != nil {
			return err
		}
	}
	if _, err := tx.Exec("UPDATE candidates SET status = 'hired', updated_at = CURRENT_TIMESTAMP WHERE id = ?", candidateID); err != nil {
		return err
	}
	reason := strings.TrimSpace(notes)
	if reason == "" {
		reason = "Accepted as employee"
	}
	if err := insertStatusHistory(tx, candidateID, status, "hired", reason, actualEvaluator); err != nil {
		return err
	}

	return tx.Commit()
}

// FindHireableOptions lists the applicants still in the pipeline - everybody
// the wizard or the assignment dialog could be pointed at.
func (r *CandidateRepository) FindHireableOptions() ([]candmodel.CandidateOption, error) {
	var options []candmodel.CandidateOption
	err := database.DB.Select(&options, `
		SELECT id, full_name, email, status FROM candidates
		WHERE LOWER(status) NOT IN ('hired')
		ORDER BY id`)
	if err != nil {
		return nil, err
	}
	return options, nil
}

// FindEvaluationCandidateOptions lists applicants who can still be given an
// evaluation. Somebody whose evaluation was submitted or completed is already
// on the evaluation side of the workflow, so they disappear from the picker
// instead of being offered for a second run. Pending/started assignments stay
// visible to the cycle wizard, which uses them to offer a drop. Rejected
// applicants stay too: a rejected run is closed, but HR may choose to evaluate
// them again rather than being permanently locked out.
func (r *CandidateRepository) FindEvaluationCandidateOptions() ([]candmodel.CandidateOption, error) {
	var options []candmodel.CandidateOption
	err := database.DB.Select(&options, `
		SELECT c.id, c.full_name, c.email, c.status,
		       COALESCE(c.job_posting_id, 0) AS job_posting_id,
		       COALESCE(jp.title, '') AS job_posting_title
		FROM candidates c
		LEFT JOIN job_postings jp ON jp.id = c.job_posting_id
		WHERE LOWER(c.status) NOT IN ('hired')
		  AND NOT EXISTS (
			SELECT 1 FROM evaluation_assignments a
			WHERE a.candidate_id = c.id
			  AND a.interview_id IS NULL
			  AND LOWER(a.status) IN ('submitted', 'completed')
		  )
		ORDER BY COALESCE(jp.title, ''), c.full_name, c.id`)
	if err != nil {
		return nil, err
	}
	return options, nil
}

// FindNames returns the display names behind a set of candidate ids.
func (r *CandidateRepository) FindNames(ids []int64) (map[int64]string, error) {
	names := map[int64]string{}
	clause, args, ok := repositories.InClause(ids)
	if !ok {
		return names, nil
	}
	rows, err := database.DB.Query("SELECT id, full_name FROM candidates WHERE id IN ("+clause+")", args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		var name string
		if err := rows.Scan(&id, &name); err == nil {
			names[id] = name
		}
	}
	return names, rows.Err()
}
