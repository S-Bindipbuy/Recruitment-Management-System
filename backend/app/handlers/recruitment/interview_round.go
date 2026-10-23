package recruitment

import (
	recmodel "ems/app/models/recruitment"
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
		JobPostingID int64   `json:"job_posting_id"`
		Name         string  `json:"name"`
		RoundNumber  int     `json:"round_number"`
		RoundType    string  `json:"round_type"`
		Description  string  `json:"description"`
		MaxScore     float64 `json:"max_score"`
		PassScore    float64 `json:"pass_score"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	round := &recmodel.InterviewRound{
		JobPostingID: req.JobPostingID,
		Name:         req.Name,
		RoundNumber:  req.RoundNumber,
		RoundType:    req.RoundType,
		Description:  req.Description,
		MaxScore:     req.MaxScore,
		PassScore:    req.PassScore,
		Status:       "active",
	}

	if err := recsvc.InterviewRoundServiceInst.Create(round); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create interview round"})
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
