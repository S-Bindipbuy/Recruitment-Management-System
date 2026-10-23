package recruitment

import (
	recmodel "ems/app/models/recruitment"
	recrepo "ems/app/repositories/recruitment"
)

type InterviewService struct{}

func (s *InterviewService) Create(interview *recmodel.Interview) error {
	return recrepo.InterviewRepo.Create(interview)
}

func (s *InterviewService) FindAll() ([]recmodel.InterviewListItem, error) {
	return recrepo.InterviewRepo.FindAll()
}

func (s *InterviewService) FindByID(id int64) (*recmodel.InterviewListItem, error) {
	return recrepo.InterviewRepo.FindByID(id)
}

func (s *InterviewService) FindByCandidateID(candidateID int64) ([]recmodel.InterviewListItem, error) {
	return recrepo.InterviewRepo.FindByCandidateID(candidateID)
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

func (s *InterviewService) GetAllCandidateSummaries() ([]recmodel.CandidateEvaluationSummary, error) {
	return recrepo.InterviewRepo.GetAllCandidateSummaries()
}

func (s *InterviewService) GetInterviewerActivity() ([]recmodel.InterviewerActivity, error) {
	return recrepo.InterviewRepo.GetInterviewerActivity()
}

func (s *InterviewService) GetHiringDecisions() ([]recmodel.HiringDecision, error) {
	return recrepo.InterviewRepo.GetHiringDecisions()
}
