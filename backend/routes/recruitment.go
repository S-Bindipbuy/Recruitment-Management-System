package routes

import (
	recruitment "ems/app/handlers/recruitment"

	"github.com/gofiber/fiber/v2"
)

func registerRecruitmentRoutes(protected fiber.Router) {
	jobPosting := &recruitment.JobPostingHandler{}
	interviewRound := &recruitment.InterviewRoundHandler{}
	interview := &recruitment.InterviewHandler{}
	feedback := &recruitment.InterviewFeedbackHandler{}
	report := &recruitment.ReportHandler{}

	// Job Postings
	protected.Get("/job-postings", jobPosting.List)
	protected.Get("/job-postings/:id", jobPosting.Get)
	protected.Post("/job-postings", jobPosting.Create)
	protected.Put("/job-postings/:id", jobPosting.Update)
	protected.Delete("/job-postings/:id", jobPosting.Delete)
	protected.Post("/job-postings/:id/publish", jobPosting.Publish)
	protected.Post("/job-postings/:id/close", jobPosting.Close)

	// Interview Rounds
	protected.Get("/job-postings/:posting_id/rounds", interviewRound.List)
	protected.Post("/interview-rounds", interviewRound.Create)
	protected.Delete("/interview-rounds/:id", interviewRound.Delete)

	// Interviews
	protected.Get("/interviews", interview.List)
	protected.Get("/interviews/:id", interview.Get)
	protected.Post("/interviews", interview.Create)
	protected.Put("/interviews/:id/status", interview.UpdateStatus)
	protected.Delete("/interviews/:id", interview.Delete)
	protected.Get("/interviews/candidate/:candidate_id", interview.ListByCandidate)

	// Interview Feedback
	protected.Get("/interviews/:interview_id/feedback", feedback.Get)
	protected.Post("/interviews/:interview_id/feedback", feedback.Create)
	protected.Put("/interview-feedback/:id", feedback.Update)

	// Reports
	protected.Get("/reports/candidate-summary/:candidate_id", report.CandidateSummary)
	protected.Get("/reports/candidate-summaries", report.AllCandidateSummaries)
	protected.Get("/reports/interviewer-activity", report.InterviewerActivity)
	protected.Get("/reports/hiring-decisions", report.HiringDecisions)
}
