package candidate

import (
	"database/sql"
	candmodel "ems/app/models/candidate"
	"ems/app/repositories"
	"ems/database"
	"fmt"
	"strings"
	"time"
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
		 COALESCE(jp.title, '') AS posting_title
		 FROM candidates c
		 LEFT JOIN job_postings jp ON jp.id = c.job_posting_id
		 ORDER BY c.created_at DESC`,
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
		 COALESCE(jp.title, '') AS posting_title
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

// Hire creates the candidate's login, employee record, role assignment, and
// assessment atomically. A partial hire would otherwise leave an unusable
// account or an employee without a role.
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

	result, err := tx.Exec(
		"INSERT INTO users (name, email, password, status) VALUES (?, ?, ?, 'active')",
		fullName, email, passwordHash,
	)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			return fmt.Errorf("a user with this email already exists")
		}
		return err
	}
	userID, err := result.LastInsertId()
	if err != nil {
		return err
	}

	employeeCode := fmt.Sprintf("EMP-%06d", candidateID)
	if _, err := tx.Exec(
		`INSERT INTO employees (user_id, employee_code, department_id, position_id, hire_date, status)
		 VALUES (?, ?, ?, ?, ?, 'active')`,
		userID, employeeCode, departmentID, positionID, time.Now(),
	); err != nil {
		return err
	}
	// Use the logged-in user as evaluator. If they no longer exist (e.g. DB was
	// recreated), fall back to the newly created user so the hire can complete.
	actualEvaluator := evaluatorID
	var evalExists int
	_ = tx.QueryRow("SELECT COUNT(*) FROM users WHERE id = ?", actualEvaluator).Scan(&evalExists)
	if evalExists == 0 {
		actualEvaluator = userID
	}
	if _, err := tx.Exec(
		`INSERT INTO candidate_assessments (candidate_id, evaluator_id, skill_score, notes, recommendation)
		 VALUES (?, ?, ?, ?, 'accepted')`,
		candidateID, actualEvaluator, rating, strings.TrimSpace(notes),
	); err != nil {
		return err
	}
	if _, err := tx.Exec("UPDATE candidates SET status = 'hired', updated_at = CURRENT_TIMESTAMP WHERE id = ?", candidateID); err != nil {
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
