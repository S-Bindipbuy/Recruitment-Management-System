package routes

import (
	audit "ems/app/handlers/audit"

	"github.com/gofiber/fiber/v2"
)

func registerAuditRoutes(protected fiber.Router) {
	h := &audit.AuditLogHandler{}
	protected.Get("/audit-logs", h.List)
}
