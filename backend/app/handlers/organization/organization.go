package organization

import (
	orgsvc "ems/app/services/organization"
	"github.com/gofiber/fiber/v2"
)

type OrganizationHandler struct{}

func (h *OrganizationHandler) List(c *fiber.Ctx) error {
	orgs, err := orgsvc.OrganizationServiceInst.FindAll()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch organizations",
		})
	}
	return c.JSON(orgs)
}

func (h *OrganizationHandler) Create(c *fiber.Ctx) error {
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	org, err := orgsvc.OrganizationServiceInst.Create(req.Name, req.Description)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create organization",
		})
	}
	return c.Status(fiber.StatusCreated).JSON(org)
}

func (h *OrganizationHandler) Update(c *fiber.Ctx) error {
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

	org, err := orgsvc.OrganizationServiceInst.Update(int64(id), req.Name, req.Description, req.Status)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update organization",
		})
	}
	return c.JSON(org)
}

func (h *OrganizationHandler) Delete(c *fiber.Ctx) error {
	id, _ := c.ParamsInt("id")
	if err := orgsvc.OrganizationServiceInst.Delete(int64(id)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete organization",
		})
	}
	return c.JSON(fiber.Map{
		"message": "Organization deleted successfully",
	})
}
