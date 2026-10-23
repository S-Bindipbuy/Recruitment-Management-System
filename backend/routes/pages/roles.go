package pages

import (
	user "ems/app/handlers/pages/user"

	"github.com/gofiber/fiber/v2"
)

func registerRolePages(pages fiber.Router, h *user.UserHandler) {
	pages.Get("/roles", h.Roles)
	pages.Post("/roles", h.CreateRole)
	pages.Put("/roles/:id", h.UpdateRole)
	pages.Delete("/roles/:id", h.DeleteRole)
	pages.Post("/roles/:id/permissions", h.SetRolePermissions)
}
