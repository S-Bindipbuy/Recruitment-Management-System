package routes

import (
	user "ems/app/handlers/user"

	"github.com/gofiber/fiber/v2"
)

func registerUserRoutes(protected fiber.Router) {
	h := &user.UserHandler{}

	protected.Get("/users", h.List)
	protected.Get("/users/:id", h.Get)
	protected.Put("/users/:id", h.Update)
	protected.Delete("/users/:id", h.Delete)
}
