package core

import (
	"bytes"
	candmodel "ems/app/models/candidate"
	"ems/config"
	"fmt"
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

const MaxCVSize = 10 * 1024 * 1024

var AllowedCVMimeTypes = map[string]bool{
	"application/pdf":    true,
	"application/msword": true,
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
}

// CreateCandidateFromForm validates and saves the uploaded document before the
// candidate record is created. Its caller removes the file if the DB write fails.
// It is shared by the /pages candidate form and the JSON candidate API.
func CreateCandidateFromForm(c *fiber.Ctx) (*candmodel.Candidate, error) {
	name := strings.TrimSpace(c.FormValue("full_name"))
	email := strings.TrimSpace(c.FormValue("email"))
	if name == "" || email == "" || !strings.Contains(email, "@") {
		return nil, fmt.Errorf("please provide a name and a valid email address")
	}
	// The name is handed back to staff in list views and toasts. Angle brackets
	// are never part of a name, and they are what a markup payload needs, so a
	// name carrying them is refused at the door instead of being stored.
	if strings.ContainsAny(name, "<>") {
		return nil, fmt.Errorf("the name may not contain the characters < or >")
	}

	var jobPostingID *int64
	if jpID := c.FormValue("job_posting_id"); jpID != "" {
		if id, err := strconv.ParseInt(jpID, 10, 64); err == nil && id > 0 {
			jobPostingID = &id
		}
	}

	header, err := c.FormFile("cv")
	if err != nil || header == nil {
		return nil, fmt.Errorf("please choose a CV file")
	}
	if header.Size == 0 || header.Size > MaxCVSize {
		return nil, fmt.Errorf("CV files must be between 1 byte and 10 MB")
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".pdf" && ext != ".doc" && ext != ".docx" {
		return nil, fmt.Errorf("CV must be a PDF, DOC, or DOCX file")
	}
	source, err := header.Open()
	if err != nil {
		return nil, fmt.Errorf("could not read the selected CV")
	}
	defer source.Close()
	peek := make([]byte, 512)
	n, _ := source.Read(peek)
	detectedType := http.DetectContentType(peek[:n])
	switch ext {
	case ".pdf":
		if detectedType != "application/pdf" {
			return nil, fmt.Errorf("the selected PDF file is invalid")
		}
	case ".doc":
		if !bytes.HasPrefix(peek[:n], []byte{0xd0, 0xcf, 0x11, 0xe0}) {
			return nil, fmt.Errorf("the selected DOC file is invalid")
		}
	case ".docx":
		if !bytes.HasPrefix(peek[:n], []byte{'P', 'K', 0x03, 0x04}) {
			return nil, fmt.Errorf("the selected DOCX file is invalid")
		}
	}
	// The Content-Type header on a multipart file part is optional; browsers
	// often omit it entirely. Treat a missing or unparseable value the same
	// as "application/octet-stream" and rely on the magic-byte checks above.
	declaredType, _, _ := mime.ParseMediaType(header.Header.Get("Content-Type"))
	if declaredType != "" && declaredType != "application/octet-stream" && !AllowedCVMimeTypes[declaredType] {
		return nil, fmt.Errorf("CV must be a PDF, DOC, or DOCX file")
	}

	uploadDir := config.AppConfig.UploadDir
	if err := os.MkdirAll(uploadDir, 0750); err != nil {
		return nil, fmt.Errorf("could not prepare CV storage")
	}
	storedPath := filepath.Join(uploadDir, uuid.NewString()+ext)
	destination, err := os.OpenFile(storedPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0640)
	if err != nil {
		return nil, fmt.Errorf("could not save the CV")
	}
	_, copyErr := io.Copy(destination, io.MultiReader(bytes.NewReader(peek[:n]), source))
	closeErr := destination.Close()
	if copyErr != nil || closeErr != nil {
		_ = os.Remove(storedPath)
		return nil, fmt.Errorf("could not save the CV")
	}

	mimeType := declaredType
	if mimeType == "" || mimeType == "application/octet-stream" {
		mimeType = mime.TypeByExtension(ext)
	}
	return &candmodel.Candidate{
		FullName: name, Email: email, Phone: strings.TrimSpace(c.FormValue("phone")),
		JobPostingID: jobPostingID, Skills: strings.TrimSpace(c.FormValue("skills")),
		// CVName stores only the extension (e.g. ".pdf") — never the user-supplied
		// filename — so the download filename is built from the candidate ID and
		// cannot carry a path-traversal or header-injection payload.
		CVPath: storedPath, CVName: ext, CVMimeType: mimeType,
		CVSize: header.Size, Status: "new", CreatedAt: time.Now(),
	}, nil
}
