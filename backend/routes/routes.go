package routes

import (
	"ems/app/middleware"
	pages "ems/routes/pages"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App) {
	api := app.Group("/api")

	registerComponentRoutes(api)

	// Server-rendered page fragments. Every /pages/ route lives in
	// the pages package - see routes/pages/.
	pages.Register(app, middleware.Protected())

	// Public routes (no auth)
	registerPublicRoutes(app)

	// Protected routes
	protected := api.Group("", middleware.Protected())

	registerAuthRoutes(api, protected)
	registerUserRoutes(protected)
	registerEmployeeRoutes(protected)
	registerCandidateRoutes(protected)
	registerOrganizationRoutes(protected)
	registerDashboardRoutes(protected)
	registerEvaluationRoutes(protected)
	registerAuditRoutes(protected)
	registerRecruitmentRoutes(protected)
}
