package pages

import (
	evaluation "ems/app/handlers/pages/evaluation"

	"github.com/gofiber/fiber/v2"
)

func registerEvaluationFillPages(pages fiber.Router, h *evaluation.EvaluationHandler) {
	pages.Get("/evaluation-fill/:id", h.EvaluationFill)
	pages.Post("/evaluation-fill/:id", h.SubmitEvaluation)
	pages.Post("/evaluation-fill/:id/start", h.StartFromFill)
	pages.Post("/evaluation-fill/:id/comments", h.AddComment)
}
