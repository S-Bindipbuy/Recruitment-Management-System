package recruitment

import (
	recmodel "ems/app/models/recruitment"
	recrepo "ems/app/repositories/recruitment"
)

type JobPostingService struct{}

func (s *JobPostingService) Create(posting *recmodel.JobPosting) error {
	return recrepo.JobPostingRepo.Create(posting)
}

func (s *JobPostingService) FindAll() ([]recmodel.JobPostingListItem, error) {
	return recrepo.JobPostingRepo.FindAll()
}

func (s *JobPostingService) FindByID(id int64) (*recmodel.JobPosting, error) {
	return recrepo.JobPostingRepo.FindByID(id)
}

func (s *JobPostingService) Update(id int64, title, description, requirements, skillsRequired string, salaryMin, salaryMax *float64, employmentType, status string) error {
	return recrepo.JobPostingRepo.Update(id, title, description, requirements, skillsRequired, salaryMin, salaryMax, employmentType, status)
}

func (s *JobPostingService) Delete(id int64) error {
	return recrepo.JobPostingRepo.Delete(id)
}

func (s *JobPostingService) Publish(id int64) error {
	return recrepo.JobPostingRepo.Publish(id)
}

func (s *JobPostingService) Close(id int64) error {
	return recrepo.JobPostingRepo.Close(id)
}

func (s *JobPostingService) Reopen(id int64) error {
	return recrepo.JobPostingRepo.Reopen(id)
}
