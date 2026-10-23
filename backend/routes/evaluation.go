package routes

import (
	evaluation "ems/app/handlers/evaluation"

	"github.com/gofiber/fiber/v2"
)

func registerEvaluationRoutes(protected fiber.Router) {
	period := &evaluation.EvaluationPeriodHandler{}
	tmpl := &evaluation.EvaluationTemplateHandler{}
	sec := &evaluation.EvaluationSectionHandler{}
	q := &evaluation.EvaluationQuestionHandler{}
	assign := &evaluation.EvaluationAssignmentHandler{}
	resp := &evaluation.EvaluationResponseHandler{}
	result := &evaluation.EvaluationResultHandler{}
	comment := &evaluation.EvaluationCommentHandler{}
	approval := &evaluation.EvaluationApprovalHandler{}

	protected.Get("/evaluation-periods", period.List)
	protected.Get("/evaluation-periods/:id", period.Get)
	protected.Post("/evaluation-periods", period.Create)
	protected.Put("/evaluation-periods/:id", period.Update)
	protected.Delete("/evaluation-periods/:id", period.Delete)

	protected.Get("/evaluation-templates", tmpl.List)
	protected.Get("/evaluation-templates/:id", tmpl.Get)
	protected.Post("/evaluation-templates", tmpl.Create)
	protected.Put("/evaluation-templates/:id", tmpl.Update)
	protected.Delete("/evaluation-templates/:id", tmpl.Delete)

	protected.Get("/evaluation-templates/:template_id/sections", sec.ListByTemplate)
	protected.Post("/evaluation-sections", sec.Create)
	protected.Put("/evaluation-sections/:id", sec.Update)
	protected.Delete("/evaluation-sections/:id", sec.Delete)

	protected.Get("/evaluation-sections/:section_id/questions", q.ListBySection)
	protected.Post("/evaluation-questions", q.Create)
	protected.Put("/evaluation-questions/:id", q.Update)
	protected.Delete("/evaluation-questions/:id", q.Delete)

	protected.Get("/evaluation-assignments", assign.List)
	protected.Get("/evaluation-assignments/:id", assign.Get)
	protected.Post("/evaluation-assignments", assign.Create)
	protected.Post("/evaluation-assignments/:id/start", assign.Start)
	protected.Post("/evaluation-assignments/:id/submit", assign.Submit)

	protected.Post("/evaluation-responses", resp.SaveResponse)
	protected.Get("/evaluation-assignments/:assignment_id/responses", resp.GetResponses)

	protected.Post("/evaluation-assignments/:assignment_id/calculate", result.Calculate)
	protected.Get("/evaluation-assignments/:assignment_id/result", result.GetResult)

	protected.Post("/evaluation-comments", comment.Create)
	protected.Get("/evaluation-assignments/:assignment_id/comments", comment.List)

	protected.Get("/evaluation-approvals/pending", approval.ListPending)
	protected.Post("/evaluation-approvals", approval.Create)
	protected.Post("/evaluation-approvals/:id/approve", approval.Approve)
	protected.Post("/evaluation-approvals/:id/reject", approval.Reject)
}
