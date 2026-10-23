package evaluation

import (
	"ems/app/enums"
	"ems/app/handlers/authz"
	evalsvc "ems/app/services/evaluation"
	orgsvc "ems/app/services/organization"
	"github.com/gofiber/fiber/v2"
)

type EvaluationApprovalHandler struct{}

func (h *EvaluationApprovalHandler) Create(c *fiber.Ctx) error {
	if err := authz.RequirePermission(c, enums.PermissionEvaluationApprove); err != nil {
		return err
	}

	var req struct {
		AssignmentID int64 `json:"assignment_id"`
		ApproverID   int64 `json:"approver_id"`
		Level        int   `json:"level"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Level <= 0 {
		req.Level = 1
	}

	assignment, err := evalsvc.AssignmentServiceInstance.FindByID(req.AssignmentID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "The selected assignment was not found",
		})
	}
	if req.ApproverID == assignment.EvaluatorID {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "The evaluator cannot approve their own evaluation",
		})
	}
	if assignment.EvaluateeID > 0 && req.ApproverID == assignment.EvaluateeID {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "The evaluated worker cannot approve their own evaluation",
		})
	}
	if !orgsvc.IsActiveWorker(req.ApproverID) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Approvals can only be requested from selected workers",
		})
	}

	appr, err := evalsvc.ApprovalServiceInstance.Create(req.AssignmentID, req.ApproverID, req.Level)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create approval request",
		})
	}
	return c.Status(fiber.StatusCreated).JSON(appr)
}

func (h *EvaluationApprovalHandler) ListPending(c *fiber.Ctx) error {
	if err := authz.RequirePermission(c, enums.PermissionEvaluationApprove); err != nil {
		return err
	}

	userID := c.Locals("user_id").(int64)
	approvals, err := evalsvc.ApprovalServiceInstance.FindPendingForApprover(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch pending approvals",
		})
	}
	return c.JSON(approvals)
}

func (h *EvaluationApprovalHandler) Approve(c *fiber.Ctx) error {
	id, _ := c.ParamsInt("id")
	var req struct {
		Comment string `json:"comment"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := evalsvc.ApprovalServiceInstance.Approve(int64(id), c.Locals("user_id").(int64), req.Comment); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.JSON(fiber.Map{
		"message": "Approved successfully",
	})
}

func (h *EvaluationApprovalHandler) Reject(c *fiber.Ctx) error {
	id, _ := c.ParamsInt("id")
	var req struct {
		Comment string `json:"comment"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := evalsvc.ApprovalServiceInstance.Reject(int64(id), c.Locals("user_id").(int64), req.Comment); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.JSON(fiber.Map{
		"message": "Rejected successfully",
	})
}
