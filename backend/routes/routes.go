package routes

import (
	"ems/app/middleware"
	"ems/config"
	pages "ems/routes/pages"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App) {
	// normal static serve
	app.Static("/static", config.AppConfig.FrontendPath+"/static")

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendFile(config.AppConfig.FrontendPath + "/static/index.html")
	})

	app.Get("/favicon.ico", func(c *fiber.Ctx) error {
		c.Set("Content-Type", "image/x-icon")
		c.Set("Cache-Control", "public, max-age=86400")
		return c.SendFile(config.AppConfig.FrontendPath + "/static/favicon.ico")
	})

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
