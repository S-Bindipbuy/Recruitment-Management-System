package employee

import (
	orgmodel "ems/app/models/organization"
	orgsvc "ems/app/services/organization"
	"time"

	"github.com/gofiber/fiber/v2"
)

type EmployeeHandler struct{}

func (h *EmployeeHandler) List(c *fiber.Ctx) error {
	employees, err := orgsvc.EmployeeServiceInstance.FindAll()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch employees",
		})
	}
	return c.JSON(employees)
}

func (h *EmployeeHandler) Get(c *fiber.Ctx) error {
	id, _ := c.ParamsInt("id")
	emp, err := orgsvc.EmployeeServiceInstance.FindByID(int64(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Employee not found",
		})
	}
	return c.JSON(emp)
}

func (h *EmployeeHandler) Create(c *fiber.Ctx) error {
	var req struct {
		UserID       int64   `json:"user_id"`
		EmployeeCode string  `json:"employee_code"`
		DepartmentID int64   `json:"department_id"`
		PositionID   int64   `json:"position_id"`
		ManagerID    *int64  `json:"manager_id"`
		HireDate     *string `json:"hire_date"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	var hireDate *time.Time
	if req.HireDate != nil {
		t, _ := time.Parse("2006-01-02", *req.HireDate)
		hireDate = &t
	}

	emp, err := orgsvc.EmployeeServiceInstance.Create(
		req.UserID, req.EmployeeCode, req.DepartmentID, req.PositionID, req.ManagerID, hireDate,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create employee",
		})
	}
	return c.Status(fiber.StatusCreated).JSON(emp)
}

func (h *EmployeeHandler) Update(c *fiber.Ctx) error {
	id, _ := c.ParamsInt("id")
	var req orgmodel.Employee

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := orgsvc.EmployeeServiceInstance.Update(int64(id), &req); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update employee",
		})
	}
	return c.JSON(req)
}

func (h *EmployeeHandler) Delete(c *fiber.Ctx) error {
	id, _ := c.ParamsInt("id")
	if err := orgsvc.EmployeeServiceInstance.Delete(int64(id)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete employee",
		})
	}
	return c.JSON(fiber.Map{
		"message": "Employee deleted successfully",
	})
}
