package routes

import (
	"ems/app/handlers"

	"github.com/gofiber/fiber/v2"
)

func registerAuthRoutes(api fiber.Router, protected fiber.Router) {
	h := &handlers.AuthHandler{}

	auth := api.Group("/auth")
	auth.Post("/login", h.Login)
	auth.Post("/register", h.Register)

	protected.Get("/auth/me", h.Me)
}
