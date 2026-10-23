package cycle

import (
	"ems/app/enums"
	core "ems/app/handlers/pages/core"
	"ems/app/services/cycles"
	evalsvc "ems/app/services/evaluation"

	"github.com/gofiber/fiber/v2"
)

// ── Step 4: review and launch ──

func (h *CycleHandler) CycleLaunch(c *fiber.Ctx) error {
	if err := core.RequirePermission(c, enums.PermissionEvaluationAssign, enums.PermissionEvaluationManage); err != nil {
		return err
	}

	form := core.ReadScopeForm(c)

	period, perr := evalsvc.PeriodServiceInstance.FindByID(form.PeriodID)
	tmpl, terr := evalsvc.TemplateServiceInstance.FindByID(form.TemplateID)
	if perr != nil || terr != nil {
		data := &cycles.CycleData{Step: cycles.CycleStepPeriod}
		data.Periods, _ = evalsvc.PeriodServiceInstance.FindAll()
		data.ErrMessage = "The period or template is no longer available. Please start again."
		return core.RenderPage(c, "page-cycle", data)
	}

	// Recompute the pairing at launch time rather than trusting the ids that
	// were previewed: the preview is display-only, and the pipeline may have
	// moved on since step 3.
	evaluatorID := form.EvaluatorID
	selected := form.CandidateIDs

	pairs := cycles.CycleCandidatePairs(period.ID, tmpl.ID, evaluatorID, selected)
	previews, _ := cycles.PreviewCandidatePairs(period.ID, tmpl.ID, evaluatorID, selected)

	if len(pairs) == 0 {
		// Send the user back to step 3 with every choice intact: the scope is
		// already fully covered, and they need to widen it (other period,
		// other evaluator) rather than re-pick everything.
		data := &cycles.CycleData{
			Step:                 cycles.CycleStepScope,
			PeriodID:             period.ID,
			PeriodName:           period.Name,
			TemplateID:           tmpl.ID,
			TemplateName:         tmpl.Name,
			DueDate:              form.DueDate,
			EvaluatorID:          evaluatorID,
			SelectedCandidateIDs: selected,
			ErrMessage:           "There is nothing left to assign. Everything in this scope is already covered for this period - try another period, another template, or someone new.",
		}
		if evaluatorID > 0 {
			data.EvaluatorName = core.UserName(evaluatorID)
		}
		return h.showCycleStepScope(c, data)
	}

	dueDate := core.DateFromForm(form.DueDate)

	if err := evalsvc.AssignmentServiceInstance.CreateMany(period.ID, tmpl.ID, pairs, dueDate); err != nil {
		data := &cycles.CycleData{
			Step:                 cycles.CycleStepReview,
			PeriodID:             period.ID,
			PeriodName:           period.Name,
			TemplateID:           tmpl.ID,
			TemplateName:         tmpl.Name,
			DueDate:              form.DueDate,
			Pairs:                previews,
			WillCreate:           len(pairs),
			EvaluatorID:          evaluatorID,
			SelectedCandidateIDs: selected,
			ErrMessage:           "The evaluations could not be created. Nothing was saved - please try again.",
		}
		if evaluatorID > 0 {
			data.EvaluatorName = core.UserName(evaluatorID)
		}
		return core.RenderPage(c, "page-cycle", data)
	}

	data := &cycles.CycleData{
		Step:                 cycles.CycleStepReview,
		Launched:             true,
		Created:              len(pairs),
		Skipped:              len(previews) - len(pairs),
		PeriodID:             period.ID,
		PeriodName:           period.Name,
		TemplateID:           tmpl.ID,
		TemplateName:         tmpl.Name,
		DueDate:              form.DueDate,
		SelectedCandidateIDs: selected,
	}
	if evaluatorID > 0 {
		data.EvaluatorName = core.UserName(evaluatorID)
	}
	return core.RenderPage(c, "page-cycle", data)
}
