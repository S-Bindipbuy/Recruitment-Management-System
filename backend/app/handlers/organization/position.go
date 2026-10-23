package organization

import (
	orgsvc "ems/app/services/organization"
	"github.com/gofiber/fiber/v2"
)

type PositionHandler struct{}

func (h *PositionHandler) List(c *fiber.Ctx) error {
	positions, err := orgsvc.PositionServiceInst.FindAll()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch positions",
		})
	}
	return c.JSON(positions)
}

func (h *PositionHandler) Create(c *fiber.Ctx) error {
	var req struct {
		DepartmentID int64  `json:"department_id"`
		Name         string `json:"name"`
		Description  string `json:"description"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	pos, err := orgsvc.PositionServiceInst.Create(req.DepartmentID, req.Name, req.Description)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create position",
		})
	}
	return c.Status(fiber.StatusCreated).JSON(pos)
}

func (h *PositionHandler) Update(c *fiber.Ctx) error {
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

	pos, err := orgsvc.PositionServiceInst.Update(int64(id), req.Name, req.Description, req.Status)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update position",
		})
	}
	return c.JSON(pos)
}

func (h *PositionHandler) Delete(c *fiber.Ctx) error {
	id, _ := c.ParamsInt("id")
	if err := orgsvc.PositionServiceInst.Delete(int64(id)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete position",
		})
	}
	return c.JSON(fiber.Map{
		"message": "Position deleted successfully",
	})
}
