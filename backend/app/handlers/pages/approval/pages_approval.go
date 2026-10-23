package approval

import (
	"ems/app/enums"
	core "ems/app/handlers/pages/core"
	evalmodel "ems/app/models/evaluation"
	candsvc "ems/app/services/candidate"
	evalsvc "ems/app/services/evaluation"
	orgsvc "ems/app/services/organization"
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type ApprovalHandler struct{}

type ApprovalView struct {
	Approval      evalmodel.EvaluationApproval
	ApproverName  string
	EvaluateeName string
}

type AssignableOption struct {
	ID    int64
	Label string
}

type ApprovalsPageData struct {
	Approvals  []ApprovalView
	Workers    []core.WorkerOption
	Assignable []AssignableOption
}

func (h *ApprovalHandler) Approvals(c *fiber.Ctx) error {
	if err := core.RequirePermission(c, enums.PermissionEvaluationApprove, enums.PermissionEvaluationManage); err != nil {
		return err
	}

	data := ApprovalsPageData{}
	data.Workers = core.LoadActiveWorkers()

	allAssignments, _ := evalsvc.AssignmentServiceInstance.FindAll()
	for _, a := range allAssignments {
		name := ""
		if a.CandidateID != nil && *a.CandidateID > 0 {
			if cand, err := candsvc.CandidateServiceInstance.FindByID(*a.CandidateID); err == nil {
				name = cand.FullName + " (Candidate)"
			}
		} else {
			name = core.UserName(a.EvaluateeID)
		}
		data.Assignable = append(data.Assignable, AssignableOption{
			ID:    a.ID,
			Label: fmt.Sprintf("Assignment #%d · %s (%s)", a.ID, name, a.Status),
		})
	}

	approvals, err := evalsvc.ApprovalServiceInstance.FindPendingForApprover(core.CurrentUserID(c))
	if err != nil {
		approvals = nil
	}

	approverIDs := make([]int64, 0, len(approvals))
	for _, ap := range approvals {
		approverIDs = append(approverIDs, ap.ApproverID)
	}
	names := core.UserNames(approverIDs)

	for _, ap := range approvals {
		evalName := ""
		if assign, err := evalsvc.AssignmentServiceInstance.FindByID(ap.EvaluationAssignmentID); err == nil && assign != nil {
			if assign.CandidateID != nil && *assign.CandidateID > 0 {
				if cand, err := candsvc.CandidateServiceInstance.FindByID(*assign.CandidateID); err == nil {
					evalName = cand.FullName + " (Candidate)"
				}
			} else {
				evalName = core.UserName(assign.EvaluateeID)
			}
		}
		data.Approvals = append(data.Approvals, ApprovalView{
			Approval:      ap,
			ApproverName:  core.DisplayName(names, ap.ApproverID),
			EvaluateeName: evalName,
		})
	}
	return core.RenderPage(c, "page-approvals", data)
}

func (h *ApprovalHandler) RequestApproval(c *fiber.Ctx) error {
	if err := core.RequirePermission(c, enums.PermissionEvaluationApprove, enums.PermissionEvaluationManage); err != nil {
		return err
	}

	assignmentID := core.FormInt64(c, "assignment_id")
	approverID := core.FormInt64(c, "approver_id")
	level, _ := strconv.Atoi(c.FormValue("level"))
	if level < 1 {
		level = 1
	}

	if assignmentID <= 0 || approverID <= 0 {
		return core.RenderError(c, fiber.StatusBadRequest, "Please provide a valid assignment and approver.")
	}
	assignment, err := evalsvc.AssignmentServiceInstance.FindByID(assignmentID)
	if err != nil {
		return core.RenderError(c, fiber.StatusBadRequest, "The selected assignment was not found.")
	}
	if approverID == assignment.EvaluatorID {
		return core.RenderError(c, fiber.StatusBadRequest,
			"The evaluator cannot approve their own evaluation. Choose somebody else as approver.")
	}
	if assignment.EvaluateeID > 0 && approverID == assignment.EvaluateeID {
		return core.RenderError(c, fiber.StatusBadRequest,
			"The evaluated worker cannot approve their own evaluation. Choose somebody else as approver.")
	}
	if !orgsvc.IsActiveWorker(approverID) {
		return core.RenderError(c, fiber.StatusBadRequest, "Approvals can only be requested from selected workers.")
	}

	if _, err := evalsvc.ApprovalServiceInstance.Create(assignmentID, approverID, level); err != nil {
		return core.RenderError(c, fiber.StatusInternalServerError, "Failed to create approval request.")
	}
	return h.Approvals(c)
}

func (h *ApprovalHandler) ApproveApproval(c *fiber.Ctx) error {
	id, _ := c.ParamsInt("id")
	if err := evalsvc.ApprovalServiceInstance.Approve(int64(id), core.CurrentUserID(c), c.FormValue("comment")); err != nil {
		return core.RenderError(c, fiber.StatusBadRequest, err.Error())
	}
	return h.Approvals(c)
}

func (h *ApprovalHandler) RejectApproval(c *fiber.Ctx) error {
	id, _ := c.ParamsInt("id")
	if err := evalsvc.ApprovalServiceInstance.Reject(int64(id), core.CurrentUserID(c), c.FormValue("comment")); err != nil {
		return core.RenderError(c, fiber.StatusBadRequest, err.Error())
	}
	return h.Approvals(c)
}
