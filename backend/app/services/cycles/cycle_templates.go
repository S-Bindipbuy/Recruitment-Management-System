package cycles

import (
	"ems/app/enums"
	evaluationsvc "ems/app/services/evaluation"
)

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

// CycleTemplateOptions lists every active template with its section and
// question counts, so step 2 can show which of them are actually usable.
func CycleTemplateOptions() []CycleTemplateOption {
	templates, _ := evaluationsvc.TemplateServiceInstance.FindAll()
	options := make([]CycleTemplateOption, 0, len(templates))
	for _, t := range templates {
		if t.Status != string(enums.TemplateStatusActive) {
			continue
		}
		sections, _ := evaluationsvc.SectionServiceInstance.FindByTemplateID(t.ID)
		options = append(options, CycleTemplateOption{
			EvaluationTemplate: t,
			SectionCount:       len(sections),
			QuestionCount:      TemplateQuestionCount(t.ID),
		})
	}
	return options
}

// TemplateQuestionCount is how many questions a template actually asks. A
// template with none would launch evaluations that can only ever score zero,
// so step 2 refuses it rather than letting that happen after the fact.
func TemplateQuestionCount(templateID int64) int {
	sections, _ := evaluationsvc.SectionServiceInstance.FindByTemplateID(templateID)
	total := 0
	for _, sec := range sections {
		questions, _ := evaluationsvc.QuestionServiceInstance.FindBySectionID(sec.ID)
		total += len(questions)
	}
	return total
}

// ── Step 3: who is in scope ──
