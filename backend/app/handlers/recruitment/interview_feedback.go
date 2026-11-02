package recruitment

import (
	"ems/app/enums"
	"ems/app/handlers/authz"
	recmodel "ems/app/models/recruitment"
	recsvc "ems/app/services/recruitment"
	"github.com/gofiber/fiber/v2"
	"time"
)

type InterviewFeedbackHandler struct{}

// canScoreInterview applies the panel rule shared by the create/update
// handlers: a manager may score on anybody's behalf, an interviewer may only
// score their own row and must be on the panel.
func canScoreInterview(c *fiber.Ctx, interviewID, interviewerID int64) error {
	if authz.CanManageInterviews(c) {
		return nil
	}
	if !authz.CanAny(c, enums.PermissionInterviewScore) {
		return authz.Forbidden(c, "You do not have permission to score interviews.")
	}
	if interviewerID != authz.CurrentUserID(c) {
		return authz.Forbidden(c, "You can only score interviews as yourself.")
	}
	member, _ := recsvc.InterviewServiceInst.IsPanelMember(interviewID, interviewerID)
	if !member {
		return authz.Forbidden(c, "Only the assigned panel can score this interview.")
	}
	return nil
}

// Get returns the feedback rows recorded on an interview. Access is limited to
// the interview panel and the evaluator assigned to that candidate. A panel
// member receives their own row unless a wider interview scope grants the whole
// panel; an assigned evaluator receives the whole panel so they can evaluate
// the candidate.
func (h *InterviewFeedbackHandler) Get(c *fiber.Ctx) error {
	if err := authz.RequirePermission(c,
		enums.PermissionRecruitmentManage,
		enums.PermissionInterviewView,
		enums.PermissionInterviewScore,
		enums.PermissionInterviewAssign,
		enums.PermissionInterviewApprove,
		enums.PermissionEvaluationAssign,
		enums.PermissionEvaluationEvaluate,
	); err != nil {
		return err
	}

	interviewID, _ := c.ParamsInt("interview_id")
	interview, err := recsvc.InterviewServiceInst.FindByID(int64(interviewID))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Interview not found"})
	}

	uid := authz.CurrentUserID(c)
	canViewAll := authz.CanViewAllInterviews(c)
	member, _ := recsvc.InterviewServiceInst.IsPanelMember(interview.ID, uid)
	assigned := false
	if !member {
		candidateIDs, _ := recsvc.InterviewServiceInst.FindAssignedCandidateIDs(uid)
		for _, id := range candidateIDs {
			if id == interview.CandidateID {
				assigned = true
				break
			}
		}
	}
	// A wider interview permission grants access, and a user without it can
	// still read feedback when they are on this panel or are the evaluator
	// assigned to the candidate.
	if !member && !assigned && !canViewAll {
		return authz.Forbidden(c, "This interview is not assigned to you.")
	}

	all, err := recsvc.InterviewFeedbackServiceInst.ListByInterviewID(int64(interviewID))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch feedback"})
	}

	visible := make([]recmodel.InterviewFeedback, 0, len(all))
	for _, fb := range all {
		// The assigned evaluator and privileged viewers see the whole panel; a
		// plain interviewer only receives their own row.
		if canViewAll || assigned || fb.InterviewerID == uid {
			visible = append(visible, fb)
		}
	}
	return c.JSON(visible)
}

func (h *InterviewFeedbackHandler) Create(c *fiber.Ctx) error {
	interviewID, _ := c.ParamsInt("interview_id")

	var req struct {
		InterviewID    int64    `json:"interview_id"`
		InterviewerID  int64    `json:"interviewer_id"`
		Score          *float64 `json:"score"`
		Recommendation string   `json:"recommendation"`
		Strengths      string   `json:"strengths"`
		Weaknesses     string   `json:"weaknesses"`
		Comments       string   `json:"comments"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	if interviewID <= 0 {
		interviewID = int(req.InterviewID)
	}

	interview, err := recsvc.InterviewServiceInst.FindByID(int64(interviewID))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Interview not found"})
	}

	userID := authz.CurrentUserID(c)
	interviewerID := req.InterviewerID
	if interviewerID <= 0 {
		interviewerID = userID
	}
	if err := canScoreInterview(c, interview.ID, interviewerID); err != nil {
		return err
	}

	now := time.Now()
	fb := &recmodel.InterviewFeedback{
		InterviewID:    interview.ID,
		InterviewerID:  interviewerID,
		RecordedByID:   userID,
		Score:          req.Score,
		Recommendation: req.Recommendation,
		Strengths:      req.Strengths,
		Weaknesses:     req.Weaknesses,
		Comments:       req.Comments,
		SubmittedAt:    &now,
	}

	// Each panel member keeps one row per interview; a second submission is an
	// update rather than a duplicate-key failure.
	if existing, err := recsvc.InterviewFeedbackServiceInst.FindByInterviewAndInterviewer(interview.ID, interviewerID); err == nil && existing.ID > 0 {
		fb.ID = existing.ID
		if err := recsvc.InterviewFeedbackServiceInst.Update(fb); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to save feedback"})
		}
		return c.JSON(fb)
	}

	if err := recsvc.InterviewFeedbackServiceInst.Create(fb); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to save feedback"})
	}
	return c.Status(fiber.StatusCreated).JSON(fb)
}

func (h *InterviewFeedbackHandler) Update(c *fiber.Ctx) error {
	id, _ := c.ParamsInt("id")
	var req struct {
		InterviewerID  int64    `json:"interviewer_id"`
		Score          *float64 `json:"score"`
		Recommendation string   `json:"recommendation"`
		Strengths      string   `json:"strengths"`
		Weaknesses     string   `json:"weaknesses"`
		Comments       string   `json:"comments"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	fb, err := recsvc.InterviewFeedbackServiceInst.FindByID(int64(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Feedback not found"})
	}

	// A manager may reassign the row to another panel member; an interviewer
	// may only edit the row they recorded themselves. Without this check a
	// panel member could swap the row's interviewer_id to their own and
	// overwrite a colleague's score.
	uid := authz.CurrentUserID(c)
	canManage := authz.CanManageInterviews(c)
	if !canManage && fb.InterviewerID != uid {
		return authz.Forbidden(c, "You can only edit your own interview score.")
	}
	interviewerID := fb.InterviewerID
	if req.InterviewerID > 0 {
		if req.InterviewerID != fb.InterviewerID && !canManage {
			return authz.Forbidden(c, "You can only edit your own interview score.")
		}
		interviewerID = req.InterviewerID
	}
	if err := canScoreInterview(c, fb.InterviewID, interviewerID); err != nil {
		return err
	}

	fb.InterviewerID = interviewerID
	fb.Score = req.Score
	fb.Recommendation = req.Recommendation
	fb.Strengths = req.Strengths
	fb.Weaknesses = req.Weaknesses
	fb.Comments = req.Comments

	if err := recsvc.InterviewFeedbackServiceInst.Update(fb); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update feedback"})
	}
	return c.JSON(fb)
}
