package pages

import (
	approval "ems/app/handlers/pages/approval"

	"github.com/gofiber/fiber/v2"
)

func registerApprovalPages(pages fiber.Router, h *approval.ApprovalHandler) {
	pages.Get("/approvals", h.Approvals)
	pages.Post("/approvals", h.RequestApproval)
	pages.Post("/approvals/:id/approve", h.ApproveApproval)
	pages.Post("/approvals/:id/reject", h.RejectApproval)
}
