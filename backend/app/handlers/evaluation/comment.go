package evaluation

import (
	"ems/app/handlers/authz"
	evalsvc "ems/app/services/evaluation"
	"github.com/gofiber/fiber/v2"
)

type EvaluationCommentHandler struct{}

func (h *EvaluationCommentHandler) Create(c *fiber.Ctx) error {
	var req struct {
		EvaluationAssignmentID int64  `json:"evaluation_assignment_id"`
		EvaluationQuestionID   *int64 `json:"evaluation_question_id"`
		Comment                string `json:"comment"`
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

	userID := c.Locals("user_id").(int64)
	comment, err := evalsvc.CommentServiceInstance.Create(req.EvaluationAssignmentID, userID, req.EvaluationQuestionID, req.Comment)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create comment",
		})
	}
	return c.Status(fiber.StatusCreated).JSON(comment)
}

func (h *EvaluationCommentHandler) List(c *fiber.Ctx) error {
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
	comments, err := evalsvc.CommentServiceInstance.FindByAssignmentID(int64(assignmentID))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch comments",
		})
	}
	return c.JSON(comments)
}
