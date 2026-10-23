package evaluation

import (
	"ems/app/enums"
	"ems/app/handlers/authz"
	evalsvc "ems/app/services/evaluation"
	"github.com/gofiber/fiber/v2"
)

type EvaluationSectionHandler struct{}

func (h *EvaluationSectionHandler) ListByTemplate(c *fiber.Ctx) error {
	templateID, _ := c.ParamsInt("template_id")
	sections, err := evalsvc.SectionServiceInstance.FindByTemplateID(int64(templateID))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch sections",
		})
	}
	return c.JSON(sections)
}

func (h *EvaluationSectionHandler) Create(c *fiber.Ctx) error {
	if err := authz.RequirePermission(c, enums.PermissionEvaluationManage); err != nil {
		return err
	}

	var req struct {
		EvaluationTemplateID int64  `json:"evaluation_template_id"`
		Name                 string `json:"name"`
		Description          string `json:"description"`
		SortOrder            int    `json:"sort_order"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	sec, err := evalsvc.SectionServiceInstance.Create(req.EvaluationTemplateID, req.Name, req.Description, req.SortOrder)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create section",
		})
	}
	return c.Status(fiber.StatusCreated).JSON(sec)
}

func (h *EvaluationSectionHandler) Update(c *fiber.Ctx) error {
	if err := authz.RequirePermission(c, enums.PermissionEvaluationManage); err != nil {
		return err
	}

	id, _ := c.ParamsInt("id")
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		SortOrder   int    `json:"sort_order"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	sec, err := evalsvc.SectionServiceInstance.Update(int64(id), req.Name, req.Description, req.SortOrder)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update section",
		})
	}
	return c.JSON(sec)
}

func (h *EvaluationSectionHandler) Delete(c *fiber.Ctx) error {
	if err := authz.RequirePermission(c, enums.PermissionEvaluationManage); err != nil {
		return err
	}

	id, _ := c.ParamsInt("id")
	if err := evalsvc.SectionServiceInstance.Delete(int64(id)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete section",
		})
	}
	return c.JSON(fiber.Map{
		"message": "Section deleted successfully",
	})
}
