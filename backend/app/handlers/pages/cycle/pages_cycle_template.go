package cycle

import (
	"ems/app/enums"
	core "ems/app/handlers/pages/core"
	"ems/app/services/cycles"
	evalsvc "ems/app/services/evaluation"

	"github.com/gofiber/fiber/v2"
)

// ── Step 2: template ──

func (h *CycleHandler) showCycleStepTemplate(c *fiber.Ctx, periodID int64, periodName string) error {
	data := &cycles.CycleData{
		Step:       cycles.CycleStepTemplate,
		PeriodID:   periodID,
		PeriodName: periodName,
	}
	data.LoadTemplateOptions()
	return core.RenderPage(c, "page-cycle", data)
}

func (h *CycleHandler) CycleChooseTemplate(c *fiber.Ctx) error {
	if err := core.RequirePermission(c, enums.PermissionEvaluationManage, enums.PermissionEvaluationAssign); err != nil {
		return err
	}

	periodID := core.FormInt64(c, "period_id")
	period, err := evalsvc.PeriodServiceInstance.FindByID(periodID)
	if err != nil {
		data := &cycles.CycleData{Step: cycles.CycleStepPeriod}
		data.Periods, _ = evalsvc.PeriodServiceInstance.FindAll()
		data.ErrMessage = "That period no longer exists. Please pick one again."
		return core.RenderPage(c, "page-cycle", data)
	}

	templateID := core.FormInt64(c, "template_id")
	tmpl, err := evalsvc.TemplateServiceInstance.FindByID(templateID)
	if err != nil {
		data := &cycles.CycleData{Step: cycles.CycleStepTemplate, PeriodID: period.ID, PeriodName: period.Name}
		data.LoadTemplateOptions()
		data.ErrMessage = "Please choose a template to continue."
		return core.RenderPage(c, "page-cycle", data)
	}
	if tmpl.Status != string(enums.TemplateStatusActive) {
		data := &cycles.CycleData{Step: cycles.CycleStepTemplate, PeriodID: period.ID, PeriodName: period.Name}
		data.LoadTemplateOptions()
		data.ErrMessage = "That template is inactive. Please choose an active one."
		return core.RenderPage(c, "page-cycle", data)
	}

	// A template with no questions would launch evaluations that can only ever
	// score zero, so stop here rather than after the fact.
	if cycles.TemplateQuestionCount(tmpl.ID) == 0 {
		data := &cycles.CycleData{Step: cycles.CycleStepTemplate, PeriodID: period.ID, PeriodName: period.Name}
		data.LoadTemplateOptions()
		data.ErrMessage = "That template has no questions yet. Add sections and questions to it first, otherwise every evaluation would score 0%."
		return core.RenderPage(c, "page-cycle", data)
	}

	return h.showCycleStepScope(c, &cycles.CycleData{
		Step:         cycles.CycleStepScope,
		PeriodID:     period.ID,
		PeriodName:   period.Name,
		TemplateID:   tmpl.ID,
		TemplateName: tmpl.Name,
	})
}
