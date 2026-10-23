package evaluation

import (
	"ems/app/enums"
	"ems/app/handlers/authz"
	evalsvc "ems/app/services/evaluation"
	"github.com/gofiber/fiber/v2"
)

type EvaluationTemplateHandler struct{}

func (h *EvaluationTemplateHandler) List(c *fiber.Ctx) error {
	templates, err := evalsvc.TemplateServiceInstance.FindAll()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch templates",
		})
	}
	return c.JSON(templates)
}

func (h *EvaluationTemplateHandler) Get(c *fiber.Ctx) error {
	id, _ := c.ParamsInt("id")
	tmpl, err := evalsvc.TemplateServiceInstance.FindByID(int64(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Template not found",
		})
	}
	return c.JSON(tmpl)
}

func (h *EvaluationTemplateHandler) Create(c *fiber.Ctx) error {
	if err := authz.RequirePermission(c, enums.PermissionEvaluationManage); err != nil {
		return err
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	userID := c.Locals("user_id").(int64)
	tmpl, err := evalsvc.TemplateServiceInstance.Create(req.Name, req.Description, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create template",
		})
	}
	return c.Status(fiber.StatusCreated).JSON(tmpl)
}

func (h *EvaluationTemplateHandler) Update(c *fiber.Ctx) error {
	if err := authz.RequirePermission(c, enums.PermissionEvaluationManage); err != nil {
		return err
	}

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

	tmpl, err := evalsvc.TemplateServiceInstance.Update(int64(id), req.Name, req.Description, req.Status)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update template",
		})
	}
	return c.JSON(tmpl)
}

func (h *EvaluationTemplateHandler) Delete(c *fiber.Ctx) error {
	if err := authz.RequirePermission(c, enums.PermissionEvaluationManage); err != nil {
		return err
	}

	id, _ := c.ParamsInt("id")
	if err := evalsvc.TemplateServiceInstance.Delete(int64(id)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete template",
		})
	}
	return c.JSON(fiber.Map{
		"message": "Template deleted successfully",
	})
}
