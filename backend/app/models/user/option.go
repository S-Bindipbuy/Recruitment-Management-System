package user

// WorkerOption is somebody who can be asked to do an evaluation.
type WorkerOption struct {
	UserID       int64  `json:"user_id" db:"user_id"`
	Name         string `json:"name" db:"name"`
	EmployeeCode string `json:"employee_code" db:"employee_code"`
}

// UserOption is an active account that holds no employee record yet.
type UserOption struct {
	ID   int64  `json:"id" db:"id"`
	Name string `json:"name" db:"name"`
}
