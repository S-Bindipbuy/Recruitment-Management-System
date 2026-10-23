package pages

import (
	user "ems/app/handlers/pages/user"

	"github.com/gofiber/fiber/v2"
)

func registerPermissionPages(pages fiber.Router, h *user.UserHandler) {
	pages.Get("/permissions", h.Permissions)
	pages.Post("/permissions", h.CreatePermission)
	pages.Delete("/permissions/:id", h.DeletePermission)
}
