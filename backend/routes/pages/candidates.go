package pages

import (
	candidate "ems/app/handlers/pages/candidate"

	"github.com/gofiber/fiber/v2"
)

func registerCandidatePages(pages fiber.Router, h *candidate.CandidateHandler) {
	pages.Get("/candidates", h.Candidates)
	pages.Post("/candidates", h.CreateCandidate)
	pages.Post("/candidates/:id/assign-eval", h.AssignCandidateEvaluation)
	pages.Post("/candidates/:id/hire", h.HireCandidate)
}
