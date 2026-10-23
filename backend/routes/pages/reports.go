package pages

import (
	report "ems/app/handlers/pages/report"

	"github.com/gofiber/fiber/v2"
)

func registerReportPages(pages fiber.Router, h *report.ReportHandler) {
	pages.Get("/reports", h.Reports)
}
