package cycles

// ── Guided evaluation cycle: the model behind the wizard ──
//
// The wizard walks four steps and carries the choices forward as hidden form
// fields. What each step may do lives here rather than in the handler, so the
// rules can be tested without a request:
//
//	1. period   -> pick an existing period or create one inline
//	2. template -> pick an active template
//	3. scope    -> tick the individual candidates to evaluate and choose who
//	               assesses them
//	4. review   -> confirm and launch the whole batch

// LoadTemplateOptions fills in the step-2 choices and reports how many of them
// are actually usable. A template with no questions is offered but disabled, so
// if none are usable the step would otherwise be a dead end.
func (d *CycleData) LoadTemplateOptions() {
	d.Templates = CycleTemplateOptions()
	d.ReadyTemplates = 0
	for _, t := range d.Templates {
		if t.QuestionCount > 0 {
			d.ReadyTemplates++
		}
	}
	if d.ErrMessage != "" {
		return
	}
	if len(d.Templates) == 0 {
		d.ErrMessage = "There are no active templates yet. Create one on the Templates page first."
	} else if d.ReadyTemplates == 0 {
		d.ErrMessage = "None of these templates have questions yet, so there is nothing to choose. Add questions to one first."
	}
}

// RefreshPreview recomputes the pairing for the current choices. It is called
// on every render of step 3 so the table the user is looking at can never be
// older than the candidates/evaluator shown next to it.
func (d *CycleData) RefreshPreview() {
	d.Pairs = nil
	d.Skipped = 0
	d.WillCreate = 0
	d.BlockedReason = ""
	d.PreviewReady = false
	if d.PeriodID <= 0 || d.TemplateID <= 0 {
		return
	}
	if len(d.SelectedCandidateIDs) == 0 {
		return
	}
	d.PreviewReady = true
	d.Pairs, d.Skipped = PreviewCandidatePairs(d.PeriodID, d.TemplateID, d.EvaluatorID, d.SelectedCandidateIDs)
	d.WillCreate = len(d.Pairs) - d.Skipped
	d.BlockedReason = d.blockedReason()
}

// blockedReason is the sentence shown where "Next: review" would otherwise be.
// An empty result means nothing is chosen yet, which is a normal empty state
// rather than a blocked one.
func (d *CycleData) blockedReason() string {
	if d.WillCreate != 0 || len(d.Pairs) == 0 || d.ErrMessage != "" {
		return ""
	}
	return CycleSkipMessage(d.Pairs)
}
