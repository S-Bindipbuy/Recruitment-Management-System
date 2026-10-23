package pages

import (
	organization "ems/app/handlers/pages/organization"

	"github.com/gofiber/fiber/v2"
)

func registerOrganizationPages(pages fiber.Router, h *organization.OrganizationHandler) {
	pages.Get("/organizations", h.Organizations)
	pages.Post("/organizations", h.CreateOrganization)
	pages.Put("/organizations/:id", h.UpdateOrganization)
	pages.Delete("/organizations/:id", h.DeleteOrganization)
	pages.Get("/departments", h.Departments)
	pages.Post("/departments", h.CreateDepartment)
	pages.Put("/departments/:id", h.UpdateDepartment)
	pages.Delete("/departments/:id", h.DeleteDepartment)
	pages.Get("/positions", h.Positions)
	pages.Post("/positions", h.CreatePosition)
	pages.Put("/positions/:id", h.UpdatePosition)
	pages.Delete("/positions/:id", h.DeletePosition)
}
