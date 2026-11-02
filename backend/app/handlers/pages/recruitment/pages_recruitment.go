package recruitment

import (
	"ems/app/enums"
	core "ems/app/handlers/pages/core"
	candmodel "ems/app/models/candidate"
	evalmodel "ems/app/models/evaluation"
	orgmodel "ems/app/models/organization"
	recmodel "ems/app/models/recruitment"
	"ems/app/services"
	candsvc "ems/app/services/candidate"
	evalsvc "ems/app/services/evaluation"
	orgsvc "ems/app/services/organization"
	recsvc "ems/app/services/recruitment"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

type RecruitmentHandler struct{}

func (h *RecruitmentHandler) JobPostings(c *fiber.Ctx) error {
	postings, _ := recsvc.JobPostingServiceInst.FindAll()
	positions, _ := orgsvc.PositionServiceInst.FindAll()
	return core.RenderPage(c, "page-job-postings", struct {
		Postings  []recmodel.JobPostingListItem
		Positions []orgmodel.Position
	}{Postings: postings, Positions: positions})
}

func (h *RecruitmentHandler) CreateJobPosting(c *fiber.Ctx) error {
	positionID := core.FormInt64(c, "position_id")
	title := strings.TrimSpace(c.FormValue("title"))
	description := strings.TrimSpace(c.FormValue("description"))
	requirements := strings.TrimSpace(c.FormValue("requirements"))
	skillsRequired := strings.TrimSpace(c.FormValue("skills_required"))
	employmentType := c.FormValue("employment_type")
	if employmentType == "" {
		employmentType = "full-time"
	}

	var salaryMin, salaryMax *float64
	if v := c.FormValue("salary_min"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			salaryMin = &f
		}
	}
	if v := c.FormValue("salary_max"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			salaryMax = &f
		}
	}

	userID := core.CurrentUserID(c)
	posting := &recmodel.JobPosting{
		PositionID:     positionID,
		Title:          title,
		Description:    description,
		Requirements:   requirements,
		SkillsRequired: skillsRequired,
		SalaryMin:      salaryMin,
		SalaryMax:      salaryMax,
		EmploymentType: employmentType,
		Status:         "draft",
		CreatedBy:      userID,
	}

	if err := recsvc.JobPostingServiceInst.Create(posting); err != nil {
		return core.RenderError(c, fiber.StatusInternalServerError, "Failed to create job posting.")
	}
	return h.JobPostings(c)
}

func (h *RecruitmentHandler) PublishJobPosting(c *fiber.Ctx) error {
	id, _ := strconv.ParseInt(c.Params("id"), 10, 64)
	if err := recsvc.JobPostingServiceInst.Publish(id); err != nil {
		return core.RenderError(c, fiber.StatusInternalServerError, "Failed to publish.")
	}
	return h.JobPostings(c)
}

func (h *RecruitmentHandler) CloseJobPosting(c *fiber.Ctx) error {
	id, _ := strconv.ParseInt(c.Params("id"), 10, 64)
	if err := recsvc.JobPostingServiceInst.Close(id); err != nil {
		return core.RenderError(c, fiber.StatusInternalServerError, "Failed to close.")
	}
	return h.JobPostings(c)
}

func (h *RecruitmentHandler) ReopenJobPosting(c *fiber.Ctx) error {
	id, _ := strconv.ParseInt(c.Params("id"), 10, 64)
	if err := recsvc.JobPostingServiceInst.Reopen(id); err != nil {
		return core.RenderError(c, fiber.StatusInternalServerError, "Failed to reopen.")
	}
	return h.JobPostings(c)
}

func (h *RecruitmentHandler) DeleteJobPosting(c *fiber.Ctx) error {
	id, _ := strconv.ParseInt(c.Params("id"), 10, 64)
	if err := recsvc.JobPostingServiceInst.Delete(id); err != nil {
		return core.RenderError(c, fiber.StatusInternalServerError, "Failed to delete.")
	}
	return h.JobPostings(c)
}

func (h *RecruitmentHandler) JobPostingCandidates(c *fiber.Ctx) error {
	postingID, _ := strconv.ParseInt(c.Params("id"), 10, 64)
	posting, err := recsvc.JobPostingServiceInst.FindByID(postingID)
	if err != nil {
		return core.RenderError(c, fiber.StatusNotFound, "Job posting not found.")
	}
	candidates, _ := candsvc.CandidateServiceInstance.FindByJobPostingID(postingID)
	periods, _ := evalsvc.PeriodServiceInstance.FindAll()
	templates, _ := evalsvc.TemplateServiceInstance.FindAll()
	workers := core.LoadActiveWorkers()
	interviewers := core.LoadInterviewers()
	if len(interviewers) == 0 {
		interviewers = workers
	}

	uid := core.CurrentUserID(c)
	canManage := core.CanManageInterviews(c)
	assignedCandidateIDs, _ := recsvc.InterviewServiceInst.FindAssignedCandidateIDs(uid)
	canSchedule := canManage || (len(assignedCandidateIDs) > 0 && core.HasAnyPermission(c, enums.PermissionEvaluationEvaluate))
	assigned := make(map[int64]bool, len(assignedCandidateIDs))
	for _, id := range assignedCandidateIDs {
		assigned[id] = true
	}

	rounds, _ := recsvc.InterviewRoundServiceInst.FindByJobPostingID(postingID)
	templateNames := make(map[int64]string, len(templates))
	for _, t := range templates {
		templateNames[t.ID] = t.Name
	}
	roundViews := make([]core.InterviewRoundView, 0, len(rounds))
	for _, round := range rounds {
		roundViews = append(roundViews, core.InterviewRoundView{
			InterviewRound: round,
			PostingTitle:   posting.Title,
			TemplateName:   templateNames[round.EvaluationTemplateID],
			Sections:       loadInterviewSections(&round),
		})
	}

	var candidatesWithEval []core.CandidateWithEvaluation
	summaries := core.LoadCandidateSummaryMap()
	for _, candidate := range candidates {
		cwe := core.CandidateWithEvaluation{CandidateListItem: candidate}
		cwe.PostingID = postingID
		cwe.Interview = summaries[candidate.ID]
		cwe.InterviewReady = core.InterviewReady(cwe.Interview)
		core.AttachCandidateContacts(&cwe)
		core.AttachCandidateEvaluation(&cwe)
		cwe.CanScheduleInterview = canSchedule && (canManage || assigned[candidate.ID]) &&
			!strings.EqualFold(candidate.Status, "hired")
		candidatesWithEval = append(candidatesWithEval, cwe)
	}

	return core.RenderPage(c, "page-job-posting-candidates", struct {
		Posting           *recmodel.JobPosting
		Candidates        []core.CandidateWithEvaluation
		Periods           []evalmodel.EvaluationPeriod
		Templates         []evalmodel.EvaluationTemplate
		Workers           []core.WorkerOption
		InterviewerGroups []core.InterviewerGroup
		Rounds            []core.InterviewRoundView
	}{
		Posting:           posting,
		Candidates:        candidatesWithEval,
		Periods:           periods,
		Templates:         templates,
		Workers:           workers,
		InterviewerGroups: core.GroupInterviewers(interviewers),
		Rounds:            roundViews,
	})
}

// InterviewRoundView carries the posting title alongside a round so the
// scheduling picker can tell rounds from different postings apart.
type InterviewRoundView = core.InterviewRoundView

// InterviewQuestionView pairs a template question with the answer already
// recorded on the interview, if any.
type InterviewQuestionView struct {
	Model    evalmodel.EvaluationQuestion
	Required bool
	Answer   *recmodel.InterviewFeedbackAnswer
}

// InterviewSectionView is one section of the round's template. Only the
// sections picked for the round reach the feedback form, which is how a large
// template is divided across interview rounds.
type InterviewSectionView struct {
	Name        string
	Description string
	Questions   []InterviewQuestionView
}

// loadInterviewSections reads the sections selected for the round together
// with their questions. A round without a template returns no sections.
func loadInterviewSections(round *recmodel.InterviewRound) []evalmodel.EvaluationSection {
	if round == nil || round.EvaluationTemplateID <= 0 {
		return nil
	}
	sections, _ := recsvc.InterviewRoundServiceInst.FindSectionsByRoundID(round.ID)
	return sections
}

// effectiveInterviewTemplateID resolves the template an interview is actually
// scored with. The scheduler can override the round's template for one
// interview, and the override is stored on the round evaluations.
func effectiveInterviewTemplateID(interviewID int64, round *recmodel.InterviewRound) int64 {
	if assignments, err := evalsvc.AssignmentServiceInstance.FindByInterviewID(interviewID); err == nil {
		for _, a := range assignments {
			if a.EvaluationTemplateID > 0 {
				return a.EvaluationTemplateID
			}
		}
	}
	if round != nil {
		return round.EvaluationTemplateID
	}
	return 0
}

// loadInterviewSectionsForFeedback returns the questions the panel answers for
// this interview. When the round's template is in effect the round's selected
// sections are used; an override template supplies all of its own sections.
func loadInterviewSectionsForFeedback(interviewID int64, round *recmodel.InterviewRound) []evalmodel.EvaluationSection {
	templateID := effectiveInterviewTemplateID(interviewID, round)
	if templateID <= 0 {
		return nil
	}
	if round != nil && templateID == round.EvaluationTemplateID {
		if sections, err := recsvc.InterviewRoundServiceInst.FindSectionsByRoundID(round.ID); err == nil && len(sections) > 0 {
			return sections
		}
	}
	sections, _ := evalsvc.SectionServiceInstance.FindByTemplateID(templateID)
	return sections
}

// CandidateGroup keeps the scheduling picker's candidates under the job
// posting they applied for, so HR is not reading one long mixed list.
type CandidateGroup struct {
	PostingID    int64
	PostingTitle string
	Candidates   []candmodel.CandidateListItem
}

// InterviewerGroup keeps the panel checklist under each department.
type InterviewerGroup = core.InterviewerGroup

func groupCandidatesByPosting(candidates []candmodel.CandidateListItem, postings []recmodel.JobPostingListItem) []CandidateGroup {
	groups := make(map[int64]*CandidateGroup)
	order := make([]int64, 0)

	ensure := func(id int64, title string) *CandidateGroup {
		if group, ok := groups[id]; ok {
			if group.PostingTitle == "" {
				group.PostingTitle = title
			}
			return group
		}
		if id == 0 {
			title = "No job posting"
		} else if title == "" {
			title = "Job posting"
		}
		group := &CandidateGroup{PostingID: id, PostingTitle: title}
		groups[id] = group
		order = append(order, id)
		return group
	}

	// Postings come first so the groups follow the job list, even before any
	// candidate has applied.
	for _, posting := range postings {
		ensure(posting.ID, posting.Title)
	}
	for _, candidate := range candidates {
		var postingID int64
		if candidate.JobPostingID != nil {
			postingID = *candidate.JobPostingID
		}
		group := ensure(postingID, candidate.PostingTitle)
		group.Candidates = append(group.Candidates, candidate)
	}

	grouped := make([]CandidateGroup, 0, len(order))
	for _, id := range order {
		if group := groups[id]; len(group.Candidates) > 0 {
			grouped = append(grouped, *group)
		}
	}
	return grouped
}

func groupInterviewers(workers []core.WorkerOption) []InterviewerGroup {
	return core.GroupInterviewers(workers)
}

// InterviewFilters is the read-only filter bar above the schedule table. The
// values are applied in SQL by the repository, not after loading every row.
type InterviewFilters struct {
	PostingID     int64
	RoundID       int64
	InterviewerID int64
	Status        string
	Search        string
}

// InterviewPageData is the whole schedule page. The Can* flags let the
// template hide controls rather than relying on the browser to hide them.
type InterviewPageData struct {
	Interviews        []recmodel.InterviewListItem
	Candidates        []candmodel.CandidateListItem
	CandidateGroups   []CandidateGroup
	Workers           []core.WorkerOption
	InterviewerGroups []InterviewerGroup
	Rounds            []InterviewRoundView
	Postings          []recmodel.JobPostingListItem
	Templates         []evalmodel.EvaluationTemplate
	Sections          []evalmodel.EvaluationSection
	Filters           InterviewFilters
	Rankings          []recmodel.CandidateEvaluationSummary

	CanManage   bool
	CanSchedule bool
	CanScore    bool
	CanApprove  bool
	CanViewAll  bool
	ReadOnly    bool
}

// queryInt64 reads an optional integer from the query string.
func queryInt64(c *fiber.Ctx, name string) int64 {
	value, _ := strconv.ParseInt(strings.TrimSpace(c.Query(name)), 10, 64)
	return value
}

// Interview Scheduling Page
func (h *RecruitmentHandler) Interviews(c *fiber.Ctx) error {
	if err := core.RequirePermission(c,
		enums.PermissionRecruitmentManage,
		enums.PermissionInterviewView,
		enums.PermissionInterviewAssign,
		enums.PermissionInterviewScore,
		enums.PermissionInterviewApprove,
		enums.PermissionEvaluationAssign,
		enums.PermissionEvaluationEvaluate,
	); err != nil {
		return err
	}

	uid := core.CurrentUserID(c)
	canManage := core.CanManageInterviews(c)
	canViewAll := core.CanViewAllInterviews(c)
	canApprove := core.CanApproveInterviews(c)
	canScore := core.CanScoreInterviews(c)

	// The evaluator sees only the candidate they were assigned. Panel members
	// additionally see the interviews they sit on; this list is the union the
	// repository uses for the viewer scope.
	assignedCandidateIDs, _ := recsvc.InterviewServiceInst.FindAssignedCandidateIDs(uid)
	canSchedule := canManage ||
		(len(assignedCandidateIDs) > 0 && core.HasAnyPermission(c, enums.PermissionEvaluationEvaluate))

	filter := InterviewFilters{
		PostingID:     queryInt64(c, "posting_id"),
		RoundID:       queryInt64(c, "round_id"),
		InterviewerID: queryInt64(c, "interviewer_id"),
		Status:        strings.TrimSpace(c.Query("status")),
		Search:        strings.TrimSpace(c.Query("q")),
	}

	repoFilter := recsvc.InterviewFilter{
		RestrictToViewer: !canViewAll,
		PostingID:        filter.PostingID,
		RoundID:          filter.RoundID,
		InterviewerID:    filter.InterviewerID,
		Status:           filter.Status,
		Search:           filter.Search,
	}
	if !canViewAll {
		repoFilter.CandidateIDs = assignedCandidateIDs
		repoFilter.PanelUserID = uid
	}
	interviews, _ := recsvc.InterviewServiceInst.FindFiltered(repoFilter)

	var candidates []candmodel.CandidateListItem
	if canViewAll {
		candidates, _ = candsvc.CandidateServiceInstance.FindAll()
	} else {
		candidates, _ = candsvc.CandidateServiceInstance.FindByIDs(assignedCandidateIDs)
	}

	// Restrict the posting/round pickers to the postings the viewer can see.
	visiblePostings := make(map[int64]bool, len(candidates))
	for _, candidate := range candidates {
		if candidate.JobPostingID != nil {
			visiblePostings[*candidate.JobPostingID] = true
		}
	}
	allPostings, _ := recsvc.JobPostingServiceInst.FindAll()
	postings := make([]recmodel.JobPostingListItem, 0, len(allPostings))
	for _, posting := range allPostings {
		if canViewAll || visiblePostings[posting.ID] {
			postings = append(postings, posting)
		}
	}

	workers := core.LoadInterviewers()
	templates, _ := evalsvc.TemplateServiceInstance.FindAll()

	// The Add Round dialog offers a template and then the sections inside it.
	// Sections are listed with their template id so the page's script can
	// show only the ones that belong to the selected template.
	templateNames := make(map[int64]string, len(templates))
	var allSections []evalmodel.EvaluationSection
	for _, template := range templates {
		templateNames[template.ID] = template.Name
		sections, _ := evalsvc.SectionServiceInstance.FindByTemplateID(template.ID)
		allSections = append(allSections, sections...)
	}

	allRounds := make([]InterviewRoundView, 0)
	for _, p := range postings {
		rounds, _ := recsvc.InterviewRoundServiceInst.FindByJobPostingID(p.ID)
		for _, round := range rounds {
			allRounds = append(allRounds, InterviewRoundView{
				InterviewRound: round,
				PostingTitle:   p.Title,
				TemplateName:   templateNames[round.EvaluationTemplateID],
				Sections:       loadInterviewSections(&round),
			})
		}
	}

	data := InterviewPageData{
		Interviews:        interviews,
		Candidates:        candidates,
		CandidateGroups:   groupCandidatesByPosting(candidates, postings),
		Workers:           workers,
		InterviewerGroups: groupInterviewers(workers),
		Rounds:            allRounds,
		Postings:          postings,
		Templates:         templates,
		Sections:          allSections,
		Filters:           filter,
		CanManage:         canManage,
		CanSchedule:       canSchedule,
		CanScore:          canScore,
		CanApprove:        canApprove,
		CanViewAll:        canViewAll,
		ReadOnly:          canViewAll && !canManage && !canSchedule,
	}

	if canApprove {
		data.Rankings = interviewRankings(filter, canViewAll, visiblePostings)
	}

	return core.RenderPage(c, "page-interviews", data)
}

// interviewRankings filters the shared candidate ranking down to what this
// viewer may see. The ranks themselves are assigned by core.CandidateRankings,
// so the Interviews page, Reports, Dashboard and Approvals all agree on who is
// first in each job posting. Filtering never renumbers candidates: #4 stays #4
// even when only a search result is shown.
func interviewRankings(filter InterviewFilters, canViewAll bool, visiblePostings map[int64]bool) []recmodel.CandidateEvaluationSummary {
	rankings := core.LoadCandidateRankings()
	filtered := make([]recmodel.CandidateEvaluationSummary, 0, len(rankings))
	for _, summary := range rankings {
		if !canViewAll && !visiblePostings[summary.JobPostingID] {
			continue
		}
		if filter.PostingID > 0 && summary.JobPostingID != filter.PostingID {
			continue
		}
		if search := strings.ToLower(strings.TrimSpace(filter.Search)); search != "" {
			if !strings.Contains(strings.ToLower(summary.CandidateName), search) &&
				!strings.Contains(strings.ToLower(summary.PostingTitle), search) {
				continue
			}
		}
		filtered = append(filtered, summary)
	}
	return filtered
}

// CreateInterviewRound adds a round to a job posting. Rounds are the fixed
// stages (screening, technical, final, ...) that interviews are scheduled
// against, and their weight controls how much each score counts.
func (h *RecruitmentHandler) CreateInterviewRound(c *fiber.Ctx) error {
	if err := core.RequirePermission(c, enums.PermissionInterviewAssign, enums.PermissionRecruitmentManage); err != nil {
		return err
	}

	postingID := core.FormInt64(c, "job_posting_id")
	templateID := core.FormInt64(c, "evaluation_template_id")
	sectionIDs := core.FormInt64s(c, "section_ids")
	name := strings.TrimSpace(c.FormValue("name"))
	roundNumber, _ := strconv.Atoi(c.FormValue("round_number"))
	roundType := strings.TrimSpace(c.FormValue("round_type"))
	description := strings.TrimSpace(c.FormValue("description"))
	maxScore, _ := strconv.ParseFloat(c.FormValue("max_score"), 64)
	passScore, _ := strconv.ParseFloat(c.FormValue("pass_score"), 64)
	weight, _ := strconv.ParseFloat(c.FormValue("weight"), 64)

	if postingID <= 0 || name == "" || roundNumber <= 0 {
		return core.RenderError(c, fiber.StatusBadRequest, "Job posting, round name, and round number are required.")
	}
	if roundType == "" {
		roundType = "technical"
	}
	if maxScore <= 0 {
		maxScore = 10
	}
	if weight <= 0 {
		weight = 1
	}

	// A round may only use sections from the template it names. Posting a
	// section from another template would ask questions the round's reviewers
	// never agreed on, so those ids are dropped.
	validSections := make([]int64, 0, len(sectionIDs))
	if templateID > 0 {
		sections, err := evalsvc.SectionServiceInstance.FindByTemplateID(templateID)
		if err != nil {
			log.Printf("[CreateInterviewRound] template %d sections: %v", templateID, err)
			return core.RenderError(c, fiber.StatusInternalServerError, "Failed to load the evaluation template.")
		}
		allowed := make(map[int64]bool, len(sections))
		for _, section := range sections {
			allowed[section.ID] = true
		}
		for _, id := range sectionIDs {
			if allowed[id] {
				validSections = append(validSections, id)
			}
		}
		if len(validSections) == 0 {
			return core.RenderError(c, fiber.StatusBadRequest, "Choose at least one section from the evaluation template for this round.")
		}
	}

	round := &recmodel.InterviewRound{
		JobPostingID:         postingID,
		EvaluationTemplateID: templateID,
		Name:                 name,
		RoundNumber:          roundNumber,
		RoundType:            roundType,
		Description:          description,
		MaxScore:             maxScore,
		PassScore:            passScore,
		Weight:               weight,
		Status:               "active",
	}
	if err := recsvc.InterviewRoundServiceInst.Create(round); err != nil {
		log.Printf("[CreateInterviewRound] %v", err)
		return core.RenderError(c, fiber.StatusInternalServerError, "Failed to create interview round.")
	}
	if len(validSections) > 0 {
		if err := recsvc.InterviewRoundServiceInst.SaveSections(round.ID, validSections); err != nil {
			log.Printf("[CreateInterviewRound] sections for round %d: %v", round.ID, err)
			// Do not leave a round behind that claims a template but asks no
			// questions because its section links failed.
			_ = recsvc.InterviewRoundServiceInst.Delete(round.ID)
			return core.RenderError(c, fiber.StatusInternalServerError, "Failed to save the round's questions.")
		}
	}
	core.TriggerToast(c, "Interview round created.", "success")
	return h.Interviews(c)
}

func (h *RecruitmentHandler) UpdateInterviewRound(c *fiber.Ctx) error {
	if err := core.RequirePermission(c, enums.PermissionInterviewAssign, enums.PermissionRecruitmentManage); err != nil {
		return err
	}
	id, _ := strconv.ParseInt(c.Params("id"), 10, 64)
	round, err := recsvc.InterviewRoundServiceInst.FindByID(id)
	if err != nil {
		return core.RenderError(c, fiber.StatusNotFound, "Interview round not found.")
	}
	round.Name = strings.TrimSpace(c.FormValue("name"))
	round.RoundNumber, _ = strconv.Atoi(c.FormValue("round_number"))
	round.RoundType = strings.TrimSpace(c.FormValue("round_type"))
	round.Description = strings.TrimSpace(c.FormValue("description"))
	round.MaxScore, _ = strconv.ParseFloat(c.FormValue("max_score"), 64)
	round.PassScore, _ = strconv.ParseFloat(c.FormValue("pass_score"), 64)
	round.Weight, _ = strconv.ParseFloat(c.FormValue("weight"), 64)
	if round.Name == "" || round.RoundNumber <= 0 || round.RoundType == "" || round.MaxScore <= 0 || round.PassScore < 0 || round.Weight <= 0 {
		return core.RenderError(c, fiber.StatusBadRequest, "Enter a name, valid order, type, max score, pass score, and weight.")
	}
	if round.PassScore > round.MaxScore {
		return core.RenderError(c, fiber.StatusBadRequest, "Pass score cannot be greater than max score.")
	}
	if err := recsvc.InterviewRoundServiceInst.Update(round); err != nil {
		return core.RenderError(c, fiber.StatusInternalServerError, "Failed to update interview round.")
	}
	core.TriggerToast(c, "Interview round updated.", "success")
	return h.Interviews(c)
}

func (h *RecruitmentHandler) DeleteInterviewRound(c *fiber.Ctx) error {
	if err := core.RequirePermission(c, enums.PermissionInterviewAssign, enums.PermissionRecruitmentManage); err != nil {
		return err
	}
	id, _ := strconv.ParseInt(c.Params("id"), 10, 64)
	if _, err := recsvc.InterviewRoundServiceInst.FindByID(id); err != nil {
		return core.RenderError(c, fiber.StatusNotFound, "Interview round not found.")
	}
	hasInterviews, err := recsvc.InterviewRoundServiceInst.HasInterviews(id)
	if err != nil {
		return core.RenderError(c, fiber.StatusInternalServerError, "Could not check whether this round is in use.")
	}
	if hasInterviews {
		return core.RenderError(c, fiber.StatusBadRequest, "This round has scheduled interviews and cannot be removed.")
	}
	if err := recsvc.InterviewRoundServiceInst.Delete(id); err != nil {
		return core.RenderError(c, fiber.StatusInternalServerError, "Failed to remove interview round.")
	}
	core.TriggerToast(c, "Interview round removed.", "success")
	return h.Interviews(c)
}

func (h *RecruitmentHandler) CreateInterview(c *fiber.Ctx) error {
	candidateIDs := core.FormInt64s(c, "candidate_ids")
	if len(candidateIDs) == 0 {
		// Keep older single-candidate forms and bookmarks working.
		if candidateID := core.FormInt64(c, "candidate_id"); candidateID > 0 {
			candidateIDs = []int64{candidateID}
		}
	}
	candidateIDs = uniquePositiveIDs(candidateIDs)

	roundID := core.FormInt64(c, "interview_round_id")
	interviewerIDs := core.FormInt64s(c, "interviewer_ids")
	if len(interviewerIDs) == 0 {
		// Keep older single-interviewer forms working.
		if interviewerID := core.FormInt64(c, "interviewer_id"); interviewerID > 0 {
			interviewerIDs = []int64{interviewerID}
		}
	}
	scheduledAt := c.FormValue("scheduled_at")
	duration, _ := strconv.Atoi(c.FormValue("duration_minutes"))
	location := strings.TrimSpace(c.FormValue("location"))
	meetingLink := strings.TrimSpace(c.FormValue("meeting_link"))
	notes := strings.TrimSpace(c.FormValue("notes"))

	if len(candidateIDs) == 0 || roundID <= 0 || len(interviewerIDs) == 0 || scheduledAt == "" {
		return core.RenderError(c, fiber.StatusBadRequest, "At least one candidate, a round, one interviewer, and a date are required.")
	}

	// A manager may schedule anybody. An assigned evaluator may schedule only
	// their own candidate - the same candidate scope that drives their list.
	if !core.CanManageInterviews(c) {
		if !core.HasAnyPermission(c, enums.PermissionEvaluationEvaluate) {
			return core.Forbidden(c, "You can only schedule interviews for candidates assigned to you.")
		}
		assigned, _ := recsvc.InterviewServiceInst.FindAssignedCandidateIDs(core.CurrentUserID(c))
		allowed := make(map[int64]bool, len(assigned))
		for _, id := range assigned {
			allowed[id] = true
		}
		for _, candidateID := range candidateIDs {
			if !allowed[candidateID] {
				return core.Forbidden(c, "You can only schedule interviews for candidates assigned to you.")
			}
		}
	}

	for _, candidateID := range candidateIDs {
		if err := recsvc.InterviewServiceInst.ValidateCandidateRound(candidateID, roundID); err != nil {
			if errors.Is(err, recsvc.ErrRoundNotForCandidate) {
				return core.RenderError(c, fiber.StatusBadRequest, "The selected round does not belong to every selected candidate's job posting. Pick candidates from one group.")
			}
			return core.RenderError(c, fiber.StatusBadRequest, "Candidate or interview round not found.")
		}
	}

	if duration <= 0 {
		duration = 60
	}

	// Parse the date once for the whole batch.
	var scheduled time.Time
	for _, format := range []string{"2006-01-02T15:04", "2006-01-02 15:04"} {
		if t, err := services.ParseTime(scheduledAt, format); err == nil {
			scheduled = t
			break
		}
	}
	if scheduled.IsZero() {
		return core.RenderError(c, fiber.StatusBadRequest, "The interview date is not valid.")
	}

	created := 0
	for _, candidateID := range candidateIDs {
		interview := &recmodel.Interview{
			CandidateID:      candidateID,
			InterviewRoundID: roundID,
			InterviewerID:    interviewerIDs[0],
			InterviewerIDs:   interviewerIDs,
			ScheduledAt:      scheduled,
			DurationMinutes:  duration,
			Location:         location,
			MeetingLink:      meetingLink,
			Status:           "scheduled",
			Notes:            notes,
		}
		if err := recsvc.InterviewServiceInst.Create(interview); err != nil {
			if created == 0 {
				return core.RenderError(c, fiber.StatusInternalServerError, "Failed to schedule interview.")
			}
			core.TriggerToast(c, fmt.Sprintf("Scheduled %d interview(s); the rest could not be saved.", created), "error")
			return h.Interviews(c)
		}
		created++
		// Scheduling an interview is the moment an applicant becomes
		// "interviewing". Failures here must not lose the scheduled interview.
		if err := candsvc.CandidateServiceInstance.MarkInterviewing(candidateID, core.CurrentUserID(c)); err != nil {
			log.Printf("[CreateInterview] could not advance candidate %d: %v", candidateID, err)
		}
		// Each panel interviewer gets a pending evaluation for this round.
		// It is best-effort: a round without a template still schedules.
		if _, err := evalsvc.AssignmentServiceInstance.CreateForInterview(interview.ID); err != nil {
			log.Printf("[CreateInterview] round evaluation for interview %d: %v", interview.ID, err)
		}
	}

	if created > 1 {
		core.TriggerToast(c, fmt.Sprintf("Scheduled %d interviews for this group.", created), "success")
	}
	return h.Interviews(c)
}

// uniquePositiveIDs drops blanks and repeats while keeping the order the
// picker submitted, so a select-all plus an individual tick cannot create the
// same interview twice.
func uniquePositiveIDs(ids []int64) []int64 {
	unique := make([]int64, 0, len(ids))
	seen := make(map[int64]bool, len(ids))
	for _, id := range ids {
		if id <= 0 || seen[id] {
			continue
		}
		seen[id] = true
		unique = append(unique, id)
	}
	return unique
}

func (h *RecruitmentHandler) UpdateInterviewStatus(c *fiber.Ctx) error {
	id, _ := strconv.ParseInt(c.Params("id"), 10, 64)
	if id <= 0 {
		return core.RenderError(c, fiber.StatusBadRequest, "Interview not found.")
	}
	if !core.CanManageInterviews(c) {
		if !core.HasAnyPermission(c, enums.PermissionInterviewScore) {
			return core.Forbidden(c, "You do not have permission to change this interview.")
		}
		member, _ := recsvc.InterviewServiceInst.IsPanelMember(id, core.CurrentUserID(c))
		if !member {
			return core.Forbidden(c, "Only the assigned panel can update this interview.")
		}
	}
	status := c.FormValue("status")
	if err := recsvc.InterviewServiceInst.UpdateStatus(id, status); err != nil {
		return core.RenderError(c, fiber.StatusInternalServerError, "Failed to update status.")
	}
	if strings.EqualFold(status, "completed") {
		if err := evalsvc.AssignmentServiceInstance.StartForInterview(id); err != nil {
			log.Printf("[UpdateInterviewStatus] round evaluations for interview %d: %v", id, err)
		}
	}
	return h.Interviews(c)
}

func (h *RecruitmentHandler) DeleteInterview(c *fiber.Ctx) error {
	if err := core.RequirePermission(c, enums.PermissionInterviewAssign, enums.PermissionRecruitmentManage); err != nil {
		return err
	}
	id, _ := strconv.ParseInt(c.Params("id"), 10, 64)
	if err := recsvc.InterviewServiceInst.Delete(id); err != nil {
		return core.RenderError(c, fiber.StatusInternalServerError, "Failed to delete interview.")
	}
	return h.Interviews(c)
}

// InterviewFeedbackPageData is one panel member's scoring form plus, for a
// manager/approver, every score already recorded on the interview.
type InterviewFeedbackPageData struct {
	Interview             *recmodel.InterviewListItem
	Feedback              *recmodel.InterviewFeedback
	AllFeedback           []recmodel.InterviewFeedback
	Panel                 []recmodel.InterviewPanelMember
	Sections              []InterviewSectionView
	TemplateName          string
	HasQuestions          bool
	SelectedInterviewerID int64
	AggregateScore        *float64
	AggregateTotalScore   *float64
	CandidateFinalTotal   *float64
	CanEdit               bool
	CanViewAll            bool
	CanApprove            bool
	CanManage             bool
}

// panelContains reports whether a user is one of the interview's panel
// members.
func panelContains(panel []recmodel.InterviewPanelMember, userID int64) bool {
	for _, member := range panel {
		if member.UserID == userID {
			return true
		}
	}
	return false
}

// averageFeedbackScore combines every panel member's score into the number the
// final approver compares candidates by.
func averageFeedbackScore(feedback []recmodel.InterviewFeedback) *float64 {
	total := 0.0
	count := 0
	for _, fb := range feedback {
		if fb.Score != nil {
			total += *fb.Score
			count++
		}
	}
	if count == 0 {
		return nil
	}
	value := total / float64(count)
	return &value
}

func totalFeedbackScore(feedback []recmodel.InterviewFeedback) *float64 {
	total := 0.0
	count := 0
	for _, fb := range feedback {
		if fb.Score != nil {
			total += *fb.Score
			count++
		}
	}
	if count == 0 {
		return nil
	}
	return &total
}

// Interview Feedback Page
func (h *RecruitmentHandler) InterviewFeedback(c *fiber.Ctx) error {
	if err := core.RequirePermission(c,
		enums.PermissionRecruitmentManage,
		enums.PermissionInterviewView,
		enums.PermissionInterviewScore,
		enums.PermissionInterviewAssign,
		enums.PermissionInterviewApprove,
		enums.PermissionEvaluationAssign,
		enums.PermissionEvaluationEvaluate,
		// Approvers reach feedback from the Approvals ranking, so they need to
		// be able to read the interview evidence behind a candidate's score.
		enums.PermissionEvaluationApprove,
		enums.PermissionEvaluationManage,
	); err != nil {
		return err
	}

	interviewID, _ := strconv.ParseInt(c.Params("id"), 10, 64)
	interview, err := recsvc.InterviewServiceInst.FindByID(interviewID)
	if err != nil {
		return core.RenderError(c, fiber.StatusNotFound, "Interview not found.")
	}

	uid := core.CurrentUserID(c)
	canManage := core.CanManageInterviews(c)
	canViewAll := core.CanViewAllInterviews(c)
	canApprove := core.CanApproveInterviews(c)

	panel, _ := recsvc.InterviewServiceInst.FindPanelMembers(interviewID)
	isPanelMember := panelContains(panel, uid)

	// Feedback is visible to users with a wider interview permission and to the
	// people assigned to this candidate: the interview panel and the evaluator
	// assigned to that candidate. A plain interviewer cannot open somebody
	// else's panel.
	isAssignedEvaluator := false
	assigned, _ := recsvc.InterviewServiceInst.FindAssignedCandidateIDs(uid)
	for _, id := range assigned {
		if id == interview.CandidateID {
			isAssignedEvaluator = true
			break
		}
	}
	if !isPanelMember && !isAssignedEvaluator && !canViewAll {
		return core.Forbidden(c, "This interview is not assigned to you.")
	}

	// Permission, or being the candidate's assigned evaluator, reveals the
	// whole panel. A plain interviewer only sees their own score.
	canViewPanel := canViewAll || isAssignedEvaluator

	selectedInterviewerID := queryInt64(c, "interviewer_id")
	if !canViewPanel {
		selectedInterviewerID = uid
	} else {
		if selectedInterviewerID <= 0 {
			if isPanelMember {
				selectedInterviewerID = uid
			} else if len(panel) > 0 {
				selectedInterviewerID = panel[0].UserID
			}
		}
		if !panelContains(panel, selectedInterviewerID) && len(panel) > 0 {
			if isPanelMember {
				selectedInterviewerID = uid
			} else {
				selectedInterviewerID = panel[0].UserID
			}
		}
	}

	fb, _ := recsvc.InterviewFeedbackServiceInst.FindByInterviewAndInterviewer(interviewID, selectedInterviewerID)
	if fb == nil && selectedInterviewerID == 0 {
		fb, _ = recsvc.InterviewFeedbackServiceInst.FindByInterviewID(interviewID)
	}

	// Answers are looked up by question id so the form can restore whatever
	// this panel member already recorded on an earlier save.
	answers := make(map[int64]recmodel.InterviewFeedbackAnswer)
	if fb != nil && fb.ID > 0 {
		saved, _ := recsvc.InterviewFeedbackServiceInst.FindAnswersByFeedbackID(fb.ID)
		for _, answer := range saved {
			answers[answer.EvaluationQuestionID] = answer
		}
	}

	round, _ := recsvc.InterviewRoundServiceInst.FindByID(interview.InterviewRoundID)
	templateName := ""
	if templateID := effectiveInterviewTemplateID(interviewID, round); templateID > 0 {
		if tmpl, err := evalsvc.TemplateServiceInstance.FindByID(templateID); err == nil {
			templateName = tmpl.Name
		}
	}

	var sectionViews []InterviewSectionView
	hasQuestions := false
	for _, sec := range loadInterviewSectionsForFeedback(interviewID, round) {
		view := InterviewSectionView{Name: sec.Name, Description: sec.Description}
		questions, _ := evalsvc.QuestionServiceInstance.FindBySectionID(sec.ID)
		for _, q := range questions {
			hasQuestions = true
			question := InterviewQuestionView{
				Model:    q,
				Required: q.IsRequired && q.Status == "active",
			}
			if answer, ok := answers[q.ID]; ok {
				saved := answer
				question.Answer = &saved
			}
			view.Questions = append(view.Questions, question)
		}
		sectionViews = append(sectionViews, view)
	}

	data := InterviewFeedbackPageData{
		Interview:             interview,
		Feedback:              fb,
		Panel:                 panel,
		Sections:              sectionViews,
		TemplateName:          templateName,
		HasQuestions:          hasQuestions,
		SelectedInterviewerID: selectedInterviewerID,
		CanViewAll:            canViewPanel,
		CanApprove:            canApprove,
		CanManage:             canManage,
		// Only a panel member who has the score permission may edit their
		// own feedback row. A manager who is not on the panel can view all
		// scores but cannot submit one attributed to someone else.
		CanEdit: core.HasAnyPermission(c, enums.PermissionInterviewScore) && isPanelMember,
	}
	if canViewPanel {
		data.AllFeedback, _ = recsvc.InterviewFeedbackServiceInst.ListByInterviewID(interviewID)
		data.AggregateScore = averageFeedbackScore(data.AllFeedback)
		data.AggregateTotalScore = totalFeedbackScore(data.AllFeedback)
		if summary, err := recsvc.InterviewServiceInst.GetCandidateSummary(interview.CandidateID); err == nil && summary.CompletedReviews > 0 {
			data.CandidateFinalTotal = &summary.TotalScore
		}
	}

	return core.RenderPage(c, "page-interview-feedback", data)
}

func (h *RecruitmentHandler) SubmitInterviewFeedback(c *fiber.Ctx) error {
	interviewID, _ := strconv.ParseInt(c.Params("id"), 10, 64)
	interview, err := recsvc.InterviewServiceInst.FindByID(interviewID)
	if err != nil {
		return core.RenderError(c, fiber.StatusNotFound, "Interview not found.")
	}

	uid := core.CurrentUserID(c)
	panel, _ := recsvc.InterviewServiceInst.FindPanelMembers(interviewID)
	isPanelMember := panelContains(panel, uid)
	canManage := core.CanManageInterviews(c)

	// Each interviewer scores for themselves only. Nobody — not even a
	// manager — may submit feedback attributed to another person.
	if !isPanelMember {
		if !canManage {
			return core.Forbidden(c, "Only the assigned panel can score this interview.")
		}
		// A manager who is not on the panel can view the interview but
		// cannot submit feedback attributed to someone else.
		return core.Forbidden(c, "You are not on the panel for this interview and cannot submit a score.")
	}
	if !core.HasAnyPermission(c, enums.PermissionInterviewScore) && !canManage {
		return core.Forbidden(c, "Only the assigned panel can score this interview.")
	}

	// Always record the feedback against the logged-in user's own row.
	interviewerID := uid

	recommendation := c.FormValue("recommendation")
	strengths := strings.TrimSpace(c.FormValue("strengths"))
	weaknesses := strings.TrimSpace(c.FormValue("weaknesses"))
	comments := strings.TrimSpace(c.FormValue("comments"))

	now := time.Now()

	type answerInput struct {
		question evalmodel.EvaluationQuestion
		answer   string
		score    *float64
	}

	// Read every question first: the feedback's single score is derived from
	// the per-question scores, so it must be known before the row is written.
	round, _ := recsvc.InterviewRoundServiceInst.FindByID(interview.InterviewRoundID)
	inputs := make([]answerInput, 0)
	var totalWeightedScore, totalWeight float64
	for _, sec := range loadInterviewSectionsForFeedback(interviewID, round) {
		questions, _ := evalsvc.QuestionServiceInstance.FindBySectionID(sec.ID)
		for _, q := range questions {
			answerRaw := c.FormValue(fmt.Sprintf("q_%d", q.ID))
			scoreRaw := c.FormValue(fmt.Sprintf("q_score_%d", q.ID))
			// Rating questions use their answer input as the score. Other
			// question types keep the answer and the score separate.
			if strings.TrimSpace(scoreRaw) == "" && isRatedQuestion(q) {
				scoreRaw = answerRaw
			}
			score, err := parseQuestionScore(scoreRaw)
			if err != nil {
				return core.RenderError(c, fiber.StatusBadRequest,
					fmt.Sprintf("Question %q: %v.", q.Question, err))
			}
			inputs = append(inputs, answerInput{question: q, answer: answerRaw, score: score})
			if score != nil && q.Weight > 0 {
				totalWeightedScore += *score * q.Weight
				totalWeight += q.Weight
			}
		}
	}

	var overall *float64
	if totalWeight > 0 {
		value := totalWeightedScore / totalWeight
		overall = &value
	} else if len(inputs) == 0 {
		// A round without questions has nothing to score per question, so it
		// can still be rated directly.
		overall, err = parseQuestionScore(c.FormValue("score"))
		if err != nil {
			return core.RenderError(c, fiber.StatusBadRequest, fmt.Sprintf("Overall score: %v.", err))
		}
	}

	fb := &recmodel.InterviewFeedback{
		InterviewID:    interviewID,
		InterviewerID:  interviewerID,
		RecordedByID:   uid,
		Score:          overall,
		Recommendation: recommendation,
		Strengths:      strengths,
		Weaknesses:     weaknesses,
		Comments:       comments,
		SubmittedAt:    &now,
	}

	existing, err := recsvc.InterviewFeedbackServiceInst.FindByInterviewAndInterviewer(interviewID, interviewerID)
	if err == nil && existing.ID > 0 {
		fb.ID = existing.ID
		if err := recsvc.InterviewFeedbackServiceInst.Update(fb); err != nil {
			log.Printf("[SubmitInterviewFeedback] update interview %d interviewer %d: %v", interviewID, interviewerID, err)
			return core.RenderError(c, fiber.StatusInternalServerError, "Failed to save feedback.")
		}
	} else {
		if err := recsvc.InterviewFeedbackServiceInst.Create(fb); err != nil {
			log.Printf("[SubmitInterviewFeedback] create interview %d interviewer %d: %v", interviewID, interviewerID, err)
			return core.RenderError(c, fiber.StatusInternalServerError, "Failed to save feedback.")
		}
	}

	// Save the round's question answers only after the feedback row exists:
	// every answer points at that row.
	answerFailed := false
	for _, input := range inputs {
		if err := saveInterviewAnswer(fb.ID, input.question, input.answer, input.score); err != nil {
			answerFailed = true
			log.Printf("[SubmitInterviewFeedback] answer for interview %d question %d: %v",
				interviewID, input.question.ID, err)
		}
	}
	if answerFailed {
		return core.RenderError(c, fiber.StatusInternalServerError,
			"Feedback was saved, but some question answers could not be stored. Please save again.")
	}

	// Saving feedback is the evaluation: score the round evaluation from the
	// answers just given and close it. The other panel members keep their own
	// open rows, which start when the interview is marked completed.
	if err := evalsvc.AssignmentServiceInstance.StartForInterview(interviewID); err != nil {
		log.Printf("[SubmitInterviewFeedback] start round evaluations for interview %d: %v", interviewID, err)
	}
	if err := evalsvc.AssignmentServiceInstance.CompleteRoundFromFeedback(interviewID, interviewerID); err != nil {
		log.Printf("[SubmitInterviewFeedback] complete round evaluation for interview %d interviewer %d: %v", interviewID, interviewerID, err)
	}

	c.Request().URI().SetQueryString(fmt.Sprintf("interviewer_id=%d", interviewerID))
	core.TriggerToast(c, "Interview score saved.", "success")
	return h.InterviewFeedback(c)
}

// isRatedQuestion reports whether the question's own answer is its score.
// Text, number, and yes/no questions keep a separate 0-10 score.
func isRatedQuestion(q evalmodel.EvaluationQuestion) bool {
	switch enums.QuestionType(q.QuestionType) {
	case enums.QuestionTypeText, enums.QuestionTypeNumber, enums.QuestionTypeBoolean:
		return false
	default:
		return true
	}
}

// parseQuestionScore reads an optional 0-10 score. A blank value means the
// question was not scored, which keeps it out of the weighted average.
func parseQuestionScore(raw string) (*float64, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	value, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return nil, fmt.Errorf("must be a number")
	}
	if value < 0 || value > 10 {
		return nil, fmt.Errorf("must be between 0 and 10")
	}
	return &value, nil
}

// saveInterviewAnswer stores one question answer and its score, keeping the
// value columns aligned with the question type. Blank inputs are saved as
// NULL/empty so clearing an answer on a later save actually clears it.
func saveInterviewAnswer(feedbackID int64, q evalmodel.EvaluationQuestion, raw string, score *float64) error {
	answer := &recmodel.InterviewFeedbackAnswer{
		FeedbackID:           feedbackID,
		EvaluationQuestionID: q.ID,
		RatingValue:          score,
	}
	raw = strings.TrimSpace(raw)

	switch enums.QuestionType(q.QuestionType) {
	case enums.QuestionTypeText:
		answer.TextValue = raw
	case enums.QuestionTypeNumber:
		if raw != "" {
			v, err := strconv.ParseFloat(raw, 64)
			if err != nil {
				return fmt.Errorf("%q is not a number", raw)
			}
			answer.NumericValue = &v
		}
	case enums.QuestionTypeBoolean:
		if raw != "" {
			v := raw == "true"
			answer.BooleanValue = &v
		}
	default:
		// The rating question's answer is already stored as RatingValue.
	}

	return recsvc.InterviewFeedbackServiceInst.SaveAnswer(answer)
}


// Reports Page
func (h *RecruitmentHandler) RecruitmentReports(c *fiber.Ctx) error {
	return c.Redirect("/pages/reports")
}

// Suppress unused import
var _ = os.Remove
