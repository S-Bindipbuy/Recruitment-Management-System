package approval

import (
	"ems/app/enums"
	core "ems/app/handlers/pages/core"
	recmodel "ems/app/models/recruitment"
	recsvc "ems/app/services/recruitment"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type ApprovalHandler struct{}

type ApprovalsPageData struct {
	Rankings             []recmodel.CandidateEvaluationSummary
	Postings             []recmodel.JobPostingListItem
	SelectedJobPostingID int64
}

func (h *ApprovalHandler) Approvals(c *fiber.Ctx) error {
	if err := core.RequirePermission(c,
		enums.PermissionEvaluationApprove,
		enums.PermissionEvaluationManage,
		enums.PermissionInterviewApprove,
		enums.PermissionRecruitmentManage,
	); err != nil {
		return err
	}

	postingID, _ := strconv.ParseInt(c.Query("job_posting_id"), 10, 64)
	if postingID == 0 {
		// Approve/reject/accept POST back to the same page; the filter rides
		// along in a hidden field so the decision keeps the current view.
		postingID, _ = strconv.ParseInt(c.FormValue("job_posting_id"), 10, 64)
	}
	postings, _ := recsvc.JobPostingServiceInst.FindAll()
	data := ApprovalsPageData{
		Postings:             postings,
		SelectedJobPostingID: postingID,
	}

	// Filter the candidate ranking by job posting.
	rankings := core.LoadCandidateRankings()
	for _, summary := range rankings {
		if postingID == 0 || summary.JobPostingID == postingID {
			data.Rankings = append(data.Rankings, summary)
		}
	}
	return core.RenderPage(c, "page-approvals", data)
}

// AcceptCandidate is the final hiring decision based on the interview scores.
// It does not need a separate candidate-level evaluation or an approval
// request: the ranking already carries the combined panel score, and the
// approver simply accepts the candidate they choose.
func (h *ApprovalHandler) AcceptCandidate(c *fiber.Ctx) error {
	if err := core.RequirePermission(c,
		enums.PermissionEvaluationApprove,
		enums.PermissionEvaluationManage,
		enums.PermissionInterviewApprove,
		enums.PermissionRecruitmentManage,
	); err != nil {
		return err
	}

	candidateID, _ := strconv.ParseInt(c.Params("id"), 10, 64)
	if candidateID <= 0 {
		return core.RenderError(c, fiber.StatusBadRequest, "That candidate no longer exists.")
	}

	notice, err := core.HireCandidate(c, candidateID)
	if err != nil {
		core.TriggerToast(c, err.Error(), "error")
		return h.Approvals(c)
	}
	core.TriggerToast(c, notice, "success")
	return h.Approvals(c)
}

