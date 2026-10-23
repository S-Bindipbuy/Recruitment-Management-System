package pages

import (
	organization "ems/app/handlers/pages/organization"

	"github.com/gofiber/fiber/v2"
)

func registerEmployeePages(pages fiber.Router, h *organization.OrganizationHandler) {
	pages.Get("/employees", h.Employees)
	pages.Post("/employees", h.CreateEmployee)
}
