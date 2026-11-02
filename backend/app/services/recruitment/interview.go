package recruitment

import (
	recmodel "ems/app/models/recruitment"
	recrepo "ems/app/repositories/recruitment"
	candsvc "ems/app/services/candidate"
	"errors"
)

// ErrRoundNotForCandidate is returned when the chosen round belongs to a
// different job posting than the candidate applied for.
var ErrRoundNotForCandidate = errors.New("interview round does not belong to the candidate's job posting")

// InterviewFilter is the repository filter exposed to page handlers without
// making them import the repository package.
type InterviewFilter = recrepo.InterviewFilter

type InterviewService struct{}

func (s *InterviewService) Create(interview *recmodel.Interview) error {
	return recrepo.InterviewRepo.Create(interview)
}

// ValidateCandidateRound keeps a scheduled interview inside the candidate's
// job posting. The scheduling picker already filters the round list, but the
// same rule has to hold for direct API calls.
func (s *InterviewService) ValidateCandidateRound(candidateID, roundID int64) error {
	candidate, err := candsvc.CandidateServiceInstance.FindByID(candidateID)
	if err != nil {
		return err
	}
	round, err := recrepo.InterviewRoundRepo.FindByID(roundID)
	if err != nil {
		return err
	}
	if candidate.JobPostingID != nil && *candidate.JobPostingID != round.JobPostingID {
		return ErrRoundNotForCandidate
	}
	return nil
}

func (s *InterviewService) FindAll() ([]recmodel.InterviewListItem, error) {
	return recrepo.InterviewRepo.FindAll()
}

// FindFiltered returns schedules narrowed by the viewer's scope and the page
// filters. The SQL lives in the repository so large pipelines stay fast.
func (s *InterviewService) FindFiltered(filter recrepo.InterviewFilter) ([]recmodel.InterviewListItem, error) {
	return recrepo.InterviewRepo.FindFiltered(filter)
}

func (s *InterviewService) FindPanelMembers(interviewID int64) ([]recmodel.InterviewPanelMember, error) {
	return recrepo.InterviewRepo.FindPanelMembers(interviewID)
}

func (s *InterviewService) IsPanelMember(interviewID, userID int64) (bool, error) {
	return recrepo.InterviewRepo.IsPanelMember(interviewID, userID)
}

func (s *InterviewService) FindAssignedCandidateIDs(userID int64) ([]int64, error) {
	return recrepo.InterviewRepo.FindAssignedCandidateIDs(userID)
}

func (s *InterviewService) FindByID(id int64) (*recmodel.InterviewListItem, error) {
	return recrepo.InterviewRepo.FindByID(id)
}

func (s *InterviewService) FindByCandidateID(candidateID int64) ([]recmodel.InterviewListItem, error) {
	return recrepo.InterviewRepo.FindByCandidateID(candidateID)
}

// FindAwaitingFeedback lists interviews that still have no score recorded.
func (s *InterviewService) FindAwaitingFeedback() ([]recmodel.InterviewListItem, error) {
	return recrepo.InterviewRepo.FindAwaitingFeedback()
}

func (s *InterviewService) UpdateStatus(id int64, status string) error {
	return recrepo.InterviewRepo.UpdateStatus(id, status)
}

func (s *InterviewService) Delete(id int64) error {
	return recrepo.InterviewRepo.Delete(id)
}

func (s *InterviewService) GetCandidateSummary(candidateID int64) (*recmodel.CandidateEvaluationSummary, error) {
	return recrepo.InterviewRepo.GetCandidateSummary(candidateID)
}

// CandidateInterviewReady reports whether every active round of the candidate's
// posting has been scored. Evaluation work starts only after that: the
// evaluator needs the full interview picture, not a partially interviewed
// candidate.
func (s *InterviewService) CandidateInterviewReady(candidateID int64) (bool, error) {
	summary, err := s.GetCandidateSummary(candidateID)
	if err != nil {
		return false, err
	}
	return InterviewSummaryReady(summary), nil
}

// InterviewSummaryReady applies the readiness rule to a summary that is
// already loaded, so a page showing many candidates does not re-query each
// one. A posting without configured rounds is ready once at least one
// interview has been scored.
func InterviewSummaryReady(s *recmodel.CandidateEvaluationSummary) bool {
	if s == nil || s.CompletedReviews == 0 {
		return false
	}
	return s.TotalRounds == 0 || s.CompletedRounds >= s.TotalRounds
}

func (s *InterviewService) GetAllCandidateSummaries() ([]recmodel.CandidateEvaluationSummary, error) {
	return recrepo.InterviewRepo.GetAllCandidateSummaries()
}

func (s *InterviewService) GetInterviewerActivity(jobPostingID int64) ([]recmodel.InterviewerActivity, error) {
	return recrepo.InterviewRepo.GetInterviewerActivity(jobPostingID)
}

func (s *InterviewService) GetHiringDecisions() ([]recmodel.HiringDecision, error) {
	return recrepo.InterviewRepo.GetHiringDecisions()
}
