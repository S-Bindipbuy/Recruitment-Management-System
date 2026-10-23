package pages

import (
	user "ems/app/handlers/pages/user"

	"github.com/gofiber/fiber/v2"
)

func registerUserPages(pages fiber.Router, h *user.UserHandler) {
	pages.Get("/users", h.Users)
	pages.Post("/users", h.CreateUser)
	pages.Put("/users/:id", h.UpdateUser)
	pages.Delete("/users/:id", h.DeleteUser)
}
