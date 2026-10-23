package cycle

import (
	"ems/app/enums"
	core "ems/app/handlers/pages/core"
	"ems/app/services/cycles"
	evalsvc "ems/app/services/evaluation"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// ── Step 1: period ──

func (h *CycleHandler) CycleStart(c *fiber.Ctx) error {
	if err := core.RequirePermission(c, enums.PermissionEvaluationManage, enums.PermissionEvaluationAssign); err != nil {
		return err
	}

	data := &cycles.CycleData{Step: cycles.CycleStepPeriod}
	periods, _ := evalsvc.PeriodServiceInstance.FindAll()
	data.Periods = periods
	return core.RenderPage(c, "page-cycle", data)
}

// CycleChoosePeriod handles both picking an existing period and creating a new
// one from the same form, so step 1 never needs a second screen.
func (h *CycleHandler) CycleChoosePeriod(c *fiber.Ctx) error {
	if err := core.RequirePermission(c, enums.PermissionEvaluationManage, enums.PermissionEvaluationAssign); err != nil {
		return err
	}

	data := &cycles.CycleData{Step: cycles.CycleStepPeriod}
	periods, _ := evalsvc.PeriodServiceInstance.FindAll()
	data.Periods = periods

	name := strings.TrimSpace(c.FormValue("new_period_name"))
	if name == "" {
		// Nothing to create: the user picked an existing period.
		id := core.FormInt64(c, "existing_period_id")
		period, err := evalsvc.PeriodServiceInstance.FindByID(id)
		if err != nil {
			data.ErrMessage = "Please choose a period to continue."
			return core.RenderPage(c, "page-cycle", data)
		}
		return h.showCycleStepTemplate(c, period.ID, period.Name)
	}

	startDate := c.FormValue("new_period_start")
	endDate := c.FormValue("new_period_end")
	if startDate == "" || endDate == "" {
		data.ErrMessage = "A new period needs both a start and an end date."
		return core.RenderPage(c, "page-cycle", data)
	}
	if endDate < startDate {
		data.ErrMessage = "The end date must be on or after the start date."
		return core.RenderPage(c, "page-cycle", data)
	}

	period, err := evalsvc.PeriodServiceInstance.Create(name, c.FormValue("new_period_description"), startDate, endDate, core.CurrentUserID(c))
	if err != nil {
		data.ErrMessage = "Failed to create the period. Please try again."
		return core.RenderPage(c, "page-cycle", data)
	}
	return h.showCycleStepTemplate(c, period.ID, period.Name)
}
