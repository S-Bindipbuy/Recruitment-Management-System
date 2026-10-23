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
