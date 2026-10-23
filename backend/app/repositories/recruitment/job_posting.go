package recruitment

import (
	recmodel "ems/app/models/recruitment"
	"ems/database"
	"time"
)

type JobPostingRepository struct{}

func (r *JobPostingRepository) Create(posting *recmodel.JobPosting) error {
	result, err := database.DB.Exec(
		`INSERT INTO job_postings (position_id, title, description, requirements, skills_required, salary_min, salary_max, employment_type, status, created_by)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		posting.PositionID, posting.Title, posting.Description, posting.Requirements, posting.SkillsRequired,
		posting.SalaryMin, posting.SalaryMax, posting.EmploymentType, posting.Status, posting.CreatedBy,
	)
	if err != nil {
		return err
	}
	posting.ID, _ = result.LastInsertId()
	return nil
}

func (r *JobPostingRepository) FindAll() ([]recmodel.JobPostingListItem, error) {
	var postings []recmodel.JobPostingListItem
	err := database.DB.Select(&postings,
		`SELECT jp.id, jp.position_id, jp.title, jp.description, jp.requirements, jp.skills_required,
		 jp.salary_min, jp.salary_max, jp.employment_type, jp.status, jp.posted_at,
		 jp.closed_at, jp.created_by, jp.created_at, jp.updated_at,
		 p.name AS position_name, d.name AS department_name
		 FROM job_postings jp
		 JOIN positions p ON p.id = jp.position_id
		 JOIN departments d ON d.id = p.department_id
		 ORDER BY jp.created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	return postings, nil
}

func (r *JobPostingRepository) FindByID(id int64) (*recmodel.JobPosting, error) {
	p := &recmodel.JobPosting{}
	err := database.DB.Get(p,
		`SELECT id, position_id, title, description, requirements, skills_required, salary_min, salary_max,
		 employment_type, status, posted_at, closed_at, created_by, created_at, updated_at
		 FROM job_postings WHERE id = ?`, id,
	)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (r *JobPostingRepository) Update(id int64, title, description, requirements, skillsRequired string, salaryMin, salaryMax *float64, employmentType, status string) error {
	_, err := database.DB.Exec(
		`UPDATE job_postings SET title=?, description=?, requirements=?, skills_required=?, salary_min=?, salary_max=?,
		 employment_type=?, status=?, updated_at=CURRENT_TIMESTAMP WHERE id=?`,
		title, description, requirements, skillsRequired, salaryMin, salaryMax, employmentType, status, id,
	)
	return err
}

func (r *JobPostingRepository) Delete(id int64) error {
	_, err := database.DB.Exec("DELETE FROM job_postings WHERE id = ?", id)
	return err
}

func (r *JobPostingRepository) Publish(id int64) error {
	now := time.Now()
	_, err := database.DB.Exec(
		"UPDATE job_postings SET status='published', posted_at=?, updated_at=CURRENT_TIMESTAMP WHERE id=?",
		now, id,
	)
	return err
}

func (r *JobPostingRepository) Close(id int64) error {
	now := time.Now()
	_, err := database.DB.Exec(
		"UPDATE job_postings SET status='closed', closed_at=?, updated_at=CURRENT_TIMESTAMP WHERE id=?",
		now, id,
	)
	return err
}

func (r *JobPostingRepository) Reopen(id int64) error {
	_, err := database.DB.Exec(
		"UPDATE job_postings SET status='published', closed_at=NULL, updated_at=CURRENT_TIMESTAMP WHERE id=?",
		id,
	)
	return err
}

// Skill Repository
