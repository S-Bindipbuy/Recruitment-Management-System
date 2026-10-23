package organization

import (
	"ems/app/enums"
	orgmodel "ems/app/models/organization"
	usermodel "ems/app/models/user"
	orgrepo "ems/app/repositories/organization"
	"time"
)

type EmployeeService struct{}

func (s *EmployeeService) Create(userID int64, employeeCode string, departmentID, positionID int64, managerID *int64, hireDate *time.Time) (*orgmodel.Employee, error) {
	emp := &orgmodel.Employee{
		UserID:       userID,
		EmployeeCode: employeeCode,
		DepartmentID: departmentID,
		PositionID:   positionID,
		ManagerID:    managerID,
		HireDate:     hireDate,
		Status:       string(enums.UserStatusActive),
	}
	if err := orgrepo.EmployeeRepo.Create(emp); err != nil {
		return nil, err
	}
	return emp, nil
}

func (s *EmployeeService) FindAll() ([]orgmodel.Employee, error) {
	return orgrepo.EmployeeRepo.FindAll()
}

func (s *EmployeeService) FindByID(id int64) (*orgmodel.Employee, error) {
	return orgrepo.EmployeeRepo.FindByID(id)
}

func (s *EmployeeService) FindByUserID(userID int64) (*orgmodel.Employee, error) {
	return orgrepo.EmployeeRepo.FindByUserID(userID)
}

func (s *EmployeeService) Update(id int64, emp *orgmodel.Employee) error {
	emp.ID = id
	return orgrepo.EmployeeRepo.Update(emp)
}

func (s *EmployeeService) Delete(id int64) error {
	return orgrepo.EmployeeRepo.Delete(id)
}

// FindActiveWorkers lists everybody who can be asked to do an evaluation:
// an active employee record behind an active user account.
func (s *EmployeeService) FindActiveWorkers() ([]usermodel.WorkerOption, error) {
	return orgrepo.EmployeeRepo.FindActiveWorkers()
}

// IsActiveWorker reports whether the user still holds an active employee
// record and an active user account.
func (s *EmployeeService) IsActiveWorker(userID int64) (bool, error) {
	return orgrepo.EmployeeRepo.IsActiveWorker(userID)
}

// IsActiveWorkerBool reports whether the user is an employee on active
// contracts - the only people who can be asked to do an evaluation. The
// error is swallowed because callers only need the yes/no.
func IsActiveWorker(userID int64) bool {
	ok, err := EmployeeServiceInstance.IsActiveWorker(userID)
	return err == nil && ok
}
