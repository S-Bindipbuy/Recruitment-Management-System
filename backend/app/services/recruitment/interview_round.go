package recruitment

import (
	evalmodel "ems/app/models/evaluation"
	recmodel "ems/app/models/recruitment"
	recrepo "ems/app/repositories/recruitment"
)

type InterviewRoundService struct{}

func (s *InterviewRoundService) Create(round *recmodel.InterviewRound) error {
	return recrepo.InterviewRoundRepo.Create(round)
}

func (s *InterviewRoundService) FindByID(id int64) (*recmodel.InterviewRound, error) {
	return recrepo.InterviewRoundRepo.FindByID(id)
}

func (s *InterviewRoundService) FindByJobPostingID(postingID int64) ([]recmodel.InterviewRound, error) {
	return recrepo.InterviewRoundRepo.FindByJobPostingID(postingID)
}

// SaveSections replaces the sections a round asks questions from. Callers
// validate that the sections belong to the round's template first.
func (s *InterviewRoundService) SaveSections(roundID int64, sectionIDs []int64) error {
	return recrepo.InterviewRoundRepo.ReplaceSections(roundID, sectionIDs)
}

func (s *InterviewRoundService) FindSectionsByRoundID(roundID int64) ([]evalmodel.EvaluationSection, error) {
	return recrepo.InterviewRoundRepo.FindSectionsByRoundID(roundID)
}

func (s *InterviewRoundService) Delete(id int64) error {
	return recrepo.InterviewRoundRepo.Delete(id)
}

func (s *InterviewRoundService) Update(round *recmodel.InterviewRound) error {
	return recrepo.InterviewRoundRepo.Update(round)
}

func (s *InterviewRoundService) HasInterviews(id int64) (bool, error) {
	return recrepo.InterviewRoundRepo.HasInterviews(id)
}
