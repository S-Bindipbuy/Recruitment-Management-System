package evaluation

import (
	"ems/app/enums"
	"ems/app/handlers/authz"
	evalmodel "ems/app/models/evaluation"
	evalsvc "ems/app/services/evaluation"
	usersvc "ems/app/services/user"
	"github.com/gofiber/fiber/v2"
	"time"
)

type EvaluationAssignmentHandler struct{}

func (h *EvaluationAssignmentHandler) List(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(int64)
	// Managers and assigners see every assignment (the pages UI does the
	// same); everyone else sees the rows they evaluate or are evaluated in.
	var assignments []evalmodel.EvaluationAssignment
	if usersvc.PolicyServiceInstance.CanAny(userID,
		enums.PermissionEvaluationManage, enums.PermissionEvaluationAssign) {
		var err error
		assignments, err = evalsvc.AssignmentServiceInstance.FindAll()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to fetch assignments",
			})
		}
	} else {
		byEvaluator, err := evalsvc.AssignmentServiceInstance.FindByEvaluator(userID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to fetch assignments",
			})
		}
		byEvaluatee, err := evalsvc.AssignmentServiceInstance.FindByEvaluatee(userID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to fetch assignments",
			})
		}
		seen := make(map[int64]bool)
		for _, a := range append(byEvaluator, byEvaluatee...) {
			if !seen[a.ID] {
				seen[a.ID] = true
				assignments = append(assignments, a)
			}
		}
	}
	return c.JSON(assignments)
}

func (h *EvaluationAssignmentHandler) Get(c *fiber.Ctx) error {
	id, _ := c.ParamsInt("id")
	assignment, err := evalsvc.AssignmentServiceInstance.FindByID(int64(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Assignment not found",
		})
	}
	if err := authz.RequireCanViewEvaluation(c, assignment); err != nil {
		return err
	}
	return c.JSON(assignment)
}

func (h *EvaluationAssignmentHandler) Create(c *fiber.Ctx) error {
	if err := authz.RequirePermission(c, enums.PermissionEvaluationAssign); err != nil {
		return err
	}

	var req struct {
		EvaluationPeriodID   int64  `json:"evaluation_period_id"`
		EvaluationTemplateID int64  `json:"evaluation_template_id"`
		EvaluatorID          int64  `json:"evaluator_id"`
		CandidateID          int64  `json:"candidate_id"`
		DueDate              string `json:"due_date"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	var dueDate *time.Time
	if req.DueDate != "" {
		t, _ := time.Parse("2006-01-02", req.DueDate)
		dueDate = &t
	}

	// Every rule for a legal assignment lives in the service, so this endpoint
	// cannot accept something the dialog would reject.
	assignment, err := evalsvc.AssignmentServiceInstance.CreateForCandidate(
		req.EvaluationPeriodID, req.EvaluationTemplateID, req.EvaluatorID, req.CandidateID, dueDate,
	)
	if err != nil {
		status := fiber.StatusInternalServerError
		if evalsvc.AssignmentRuleError(err) {
			status = fiber.StatusBadRequest
		}
		return c.Status(status).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(assignment)
}

func (h *EvaluationAssignmentHandler) Start(c *fiber.Ctx) error {
	id, _ := c.ParamsInt("id")
	assignment, err := evalsvc.AssignmentServiceInstance.FindByID(int64(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Assignment not found",
		})
	}
	if err := authz.RequireCanFillEvaluation(c, assignment); err != nil {
		return err
	}
	if err := evalsvc.AssignmentServiceInstance.Start(int64(id)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to start evaluation",
		})
	}
	return c.JSON(fiber.Map{
		"message": "Evaluation started",
	})
}

func (h *EvaluationAssignmentHandler) Submit(c *fiber.Ctx) error {
	id, _ := c.ParamsInt("id")
	assignment, err := evalsvc.AssignmentServiceInstance.FindByID(int64(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Assignment not found",
		})
	}
	if err := authz.RequireCanFillEvaluation(c, assignment); err != nil {
		return err
	}
	if err := evalsvc.AssignmentServiceInstance.Submit(int64(id)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to submit evaluation",
		})
	}
	return c.JSON(fiber.Map{
		"message": "Evaluation submitted",
	})
}
