package public

import (
	"bytes"
	pages "ems/app/handlers/pages/core"
	candmodel "ems/app/models/candidate"
	candsvc "ems/app/services/candidate"
	recsvc "ems/app/services/recruitment"
	"ems/config"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type PublicHandler struct{}

func (h *PublicHandler) SubmitCV(c *fiber.Ctx) error {
	name := strings.TrimSpace(c.FormValue("full_name"))
	email := strings.TrimSpace(c.FormValue("email"))
	if name == "" || email == "" || !strings.Contains(email, "@") {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "please provide a name and a valid email address"})
	}

	var jobPostingID *int64
	if jpID := c.FormValue("job_posting_id"); jpID != "" {
		if id, err := strconv.ParseInt(jpID, 10, 64); err == nil && id > 0 {
			jobPostingID = &id
			posting, err := recsvc.JobPostingServiceInst.FindByID(id)
			if err != nil || posting.Status != "published" {
				jobPostingID = nil
			}
		}
	}

	header, err := c.FormFile("cv")
	if err != nil || header == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "please choose a CV file"})
	}
	if header.Size == 0 || header.Size > pages.MaxCVSize {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "CV files must be between 1 byte and 10 MB"})
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".pdf" && ext != ".doc" && ext != ".docx" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "CV must be a PDF, DOC, or DOCX file"})
	}
	source, err := header.Open()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "could not read the selected CV"})
	}
	defer source.Close()
	peek := make([]byte, 512)
	n, _ := source.Read(peek)
	detectedType := http.DetectContentType(peek[:n])
	switch ext {
	case ".pdf":
		if detectedType != "application/pdf" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "the selected PDF file is invalid"})
		}
	case ".doc":
		if !bytes.HasPrefix(peek[:n], []byte{0xd0, 0xcf, 0x11, 0xe0}) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "the selected DOC file is invalid"})
		}
	case ".docx":
		if !bytes.HasPrefix(peek[:n], []byte{'P', 'K', 0x03, 0x04}) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "the selected DOCX file is invalid"})
		}
	}
	declaredType, _, _ := mime.ParseMediaType(header.Header.Get("Content-Type"))
	if declaredType != "" && declaredType != "application/octet-stream" && !pages.AllowedCVMimeTypes[declaredType] {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "CV must be a PDF, DOC, or DOCX file"})
	}

	uploadDir := config.AppConfig.UploadDir
	if err := os.MkdirAll(uploadDir, 0750); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not prepare CV storage"})
	}
	storedPath := filepath.Join(uploadDir, uuid.NewString()+ext)
	destination, err := os.OpenFile(storedPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0640)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not save the CV"})
	}
	_, copyErr := io.Copy(destination, io.MultiReader(bytes.NewReader(peek[:n]), source))
	closeErr := destination.Close()
	if copyErr != nil || closeErr != nil {
		_ = os.Remove(storedPath)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not save the CV"})
	}

	mimeType := declaredType
	if mimeType == "" || mimeType == "application/octet-stream" {
		mimeType = mime.TypeByExtension(ext)
	}

	candidate := &candmodel.Candidate{
		FullName: name, Email: email, Phone: strings.TrimSpace(c.FormValue("phone")),
		JobPostingID: jobPostingID, Skills: strings.TrimSpace(c.FormValue("skills")),
		CVPath: storedPath, CVName: ext, CVMimeType: mimeType,
		CVSize: header.Size, Status: "new", CreatedAt: time.Now(),
	}
	if err := candsvc.CandidateServiceInstance.Create(candidate); err != nil {
		_ = os.Remove(storedPath)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to save application"})
	}
	return c.JSON(fiber.Map{"message": "application submitted successfully"})
}
