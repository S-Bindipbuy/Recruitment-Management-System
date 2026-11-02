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

func (s *InterviewFeedbackService) FindByID(id int64) (*recmodel.InterviewFeedback, error) {
	return recrepo.InterviewFeedbackRepo.FindByID(id)
}

func (s *InterviewFeedbackService) FindByInterviewAndInterviewer(interviewID, interviewerID int64) (*recmodel.InterviewFeedback, error) {
	return recrepo.InterviewFeedbackRepo.FindByInterviewAndInterviewer(interviewID, interviewerID)
}

func (s *InterviewFeedbackService) ListByInterviewID(interviewID int64) ([]recmodel.InterviewFeedback, error) {
	return recrepo.InterviewFeedbackRepo.ListByInterviewID(interviewID)
}

func (s *InterviewFeedbackService) Update(fb *recmodel.InterviewFeedback) error {
	return recrepo.InterviewFeedbackRepo.Update(fb)
}

func (s *InterviewFeedbackService) SaveAnswer(answer *recmodel.InterviewFeedbackAnswer) error {
	return recrepo.InterviewFeedbackRepo.UpsertAnswer(answer)
}

func (s *InterviewFeedbackService) FindAnswersByFeedbackID(feedbackID int64) ([]recmodel.InterviewFeedbackAnswer, error) {
	return recrepo.InterviewFeedbackRepo.FindAnswersByFeedbackID(feedbackID)
}
