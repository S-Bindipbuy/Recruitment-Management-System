package authz

import (
	"ems/app/enums"
	"ems/app/httperr"
	evalmodel "ems/app/models/evaluation"
	usersvc "ems/app/services/user"

	"github.com/gofiber/fiber/v2"
)

// Forbidden aborts with 403. The ResponseError is non-nil so the caller's
// handler chain stops; the app-wide error handler serialises it as JSON and the
// frontend decides how to show it.
func Forbidden(c *fiber.Ctx, msg string) error {
	return httperr.New(fiber.StatusForbidden, msg)
}

// RequirePermission returns an error when the current user lacks any of the
// given permissions. It is the policy gate used inside handlers.
func RequirePermission(c *fiber.Ctx, permissions ...enums.Permission) error {
	if !usersvc.PolicyServiceInstance.CanAny(CurrentUserID(c), permissions...) {
		return Forbidden(c, "You do not have permission to perform this action.")
	}
	return nil
}

// RequireAdmin restricts a handler to administrators only. Managing users,
// roles and permissions is reserved for the admin role, independent of what
// individual permissions a role happens to carry.
func RequireAdmin(c *fiber.Ctx) error {
	if !usersvc.PolicyServiceInstance.IsAdmin(CurrentUserID(c)) {
		return Forbidden(c, "This action is only allowed for administrators.")
	}
	return nil
}

// RequireOwnership enforces that the current user is the one a record belongs
// to (e.g. the assigned evaluator on an evaluation assignment).
func RequireOwnership(c *fiber.Ctx, ownerID int64) error {
	if CurrentUserID(c) != ownerID {
		return Forbidden(c, "This action is only allowed for the assigned user.")
	}
	return nil
}

// RequireCanFillEvaluation enforces the policy that an evaluation assignment
// can only be started, filled, saved, commented on or submitted by either:
//   - a user privileged to manage or assign evaluations (admins, managers), or
//   - the assigned evaluator holding the evaluation.evaluate permission.
//
// Without this, an admin who creates an assignment but is not itself the
// chosen evaluator would never be able to drive the evaluation.
func RequireCanFillEvaluation(c *fiber.Ctx, assignment *evalmodel.EvaluationAssignment) error {
	if usersvc.PolicyServiceInstance.CanAny(CurrentUserID(c),
		enums.PermissionEvaluationManage, enums.PermissionEvaluationAssign) {
		return nil
	}
	if err := RequirePermission(c, enums.PermissionEvaluationEvaluate); err != nil {
		return err
	}
	return RequireOwnership(c, assignment.EvaluatorID)
}

// RequireCanViewEvaluation allows managers/auditors to review an evaluation
// while still letting the assigned evaluator see their own work.
func RequireCanViewEvaluation(c *fiber.Ctx, assignment *evalmodel.EvaluationAssignment) error {
	if usersvc.PolicyServiceInstance.CanAny(CurrentUserID(c), enums.PermissionEvaluationManage, enums.PermissionAuditView) {
		return nil
	}
	return RequireCanFillEvaluation(c, assignment)
}

// The page-rendering handlers moved to the pages subpackage, which cannot
// import this package without a cycle. This lookup is still needed by the API
// handlers here, so it stays as a small local copy.

// CurrentUserID reads the identity the auth middleware stashed on the request.
func CurrentUserID(c *fiber.Ctx) int64 {
	if id, ok := c.Locals("user_id").(int64); ok {
		return id
	}
	return 0
}
