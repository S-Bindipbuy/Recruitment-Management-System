package cycle

import (
	"ems/app/enums"
	core "ems/app/handlers/pages/core"
	"ems/app/services/cycles"
	evalsvc "ems/app/services/evaluation"

	"github.com/gofiber/fiber/v2"
)

type CycleHandler struct{}

// ── Guided evaluation cycle ──
//
// Running an evaluation cycle used to mean visiting four separate pages in a
// fixed order (Periods, Templates, Evaluations, then Approvals) and holding
// the order in your head. The wizard walks the same steps one at a time and
// carries the choices forward as hidden form fields, so the user only ever has
// to decide the current step. What each step may do lives in services/cycle.go;
// this file is the HTTP surface around it.

// ── Back navigation ──
//
// Each step is rendered by a distinct endpoint so the wizard can go backwards
// without re-submitting (or losing) the choices already made.

func (h *CycleHandler) CycleStepPeriod(c *fiber.Ctx) error {
	if err := core.RequirePermission(c, enums.PermissionEvaluationManage, enums.PermissionEvaluationAssign); err != nil {
		return err
	}
	data := &cycles.CycleData{Step: cycles.CycleStepPeriod}
	data.Periods, _ = evalsvc.PeriodServiceInstance.FindAll()
	return core.RenderPage(c, "page-cycle", data)
}

func (h *CycleHandler) CycleStepTemplateShow(c *fiber.Ctx) error {
	if err := core.RequirePermission(c, enums.PermissionEvaluationManage, enums.PermissionEvaluationAssign); err != nil {
		return err
	}
	periodID := core.FormInt64(c, "period_id")
	period, err := evalsvc.PeriodServiceInstance.FindByID(periodID)
	if err != nil {
		return h.CycleStepPeriod(c)
	}
	return h.showCycleStepTemplate(c, period.ID, period.Name)
}

func (h *CycleHandler) CycleStepScopeShow(c *fiber.Ctx) error {
	if err := core.RequirePermission(c, enums.PermissionEvaluationManage, enums.PermissionEvaluationAssign); err != nil {
		return err
	}
	periodID := core.FormInt64(c, "period_id")
	templateID := core.FormInt64(c, "template_id")
	period, perr := evalsvc.PeriodServiceInstance.FindByID(periodID)
	tmpl, terr := evalsvc.TemplateServiceInstance.FindByID(templateID)
	if perr != nil || terr != nil {
		return h.CycleStepPeriod(c)
	}

	// Carry the scope choices back so the step renders as the user left it
	// instead of making them pick the same three things a second time.
	data := &cycles.CycleData{
		Step:         cycles.CycleStepScope,
		PeriodID:     period.ID,
		PeriodName:   period.Name,
		TemplateID:   tmpl.ID,
		TemplateName: tmpl.Name,
		DueDate:      c.FormValue("due_date"),
		EvaluatorID:  core.FormInt64(c, "evaluator_id"),
	}
	return h.showCycleStepScope(c, data)
}
