package pages

import (
	recruitment "ems/app/handlers/pages/recruitment"

	"github.com/gofiber/fiber/v2"
)

func registerRecruitmentPages(pages fiber.Router, h *recruitment.RecruitmentHandler) {
	pages.Get("/job-postings", h.JobPostings)
	pages.Post("/job-postings", h.CreateJobPosting)
	pages.Post("/job-postings/:id/publish", h.PublishJobPosting)
	pages.Post("/job-postings/:id/close", h.CloseJobPosting)
	pages.Post("/job-postings/:id/reopen", h.ReopenJobPosting)
	pages.Post("/job-postings/:id/delete", h.DeleteJobPosting)
	pages.Get("/job-postings/:id/candidates", h.JobPostingCandidates)

	// Job posting wizard. Same shape as the evaluation cycle wizard: one
	// decision per screen, answers carried forward as hidden fields.
	pages.Get("/job-wizard", h.JobWizard)
	pages.Post("/job-wizard/basics", h.JobWizardBasics)
	pages.Post("/job-wizard/details", h.JobWizardDetails)
	pages.Post("/job-wizard/terms", h.JobWizardTerms)
	pages.Post("/job-wizard/launch", h.JobWizardLaunch)
	pages.Post("/job-wizard/back/basics", h.JobWizardBackBasics)
	pages.Post("/job-wizard/back/details", h.JobWizardBackDetails)
	pages.Post("/job-wizard/back/terms", h.JobWizardBackTerms)

	pages.Get("/interviews", h.Interviews)
	pages.Post("/interviews", h.CreateInterview)
	pages.Post("/interviews/:id/status", h.UpdateInterviewStatus)
	pages.Post("/interviews/:id/delete", h.DeleteInterview)

	pages.Get("/interviews/:id/feedback", h.InterviewFeedback)
	pages.Post("/interviews/:id/feedback", h.SubmitInterviewFeedback)

	pages.Get("/recruitment-reports", func(c *fiber.Ctx) error {
		return c.Redirect("/pages/reports")
	})
}
