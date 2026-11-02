package recruitment

import (
	recmodel "ems/app/models/recruitment"
	evalsvc "ems/app/services/evaluation"
	recsvc "ems/app/services/recruitment"
	"github.com/gofiber/fiber/v2"
)

type InterviewRoundHandler struct{}

func (h *InterviewRoundHandler) List(c *fiber.Ctx) error {
	postingID, _ := c.ParamsInt("posting_id")
	rounds, err := recsvc.InterviewRoundServiceInst.FindByJobPostingID(int64(postingID))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch interview rounds"})
	}
	return c.JSON(rounds)
}

func (h *InterviewRoundHandler) Create(c *fiber.Ctx) error {
	var req struct {
		JobPostingID         int64   `json:"job_posting_id"`
		EvaluationTemplateID int64   `json:"evaluation_template_id"`
		SectionIDs           []int64 `json:"section_ids"`
		Name                 string  `json:"name"`
		RoundNumber          int     `json:"round_number"`
		RoundType            string  `json:"round_type"`
		Description          string  `json:"description"`
		MaxScore             float64 `json:"max_score"`
		PassScore            float64 `json:"pass_score"`
		Weight               float64 `json:"weight"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// An omitted weight must not silently become 0, which would drop every
	// score in this round out of the candidate's weighted average.
	weight := req.Weight
	if weight <= 0 {
		weight = 1
	}

	// Keep only sections that belong to the chosen template; a section from a
	// different template would put questions on a round nobody selected.
	sectionIDs := make([]int64, 0, len(req.SectionIDs))
	if req.EvaluationTemplateID > 0 {
		sections, err := evalsvc.SectionServiceInstance.FindByTemplateID(req.EvaluationTemplateID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to load the evaluation template"})
		}
		allowed := make(map[int64]bool, len(sections))
		for _, section := range sections {
			allowed[section.ID] = true
		}
		for _, id := range req.SectionIDs {
			if allowed[id] {
				sectionIDs = append(sectionIDs, id)
			}
		}
		if len(sectionIDs) == 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Choose at least one section from the evaluation template"})
		}
	}

	round := &recmodel.InterviewRound{
		JobPostingID:         req.JobPostingID,
		EvaluationTemplateID: req.EvaluationTemplateID,
		Name:                 req.Name,
		RoundNumber:          req.RoundNumber,
		RoundType:            req.RoundType,
		Description:          req.Description,
		MaxScore:             req.MaxScore,
		PassScore:            req.PassScore,
		Weight:               weight,
		Status:               "active",
	}

	if err := recsvc.InterviewRoundServiceInst.Create(round); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create interview round"})
	}
	if len(sectionIDs) > 0 {
		if err := recsvc.InterviewRoundServiceInst.SaveSections(round.ID, sectionIDs); err != nil {
			_ = recsvc.InterviewRoundServiceInst.Delete(round.ID)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to save the round's questions"})
		}
	}
	return c.Status(fiber.StatusCreated).JSON(round)
}

func (h *InterviewRoundHandler) Delete(c *fiber.Ctx) error {
	id, _ := c.ParamsInt("id")
	if err := recsvc.InterviewRoundServiceInst.Delete(int64(id)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete interview round"})
	}
	return c.JSON(fiber.Map{"message": "Deleted successfully"})
}

// Interview Handlers
