package organization

import (
	orgmodel "ems/app/models/organization"
	usermodel "ems/app/models/user"
	"ems/database"
	"time"
)

type EmployeeRepository struct{}

// Employees may point at NULL department/position (hire without a job
// posting), so the ids are coalesced to 0 instead of failing the scan.
const employeeColumns = `COALESCE(e.id, 0) AS id, e.user_id, e.employee_code,
	COALESCE(e.department_id, 0) AS department_id, COALESCE(e.position_id, 0) AS position_id,
	e.manager_id, e.hire_date, e.status, e.created_at, e.updated_at`

func (r *EmployeeRepository) Create(emp *orgmodel.Employee) error {
	// A worker can be added before a position exists for them, so 0 is stored
	// as NULL rather than as a position id that points nowhere.
	var position any
	if emp.PositionID > 0 {
		position = emp.PositionID
	}
	var department any
	if emp.DepartmentID > 0 {
		department = emp.DepartmentID
	}
	result, err := database.DB.Exec(
		`INSERT INTO employees (user_id, employee_code, department_id, position_id, manager_id, hire_date, status) 
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		emp.UserID, emp.EmployeeCode, department, position, emp.ManagerID, emp.HireDate, emp.Status,
	)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	emp.ID = id
	return nil
}

func (r *EmployeeRepository) FindByID(id int64) (*orgmodel.Employee, error) {
	emp := &orgmodel.Employee{}
	err := database.DB.Get(emp,
		`SELECT `+employeeColumns+` FROM employees e WHERE e.id = ?`, id,
	)
	if err != nil {
		return nil, err
	}
	return emp, nil
}

func (r *EmployeeRepository) FindAll() ([]orgmodel.Employee, error) {
	var employees []orgmodel.Employee
	err := database.DB.Select(&employees,
		`SELECT `+employeeColumns+` FROM employees e`,
	)
	if err != nil {
		return nil, err
	}
	return employees, nil
}

func (r *EmployeeRepository) Update(emp *orgmodel.Employee) error {
	_, err := database.DB.Exec(
		`UPDATE employees SET employee_code = ?, department_id = ?, position_id = ?, manager_id = ?, 
		 hire_date = ?, status = ?, updated_at = ? WHERE id = ?`,
		emp.EmployeeCode, emp.DepartmentID, emp.PositionID, emp.ManagerID,
		emp.HireDate, emp.Status, time.Now(), emp.ID,
	)
	return err
}

func (r *EmployeeRepository) Delete(id int64) error {
	_, err := database.DB.Exec("DELETE FROM employees WHERE id = ?", id)
	return err
}

// SetManager points an employee at another employee (or clears the link with
// nil). It only touches manager_id so the rest of the record - code,
// department, position, hire date - cannot be lost by an assign-manager call.
func (r *EmployeeRepository) FindByUserID(userID int64) (*orgmodel.Employee, error) {
	emp := &orgmodel.Employee{}
	err := database.DB.Get(emp,
		`SELECT `+employeeColumns+` FROM employees e WHERE e.user_id = ?`, userID,
	)
	if err != nil {
		return nil, err
	}
	return emp, nil
}

// FindActiveWorkers lists everybody who can be asked to do an evaluation: an
// active employee record behind an active user account.
func (r *EmployeeRepository) FindActiveWorkers() ([]usermodel.WorkerOption, error) {
	var workers []usermodel.WorkerOption
	err := database.DB.Select(&workers, `
		SELECT e.user_id, u.name, e.employee_code
		FROM employees e JOIN users u ON u.id = e.user_id
		WHERE e.status = 'active' AND u.status = 'active' AND e.user_id IS NOT NULL
		ORDER BY u.name`)
	if err != nil {
		return nil, err
	}
	return workers, nil
}

// IsActiveWorker reports whether the user still holds an active employee
// record and an active user account. Evaluations may only be assigned between
// real, active workers.
func (r *EmployeeRepository) IsActiveWorker(userID int64) (bool, error) {
	if userID <= 0 {
		return false, nil
	}
	var ok bool
	err := database.DB.Get(&ok, `
		SELECT EXISTS(
			SELECT 1 FROM employees e JOIN users u ON u.id = e.user_id
			WHERE e.user_id = ? AND e.user_id IS NOT NULL AND e.status = 'active' AND u.status = 'active')`, userID)
	return ok, err
}
