package organization

import "time"

type Employee struct {
	ID           int64      `json:"id" db:"id"`
	UserID       int64      `json:"user_id" db:"user_id"`
	EmployeeCode string     `json:"employee_code" db:"employee_code"`
	DepartmentID int64      `json:"department_id" db:"department_id"`
	PositionID   int64      `json:"position_id" db:"position_id"`
	ManagerID    *int64     `json:"manager_id" db:"manager_id"`
	HireDate     *time.Time `json:"hire_date" db:"hire_date"`
	Status       string     `json:"status" db:"status"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
}
