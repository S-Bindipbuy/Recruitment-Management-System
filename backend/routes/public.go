package routes

import (
	public "ems/app/handlers/public"

	"github.com/gofiber/fiber/v2"
)

func registerPublicRoutes(app *fiber.App) {
	h := &public.PublicHandler{}
	jobH := &public.PublicJobHandler{}

	// Public job board and apply (no auth required, with layout)
	app.Get("/jobs", jobH.PublicJobs)
	app.Get("/jobs/:id", jobH.PublicJobDetail)
	app.Get("/apply", jobH.PublicApply)
	app.Post("/apply", h.SubmitCV)
}
