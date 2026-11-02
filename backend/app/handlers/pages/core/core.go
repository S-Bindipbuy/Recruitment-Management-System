package core

import (
	"bytes"
	"ems/app/enums"
	recmodel "ems/app/models/recruitment"
	usermodel "ems/app/models/user"
	"ems/app/repositories"
	candsvc "ems/app/services/candidate"
	evalsvc "ems/app/services/evaluation"
	orgsvc "ems/app/services/organization"
	recsvc "ems/app/services/recruitment"
	usersvc "ems/app/services/user"
	"ems/config"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
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
		"derefBool": func(b *bool) bool {
			if b == nil {
				return false
			}
			return *b
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

// LoadInterviewers lists active workers flagged as eligible interviewers.
func LoadInterviewers() []usermodel.WorkerOption {
	workers, err := orgsvc.EmployeeServiceInstance.FindActiveInterviewers()
	if err != nil {
		return nil
	}
	return workers
}

// PipelineCandidates drops applicants who have already been hired. Hiring
// removes a candidate from the evaluation/ranking pipeline; they remain in the
// hiring-decision history instead.
func PipelineCandidates(summaries []recmodel.CandidateEvaluationSummary) []recmodel.CandidateEvaluationSummary {
	pipeline := make([]recmodel.CandidateEvaluationSummary, 0, len(summaries))
	for _, summary := range summaries {
		if !strings.EqualFold(summary.CandidateStatus, "hired") {
			pipeline = append(pipeline, summary)
		}
	}
	return pipeline
}

// candidateRanked reports whether a candidate belongs in a ranking table. The
// ranking is for people still being evaluated: a hired applicant has left the
// pipeline and must not occupy a place (or hide the rank of) the remaining
// candidates. Candidates without a score are unranked as well.
func candidateRanked(s recmodel.CandidateEvaluationSummary) bool {
	return s.CompletedReviews > 0 && !strings.EqualFold(s.CandidateStatus, "hired")
}

// AssignCandidateRanks fills Rank for every candidate that has at least one
// interviewer score. Candidates are ranked inside their own job posting, so two
// roles never compete for the same place. Candidates without a score keep rank
// zero and stay unranked.
func AssignCandidateRanks(summaries []recmodel.CandidateEvaluationSummary) []recmodel.CandidateEvaluationSummary {
	scored := make([]recmodel.CandidateEvaluationSummary, 0, len(summaries))
	for _, summary := range summaries {
		if candidateRanked(summary) {
			scored = append(scored, summary)
		}
	}

	sort.SliceStable(scored, func(i, j int) bool {
		if scored[i].JobPostingID != scored[j].JobPostingID {
			if scored[i].PostingTitle != scored[j].PostingTitle {
				return scored[i].PostingTitle < scored[j].PostingTitle
			}
			return scored[i].JobPostingID < scored[j].JobPostingID
		}
		if scored[i].AverageScore != scored[j].AverageScore {
			return scored[i].AverageScore > scored[j].AverageScore
		}
		return scored[i].CandidateName < scored[j].CandidateName
	})

	ranks := make(map[int64]int, len(scored))
	counts := make(map[int64]int, len(scored))
	for _, summary := range scored {
		counts[summary.JobPostingID]++
		ranks[summary.CandidateID] = counts[summary.JobPostingID]
	}
	for i := range summaries {
		summaries[i].Rank = ranks[summaries[i].CandidateID]
	}
	return summaries
}

// CandidateRankings returns only the candidates that have at least one score
// and have not been hired, sorted by job posting and then by their rank. It is
// the same ordering the Interviews page uses, so every page tells the same
// story: the ranking is the shortlist still waiting for a decision.
func CandidateRankings(summaries []recmodel.CandidateEvaluationSummary) []recmodel.CandidateEvaluationSummary {
	scored := make([]recmodel.CandidateEvaluationSummary, 0, len(summaries))
	for _, summary := range summaries {
		if candidateRanked(summary) {
			scored = append(scored, summary)
		}
	}
	scored = AssignCandidateRanks(scored)
	sort.SliceStable(scored, func(i, j int) bool {
		if scored[i].PostingTitle != scored[j].PostingTitle {
			return scored[i].PostingTitle < scored[j].PostingTitle
		}
		if scored[i].JobPostingID != scored[j].JobPostingID {
			return scored[i].JobPostingID < scored[j].JobPostingID
		}
		if scored[i].Rank != scored[j].Rank {
			return scored[i].Rank < scored[j].Rank
		}
		return scored[i].CandidateName < scored[j].CandidateName
	})
	return scored
}

// LoadCandidateRankings loads every candidate and returns the ranked ones,
// ordered by job posting and combined interviewer score.
func LoadCandidateRankings() []recmodel.CandidateEvaluationSummary {
	summaries, err := recsvc.InterviewServiceInst.GetAllCandidateSummaries()
	if err != nil {
		return nil
	}
	return CandidateRankings(summaries)
}

// LoadCandidateSummaryMap returns per-candidate interview progress and ranking
// keyed by candidate id, for pages that show a candidate list.
func LoadCandidateSummaryMap() map[int64]*recmodel.CandidateEvaluationSummary {
	summaries, err := recsvc.InterviewServiceInst.GetAllCandidateSummaries()
	if err != nil {
		return nil
	}
	summaries = AssignCandidateRanks(summaries)
	byCandidate := make(map[int64]*recmodel.CandidateEvaluationSummary, len(summaries))
	for i := range summaries {
		byCandidate[summaries[i].CandidateID] = &summaries[i]
	}
	return byCandidate
}

// InterviewReady reports whether a candidate has a review for every active
// round of their posting (or at least one review when no rounds are configured).
func InterviewReady(s *recmodel.CandidateEvaluationSummary) bool {
	return recsvc.InterviewSummaryReady(s)
}

// AttachCandidateEvaluation fills a candidate row with where their evaluation
// stands. A submitted evaluation is not finished: it is waiting for a named
// approver, so the row can say who still has to act instead of treating the
// candidate as evaluated.
func AttachCandidateEvaluation(cwe *CandidateWithEvaluation) {
	assignments, err := evalsvc.AssignmentServiceInstance.FindByCandidate(cwe.ID)
	if err != nil || len(assignments) == 0 {
		return
	}

	for _, a := range assignments {
		// Round evaluations are tracked separately: they complete on their own
		// and must not make the candidate look evaluated or approved.
		if a.InterviewID != nil {
			continue
		}
		switch a.Status {
		case string(enums.AssignmentStatusCompleted), string(enums.AssignmentStatusSubmitted):
			cwe.EvaluationCompleted = true
		case string(enums.AssignmentStatusPending), string(enums.AssignmentStatusStarted), string(enums.AssignmentStatusOverdue):
			cwe.EvaluationPending = true
		default:
			// A rejected run is closed: it must not keep the candidate stuck on
			// "Evaluation in progress" and hide the button that starts a new one.
			continue
		}
		cwe.HasEvaluation = true
	}
}

// AttachCandidateContacts fills a candidate row with its contact history. Only
// the newest row can schedule a follow-up: an older follow-up is superseded as
// soon as HR records the next conversation.
func AttachCandidateContacts(cwe *CandidateWithEvaluation) {
	contacts, err := candsvc.CandidateServiceInstance.FindContacts(cwe.ID)
	if err != nil {
		return
	}
	cwe.Contacts = contacts
	if len(contacts) == 0 {
		return
	}
	cwe.LastContact = &contacts[0]
	if contacts[0].NextFollowUpAt != nil {
		cwe.NextFollowUp = &contacts[0]
		now := time.Now()
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		due := *contacts[0].NextFollowUpAt
		dueDay := time.Date(due.Year(), due.Month(), due.Day(), 0, 0, 0, 0, now.Location())
		cwe.FollowUpOverdue = dueDay.Before(today)
	}
}

// HireCandidate applies the accept-worker rules shared by the candidates page
// and the approvals page. Acceptance needs a rating of 1-5; the interview
// score already evaluates the candidate, so a completed candidate-level
// evaluation is not required. It returns the success notice shown to the user.
func HireCandidate(c *fiber.Ctx, candidateID int64) (string, error) {
	rating, ratingErr := strconv.ParseFloat(c.FormValue("rating"), 64)
	if candidateID <= 0 {
		return "", errors.New("That candidate no longer exists.")
	}
	if ratingErr != nil || rating < 1 || rating > 5 {
		return "", errors.New("The interview rating has to be a number between 1 and 5.")
	}

	uid := CurrentUserID(c)
	if uid <= 0 {
		log.Printf("[Hire] CurrentUserID returned %d", uid)
		return "", errors.New("Your session has expired. Sign in again and retry.")
	}

	candidate, err := candsvc.CandidateServiceInstance.FindByID(candidateID)
	if err != nil {
		return "", errors.New("That candidate no longer exists.")
	}
	if candidate.Status == "rejected" {
		return "", errors.New("Move this candidate back into the pipeline before accepting them.")
	}
	if candidate.Status == "hired" {
		return "", errors.New("This candidate has already been accepted as an employee.")
	}
	interview, err := recsvc.InterviewServiceInst.GetCandidateSummary(candidateID)
	if err != nil || interview == nil || interview.CompletedReviews == 0 {
		return "", errors.New("This candidate needs at least one completed interview review before they can be hired.")
	}

	if err := candsvc.CandidateServiceInstance.Hire(candidateID, uid, 0, rating, c.FormValue("notes"), ""); err != nil {
		log.Printf("[Hire] error hiring candidate %d: %v", candidateID, err)
		return "", errors.New(HireErrorMessage(err))
	}

	if candidate.FullName != "" {
		return candidate.FullName + " is now an employee — they appear on the Employees page.", nil
	}
	return "Employee created.", nil
}

// HireErrorMessage turns the repository's raw errors into something the person
// hiring can act on. The underlying causes were only ever written to the log.
func HireErrorMessage(err error) string {
	msg := err.Error()
	switch {
	case strings.Contains(strings.ToLower(msg), "already been accepted"):
		return "This candidate has already been accepted, so there is nothing left to create."
	case strings.Contains(strings.ToLower(msg), "not found"):
		return "That candidate no longer exists."
	default:
		return "Could not create employee: " + msg
	}
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
