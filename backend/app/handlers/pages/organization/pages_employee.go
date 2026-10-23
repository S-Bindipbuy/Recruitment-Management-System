package organization

import (
	"ems/app/enums"
	core "ems/app/handlers/pages/core"
	orgmodel "ems/app/models/organization"
	orgsvc "ems/app/services/organization"
	usersvc "ems/app/services/user"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type EmployeeListItem struct {
	Employee       orgmodel.Employee
	UserName       string
	DepartmentName string
	PositionName   string
}

type EmployeesPageData struct {
	Employees   []EmployeeListItem
	Candidates  []core.UserOption
	Departments []orgmodel.Department
	Positions   []orgmodel.Position
}

func (h *OrganizationHandler) Employees(c *fiber.Ctx) error {
	data := EmployeesPageData{}
	employees, _ := orgsvc.EmployeeServiceInstance.FindAll()
	data.Candidates = core.LoadWorkerCandidates()
	data.Departments, _ = orgsvc.DepartmentServiceInst.FindAll()
	data.Positions, _ = orgsvc.PositionServiceInst.FindAll()

	deptNames := map[int64]string{}
	for _, d := range data.Departments {
		deptNames[d.ID] = d.Name
	}
	posNames := map[int64]string{}
	for _, p := range data.Positions {
		posNames[p.ID] = p.Name
	}

	userIDs := make([]int64, 0, len(employees))
	for _, e := range employees {
		userIDs = append(userIDs, e.UserID)
	}
	names := core.UserNames(userIDs)

	for _, e := range employees {
		item := EmployeeListItem{
			Employee:       e,
			UserName:       core.DisplayName(names, e.UserID),
			DepartmentName: deptNames[e.DepartmentID],
			PositionName:   posNames[e.PositionID],
		}
		if item.DepartmentName == "" {
			item.DepartmentName = "Department #" + strconv.FormatInt(e.DepartmentID, 10)
		}
		if item.PositionName == "" {
			item.PositionName = "Position #" + strconv.FormatInt(e.PositionID, 10)
		}
		data.Employees = append(data.Employees, item)
	}

	return core.RenderPage(c, "page-employees", data)
}

func (h *OrganizationHandler) CreateEmployee(c *fiber.Ctx) error {
	userID := core.FormInt64(c, "user_id")
	code := strings.TrimSpace(c.FormValue("employee_code"))
	deptID := core.FormInt64(c, "department_id")
	posID := core.FormInt64(c, "position_id")

	if userID <= 0 || code == "" || deptID <= 0 || posID <= 0 {
		return core.RenderError(c, fiber.StatusBadRequest, "Please fill in all required fields.")
	}
	user, err := usersvc.UserServiceInstance.FindByID(userID)
	if err != nil {
		return core.RenderError(c, fiber.StatusBadRequest, "The selected user was not found.")
	}
	if user.Status != string(enums.UserStatusActive) {
		return core.RenderError(c, fiber.StatusBadRequest, "The selected user is not active.")
	}
	if _, err := orgsvc.EmployeeServiceInstance.FindByUserID(userID); err == nil {
		return core.RenderError(c, fiber.StatusBadRequest, "This user already has a worker record.")
	}
	if !core.ActiveRecordExists("departments", deptID) {
		return core.RenderError(c, fiber.StatusBadRequest, "The selected department was not found.")
	}
	if !core.ActiveRecordExists("positions", posID) {
		return core.RenderError(c, fiber.StatusBadRequest, "The selected position was not found.")
	}

	hireDate := core.DateFromForm(c.FormValue("hire_date"))

	if _, err := orgsvc.EmployeeServiceInstance.Create(userID, code, deptID, posID, nil, hireDate); err != nil {
		return core.RenderError(c, fiber.StatusInternalServerError, "Failed to create employee.")
	}

	return h.Employees(c)
}
