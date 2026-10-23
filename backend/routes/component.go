package routes

import (
	"ems/app/handlers"

	"github.com/gofiber/fiber/v2"
)

func registerComponentRoutes(api fiber.Router) {
	h := &handlers.ComponentHandler{}
	api.Get("/components/:name", h.GetComponent)
}
