package pages

import (
	approval "ems/app/handlers/pages/approval"

	"github.com/gofiber/fiber/v2"
)

func registerApprovalPages(pages fiber.Router, h *approval.ApprovalHandler) {
	pages.Get("/approvals", h.Approvals)
	pages.Post("/approvals/candidates/:id/accept", h.AcceptCandidate)
}
