package candidate

import (
	"ems/app/enums"
	"ems/app/handlers/authz"
	pages "ems/app/handlers/pages/core"
	candsvc "ems/app/services/candidate"
	"ems/config"
	"fmt"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type CandidateHandler struct{}

func (h *CandidateHandler) List(c *fiber.Ctx) error {
	candidates, err := candsvc.CandidateServiceInstance.FindAll()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch candidates"})
	}
	return c.JSON(candidates)
}

func (h *CandidateHandler) Create(c *fiber.Ctx) error {
	candidate, err := pages.CreateCandidateFromForm(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	if err := candsvc.CandidateServiceInstance.Create(candidate); err != nil {
		_ = os.Remove(candidate.CVPath)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to save candidate"})
	}
	return c.Status(fiber.StatusCreated).JSON(candidate)
}

func (h *CandidateHandler) DownloadCV(c *fiber.Ctx) error {
	if err := authz.RequirePermission(c, enums.PermissionRecruitmentManage); err != nil {
		return err
	}
	id, _ := c.ParamsInt("id")
	candidate, err := candsvc.CandidateServiceInstance.FindByID(int64(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Candidate not found"})
	}
	filePath := candidate.CVPath
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		filePath = filepath.Join(config.AppConfig.UploadDir, filepath.Base(candidate.CVPath))
	}
	// Prevent path traversal outside upload dir
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "CV file not found"})
	}
	absUpload, err := filepath.Abs(config.AppConfig.UploadDir)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "CV file not found"})
	}
	if filepath.Dir(absPath) != absUpload {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Access denied"})
	}
	data, err := os.ReadFile(absPath)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "CV file not found"})
	}

	downloadName := fmt.Sprintf("cv-%d%s", candidate.ID, filepath.Ext(candidate.CVName))
	if candidate.CVName != "" && strings.HasPrefix(candidate.CVName, ".") {
		downloadName = fmt.Sprintf("cv-%d%s", candidate.ID, candidate.CVName)
	} else if candidate.CVName == "" {
		downloadName = fmt.Sprintf("cv-%d", candidate.ID)
	}
	c.Type(candidate.CVMimeType)
	c.Set(fiber.HeaderContentDisposition, mime.FormatMediaType("attachment", map[string]string{"filename": downloadName}))
	return c.Send(data)
}

// createCandidateFromForm validates and saves the uploaded document before the
// candidate record is created. Its caller removes the file if the DB write fails.
