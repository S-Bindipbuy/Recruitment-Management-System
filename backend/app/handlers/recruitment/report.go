package recruitment

import (
	recsvc "ems/app/services/recruitment"
	"github.com/gofiber/fiber/v2"
)

type ReportHandler struct{}

func (h *ReportHandler) CandidateSummary(c *fiber.Ctx) error {
	candidateID, _ := c.ParamsInt("candidate_id")
	summary, err := recsvc.InterviewServiceInst.GetCandidateSummary(int64(candidateID))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Summary not found"})
	}
	return c.JSON(summary)
}

func (h *ReportHandler) AllCandidateSummaries(c *fiber.Ctx) error {
	summaries, err := recsvc.InterviewServiceInst.GetAllCandidateSummaries()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch summaries"})
	}
	return c.JSON(summaries)
}

func (h *ReportHandler) InterviewerActivity(c *fiber.Ctx) error {
	activities, err := recsvc.InterviewServiceInst.GetInterviewerActivity()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch activity"})
	}
	return c.JSON(activities)
}

func (h *ReportHandler) HiringDecisions(c *fiber.Ctx) error {
	decisions, err := recsvc.InterviewServiceInst.GetHiringDecisions()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch decisions"})
	}
	return c.JSON(decisions)
}
