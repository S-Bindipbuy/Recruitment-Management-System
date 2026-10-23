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

type TemplateListItem struct {
	evalmodel.EvaluationTemplate
	SectionCount  int
	QuestionCount int
}

type TemplatesPageData struct {
	Templates []TemplateListItem
	// PeriodID mirrors the wizard's current period so a detour through this
	// page does not strand the visitor mid-cycle.
	PeriodID int64
}

func (h *EvaluationHandler) TemplatesEval(c *fiber.Ctx) error {
	periodID, _ := strconv.ParseInt(c.Query("period_id"), 10, 64)
	return h.templatesEvalPage(c, periodID)
}

func (h *EvaluationHandler) templatesEvalPage(c *fiber.Ctx, periodID int64) error {
	data := TemplatesPageData{PeriodID: periodID}

	templates, _ := evalsvc.TemplateServiceInstance.FindAll()
	counts, _ := evalsvc.TemplateServiceInstance.StructureCounts()
	for _, t := range templates {
		item := TemplateListItem{EvaluationTemplate: t}
		if st, ok := counts[t.ID]; ok {
			item.SectionCount = st.Sections
			item.QuestionCount = st.Questions
		}
		data.Templates = append(data.Templates, item)
	}
	return core.RenderPage(c, "page-templates-eval", data)
}

func (h *EvaluationHandler) CreateTemplatePage(c *fiber.Ctx) error {
	if err := core.RequirePermission(c, enums.PermissionEvaluationManage); err != nil {
		return err
	}

	if _, err := evalsvc.TemplateServiceInstance.Create(
		c.FormValue("name"),
		c.FormValue("description"),
		core.CurrentUserID(c),
	); err != nil {
		return core.RenderError(c, fiber.StatusInternalServerError, "Failed to create template.")
	}
	return h.templatesEvalPage(c, periodIDFromForm(c))
}

func (h *EvaluationHandler) UpdateTemplate(c *fiber.Ctx) error {
	if err := core.RequirePermission(c, enums.PermissionEvaluationManage); err != nil {
		return err
	}

	id, _ := strconv.ParseInt(c.Params("id"), 10, 64)
	name := strings.TrimSpace(c.FormValue("name"))
	if name == "" {
		return core.RenderError(c, fiber.StatusBadRequest, "Please provide a name.")
	}
	if _, err := evalsvc.TemplateServiceInstance.Update(id, name, c.FormValue("description"), c.FormValue("status")); err != nil {
		return core.RenderError(c, fiber.StatusInternalServerError, "Failed to update template.")
	}
	return h.templatesEvalPage(c, periodIDFromForm(c))
}

func (h *EvaluationHandler) DeleteTemplate(c *fiber.Ctx) error {
	if err := core.RequirePermission(c, enums.PermissionEvaluationManage); err != nil {
		return err
	}

	id, _ := strconv.ParseInt(c.Params("id"), 10, 64)
	if err := evalsvc.TemplateServiceInstance.Delete(id); err != nil {
		return core.RenderError(c, fiber.StatusInternalServerError, "Failed to delete template.")
	}
	// A DELETE carries no body, so the cycle context is dropped here; the list
	// reopens without a "continue" target rather than failing.
	return h.TemplatesEval(c)
}
