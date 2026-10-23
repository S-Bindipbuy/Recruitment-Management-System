package organization

import (
	orgsvc "ems/app/services/organization"
	"github.com/gofiber/fiber/v2"
)

type DepartmentHandler struct{}

func (h *DepartmentHandler) List(c *fiber.Ctx) error {
	depts, err := orgsvc.DepartmentServiceInst.FindAll()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch departments",
		})
	}
	return c.JSON(depts)
}

func (h *DepartmentHandler) Create(c *fiber.Ctx) error {
	var req struct {
		OrganizationID int64  `json:"organization_id"`
		Name           string `json:"name"`
		Description    string `json:"description"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	dept, err := orgsvc.DepartmentServiceInst.Create(req.OrganizationID, req.Name, req.Description)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create department",
		})
	}
	return c.Status(fiber.StatusCreated).JSON(dept)
}

func (h *DepartmentHandler) Update(c *fiber.Ctx) error {
	id, _ := c.ParamsInt("id")
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Status      string `json:"status"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	dept, err := orgsvc.DepartmentServiceInst.Update(int64(id), req.Name, req.Description, req.Status)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update department",
		})
	}
	return c.JSON(dept)
}

func (h *DepartmentHandler) Delete(c *fiber.Ctx) error {
	id, _ := c.ParamsInt("id")
	if err := orgsvc.DepartmentServiceInst.Delete(int64(id)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete department",
		})
	}
	return c.JSON(fiber.Map{
		"message": "Department deleted successfully",
	})
}
