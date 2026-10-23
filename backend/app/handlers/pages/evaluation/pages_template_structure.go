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

// ── Template structure editor ──
//
// The cycle wizard refuses a template with no questions, because such a
// template would score every evaluation 0%. The API has always had section and
// question CRUD, but nothing in the UI exposed it, so templates stayed empty
// with no way forward. This page is that missing editor.

// QuestionTypeOption describes one selectable question type. Only the types the
// evaluation-fill page actually renders are offered; offering a type the form
// cannot collect an answer for would silently produce a broken evaluation.
type QuestionTypeOption struct {
	Value string
	Label string
	Hint  string
}

var questionTypeOptions = []QuestionTypeOption{
	{Value: "rating", Label: "Rating (0 - 10)", Hint: "The only type that produces a score."},
	{Value: "text", Label: "Long text", Hint: "Free comment, not scored."},
	{Value: "number", Label: "Number", Hint: "Not scored."},
	{Value: "boolean", Label: "Yes / No", Hint: "Not scored."},
}

func isSupportedQuestionType(v string) bool {
	for _, o := range questionTypeOptions {
		if o.Value == v {
			return true
		}
	}
	return false
}

type SectionWithQuestions struct {
	Section   evalmodel.EvaluationSection
	Questions []evalmodel.EvaluationQuestion
	// Position is the 1-based display number.
	Position int
}

type TemplateDetailData struct {
	Template      evalmodel.EvaluationTemplate
	Sections      []SectionWithQuestions
	QuestionCount int
	QuestionTypes []QuestionTypeOption
	Ready         bool
	ErrMessage    string
	Notice        string
	// PeriodID is the cycle the visitor came from, if any. It lets the editor
	// offer a way back to the wizard step instead of making them start over.
	PeriodID int64
}

// periodIDFromForm carries the wizard's period across every mutation so that
// landing back on the editor does not drop the user out of their cycle.
func periodIDFromForm(c *fiber.Ctx) int64 {
	periodID := core.FormInt64(c, "period_id")
	return periodID
}

func (h *EvaluationHandler) TemplateDetail(c *fiber.Ctx) error {
	id, _ := strconv.ParseInt(c.Params("id"), 10, 64)
	periodID, _ := strconv.ParseInt(c.Query("period_id"), 10, 64)
	return h.renderTemplateDetail(c, id, periodID, "", "")
}

// renderTemplateDetail rebuilds the whole editor after a mutation so the counts
// and the "ready to use" state are always in sync with the database.
func (h *EvaluationHandler) renderTemplateDetail(c *fiber.Ctx, id, periodID int64, notice, errMessage string) error {
	tmpl, err := evalsvc.TemplateServiceInstance.FindByID(id)
	if err != nil {
		return core.RenderError(c, fiber.StatusNotFound, "That template no longer exists.")
	}

	data := TemplateDetailData{
		Template:      *tmpl,
		QuestionTypes: questionTypeOptions,
		Notice:        notice,
		ErrMessage:    errMessage,
		PeriodID:      periodID,
	}

	sections, _ := evalsvc.SectionServiceInstance.FindByTemplateID(tmpl.ID)
	for i, sec := range sections {
		entry := SectionWithQuestions{Section: sec, Position: i + 1}
		entry.Questions, _ = evalsvc.QuestionServiceInstance.FindBySectionID(sec.ID)
		data.QuestionCount += len(entry.Questions)
		data.Sections = append(data.Sections, entry)
	}
	data.Ready = data.QuestionCount > 0

	return core.RenderPage(c, "page-template-detail", data)
}

func (h *EvaluationHandler) CreateSectionPage(c *fiber.Ctx) error {
	if err := core.RequirePermission(c, enums.PermissionEvaluationManage); err != nil {
		return err
	}

	templateID, _ := strconv.ParseInt(c.Params("id"), 10, 64)
	periodID := periodIDFromForm(c)
	name := strings.TrimSpace(c.FormValue("name"))
	if name == "" {
		return h.renderTemplateDetail(c, templateID, periodID, "", "Please give the section a name.")
	}

	// Sort after the existing sections so ordering stays predictable without
	// asking the user to number them.
	existing, _ := evalsvc.SectionServiceInstance.FindByTemplateID(templateID)
	if _, err := evalsvc.SectionServiceInstance.Create(
		templateID, name, strings.TrimSpace(c.FormValue("description")), len(existing),
	); err != nil {
		return h.renderTemplateDetail(c, templateID, periodID, "", "Failed to add the section.")
	}
	return h.renderTemplateDetail(c, templateID, periodID, "Section added. Now give it a question.", "")
}

func (h *EvaluationHandler) CreateQuestionPage(c *fiber.Ctx) error {
	if err := core.RequirePermission(c, enums.PermissionEvaluationManage); err != nil {
		return err
	}

	sectionID, _ := strconv.ParseInt(c.Params("id"), 10, 64)
	periodID := periodIDFromForm(c)
	section, err := evalsvc.SectionServiceInstance.FindByID(sectionID)
	if err != nil {
		return core.RenderError(c, fiber.StatusNotFound, "That section no longer exists.")
	}

	question := strings.TrimSpace(c.FormValue("question"))
	if question == "" {
		return h.renderTemplateDetail(c, section.EvaluationTemplateID, periodID, "", "Please write the question.")
	}

	questionType := c.FormValue("question_type")
	if questionType == "" {
		questionType = "rating"
	}
	if !isSupportedQuestionType(questionType) {
		return h.renderTemplateDetail(c, section.EvaluationTemplateID, periodID, "", "That question type is not supported.")
	}

	// Weight only matters for rating questions, where it multiplies their value
	// in the weighted average. Default to 1 so a blank field is harmless.
	weight := 1.0
	if raw := strings.TrimSpace(c.FormValue("weight")); raw != "" {
		if w, err := strconv.ParseFloat(raw, 64); err == nil && w > 0 {
			weight = w
		}
	}

	existing, _ := evalsvc.QuestionServiceInstance.FindBySectionID(sectionID)
	if _, err := evalsvc.QuestionServiceInstance.Create(
		sectionID, question, strings.TrimSpace(c.FormValue("description")), questionType,
		c.FormValue("is_required") != "", weight, len(existing),
	); err != nil {
		return h.renderTemplateDetail(c, section.EvaluationTemplateID, periodID, "", "Failed to add the question.")
	}
	return h.renderTemplateDetail(c, section.EvaluationTemplateID, periodID, "Question added.", "")
}

func (h *EvaluationHandler) DeleteSectionPage(c *fiber.Ctx) error {
	if err := core.RequirePermission(c, enums.PermissionEvaluationManage); err != nil {
		return err
	}

	sectionID, _ := strconv.ParseInt(c.Params("id"), 10, 64)
	periodID, _ := strconv.ParseInt(c.Query("period_id"), 10, 64)
	section, err := evalsvc.SectionServiceInstance.FindByID(sectionID)
	if err != nil {
		return core.RenderError(c, fiber.StatusNotFound, "That section no longer exists.")
	}
	if err := evalsvc.SectionServiceInstance.Delete(sectionID); err != nil {
		return h.renderTemplateDetail(c, section.EvaluationTemplateID, periodID, "", "Failed to delete the section.")
	}
	// The section's questions go with it (ON DELETE CASCADE).
	return h.renderTemplateDetail(c, section.EvaluationTemplateID, periodID, "Section and its questions deleted.", "")
}

func (h *EvaluationHandler) DeleteQuestionPage(c *fiber.Ctx) error {
	if err := core.RequirePermission(c, enums.PermissionEvaluationManage); err != nil {
		return err
	}

	questionID, _ := strconv.ParseInt(c.Params("id"), 10, 64)
	periodID, _ := strconv.ParseInt(c.Query("period_id"), 10, 64)
	question, err := evalsvc.QuestionServiceInstance.FindByID(questionID)
	if err != nil {
		return core.RenderError(c, fiber.StatusNotFound, "That question no longer exists.")
	}

	section, err := evalsvc.SectionServiceInstance.FindByID(question.EvaluationSectionID)
	if err != nil {
		return core.RenderError(c, fiber.StatusNotFound, "That section no longer exists.")
	}
	if err := evalsvc.QuestionServiceInstance.Delete(questionID); err != nil {
		return h.renderTemplateDetail(c, section.EvaluationTemplateID, periodID, "", "Failed to delete the question.")
	}
	return h.renderTemplateDetail(c, section.EvaluationTemplateID, periodID, "Question deleted.", "")
}
