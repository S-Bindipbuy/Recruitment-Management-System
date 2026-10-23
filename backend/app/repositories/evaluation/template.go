package evaluation

import (
	evalmodel "ems/app/models/evaluation"
	"ems/database"
	"time"
)

type EvaluationTemplateRepository struct{}

func (r *EvaluationTemplateRepository) Create(tmpl *evalmodel.EvaluationTemplate) error {
	result, err := database.DB.Exec(
		`INSERT INTO evaluation_templates (name, description, status, created_by) 
		 VALUES (?, ?, ?, ?)`,
		tmpl.Name, tmpl.Description, tmpl.Status, tmpl.CreatedBy,
	)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	tmpl.ID = id
	return nil
}

func (r *EvaluationTemplateRepository) FindAll() ([]evalmodel.EvaluationTemplate, error) {
	var templates []evalmodel.EvaluationTemplate
	err := database.DB.Select(&templates,
		`SELECT id, name, COALESCE(description, '') AS description, status, created_by, created_at, updated_at
		 FROM evaluation_templates`,
	)
	if err != nil {
		return nil, err
	}
	return templates, nil
}

func (r *EvaluationTemplateRepository) StructureCounts() (map[int64]evalmodel.TemplateStructure, error) {
	var rowsData []struct {
		ID        int64 `db:"id"`
		Sections  int   `db:"sections"`
		Questions int   `db:"questions"`
	}
	err := database.DB.Select(&rowsData, `
		SELECT t.id AS id,
		       (SELECT COUNT(*) FROM evaluation_sections s
		         WHERE s.evaluation_template_id = t.id) AS sections,
		       (SELECT COUNT(*) FROM evaluation_questions q
		          JOIN evaluation_sections s2 ON s2.id = q.evaluation_section_id
		         WHERE s2.evaluation_template_id = t.id) AS questions
		  FROM evaluation_templates t`)
	if err != nil {
		return nil, err
	}

	counts := make(map[int64]evalmodel.TemplateStructure, len(rowsData))
	for _, row := range rowsData {
		counts[row.ID] = evalmodel.TemplateStructure{Sections: row.Sections, Questions: row.Questions}
	}
	return counts, nil
}

func (r *EvaluationTemplateRepository) FindByID(id int64) (*evalmodel.EvaluationTemplate, error) {
	t := &evalmodel.EvaluationTemplate{}
	err := database.DB.Get(t,
		"SELECT id, name, COALESCE(description, '') AS description, status, created_by, created_at, updated_at FROM evaluation_templates WHERE id = ?", id,
	)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (r *EvaluationTemplateRepository) Update(tmpl *evalmodel.EvaluationTemplate) error {
	_, err := database.DB.Exec(
		`UPDATE evaluation_templates SET name = ?, description = ?, 
		 status = ?, updated_at = ? WHERE id = ?`,
		tmpl.Name, tmpl.Description, tmpl.Status, time.Now(), tmpl.ID,
	)
	return err
}

func (r *EvaluationTemplateRepository) Delete(id int64) error {
	_, err := database.DB.Exec("DELETE FROM evaluation_templates WHERE id = ?", id)
	return err
}
