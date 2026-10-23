package recruitment

import (
	recmodel "ems/app/models/recruitment"
	recrepo "ems/app/repositories/recruitment"
)

type InterviewFeedbackService struct{}

func (s *InterviewFeedbackService) Create(fb *recmodel.InterviewFeedback) error {
	return recrepo.InterviewFeedbackRepo.Create(fb)
}

func (s *InterviewFeedbackService) FindByInterviewID(interviewID int64) (*recmodel.InterviewFeedback, error) {
	return recrepo.InterviewFeedbackRepo.FindByInterviewID(interviewID)
}

func (s *InterviewFeedbackService) Update(fb *recmodel.InterviewFeedback) error {
	return recrepo.InterviewFeedbackRepo.Update(fb)
}
