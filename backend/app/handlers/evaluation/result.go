package evaluation

import (
	"ems/app/handlers/authz"
	evalsvc "ems/app/services/evaluation"
	"github.com/gofiber/fiber/v2"
)

type EvaluationResultHandler struct{}

func (h *EvaluationResultHandler) Calculate(c *fiber.Ctx) error {
	assignmentID, _ := c.ParamsInt("assignment_id")
	assignment, err := evalsvc.AssignmentServiceInstance.FindByID(int64(assignmentID))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Assignment not found",
		})
	}
	if err := authz.RequireCanViewEvaluation(c, assignment); err != nil {
		return err
	}
	result, err := evalsvc.ResultServiceInstance.CalculateAndSave(int64(assignmentID))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to calculate result",
		})
	}
	if result == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "No rating has been recorded for this evaluation, so there is no result yet",
		})
	}
	return c.JSON(result)
}

func (h *EvaluationResultHandler) GetResult(c *fiber.Ctx) error {
	assignmentID, _ := c.ParamsInt("assignment_id")
	assignment, err := evalsvc.AssignmentServiceInstance.FindByID(int64(assignmentID))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Assignment not found",
		})
	}
	if err := authz.RequireCanViewEvaluation(c, assignment); err != nil {
		return err
	}
	result, err := evalsvc.ResultServiceInstance.GetResult(int64(assignmentID))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Result not found",
		})
	}
	return c.JSON(result)
}
