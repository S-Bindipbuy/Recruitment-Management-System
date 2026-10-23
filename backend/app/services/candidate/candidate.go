package candidate

import (
	candmodel "ems/app/models/candidate"
	candrepo "ems/app/repositories/candidate"
)

type CandidateService struct{}

func (s *CandidateService) Create(candidate *candmodel.Candidate) error {
	return candrepo.CandidateRepo.Create(candidate)
}

func (s *CandidateService) FindAll() ([]candmodel.CandidateListItem, error) {
	return candrepo.CandidateRepo.FindAll()
}

func (s *CandidateService) FindByID(id int64) (*candmodel.Candidate, error) {
	return candrepo.CandidateRepo.FindByID(id)
}

func (s *CandidateService) FindByJobPostingID(postingID int64) ([]candmodel.CandidateListItem, error) {
	return candrepo.CandidateRepo.FindByJobPostingID(postingID)
}

func (s *CandidateService) UpdateStatus(id int64, status string) error {
	return candrepo.CandidateRepo.UpdateStatus(id, status)
}

// Hire converts an eligible candidate into an active user and employee, and
// assigns the selected application role as part of the same transaction.
func (s *CandidateService) Hire(candidateID, evaluatorID, roleID int64, rating float64, notes, passwordHash string) error {
	return candrepo.CandidateRepo.Hire(candidateID, evaluatorID, roleID, rating, notes, passwordHash)
}

// FindHireableOptions lists the applicants still in the pipeline - everybody
// the wizard or the assignment dialog could be pointed at.
func (s *CandidateService) FindHireableOptions() ([]candmodel.CandidateOption, error) {
	return candrepo.CandidateRepo.FindHireableOptions()
}

// FindNames returns the display names behind a set of candidate ids in one query.
func (s *CandidateService) FindNames(ids []int64) (map[int64]string, error) {
	return candrepo.CandidateRepo.FindNames(ids)
}
