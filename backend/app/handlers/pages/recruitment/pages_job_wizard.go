package recruitment

import (
	"ems/app/enums"
	core "ems/app/handlers/pages/core"
	orgmodel "ems/app/models/organization"
	recmodel "ems/app/models/recruitment"
	orgsvc "ems/app/services/organization"
	recsvc "ems/app/services/recruitment"
	"log"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// ── Job posting wizard ──
//
// The same shape as the evaluation cycle wizard: one decision per screen, the
// answers carried forward as hidden fields, and a review step before anything
// is written to the database. Posting a job used to mean filling a nine field
// dialog in one go, which is a lot of decisions at once for what is really a
// four part job.

const (
	jobStepBasics  = 1
	jobStepDetails = 2
	jobStepTerms   = 3
	jobStepReview  = 4
)

// JobTypeOption keeps the option values and their display labels in one place
// so the form cannot drift from what the handler accepts.
type JobTypeOption struct {
	Value string
	Label string
}

var jobEmploymentTypeOptions = []JobTypeOption{
	{Value: "full-time", Label: "Full-time"},
	{Value: "part-time", Label: "Part-time"},
	{Value: "contract", Label: "Contract"},
	{Value: "internship", Label: "Internship"},
}

type JobWizardData struct {
	Step       int
	Positions  []orgmodel.Position
	ErrMessage string
	Notice     string

	// Step 1: basics
	Title           string
	PositionID      int64
	EmploymentType  string
	EmploymentTypes []JobTypeOption
	PositionName    string

	// Step 2: details
	Description  string
	Requirements string

	// Step 3: skills and terms. Salary stays as text until the final step so
	// whatever the user typed survives a trip backwards through the wizard.
	SkillsRequired string
	SalaryMin      string
	SalaryMax      string

	// Step 4
	Status string

	// Result
	Created   bool
	JobID     int64
	Published bool
}

func (d *JobWizardData) loadPositions() {
	d.Positions, _ = orgsvc.PositionServiceInst.FindAll()
	for _, p := range d.Positions {
		if p.ID == d.PositionID {
			d.PositionName = p.Name
			break
		}
	}
}

func isJobEmploymentType(v string) bool {
	for _, t := range jobEmploymentTypeOptions {
		if t.Value == v {
			return true
		}
	}
	return false
}

func (h *RecruitmentHandler) jobWizardPage(c *fiber.Ctx, data *JobWizardData) error {
	data.loadPositions()
	data.EmploymentTypes = jobEmploymentTypeOptions
	return core.RenderPage(c, "page-job-wizard", data)
}

// jobWizardFromForm rebuilds the review step from the hidden fields the earlier
// steps posted forward, so launching (and recovering from a launch error) sees
// exactly what the user last saw.
func jobWizardFromForm(c *fiber.Ctx) *JobWizardData {
	status := c.FormValue("status")
	if status != "publish" {
		status = "draft"
	}
	return &JobWizardData{
		Step:           jobStepReview,
		Title:          strings.TrimSpace(c.FormValue("title")),
		PositionID:     core.FormInt64(c, "position_id"),
		EmploymentType: c.FormValue("employment_type"),
		Description:    strings.TrimSpace(c.FormValue("description")),
		Requirements:   strings.TrimSpace(c.FormValue("requirements")),
		SkillsRequired: strings.TrimSpace(c.FormValue("skills_required")),
		SalaryMin:      strings.TrimSpace(c.FormValue("salary_min")),
		SalaryMax:      strings.TrimSpace(c.FormValue("salary_max")),
		Status:         status,
	}
}

// jobSalaries parses the optional range. An unusable value is reported rather
// than silently dropped, so a typo cannot quietly produce a posting with no pay.
func jobSalaries(minRaw, maxRaw string) (min, max *float64, errMsg string) {
	parse := func(raw, label string) (*float64, string) {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			return nil, ""
		}
		f, err := strconv.ParseFloat(raw, 64)
		if err != nil || f < 0 {
			return nil, label + " has to be a number."
		}
		return &f, ""
	}

	var msg string
	if min, msg = parse(minRaw, "Minimum salary"); msg != "" {
		return nil, nil, msg
	}
	if max, msg = parse(maxRaw, "Maximum salary"); msg != "" {
		return nil, nil, msg
	}
	if min != nil && max != nil && *max < *min {
		return nil, nil, "The maximum salary cannot be lower than the minimum."
	}
	return min, max, ""
}

func (h *RecruitmentHandler) JobWizard(c *fiber.Ctx) error {
	if err := core.RequirePermission(c, enums.PermissionRecruitmentManage); err != nil {
		return err
	}
	return h.jobWizardPage(c, &JobWizardData{
		Step:           jobStepBasics,
		EmploymentType: "full-time",
	})
}

func (h *RecruitmentHandler) JobWizardBasics(c *fiber.Ctx) error {
	if err := core.RequirePermission(c, enums.PermissionRecruitmentManage); err != nil {
		return err
	}

	data := &JobWizardData{
		Step:           jobStepDetails,
		Title:          strings.TrimSpace(c.FormValue("title")),
		PositionID:     core.FormInt64(c, "position_id"),
		EmploymentType: c.FormValue("employment_type"),
	}
	if !isJobEmploymentType(data.EmploymentType) {
		data.EmploymentType = "full-time"
	}

	if data.Title == "" {
		data.Step = jobStepBasics
		data.ErrMessage = "Give the job a title so people know what it is."
		return h.jobWizardPage(c, data)
	}
	if data.PositionID <= 0 {
		data.Step = jobStepBasics
		data.ErrMessage = "Choose the position this posting is for."
		return h.jobWizardPage(c, data)
	}
	return h.jobWizardPage(c, data)
}

func (h *RecruitmentHandler) JobWizardDetails(c *fiber.Ctx) error {
	if err := core.RequirePermission(c, enums.PermissionRecruitmentManage); err != nil {
		return err
	}

	return h.jobWizardPage(c, &JobWizardData{
		Step:           jobStepTerms,
		Title:          strings.TrimSpace(c.FormValue("title")),
		PositionID:     core.FormInt64(c, "position_id"),
		EmploymentType: c.FormValue("employment_type"),
		Description:    strings.TrimSpace(c.FormValue("description")),
		Requirements:   strings.TrimSpace(c.FormValue("requirements")),
	})
}

func (h *RecruitmentHandler) JobWizardTerms(c *fiber.Ctx) error {
	if err := core.RequirePermission(c, enums.PermissionRecruitmentManage); err != nil {
		return err
	}

	data := &JobWizardData{
		Step:           jobStepReview,
		Title:          strings.TrimSpace(c.FormValue("title")),
		PositionID:     core.FormInt64(c, "position_id"),
		EmploymentType: c.FormValue("employment_type"),
		Description:    strings.TrimSpace(c.FormValue("description")),
		Requirements:   strings.TrimSpace(c.FormValue("requirements")),
		SkillsRequired: strings.TrimSpace(c.FormValue("skills_required")),
		SalaryMin:      strings.TrimSpace(c.FormValue("salary_min")),
		SalaryMax:      strings.TrimSpace(c.FormValue("salary_max")),
		Status:         "draft",
	}
	if data.EmploymentType == "" {
		data.EmploymentType = "full-time"
	}

	if _, _, errMsg := jobSalaries(data.SalaryMin, data.SalaryMax); errMsg != "" {
		data.Step = jobStepTerms
		data.ErrMessage = errMsg
		return h.jobWizardPage(c, data)
	}
	return h.jobWizardPage(c, data)
}

// ── Back navigation ──

func (h *RecruitmentHandler) JobWizardBackBasics(c *fiber.Ctx) error {
	if err := core.RequirePermission(c, enums.PermissionRecruitmentManage); err != nil {
		return err
	}
	data := jobWizardFromForm(c)
	data.Step = jobStepBasics
	return h.jobWizardPage(c, data)
}

func (h *RecruitmentHandler) JobWizardBackDetails(c *fiber.Ctx) error {
	if err := core.RequirePermission(c, enums.PermissionRecruitmentManage); err != nil {
		return err
	}
	data := jobWizardFromForm(c)
	data.Step = jobStepDetails
	return h.jobWizardPage(c, data)
}

func (h *RecruitmentHandler) JobWizardBackTerms(c *fiber.Ctx) error {
	if err := core.RequirePermission(c, enums.PermissionRecruitmentManage); err != nil {
		return err
	}
	data := jobWizardFromForm(c)
	data.Step = jobStepTerms
	return h.jobWizardPage(c, data)
}

// ── Launch ──

func (h *RecruitmentHandler) JobWizardLaunch(c *fiber.Ctx) error {
	if err := core.RequirePermission(c, enums.PermissionRecruitmentManage); err != nil {
		return err
	}

	data := jobWizardFromForm(c)
	fail := func(msg string) error {
		core.TriggerToast(c, msg, "error")
		data.ErrMessage = msg
		return h.jobWizardPage(c, data)
	}

	if data.Title == "" {
		return fail("The job needs a title before it can be created.")
	}
	if data.PositionID <= 0 {
		return fail("Choose the position this posting is for.")
	}
	salaryMin, salaryMax, errMsg := jobSalaries(data.SalaryMin, data.SalaryMax)
	if errMsg != "" {
		return fail(errMsg)
	}
	if data.Description == "" && data.Requirements == "" {
		return fail("Add a description or a list of requirements so applicants know what the job involves.")
	}

	posting := &recmodel.JobPosting{
		PositionID:     data.PositionID,
		Title:          data.Title,
		Description:    data.Description,
		Requirements:   data.Requirements,
		SkillsRequired: data.SkillsRequired,
		SalaryMin:      salaryMin,
		SalaryMax:      salaryMax,
		EmploymentType: data.EmploymentType,
		Status:         "draft",
		CreatedBy:      core.CurrentUserID(c),
	}
	if err := recsvc.JobPostingServiceInst.Create(posting); err != nil {
		log.Printf("[JobWizard] create failed: %v", err)
		return fail("Could not create the job posting. Please try again.")
	}

	data.JobID = posting.ID
	if data.Status == "publish" {
		if err := recsvc.JobPostingServiceInst.Publish(posting.ID); err != nil {
			log.Printf("[JobWizard] publish failed for %d: %v", posting.ID, err)
			// The posting exists as a draft, so say so rather than reporting a
			// total failure the user would then retry and duplicate.
			core.TriggerToast(c, "The posting was created as a draft, but publishing it failed. Publish it from the list.", "error")
			data.Published = false
		} else {
			data.Published = true
		}
	}

	data.Created = true
	data.EmploymentType = posting.EmploymentType
	notice := data.Title + " was created."
	if data.Published {
		notice = data.Title + " is now published and taking applications."
	}
	core.TriggerToast(c, notice, "success")
	data.Notice = notice
	return h.jobWizardPage(c, data)
}
