package evaluation

import (
	"ems/app/enums"
	core "ems/app/handlers/pages/core"
	evalmodel "ems/app/models/evaluation"
	evalsvc "ems/app/services/evaluation"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type PeriodsPageData struct {
	Periods []evalmodel.EvaluationPeriod
}

func (h *EvaluationHandler) Periods(c *fiber.Ctx) error {
	data := PeriodsPageData{}
	data.Periods, _ = evalsvc.PeriodServiceInstance.FindAll()
	return core.RenderPage(c, "page-periods", data)
}

func (h *EvaluationHandler) CreatePeriodPage(c *fiber.Ctx) error {
	if err := core.RequirePermission(c, enums.PermissionEvaluationManage); err != nil {
		return err
	}

	if _, err := evalsvc.PeriodServiceInstance.Create(
		c.FormValue("name"),
		c.FormValue("description"),
		c.FormValue("start_date"),
		c.FormValue("end_date"),
		core.CurrentUserID(c),
	); err != nil {
		return core.RenderError(c, fiber.StatusInternalServerError, "Failed to create period.")
	}
	return h.Periods(c)
}

func (h *EvaluationHandler) UpdatePeriod(c *fiber.Ctx) error {
	if err := core.RequirePermission(c, enums.PermissionEvaluationManage); err != nil {
		return err
	}

	id, _ := strconv.ParseInt(c.Params("id"), 10, 64)
	name := strings.TrimSpace(c.FormValue("name"))
	if name == "" {
		return core.RenderError(c, fiber.StatusBadRequest, "Please provide a name.")
	}
	if _, err := evalsvc.PeriodServiceInstance.Update(id, name, c.FormValue("description"), c.FormValue("start_date"), c.FormValue("end_date"), c.FormValue("status")); err != nil {
		return core.RenderError(c, fiber.StatusInternalServerError, "Failed to update period.")
	}
	return h.Periods(c)
}

func (h *EvaluationHandler) DeletePeriod(c *fiber.Ctx) error {
	if err := core.RequirePermission(c, enums.PermissionEvaluationManage); err != nil {
		return err
	}

	id, _ := strconv.ParseInt(c.Params("id"), 10, 64)
	if err := evalsvc.PeriodServiceInstance.Delete(id); err != nil {
		return core.RenderError(c, fiber.StatusInternalServerError, "Failed to delete period.")
	}
	return h.Periods(c)
}
