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

func (s *CandidateService) FindByIDs(ids []int64) ([]candmodel.CandidateListItem, error) {
	return candrepo.CandidateRepo.FindByIDs(ids)
}

func (s *CandidateService) UpdateStatus(id int64, status string) error {
	return candrepo.CandidateRepo.UpdateStatus(id, status)
}

// SetStatus moves a candidate through the pipeline and records who changed it
// and why.
func (s *CandidateService) SetStatus(id int64, status, reason string, changedBy int64) error {
	return candrepo.CandidateRepo.SetStatus(id, status, reason, changedBy)
}

// MarkInterviewing advances a fresh applicant when HR schedules an interview.
func (s *CandidateService) MarkInterviewing(id, changedBy int64) error {
	return candrepo.CandidateRepo.MarkInterviewing(id, changedBy)
}

// FindStatusHistory returns the audit trail for a candidate's status changes.
func (s *CandidateService) FindStatusHistory(id int64) ([]candmodel.StatusHistory, error) {
	return candrepo.CandidateRepo.FindStatusHistory(id)
}

// LogContact records a contact attempt made outside the system.
func (s *CandidateService) LogContact(contact *candmodel.ContactLog) error {
	return candrepo.CandidateRepo.CreateContact(contact)
}

// FindContacts returns a candidate's contact history, newest first.
func (s *CandidateService) FindContacts(id int64) ([]candmodel.ContactLog, error) {
	return candrepo.CandidateRepo.FindContactsByCandidate(id)
}

// FindActiveContacts returns contact rows for candidates still in the pipeline.
func (s *CandidateService) FindActiveContacts() ([]candmodel.ContactLog, error) {
	return candrepo.CandidateRepo.FindContactsForActiveCandidates()
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

// FindEvaluationCandidateOptions lists applicants who do not already have a
// submitted or completed evaluation, for the evaluation picker.
func (s *CandidateService) FindEvaluationCandidateOptions() ([]candmodel.CandidateOption, error) {
	return candrepo.CandidateRepo.FindEvaluationCandidateOptions()
}

// FindNames returns the display names behind a set of candidate ids in one query.
func (s *CandidateService) FindNames(ids []int64) (map[int64]string, error) {
	return candrepo.CandidateRepo.FindNames(ids)
}
