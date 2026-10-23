package evaluation

import (
	evalmodel "ems/app/models/evaluation"
	"ems/database"
	"time"
)

type EvaluationSectionRepository struct{}

func (r *EvaluationSectionRepository) Create(sec *evalmodel.EvaluationSection) error {
	result, err := database.DB.Exec(
		`INSERT INTO evaluation_sections (evaluation_template_id, name, description, sort_order) 
		 VALUES (?, ?, ?, ?)`,
		sec.EvaluationTemplateID, sec.Name, sec.Description, sec.SortOrder,
	)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	sec.ID = id
	return nil
}

func (r *EvaluationSectionRepository) FindByID(id int64) (*evalmodel.EvaluationSection, error) {
	s := &evalmodel.EvaluationSection{}
	err := database.DB.Get(s,
		"SELECT id, evaluation_template_id, name, description, sort_order, created_at, updated_at FROM evaluation_sections WHERE id = ?", id,
	)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (r *EvaluationSectionRepository) FindByTemplateID(templateID int64) ([]evalmodel.EvaluationSection, error) {
	var sections []evalmodel.EvaluationSection
	err := database.DB.Select(&sections,
		"SELECT id, evaluation_template_id, name, COALESCE(description, '') AS description, sort_order, created_at, updated_at FROM evaluation_sections WHERE evaluation_template_id = ? ORDER BY sort_order",
		templateID,
	)
	if err != nil {
		return nil, err
	}
	return sections, nil
}

func (r *EvaluationSectionRepository) Update(sec *evalmodel.EvaluationSection) error {
	_, err := database.DB.Exec(
		"UPDATE evaluation_sections SET name = ?, description = ?, sort_order = ?, updated_at = ? WHERE id = ?",
		sec.Name, sec.Description, sec.SortOrder, time.Now(), sec.ID,
	)
	return err
}

func (r *EvaluationSectionRepository) Delete(id int64) error {
	_, err := database.DB.Exec("DELETE FROM evaluation_sections WHERE id = ?", id)
	return err
}
