package routes

import (
	dashboard "ems/app/handlers/dashboard"

	"github.com/gofiber/fiber/v2"
)

func registerDashboardRoutes(protected fiber.Router) {
	h := &dashboard.DashboardHandler{}
	protected.Get("/dashboard/stats", h.Stats)
}
