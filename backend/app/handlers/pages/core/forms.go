package core

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

// ── Reading the request's forms ──
//
// Each struct names every field one step accepts, so a handler reads the form
// once instead of fetching values by string wherever it happens to need them,
// and a new field has one place to be added.

// ScopeForm is what steps 3 and 4 of the wizard post between them.
type ScopeForm struct {
	PeriodID         int64
	TemplateID       int64
	EvaluatorID      int64
	DueDate          string
	CandidateIDs     []int64
	DropAssignmentID int64
}

func ReadScopeForm(c *fiber.Ctx) ScopeForm {
	return ScopeForm{
		PeriodID:         FormInt64(c, "period_id"),
		TemplateID:       FormInt64(c, "template_id"),
		EvaluatorID:      FormInt64(c, "evaluator_id"),
		DueDate:          c.FormValue("due_date"),
		CandidateIDs:     FormInt64s(c, "candidate_ids"),
		DropAssignmentID: FormInt64(c, "drop_assignment_id"),
	}
}

// AssignmentForm is the New Assignment dialog on the Evaluations page (and the
// same three choices offered on a candidate's row).
type AssignmentForm struct {
	PeriodID    int64
	TemplateID  int64
	EvaluatorID int64
	CandidateID int64
	DueDate     string
}

func ReadAssignmentForm(c *fiber.Ctx) AssignmentForm {
	return AssignmentForm{
		PeriodID:    FormInt64(c, "evaluation_period_id"),
		TemplateID:  FormInt64(c, "evaluation_template_id"),
		EvaluatorID: FormInt64(c, "evaluator_id"),
		CandidateID: FormInt64(c, "candidate_id"),
		DueDate:     c.FormValue("due_date"),
	}
}

// FormInt64 reads one id-shaped field, tolerating empty and junk values: a
// half-filled form is a validation error, not a parse failure.
func FormInt64(c *fiber.Ctx, field string) int64 {
	v, _ := strconv.ParseInt(c.FormValue(field), 10, 64)
	return v
}

// FormInt64s collects every value of a repeated field. Checkboxes all share
// one name, so a single-value lookup would only ever see the first tick.
func FormInt64s(c *fiber.Ctx, field string) []int64 {
	raw := append(c.Context().QueryArgs().PeekMulti(field), c.Context().PostArgs().PeekMulti(field)...)
	ids := make([]int64, 0, len(raw))
	seen := make(map[int64]bool, len(raw))
	for _, r := range raw {
		v, err := strconv.ParseInt(string(r), 10, 64)
		if err != nil || v <= 0 || seen[v] {
			continue
		}
		seen[v] = true
		ids = append(ids, v)
	}
	return ids
}

// DateFromForm parses the date picker's value, or nil when it is empty or
// malformed - a due date is optional, and a bad one is simply no due date.
func DateFromForm(value string) *time.Time {
	if value == "" {
		return nil
	}
	if t, err := time.Parse("2006-01-02", value); err == nil {
		return &t
	}
	return nil
}
