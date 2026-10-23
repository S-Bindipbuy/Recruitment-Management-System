package dashboard

import (
	core "ems/app/handlers/pages/core"

	"github.com/gofiber/fiber/v2"
)

type DashboardHandler struct{}

func (h *DashboardHandler) Dashboard(c *fiber.Ctx) error {
	data := struct{ Stats core.DashboardStats }{Stats: core.LoadDashboardStats(core.CurrentUserID(c))}
	return core.RenderPage(c, "page-dashboard", data)
}
