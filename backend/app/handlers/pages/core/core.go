package core

import (
	"bytes"
	usermodel "ems/app/models/user"
	"ems/app/repositories"
	candsvc "ems/app/services/candidate"
	orgsvc "ems/app/services/organization"
	usersvc "ems/app/services/user"
	"ems/config"
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"sync"
)

// ── Template loading ──
//
// Each file under templates/pages is parsed into a set of its own. Sharing one
// parsed set meant a single syntax error anywhere in that directory made every
// route that renders a page answer 500; now the damage stays inside the file
// that caused it, and the error names that file.

var (
	pageTemplateSets   map[string]*template.Template // template name -> the file's set
	pageTemplateOwners map[string]string             // template name -> file it is defined in
	pageTemplateErrs   map[string]error              // file -> parse error
	pageTemplatesDir   string                        // directory the maps above came from
	pageTemplatesOnce  sync.Once
)

func pagesTemplateDir() string {
	return filepath.Join(config.AppConfig.FrontendPath, "templates", "pages")
}

func pageFuncs() template.FuncMap {
	return template.FuncMap{
		"dict": func(kv ...any) map[string]any {
			m := make(map[string]any, len(kv)/2)
			for i := 0; i+1 < len(kv); i += 2 {
				m[kv[i].(string)] = kv[i+1]
			}
			return m
		},
		"derefFloat": func(f *float64) float64 {
			if f == nil {
				return 0
			}
			return *f
		},
	}
}

func loadPageTemplates() {
	loadPageTemplatesFrom(pagesTemplateDir())
}

// loadPageTemplatesFrom parses every .html file in dir on its own. It takes the
// directory as an argument so a test can hand it a directory holding one broken
// file next to one working one.
func loadPageTemplatesFrom(dir string) {
	pageTemplateSets = map[string]*template.Template{}
	pageTemplateOwners = map[string]string{}
	pageTemplateErrs = map[string]error{}
	pageTemplatesDir = dir

	matches, err := filepath.Glob(filepath.Join(dir, "*.html"))
	if err != nil {
		pageTemplateErrs["*.html"] = err
		return
	}
	for _, match := range matches {
		base := filepath.Base(match)
		set, err := template.New(base).Funcs(pageFuncs()).ParseFiles(match)
		if err != nil {
			pageTemplateErrs[base] = err
			log.Printf("[templates] %s: %v", base, err)
			continue
		}
		for _, t := range set.Templates() {
			name := t.Name()
			if name == "" || name == base {
				continue // the file itself, not something it defines
			}
			pageTemplateSets[name] = set
			pageTemplateOwners[name] = base
		}
	}
}

// pageTemplate returns the set holding a template name together with the file
// it was parsed from, so a broken file can be named in the error instead of
// taking every page down with a bare 500.
func pageTemplate(name string) (set *template.Template, owner string, err error) {
	pageTemplatesOnce.Do(loadPageTemplates)
	return lookupPageTemplate(name)
}

// lookupPageTemplate reads the maps loadPageTemplatesFrom filled in. It does
// no loading of its own, so a test can point it at a directory of its own.
func lookupPageTemplate(name string) (set *template.Template, owner string, err error) {
	if set, ok := pageTemplateSets[name]; ok {
		return set, pageTemplateOwners[name], nil
	}
	// Either the file that defines it failed to parse, or nothing defines it.
	for base, parseErr := range pageTemplateErrs {
		if templateDefines(filepath.Join(pageTemplatesDir, base), name) {
			return nil, base, parseErr
		}
	}
	return nil, "", fmt.Errorf("template %q is not defined", name)
}

// templateDefines reports whether a file writes {{define "name"}}. It is only
// consulted when a name cannot be found in any parsed set.
func templateDefines(path, name string) bool {
	body, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	return bytes.Contains(body, []byte(`{{define "`+name+`"`))
}

// RenderTemplate renders one named template to the response, for handlers that
// answer with a fragment rather than a whole page.
func RenderTemplate(c *fiber.Ctx, name string, data any) error {
	tpl, owner, err := pageTemplate(name)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "template error in "+owner)
	}
	var buf bytes.Buffer
	if err := tpl.ExecuteTemplate(&buf, name, data); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to render "+name)
	}
	c.Set("Content-Type", "text/html; charset=utf-8")
	return c.Send(buf.Bytes())
}

func RenderPage(c *fiber.Ctx, name string, data any) error {
	tpl, owner, err := pageTemplate(name)
	if err != nil {
		if owner != "" {
			return fiber.NewError(fiber.StatusInternalServerError, name+" comes from "+owner+" and could not be parsed: "+err.Error())
		}
		return fiber.NewError(fiber.StatusInternalServerError, "failed to render "+name)
	}
	var buf bytes.Buffer
	if err := tpl.ExecuteTemplate(&buf, name, data); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to render "+name)
	}
	c.Set("Content-Type", "text/html; charset=utf-8")
	return c.Send(buf.Bytes())
}

// RenderError aborts a page request with a status and a message. Like every
// other error it leaves the server as JSON; the frontend decides how to show it.
func RenderError(c *fiber.Ctx, status int, msg string) error {
	return c.Status(status).JSON(fiber.Map{"error": msg})
}

// TriggerToast makes the front end raise a toast without navigating away.
// Without this, handlers that return HTTP 200 alongside an error message leave
// the user looking at an unchanged page with no idea what went wrong.
func TriggerToast(c *fiber.Ctx, message, kind string) {
	payload, err := json.Marshal(map[string]any{
		"showToast": map[string]string{"message": message, "type": kind},
	})
	if err != nil {
		return
	}
	c.Set("X-EMS-Trigger", string(payload))
}

// ── Worker / user lookups shared across page handlers ──
//
// The queries live in the repositories and the services; what stays here is
// the vocabulary the templates are written against.

// LoadActiveWorkers lists everybody who can be asked to do an evaluation.
func LoadActiveWorkers() []usermodel.WorkerOption {
	workers, err := orgsvc.EmployeeServiceInstance.FindActiveWorkers()
	if err != nil {
		return nil
	}
	return workers
}

// LoadWorkerCandidates lists active accounts with no employee record yet.
func LoadWorkerCandidates() []usermodel.UserOption {
	users, err := usersvc.UserServiceInstance.FindActiveNonEmployees()
	if err != nil {
		return nil
	}
	return users
}

func UserNames(ids []int64) map[int64]string {
	names, err := usersvc.UserServiceInstance.FindNames(ids)
	if err != nil {
		return map[int64]string{}
	}
	return names
}

func CandidateNames(ids []int64) map[int64]string {
	names, err := candsvc.CandidateServiceInstance.FindNames(ids)
	if err != nil {
		return map[int64]string{}
	}
	return names
}

func UserName(id int64) string {
	name, _ := usersvc.UserServiceInstance.UserName(id)
	if name == "" {
		return fmt.Sprintf("User #%d", id)
	}
	return name
}

func DisplayName(names map[int64]string, userID int64) string {
	if name := names[userID]; name != "" {
		return name
	}
	return fmt.Sprintf("User #%d", userID)
}

func ActiveRecordExists(table string, id int64) bool {
	ok, err := repositories.ActiveRecordExists(table, id)
	return err == nil && ok
}
