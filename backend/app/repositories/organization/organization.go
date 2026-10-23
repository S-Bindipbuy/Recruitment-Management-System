package organization

import (
	orgmodel "ems/app/models/organization"
	"ems/database"
	"time"
)

type OrganizationRepository struct{}

func (r *OrganizationRepository) Create(org *orgmodel.Organization) error {
	result, err := database.DB.Exec(
		"INSERT INTO organizations (name, description, status) VALUES (?, ?, ?)",
		org.Name, org.Description, org.Status,
	)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	org.ID = id
	return nil
}

func (r *OrganizationRepository) FindByID(id int64) (*orgmodel.Organization, error) {
	org := &orgmodel.Organization{}
	err := database.DB.Get(org,
		"SELECT id, name, description, status, created_at, updated_at FROM organizations WHERE id = ?", id,
	)
	if err != nil {
		return nil, err
	}
	return org, nil
}

func (r *OrganizationRepository) FindAll() ([]orgmodel.Organization, error) {
	var orgs []orgmodel.Organization
	err := database.DB.Select(&orgs,
		"SELECT id, name, COALESCE(description, '') AS description, status, created_at, updated_at FROM organizations",
	)
	if err != nil {
		return nil, err
	}
	return orgs, nil
}

func (r *OrganizationRepository) Update(org *orgmodel.Organization) error {
	_, err := database.DB.Exec(
		"UPDATE organizations SET name = ?, description = ?, status = ?, updated_at = ? WHERE id = ?",
		org.Name, org.Description, org.Status, time.Now(), org.ID,
	)
	return err
}

func (r *OrganizationRepository) Delete(id int64) error {
	_, err := database.DB.Exec("DELETE FROM organizations WHERE id = ?", id)
	return err
}

type DepartmentRepository struct{}

func (r *DepartmentRepository) Create(dept *orgmodel.Department) error {
	result, err := database.DB.Exec(
		"INSERT INTO departments (organization_id, name, description, status) VALUES (?, ?, ?, ?)",
		dept.OrganizationID, dept.Name, dept.Description, dept.Status,
	)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	dept.ID = id
	return nil
}

func (r *DepartmentRepository) FindByID(id int64) (*orgmodel.Department, error) {
	dept := &orgmodel.Department{}
	err := database.DB.Get(dept,
		"SELECT id, organization_id, name, description, status, created_at, updated_at FROM departments WHERE id = ?", id,
	)
	if err != nil {
		return nil, err
	}
	return dept, nil
}

func (r *DepartmentRepository) FindAll() ([]orgmodel.Department, error) {
	var departments []orgmodel.Department
	err := database.DB.Select(&departments,
		"SELECT id, organization_id, name, COALESCE(description, '') AS description, status, created_at, updated_at FROM departments",
	)
	if err != nil {
		return nil, err
	}
	return departments, nil
}

func (r *DepartmentRepository) Update(dept *orgmodel.Department) error {
	_, err := database.DB.Exec(
		"UPDATE departments SET name = ?, description = ?, status = ?, updated_at = ? WHERE id = ?",
		dept.Name, dept.Description, dept.Status, time.Now(), dept.ID,
	)
	return err
}

func (r *DepartmentRepository) Delete(id int64) error {
	_, err := database.DB.Exec("DELETE FROM departments WHERE id = ?", id)
	return err
}

type PositionRepository struct{}

func (r *PositionRepository) Create(pos *orgmodel.Position) error {
	result, err := database.DB.Exec(
		"INSERT INTO positions (department_id, name, description, status) VALUES (?, ?, ?, ?)",
		pos.DepartmentID, pos.Name, pos.Description, pos.Status,
	)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	pos.ID = id
	return nil
}

func (r *PositionRepository) FindByID(id int64) (*orgmodel.Position, error) {
	pos := &orgmodel.Position{}
	err := database.DB.Get(pos,
		"SELECT id, department_id, name, description, status, created_at, updated_at FROM positions WHERE id = ?", id,
	)
	if err != nil {
		return nil, err
	}
	return pos, nil
}

func (r *PositionRepository) FindAll() ([]orgmodel.Position, error) {
	var positions []orgmodel.Position
	err := database.DB.Select(&positions,
		"SELECT id, department_id, name, COALESCE(description, '') AS description, status, created_at, updated_at FROM positions",
	)
	if err != nil {
		return nil, err
	}
	return positions, nil
}

func (r *PositionRepository) Update(pos *orgmodel.Position) error {
	_, err := database.DB.Exec(
		"UPDATE positions SET name = ?, description = ?, status = ?, updated_at = ? WHERE id = ?",
		pos.Name, pos.Description, pos.Status, time.Now(), pos.ID,
	)
	return err
}

func (r *PositionRepository) Delete(id int64) error {
	_, err := database.DB.Exec("DELETE FROM positions WHERE id = ?", id)
	return err
}
