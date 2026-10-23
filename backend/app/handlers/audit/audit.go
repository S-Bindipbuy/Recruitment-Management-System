package audit

import (
	"ems/app/enums"
	"ems/app/handlers/authz"
	auditrepo "ems/app/repositories/audit"
	"github.com/gofiber/fiber/v2"
)

type AuditLogHandler struct{}

func (h *AuditLogHandler) List(c *fiber.Ctx) error {
	if err := authz.RequirePermission(c, enums.PermissionAuditView); err != nil {
		return err
	}

	limit := c.QueryInt("limit", 50)
	offset := c.QueryInt("offset", 0)

	logs, err := auditrepo.AuditLogRepo.FindAll(limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch audit logs",
		})
	}
	return c.JSON(logs)
}
