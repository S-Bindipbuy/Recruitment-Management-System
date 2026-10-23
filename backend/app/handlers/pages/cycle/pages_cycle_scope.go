package cycle

import (
	"ems/app/enums"
	core "ems/app/handlers/pages/core"
	"ems/app/services/cycles"
	evalsvc "ems/app/services/evaluation"
	orgsvc "ems/app/services/organization"

	"github.com/gofiber/fiber/v2"
)

// ── Step 3: scope ──

func (h *CycleHandler) showCycleStepScope(c *fiber.Ctx, data *cycles.CycleData) error {
	h.loadScope(c, data)
	return core.RenderPage(c, "page-cycle", data)
}

// loadScope fills the parts of step 3 that follow from the form: which
// candidates exist, which are ticked and what the preview will create.
func (h *CycleHandler) loadScope(c *fiber.Ctx, data *cycles.CycleData) {
	// Whatever was ticked survives the round trip; the rest of the step is
	// carried by the caller so a failed submission can be re-drawn typed.
	data.SelectedCandidateIDs = core.ReadScopeForm(c).CandidateIDs
	data.Workers = core.LoadActiveWorkers()
	// Candidates have no manager and no department, so they are always
	// assessed by one chosen evaluator.
	data.Candidates = cycles.LoadCycleCandidates(data.PeriodID, data.TemplateID, data.SelectedCandidateIDs)
	data.SelectedCandidateIDs = cycles.PruneCycleCandidateIDs(data.SelectedCandidateIDs, data.Candidates)
	data.AllSelected = len(data.Candidates) > 0 && len(data.SelectedCandidateIDs) == len(data.Candidates)
	data.RefreshPreview()
}

// cycleScopeContext rebuilds step 3 from the values the wizard carries in its
// form fields. Every endpoint that can only partly succeed re-renders through
// it, so the user corrects the one wrong field instead of restarting four
// steps.
func (h *CycleHandler) cycleScopeContext(c *fiber.Ctx) *cycles.CycleData {
	form := core.ReadScopeForm(c)
	data := &cycles.CycleData{
		Step:                 cycles.CycleStepScope,
		DueDate:              form.DueDate,
		EvaluatorID:          form.EvaluatorID,
		SelectedCandidateIDs: form.CandidateIDs,
	}
	if period, err := evalsvc.PeriodServiceInstance.FindByID(form.PeriodID); err == nil {
		data.PeriodID, data.PeriodName = period.ID, period.Name
	}
	if tmpl, err := evalsvc.TemplateServiceInstance.FindByID(form.TemplateID); err == nil {
		data.TemplateID, data.TemplateName = tmpl.ID, tmpl.Name
	}
	if data.EvaluatorID > 0 {
		data.EvaluatorName = core.UserName(data.EvaluatorID)
	}
	return data
}

func (h *CycleHandler) CycleChooseScope(c *fiber.Ctx) error {
	if err := core.RequirePermission(c, enums.PermissionEvaluationManage, enums.PermissionEvaluationAssign); err != nil {
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

	evaluatorID := form.EvaluatorID
	selected := form.CandidateIDs
	data := &cycles.CycleData{
		Step:                 cycles.CycleStepReview,
		PeriodID:             period.ID,
		PeriodName:           period.Name,
		TemplateID:           tmpl.ID,
		TemplateName:         tmpl.Name,
		DueDate:              form.DueDate,
		EvaluatorID:          evaluatorID,
		SelectedCandidateIDs: selected,
	}
	if evaluatorID > 0 {
		data.EvaluatorName = core.UserName(evaluatorID)
	}

	if len(selected) == 0 {
		data.Step = cycles.CycleStepScope
		data.ErrMessage = "Tick at least one candidate to evaluate."
		return h.showCycleStepScope(c, data)
	}
	if evaluatorID <= 0 {
		data.Step = cycles.CycleStepScope
		data.ErrMessage = "Choose who will evaluate these candidates."
		return h.showCycleStepScope(c, data)
	}
	if !orgsvc.IsActiveWorker(evaluatorID) {
		data.Step = cycles.CycleStepScope
		data.ErrMessage = "The chosen evaluator is not an active worker - pick somebody else."
		return h.showCycleStepScope(c, data)
	}

	data.RefreshPreview()
	if data.WillCreate == 0 {
		// Step 3 keeps everything typed so far and now shows the pairing
		// table: the reason each row was skipped travels as BlockedReason, so
		// it lands beside the Next button the user just pressed. Swapping only
		// the markup - without jumping to the top - keeps them next to it.
		data.Step = cycles.CycleStepScope
		c.Set("HX-Reswap", "innerHTML")
		return h.showCycleStepScope(c, data)
	}
	return core.RenderPage(c, "page-cycle", data)
}

// ── Step 3: taking a row back out ──

// CycleDropAssignment takes one person back out of the period from the step-3
// picker. An evaluation left behind by an earlier run used to lock a candidate
// out of the wizard for good, with the only way forward being another period.
// Pending and started rows only: anything already handed in stays.
func (h *CycleHandler) CycleDropAssignment(c *fiber.Ctx) error {
	if err := core.RequirePermission(c, enums.PermissionEvaluationManage, enums.PermissionEvaluationAssign); err != nil {
		return err
	}

	data := h.cycleScopeContext(c)
	fail := func(msg string) error {
		data.Step = cycles.CycleStepScope
		data.ErrMessage = msg
		return h.showCycleStepScope(c, data)
	}

	id := core.ReadScopeForm(c).DropAssignmentID
	if id <= 0 {
		return fail("That evaluation could not be found - refresh this step and try again.")
	}
	a, err := evalsvc.AssignmentServiceInstance.FindByID(id)
	if err != nil {
		return fail("That evaluation no longer exists.")
	}
	if a.CandidateID == nil || *a.CandidateID <= 0 {
		return fail("Only a candidate evaluation can be dropped from this list.")
	}
	// The picker only ever offers the row it is showing, so anything pointing
	// at another period or template is somebody else's business.
	if a.EvaluationPeriodID != data.PeriodID || a.EvaluationTemplateID != data.TemplateID {
		return fail("That evaluation belongs to a different period or template - it cannot be dropped from here.")
	}
	if !cycles.IsDroppableStatus(a.Status) {
		return fail("This evaluation is already " + a.Status + " - only a pending or started one can be dropped.")
	}
	if err := evalsvc.AssignmentServiceInstance.Delete(id); err != nil {
		return fail("Failed to drop that evaluation. Please try again.")
	}

	who := core.CandidateNames([]int64{*a.CandidateID})[*a.CandidateID]
	if who == "" {
		who = "the candidate"
	}
	if a.Status == string(enums.AssignmentStatusStarted) {
		core.TriggerToast(c, "Dropped "+who+"'s started evaluation - the answers saved so far are gone.", "success")
	} else {
		core.TriggerToast(c, "Dropped "+who+"'s pending evaluation from this period.", "success")
	}
	return h.showCycleStepScope(c, data)
}
