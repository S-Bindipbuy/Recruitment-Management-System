package evaluation

import (
	"ems/app/handlers/authz"
	evalsvc "ems/app/services/evaluation"
	"github.com/gofiber/fiber/v2"
)

type EvaluationResponseHandler struct{}

func (h *EvaluationResponseHandler) SaveResponse(c *fiber.Ctx) error {
	var req struct {
		EvaluationAssignmentID int64    `json:"evaluation_assignment_id"`
		EvaluationQuestionID   int64    `json:"evaluation_question_id"`
		RatingValue            *float64 `json:"rating_value"`
		TextValue              string   `json:"text_value"`
		NumericValue           *float64 `json:"numeric_value"`
		BooleanValue           *bool    `json:"boolean_value"`
		SelectedOptionID       *int64   `json:"selected_option_id"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	assignment, err := evalsvc.AssignmentServiceInstance.FindByID(req.EvaluationAssignmentID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Assignment not found",
		})
	}
	if err := authz.RequireCanFillEvaluation(c, assignment); err != nil {
		return err
	}

	if err := evalsvc.ResponseServiceInstance.SaveResponse(
		req.EvaluationAssignmentID, req.EvaluationQuestionID,
		req.RatingValue, req.TextValue, req.NumericValue,
		req.BooleanValue, req.SelectedOptionID,
	); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save response",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Response saved",
	})
}

func (h *EvaluationResponseHandler) GetResponses(c *fiber.Ctx) error {
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
	responses, err := evalsvc.ResponseServiceInstance.GetResponses(int64(assignmentID))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch responses",
		})
	}
	return c.JSON(responses)
}
