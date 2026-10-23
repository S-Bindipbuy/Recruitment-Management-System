package recruitment

import (
	recmodel "ems/app/models/recruitment"
	recsvc "ems/app/services/recruitment"
	"github.com/gofiber/fiber/v2"
	"time"
)

type InterviewHandler struct{}

func (h *InterviewHandler) List(c *fiber.Ctx) error {
	interviews, err := recsvc.InterviewServiceInst.FindAll()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch interviews"})
	}
	return c.JSON(interviews)
}

func (h *InterviewHandler) Get(c *fiber.Ctx) error {
	id, _ := c.ParamsInt("id")
	interview, err := recsvc.InterviewServiceInst.FindByID(int64(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Interview not found"})
	}
	return c.JSON(interview)
}

func (h *InterviewHandler) Create(c *fiber.Ctx) error {
	var req struct {
		CandidateID      int64  `json:"candidate_id"`
		InterviewRoundID int64  `json:"interview_round_id"`
		InterviewerID    int64  `json:"interviewer_id"`
		ScheduledAt      string `json:"scheduled_at"`
		DurationMinutes  int    `json:"duration_minutes"`
		Location         string `json:"location"`
		MeetingLink      string `json:"meeting_link"`
		Notes            string `json:"notes"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	scheduledAt, err := time.Parse("2006-01-02T15:04", req.ScheduledAt)
	if err != nil {
		scheduledAt, err = time.Parse("2006-01-02 15:04", req.ScheduledAt)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid date format. Use YYYY-MM-DD HH:MM"})
		}
	}

	if req.DurationMinutes <= 0 {
		req.DurationMinutes = 60
	}

	interview := &recmodel.Interview{
		CandidateID:      req.CandidateID,
		InterviewRoundID: req.InterviewRoundID,
		InterviewerID:    req.InterviewerID,
		ScheduledAt:      scheduledAt,
		DurationMinutes:  req.DurationMinutes,
		Location:         req.Location,
		MeetingLink:      req.MeetingLink,
		Status:           "scheduled",
		Notes:            req.Notes,
	}

	if err := recsvc.InterviewServiceInst.Create(interview); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create interview"})
	}
	return c.Status(fiber.StatusCreated).JSON(interview)
}

func (h *InterviewHandler) UpdateStatus(c *fiber.Ctx) error {
	id, _ := c.ParamsInt("id")
	var req struct {
		Status string `json:"status"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if err := recsvc.InterviewServiceInst.UpdateStatus(int64(id), req.Status); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update status"})
	}
	return c.JSON(fiber.Map{"message": "Status updated"})
}

func (h *InterviewHandler) Delete(c *fiber.Ctx) error {
	id, _ := c.ParamsInt("id")
	if err := recsvc.InterviewServiceInst.Delete(int64(id)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete interview"})
	}
	return c.JSON(fiber.Map{"message": "Deleted successfully"})
}

func (h *InterviewHandler) ListByCandidate(c *fiber.Ctx) error {
	candidateID, _ := c.ParamsInt("candidate_id")
	interviews, err := recsvc.InterviewServiceInst.FindByCandidateID(int64(candidateID))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch interviews"})
	}
	return c.JSON(interviews)
}
