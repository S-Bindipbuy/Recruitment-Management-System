package pages

import (
	cycle "ems/app/handlers/pages/cycle"

	"github.com/gofiber/fiber/v2"
)

func registerCyclePages(pages fiber.Router, h *cycle.CycleHandler) {
	// Guided cycle wizard. Each step is its own endpoint and hands the chosen
	// values to the next one as hidden fields.
	pages.Get("/cycle", h.CycleStart)
	pages.Post("/cycle/period", h.CycleChoosePeriod)
	pages.Post("/cycle/template", h.CycleChooseTemplate)
	pages.Post("/cycle/scope", h.CycleChooseScope)
	pages.Post("/cycle/launch", h.CycleLaunch)
	pages.Post("/cycle/back/period", h.CycleStepPeriod)
	pages.Post("/cycle/back/template", h.CycleStepTemplateShow)
	pages.Post("/cycle/back/scope", h.CycleStepScopeShow)

	// Step 3 no longer has a GET (ticking is local now), but old history
	// entries and bookmarks to it must still resolve instead of 404ing.
	pages.Get("/cycle/scope", func(c *fiber.Ctx) error {
		return c.Redirect("/pages/cycle", fiber.StatusFound)
	})

	// Taking back a row the picker already flags as assigned. The button lives
	// inside the scope form, so the step's own fields ride along and the
	// response can redraw the step with nothing else to retype.
	pages.Post("/cycle/assignment/drop", h.CycleDropAssignment)
}
