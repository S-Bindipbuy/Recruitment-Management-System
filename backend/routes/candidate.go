package routes

import (
	candidate "ems/app/handlers/candidate"

	"github.com/gofiber/fiber/v2"
)

func registerCandidateRoutes(protected fiber.Router) {
	h := &candidate.CandidateHandler{}
	protected.Get("/candidates", h.List)
	protected.Post("/candidates", h.Create)
	protected.Get("/candidates/:id/cv", h.DownloadCV)
}
