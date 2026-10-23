package dashboard

import (
	"ems/app/models"
	statsrepo "ems/app/repositories/stats"
	"github.com/gofiber/fiber/v2"
)

type DashboardHandler struct{}

func (h *DashboardHandler) Stats(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(int64)

	stats, err := statsrepo.StatsRepo.Dashboard(userID)
	if err != nil {
		stats = models.DashboardStats{}
	}
	return c.JSON(stats)
}
