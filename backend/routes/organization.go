package routes

import (
	organization "ems/app/handlers/organization"

	"github.com/gofiber/fiber/v2"
)

func registerOrganizationRoutes(protected fiber.Router) {
	org := &organization.OrganizationHandler{}
	dept := &organization.DepartmentHandler{}
	pos := &organization.PositionHandler{}

	protected.Get("/organizations", org.List)
	protected.Post("/organizations", org.Create)
	protected.Put("/organizations/:id", org.Update)
	protected.Delete("/organizations/:id", org.Delete)

	protected.Get("/departments", dept.List)
	protected.Post("/departments", dept.Create)
	protected.Put("/departments/:id", dept.Update)
	protected.Delete("/departments/:id", dept.Delete)

	protected.Get("/positions", pos.List)
	protected.Post("/positions", pos.Create)
	protected.Put("/positions/:id", pos.Update)
	protected.Delete("/positions/:id", pos.Delete)
}
