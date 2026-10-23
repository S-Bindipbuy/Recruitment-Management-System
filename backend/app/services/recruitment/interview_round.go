package recruitment

import (
	recmodel "ems/app/models/recruitment"
	recrepo "ems/app/repositories/recruitment"
)

type InterviewRoundService struct{}

func (s *InterviewRoundService) Create(round *recmodel.InterviewRound) error {
	return recrepo.InterviewRoundRepo.Create(round)
}

func (s *InterviewRoundService) FindByJobPostingID(postingID int64) ([]recmodel.InterviewRound, error) {
	return recrepo.InterviewRoundRepo.FindByJobPostingID(postingID)
}

func (s *InterviewRoundService) Delete(id int64) error {
	return recrepo.InterviewRoundRepo.Delete(id)
}
