package pages

import (
	"ems/app/handlers/pages/approval"
	"ems/app/handlers/pages/candidate"
	"ems/app/handlers/pages/cycle"
	"ems/app/handlers/pages/dashboard"
	"ems/app/handlers/pages/evaluation"
	"ems/app/handlers/pages/organization"
	"ems/app/handlers/pages/recruitment"
	"ems/app/handlers/pages/report"
	"ems/app/handlers/pages/user"
	"ems/config"

	"github.com/gofiber/fiber/v2"
)

// Register is the whole server-rendered surface: one /pages/ group and one
// file per area of the app behind it, each driven by the handler that owns
// that area. The pages answer app requests (fetch with a bearer token) with
// fragments and direct browser loads with the app shell.
func Register(app *fiber.App, auth fiber.Handler) {
	pages := app.Group("/pages", serveShellForNavigations, auth)

	registerDashboardPages(pages, &dashboard.DashboardHandler{})
	registerEmployeePages(pages, &organization.OrganizationHandler{})
	registerOrganizationPages(pages, &organization.OrganizationHandler{})
	registerCandidatePages(pages, &candidate.CandidateHandler{})
	registerEvaluationPages(pages, &evaluation.EvaluationHandler{})
	registerCyclePages(pages, &cycle.CycleHandler{})
	registerEvaluationFillPages(pages, &evaluation.EvaluationHandler{})
	registerApprovalPages(pages, &approval.ApprovalHandler{})
	registerReportPages(pages, &report.ReportHandler{})
	registerUserPages(pages, &user.UserHandler{})
	registerRolePages(pages, &user.UserHandler{})
	registerPermissionPages(pages, &user.UserHandler{})
	registerRecruitmentPages(pages, &recruitment.RecruitmentHandler{})
}

// serveShellForNavigations answers a plain browser navigation (refresh, pasted
// deep link) with the app shell instead of the bare fragment. Requests made by
// the app itself carry the bearer token and get the fragment.
func serveShellForNavigations(c *fiber.Ctx) error {
	if c.Method() != fiber.MethodGet {
		return c.Next()
	}
	if c.Get("Authorization") != "" || c.Get("X-Requested-With") == "fetch" {
		return c.Next()
	}
	return c.SendFile(config.AppConfig.FrontendPath + "/static/index.html")
}
