package recruitment

import (
	recmodel "ems/app/models/recruitment"
	recsvc "ems/app/services/recruitment"
	"github.com/gofiber/fiber/v2"
	"time"
)

type InterviewFeedbackHandler struct{}

func (h *InterviewFeedbackHandler) Get(c *fiber.Ctx) error {
	interviewID, _ := c.ParamsInt("interview_id")
	fb, err := recsvc.InterviewFeedbackServiceInst.FindByInterviewID(int64(interviewID))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Feedback not found"})
	}
	return c.JSON(fb)
}

func (h *InterviewFeedbackHandler) Create(c *fiber.Ctx) error {
	var req struct {
		InterviewID    int64    `json:"interview_id"`
		Score          *float64 `json:"score"`
		Recommendation string   `json:"recommendation"`
		Strengths      string   `json:"strengths"`
		Weaknesses     string   `json:"weaknesses"`
		Comments       string   `json:"comments"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	userID := c.Locals("user_id").(int64)
	now := time.Now()
	fb := &recmodel.InterviewFeedback{
		InterviewID:    req.InterviewID,
		InterviewerID:  userID,
		Score:          req.Score,
		Recommendation: req.Recommendation,
		Strengths:      req.Strengths,
		Weaknesses:     req.Weaknesses,
		Comments:       req.Comments,
		SubmittedAt:    &now,
	}

	if err := recsvc.InterviewFeedbackServiceInst.Create(fb); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to save feedback"})
	}
	return c.Status(fiber.StatusCreated).JSON(fb)
}

func (h *InterviewFeedbackHandler) Update(c *fiber.Ctx) error {
	id, _ := c.ParamsInt("id")
	var req struct {
		Score          *float64 `json:"score"`
		Recommendation string   `json:"recommendation"`
		Strengths      string   `json:"strengths"`
		Weaknesses     string   `json:"weaknesses"`
		Comments       string   `json:"comments"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	fb, err := recsvc.InterviewFeedbackServiceInst.FindByInterviewID(int64(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Feedback not found"})
	}

	fb.Score = req.Score
	fb.Recommendation = req.Recommendation
	fb.Strengths = req.Strengths
	fb.Weaknesses = req.Weaknesses
	fb.Comments = req.Comments

	if err := recsvc.InterviewFeedbackServiceInst.Update(fb); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update feedback"})
	}
	return c.JSON(fb)
}
