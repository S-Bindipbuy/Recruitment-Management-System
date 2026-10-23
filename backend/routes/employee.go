package routes

import (
	employee "ems/app/handlers/employee"

	"github.com/gofiber/fiber/v2"
)

func registerEmployeeRoutes(protected fiber.Router) {
	h := &employee.EmployeeHandler{}

	protected.Get("/employees", h.List)
	protected.Get("/employees/:id", h.Get)
	protected.Post("/employees", h.Create)
	protected.Put("/employees/:id", h.Update)
	protected.Delete("/employees/:id", h.Delete)
}
