package pages

import (
	evaluation "ems/app/handlers/pages/evaluation"

	"github.com/gofiber/fiber/v2"
)

func registerEvaluationPages(pages fiber.Router, h *evaluation.EvaluationHandler) {
	pages.Get("/evaluations", h.Evaluations)
	pages.Post("/evaluations/assignments", h.CreateAssignment)
	pages.Post("/evaluations/:id/start", h.StartAssignment)

	pages.Get("/periods", h.Periods)
	pages.Post("/periods", h.CreatePeriodPage)
	pages.Put("/periods/:id", h.UpdatePeriod)
	pages.Delete("/periods/:id", h.DeletePeriod)

	pages.Get("/templates-eval", h.TemplatesEval)
	pages.Get("/templates-eval/:id", h.TemplateDetail)
	pages.Post("/templates", h.CreateTemplatePage)
	pages.Put("/templates/:id", h.UpdateTemplate)
	pages.Delete("/templates/:id", h.DeleteTemplate)

	// Sections and questions of a template. The cycle wizard refuses a template
	// with no questions, so this is the only way to make one usable.
	pages.Post("/templates-eval/:id/sections", h.CreateSectionPage)
	pages.Post("/sections/:id/questions", h.CreateQuestionPage)
	pages.Delete("/sections/:id", h.DeleteSectionPage)
	pages.Delete("/questions/:id", h.DeleteQuestionPage)

	pages.Get("/periods/:id/history", h.PeriodHistory)
}
