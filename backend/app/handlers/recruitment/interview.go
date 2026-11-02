package recruitment

import (
	"ems/app/enums"
	"ems/app/handlers/authz"
	recmodel "ems/app/models/recruitment"
	candsvc "ems/app/services/candidate"
	evalsvc "ems/app/services/evaluation"
	recsvc "ems/app/services/recruitment"
	"github.com/gofiber/fiber/v2"
	"log"
	"time"
)

type InterviewHandler struct{}

func (h *InterviewHandler) List(c *fiber.Ctx) error {
	if err := authz.RequirePermission(c,
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

	filter := recsvc.InterviewFilter{RestrictToViewer: !authz.CanViewAllInterviews(c)}
	if filter.RestrictToViewer {
		filter.CandidateIDs, _ = recsvc.InterviewServiceInst.FindAssignedCandidateIDs(authz.CurrentUserID(c))
		filter.PanelUserID = authz.CurrentUserID(c)
	}
	interviews, err := recsvc.InterviewServiceInst.FindFiltered(filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch interviews"})
	}
	return c.JSON(interviews)
}

// canAccessInterview applies the same scope as the schedule page: managers and
// approvers see everything, everybody else sees their own panel interviews or
// interviews for candidates they are assigned to evaluate.
func canAccessInterview(c *fiber.Ctx, interview *recmodel.InterviewListItem) bool {
	if authz.CanViewAllInterviews(c) {
		return true
	}
	uid := authz.CurrentUserID(c)
	member, _ := recsvc.InterviewServiceInst.IsPanelMember(interview.ID, uid)
	if member {
		return true
	}
	assigned, _ := recsvc.InterviewServiceInst.FindAssignedCandidateIDs(uid)
	for _, id := range assigned {
		if id == interview.CandidateID {
			return true
		}
	}
	return false
}

func (h *InterviewHandler) Get(c *fiber.Ctx) error {
	id, _ := c.ParamsInt("id")
	interview, err := recsvc.InterviewServiceInst.FindByID(int64(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Interview not found"})
	}
	if !canAccessInterview(c, interview) {
		return authz.Forbidden(c, "This interview is not assigned to you.")
	}
	return c.JSON(interview)
}

func (h *InterviewHandler) Create(c *fiber.Ctx) error {
	var req struct {
		CandidateID      int64   `json:"candidate_id"`
		CandidateIDs     []int64 `json:"candidate_ids"`
		InterviewRoundID int64   `json:"interview_round_id"`
		InterviewerID    int64   `json:"interviewer_id"`
		InterviewerIDs   []int64 `json:"interviewer_ids"`
		ScheduledAt      string  `json:"scheduled_at"`
		DurationMinutes  int     `json:"duration_minutes"`
		Location         string  `json:"location"`
		MeetingLink      string  `json:"meeting_link"`
		Notes            string  `json:"notes"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	candidates := uniqueIDs(req.CandidateIDs)
	if len(candidates) == 0 && req.CandidateID > 0 {
		candidates = []int64{req.CandidateID}
	}
	panel := uniqueIDs(req.InterviewerIDs)
	if len(panel) == 0 && req.InterviewerID > 0 {
		panel = []int64{req.InterviewerID}
	}
	if len(candidates) == 0 || req.InterviewRoundID <= 0 || len(panel) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "At least one candidate, a round, and one interviewer are required"})
	}

	// Managers schedule anybody. An assigned evaluator schedules only the
	// candidates whose evaluation is assigned to them.
	if !authz.CanManageInterviews(c) {
		if !authz.CanAny(c, enums.PermissionEvaluationEvaluate) {
			return authz.Forbidden(c, "You can only schedule interviews for candidates assigned to you.")
		}
		assigned, _ := recsvc.InterviewServiceInst.FindAssignedCandidateIDs(authz.CurrentUserID(c))
		allowed := make(map[int64]bool, len(assigned))
		for _, id := range assigned {
			allowed[id] = true
		}
		for _, candidateID := range candidates {
			if !allowed[candidateID] {
				return authz.Forbidden(c, "You can only schedule interviews for candidates assigned to you.")
			}
		}
	}

	for _, candidateID := range candidates {
		if err := recsvc.InterviewServiceInst.ValidateCandidateRound(candidateID, req.InterviewRoundID); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
	}

	scheduledAt, err := time.Parse("2006-01-02T15:04", req.ScheduledAt)
	if err != nil {
		scheduledAt, err = time.Parse("2006-01-02 15:04", req.ScheduledAt)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid date format. Use YYYY-MM-DD HH:MM"})
		}
	}

	if req.DurationMinutes <= 0 {
		req.DurationMinutes = 60
	}

	created := make([]*recmodel.Interview, 0, len(candidates))
	uid := authz.CurrentUserID(c)
	for _, candidateID := range candidates {
		interview := &recmodel.Interview{
			CandidateID:      candidateID,
			InterviewRoundID: req.InterviewRoundID,
			InterviewerID:    panel[0],
			InterviewerIDs:   panel,
			ScheduledAt:      scheduledAt,
			DurationMinutes:  req.DurationMinutes,
			Location:         req.Location,
			MeetingLink:      req.MeetingLink,
			Status:           "scheduled",
			Notes:            req.Notes,
		}
		if err := recsvc.InterviewServiceInst.Create(interview); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create interview"})
		}
		created = append(created, interview)
		if err := candsvc.CandidateServiceInstance.MarkInterviewing(candidateID, uid); err != nil {
			log.Printf("[CreateInterview] could not advance candidate %d: %v", candidateID, err)
		}
		if _, err := evalsvc.AssignmentServiceInstance.CreateForInterview(interview.ID); err != nil {
			log.Printf("[CreateInterview] round evaluation for interview %d: %v", interview.ID, err)
		}
	}

	if len(created) == 1 {
		return c.Status(fiber.StatusCreated).JSON(created[0])
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"created": len(created), "interviews": created})
}

func uniqueIDs(ids []int64) []int64 {
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

func (h *InterviewHandler) UpdateStatus(c *fiber.Ctx) error {
	id, _ := c.ParamsInt("id")
	var req struct {
		Status string `json:"status"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if !authz.CanManageInterviews(c) {
		if !authz.CanAny(c, enums.PermissionInterviewScore) {
			return authz.Forbidden(c, "You do not have permission to change this interview.")
		}
		member, _ := recsvc.InterviewServiceInst.IsPanelMember(int64(id), authz.CurrentUserID(c))
		if !member {
			return authz.Forbidden(c, "Only the assigned panel can update this interview.")
		}
	}
	if err := recsvc.InterviewServiceInst.UpdateStatus(int64(id), req.Status); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update status"})
	}
	if req.Status == "completed" {
		if err := evalsvc.AssignmentServiceInstance.StartForInterview(int64(id)); err != nil {
			log.Printf("[UpdateInterviewStatus] round evaluations for interview %d: %v", id, err)
		}
	}
	return c.JSON(fiber.Map{"message": "Status updated"})
}

func (h *InterviewHandler) Delete(c *fiber.Ctx) error {
	if err := authz.RequirePermission(c, enums.PermissionInterviewAssign, enums.PermissionRecruitmentManage); err != nil {
		return err
	}
	id, _ := c.ParamsInt("id")
	if err := recsvc.InterviewServiceInst.Delete(int64(id)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete interview"})
	}
	return c.JSON(fiber.Map{"message": "Deleted successfully"})
}

func (h *InterviewHandler) ListByCandidate(c *fiber.Ctx) error {
	candidateID, _ := c.ParamsInt("candidate_id")
	if !authz.CanViewAllInterviews(c) {
		uid := authz.CurrentUserID(c)
		assigned, _ := recsvc.InterviewServiceInst.FindAssignedCandidateIDs(uid)
		allowed := false
		for _, id := range assigned {
			if id == int64(candidateID) {
				allowed = true
				break
			}
		}
		if !allowed {
			interviews, _ := recsvc.InterviewServiceInst.FindByCandidateID(int64(candidateID))
			for _, interview := range interviews {
				member, _ := recsvc.InterviewServiceInst.IsPanelMember(interview.ID, uid)
				if member {
					allowed = true
					break
				}
			}
		}
		if !allowed {
			return authz.Forbidden(c, "This candidate is not assigned to you.")
		}
	}

	interviews, err := recsvc.InterviewServiceInst.FindByCandidateID(int64(candidateID))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch interviews"})
	}
	return c.JSON(interviews)
}
