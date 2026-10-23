package public

import (
	"bytes"
	recmodel "ems/app/models/recruitment"
	recsvc "ems/app/services/recruitment"
	"ems/config"
	"fmt"
	"html/template"
	"log"
	"path/filepath"
	"sync"

	"github.com/gofiber/fiber/v2"
)

type PublicJobHandler struct{}

// ── Public template loading ──
//
// Each content page is parsed together with the layout it is wrapped in, into a
// set of its own. One broken page therefore takes down that page only; the
// layout is shared, so when it is the thing that broke, every public page says
// so instead of blaming its own file.

const publicLayoutFile = "public_layout.html"

var (
	publicTemplateSets  map[string]*template.Template // page name -> layout + page
	publicTemplateErrs  map[string]error              // file -> parse error
	publicTemplatesOnce sync.Once
)

func publicFuncs() template.FuncMap {
	return template.FuncMap{
		"derefFloat": func(f *float64) float64 {
			if f == nil {
				return 0
			}
			return *f
		},
	}
}

func publicTemplatesDirPath() string {
	return filepath.Join(config.AppConfig.FrontendPath, "templates")
}

func loadPublicTemplates() {
	loadPublicTemplatesFrom(publicTemplatesDirPath())
}

// loadPublicTemplatesFrom parses every content page in dir next to the layout.
// The directory is an argument so a test can hand it one working page and one
// broken one.
func loadPublicTemplatesFrom(dir string) {
	publicTemplateSets = map[string]*template.Template{}
	publicTemplateErrs = map[string]error{}

	layout := filepath.Join(dir, publicLayoutFile)
	if _, err := template.New(publicLayoutFile).Funcs(publicFuncs()).ParseFiles(layout); err != nil {
		publicTemplateErrs[publicLayoutFile] = err
		log.Printf("[templates] %s: %v", publicLayoutFile, err)
		return
	}

	pages, err := filepath.Glob(filepath.Join(dir, "content_*.html"))
	if err != nil {
		publicTemplateErrs["content_*.html"] = err
		return
	}
	for _, page := range pages {
		name := filepath.Base(page)
		set, err := template.New(publicLayoutFile).Funcs(publicFuncs()).ParseFiles(layout, page)
		if err != nil {
			publicTemplateErrs[name] = err
			log.Printf("[templates] %s: %v", name, err)
			continue
		}
		publicTemplateSets[name] = set
	}
}

// publicTemplate returns the set holding a page together with the file to
// blame when it cannot be rendered.
func publicTemplate(name string) (set *template.Template, owner string, err error) {
	publicTemplatesOnce.Do(loadPublicTemplates)
	return lookupPublicTemplate(name)
}

// lookupPublicTemplate reads the maps loadPublicTemplatesFrom filled in. It
// does no loading of its own, so a test can point it at a directory of its own.
func lookupPublicTemplate(name string) (set *template.Template, owner string, err error) {
	if err, ok := publicTemplateErrs[publicLayoutFile]; ok {
		return nil, publicLayoutFile, err
	}
	if set, ok := publicTemplateSets[name]; ok {
		return set, name, nil
	}
	if err, ok := publicTemplateErrs[name]; ok {
		return nil, name, err
	}
	return nil, "", fmt.Errorf("template %q is not defined", name)
}

func renderPublicPage(c *fiber.Ctx, contentTemplate string, data map[string]any) error {
	tpl, owner, err := publicTemplate(contentTemplate)
	if err != nil {
		if owner != "" {
			return fiber.NewError(fiber.StatusInternalServerError, contentTemplate+" comes from "+owner+" and could not be parsed: "+err.Error())
		}
		return fiber.NewError(fiber.StatusInternalServerError, "failed to render "+contentTemplate)
	}

	var contentBuf bytes.Buffer
	if err := tpl.ExecuteTemplate(&contentBuf, contentTemplate, data); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to render "+contentTemplate+": "+err.Error())
	}

	layoutData := map[string]any{
		"Title":   data["Title"],
		"Content": template.HTML(contentBuf.String()),
	}

	var layoutBuf bytes.Buffer
	if err := tpl.ExecuteTemplate(&layoutBuf, "public_layout.html", layoutData); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to render layout: "+err.Error())
	}

	c.Set("Content-Type", "text/html; charset=utf-8")
	return c.Send(layoutBuf.Bytes())
}

func (h *PublicJobHandler) PublicJobs(c *fiber.Ctx) error {
	postings, _ := recsvc.JobPostingServiceInst.FindAll()
	var published []recmodel.JobPostingListItem
	for _, p := range postings {
		if p.Status == "published" {
			published = append(published, p)
		}
	}
	return renderPublicPage(c, "content_jobs.html", map[string]any{
		"Title":    "Open Positions",
		"Postings": published,
	})
}

func (h *PublicJobHandler) PublicJobDetail(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil || id <= 0 {
		return c.Redirect("/jobs", fiber.StatusTemporaryRedirect)
	}
	posting, err := recsvc.JobPostingServiceInst.FindByID(int64(id))
	if err != nil || posting == nil {
		return c.Redirect("/jobs", fiber.StatusTemporaryRedirect)
	}
	if posting.Status != "published" {
		return c.Redirect("/jobs", fiber.StatusTemporaryRedirect)
	}
	allPostings, _ := recsvc.JobPostingServiceInst.FindAll()
	var listItem *recmodel.JobPostingListItem
	for _, p := range allPostings {
		if p.ID == posting.ID {
			listItem = &p
			break
		}
	}
	return renderPublicPage(c, "content_job_detail.html", map[string]any{
		"Title":   posting.Title,
		"Posting": listItem,
	})
}

func (h *PublicJobHandler) PublicApply(c *fiber.Ctx) error {
	postingID := int64(0)
	var posting *recmodel.JobPosting
	if id := c.Query("posting_id", ""); id != "" {
		fmt.Sscanf(id, "%d", &postingID)
		if postingID > 0 {
			posting, _ = recsvc.JobPostingServiceInst.FindByID(postingID)
		}
	}
	return renderPublicPage(c, "content_apply.html", map[string]any{
		"Title":     "Apply Now",
		"PostingID": postingID,
		"Posting":   posting,
	})
}
