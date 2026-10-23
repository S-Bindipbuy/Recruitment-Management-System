package pages

import (
	dashboard "ems/app/handlers/pages/dashboard"

	"github.com/gofiber/fiber/v2"
)

func registerDashboardPages(pages fiber.Router, h *dashboard.DashboardHandler) {
	pages.Get("", h.Dashboard)
	pages.Get("/", h.Dashboard)
	pages.Get("/dashboard", h.Dashboard)
}
