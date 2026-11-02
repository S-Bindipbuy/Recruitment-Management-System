package evaluation

import (
	evalmodel "ems/app/models/evaluation"
	"ems/database"
	"time"
)

type EvaluationPeriodRepository struct{}

func (r *EvaluationPeriodRepository) Create(period *evalmodel.EvaluationPeriod) error {
	result, err := database.DB.Exec(
		`INSERT INTO evaluation_periods (name, description, start_date, end_date, status, created_by) 
		 VALUES (?, ?, ?, ?, ?, ?)`,
		period.Name, period.Description, period.StartDate, period.EndDate, period.Status, period.CreatedBy,
	)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	period.ID = id
	return nil
}

func (r *EvaluationPeriodRepository) FindAll() ([]evalmodel.EvaluationPeriod, error) {
	var periods []evalmodel.EvaluationPeriod
	err := database.DB.Select(&periods,
		`SELECT id, name, COALESCE(description, '') AS description, start_date, end_date, status, created_by, created_at, updated_at
		 FROM evaluation_periods`,
	)
	if err != nil {
		return nil, err
	}
	return periods, nil
}

func (r *EvaluationPeriodRepository) FindByID(id int64) (*evalmodel.EvaluationPeriod, error) {
	p := &evalmodel.EvaluationPeriod{}
	err := database.DB.Get(p,
		"SELECT id, name, description, start_date, end_date, status, created_by, created_at, updated_at FROM evaluation_periods WHERE id = ?", id,
	)
	if err != nil {
		return nil, err
	}
	return p, nil
}

// FindActive returns the active period that contains the given date. Round
// evaluations are created automatically, without a period picker, so the
// interview's scheduled date decides which period they belong to.
func (r *EvaluationPeriodRepository) FindActive(on time.Time) (*evalmodel.EvaluationPeriod, error) {
	p := &evalmodel.EvaluationPeriod{}
	date := on.Format("2006-01-02")
	err := database.DB.Get(p,
		`SELECT id, COALESCE(name, '') AS name, COALESCE(description, '') AS description,
		        COALESCE(start_date, '') AS start_date, COALESCE(end_date, '') AS end_date,
		        status, created_by, created_at, updated_at
		 FROM evaluation_periods
		 WHERE LOWER(status) = 'active'
		   AND (start_date IS NULL OR start_date = '' OR start_date <= ?)
		   AND (end_date IS NULL OR end_date = '' OR end_date >= ?)
		 ORDER BY start_date DESC, id DESC
		 LIMIT 1`,
		date, date,
	)
	if err != nil {
		return nil, err
	}
	return p, nil
}

// FindLatestActive is the fallback when no active period covers the interview
// date: the newest active period is still a better home for the assignment
// than dropping it.
func (r *EvaluationPeriodRepository) FindLatestActive() (*evalmodel.EvaluationPeriod, error) {
	p := &evalmodel.EvaluationPeriod{}
	err := database.DB.Get(p,
		`SELECT id, COALESCE(name, '') AS name, COALESCE(description, '') AS description,
		        COALESCE(start_date, '') AS start_date, COALESCE(end_date, '') AS end_date,
		        status, created_by, created_at, updated_at
		 FROM evaluation_periods
		 WHERE LOWER(status) = 'active'
		 ORDER BY id DESC
		 LIMIT 1`,
	)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (r *EvaluationPeriodRepository) Update(period *evalmodel.EvaluationPeriod) error {
	_, err := database.DB.Exec(
		`UPDATE evaluation_periods SET name = ?, description = ?, start_date = ?, end_date = ?, 
		 status = ?, updated_at = ? WHERE id = ?`,
		period.Name, period.Description, period.StartDate, period.EndDate, period.Status, time.Now(), period.ID,
	)
	return err
}

func (r *EvaluationPeriodRepository) Delete(id int64) error {
	_, err := database.DB.Exec("DELETE FROM evaluation_periods WHERE id = ?", id)
	return err
}
