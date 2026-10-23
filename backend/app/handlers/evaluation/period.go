package evaluation

import (
	"ems/app/enums"
	"ems/app/handlers/authz"
	evalsvc "ems/app/services/evaluation"
	"github.com/gofiber/fiber/v2"
)

type EvaluationPeriodHandler struct{}

func (h *EvaluationPeriodHandler) List(c *fiber.Ctx) error {
	periods, err := evalsvc.PeriodServiceInstance.FindAll()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch periods",
		})
	}
	return c.JSON(periods)
}

func (h *EvaluationPeriodHandler) Get(c *fiber.Ctx) error {
	id, _ := c.ParamsInt("id")
	period, err := evalsvc.PeriodServiceInstance.FindByID(int64(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Period not found",
		})
	}
	return c.JSON(period)
}

func (h *EvaluationPeriodHandler) Create(c *fiber.Ctx) error {
	if err := authz.RequirePermission(c, enums.PermissionEvaluationManage); err != nil {
		return err
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		StartDate   string `json:"start_date"`
		EndDate     string `json:"end_date"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	userID := c.Locals("user_id").(int64)
	period, err := evalsvc.PeriodServiceInstance.Create(req.Name, req.Description, req.StartDate, req.EndDate, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create period",
		})
	}
	return c.Status(fiber.StatusCreated).JSON(period)
}

func (h *EvaluationPeriodHandler) Update(c *fiber.Ctx) error {
	if err := authz.RequirePermission(c, enums.PermissionEvaluationManage); err != nil {
		return err
	}

	id, _ := c.ParamsInt("id")
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		StartDate   string `json:"start_date"`
		EndDate     string `json:"end_date"`
		Status      string `json:"status"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	period, err := evalsvc.PeriodServiceInstance.Update(int64(id), req.Name, req.Description, req.StartDate, req.EndDate, req.Status)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update period",
		})
	}
	return c.JSON(period)
}

func (h *EvaluationPeriodHandler) Delete(c *fiber.Ctx) error {
	if err := authz.RequirePermission(c, enums.PermissionEvaluationManage); err != nil {
		return err
	}

	id, _ := c.ParamsInt("id")
	if err := evalsvc.PeriodServiceInstance.Delete(int64(id)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete period",
		})
	}
	return c.JSON(fiber.Map{
		"message": "Period deleted successfully",
	})
}
