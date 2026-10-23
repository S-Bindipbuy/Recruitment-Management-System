package evaluation

import (
	"ems/app/enums"
	"ems/app/handlers/authz"
	evalsvc "ems/app/services/evaluation"
	"github.com/gofiber/fiber/v2"
)

type EvaluationQuestionHandler struct{}

func (h *EvaluationQuestionHandler) ListBySection(c *fiber.Ctx) error {
	sectionID, _ := c.ParamsInt("section_id")
	questions, err := evalsvc.QuestionServiceInstance.FindBySectionID(int64(sectionID))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch questions",
		})
	}
	return c.JSON(questions)
}

func (h *EvaluationQuestionHandler) Create(c *fiber.Ctx) error {
	if err := authz.RequirePermission(c, enums.PermissionEvaluationManage); err != nil {
		return err
	}

	var req struct {
		EvaluationSectionID int64   `json:"evaluation_section_id"`
		Question            string  `json:"question"`
		Description         string  `json:"description"`
		QuestionType        string  `json:"question_type"`
		IsRequired          bool    `json:"is_required"`
		Weight              float64 `json:"weight"`
		SortOrder           int     `json:"sort_order"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	q, err := evalsvc.QuestionServiceInstance.Create(
		req.EvaluationSectionID, req.Question, req.Description, req.QuestionType,
		req.IsRequired, req.Weight, req.SortOrder,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create question",
		})
	}
	return c.Status(fiber.StatusCreated).JSON(q)
}

func (h *EvaluationQuestionHandler) Update(c *fiber.Ctx) error {
	if err := authz.RequirePermission(c, enums.PermissionEvaluationManage); err != nil {
		return err
	}

	id, _ := c.ParamsInt("id")
	var req struct {
		Question     string  `json:"question"`
		Description  string  `json:"description"`
		QuestionType string  `json:"question_type"`
		IsRequired   bool    `json:"is_required"`
		Weight       float64 `json:"weight"`
		SortOrder    int     `json:"sort_order"`
		Status       string  `json:"status"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	q, err := evalsvc.QuestionServiceInstance.Update(
		int64(id), req.Question, req.Description, req.QuestionType,
		req.IsRequired, req.Weight, req.SortOrder, req.Status,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update question",
		})
	}
	return c.JSON(q)
}

func (h *EvaluationQuestionHandler) Delete(c *fiber.Ctx) error {
	if err := authz.RequirePermission(c, enums.PermissionEvaluationManage); err != nil {
		return err
	}

	id, _ := c.ParamsInt("id")
	if err := evalsvc.QuestionServiceInstance.Delete(int64(id)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete question",
		})
	}
	return c.JSON(fiber.Map{
		"message": "Question deleted successfully",
	})
}
