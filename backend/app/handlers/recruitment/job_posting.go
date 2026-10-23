package recruitment

import (
	recmodel "ems/app/models/recruitment"
	recsvc "ems/app/services/recruitment"
	"github.com/gofiber/fiber/v2"
)

type JobPostingHandler struct{}

func (h *JobPostingHandler) List(c *fiber.Ctx) error {
	postings, err := recsvc.JobPostingServiceInst.FindAll()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch job postings"})
	}
	return c.JSON(postings)
}

func (h *JobPostingHandler) Get(c *fiber.Ctx) error {
	id, _ := c.ParamsInt("id")
	posting, err := recsvc.JobPostingServiceInst.FindByID(int64(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Job posting not found"})
	}
	return c.JSON(posting)
}

func (h *JobPostingHandler) Create(c *fiber.Ctx) error {
	var req struct {
		PositionID     int64    `json:"position_id"`
		Title          string   `json:"title"`
		Description    string   `json:"description"`
		Requirements   string   `json:"requirements"`
		SalaryMin      *float64 `json:"salary_min"`
		SalaryMax      *float64 `json:"salary_max"`
		EmploymentType string   `json:"employment_type"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	userID := c.Locals("user_id").(int64)
	posting := &recmodel.JobPosting{
		PositionID:     req.PositionID,
		Title:          req.Title,
		Description:    req.Description,
		Requirements:   req.Requirements,
		SalaryMin:      req.SalaryMin,
		SalaryMax:      req.SalaryMax,
		EmploymentType: req.EmploymentType,
		Status:         "draft",
		CreatedBy:      userID,
	}

	if err := recsvc.JobPostingServiceInst.Create(posting); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create job posting"})
	}
	return c.Status(fiber.StatusCreated).JSON(posting)
}

func (h *JobPostingHandler) Update(c *fiber.Ctx) error {
	id, _ := c.ParamsInt("id")
	var req struct {
		Title          string   `json:"title"`
		Description    string   `json:"description"`
		Requirements   string   `json:"requirements"`
		SkillsRequired string   `json:"skills_required"`
		SalaryMin      *float64 `json:"salary_min"`
		SalaryMax      *float64 `json:"salary_max"`
		EmploymentType string   `json:"employment_type"`
		Status         string   `json:"status"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if err := recsvc.JobPostingServiceInst.Update(int64(id), req.Title, req.Description, req.Requirements, req.SkillsRequired,
		req.SalaryMin, req.SalaryMax, req.EmploymentType, req.Status); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update job posting"})
	}
	return c.JSON(fiber.Map{"message": "Updated successfully"})
}

func (h *JobPostingHandler) Delete(c *fiber.Ctx) error {
	id, _ := c.ParamsInt("id")
	if err := recsvc.JobPostingServiceInst.Delete(int64(id)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete job posting"})
	}
	return c.JSON(fiber.Map{"message": "Deleted successfully"})
}

func (h *JobPostingHandler) Publish(c *fiber.Ctx) error {
	id, _ := c.ParamsInt("id")
	if err := recsvc.JobPostingServiceInst.Publish(int64(id)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to publish job posting"})
	}
	return c.JSON(fiber.Map{"message": "Published successfully"})
}

func (h *JobPostingHandler) Close(c *fiber.Ctx) error {
	id, _ := c.ParamsInt("id")
	if err := recsvc.JobPostingServiceInst.Close(int64(id)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to close job posting"})
	}
	return c.JSON(fiber.Map{"message": "Closed successfully"})
}
