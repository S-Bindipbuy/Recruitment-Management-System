package recruitment

import (
	recmodel "ems/app/models/recruitment"
	"ems/database"
)

type InterviewFeedbackRepository struct{}

func (r *InterviewFeedbackRepository) Create(fb *recmodel.InterviewFeedback) error {
	result, err := database.DB.Exec(
		`INSERT INTO interview_feedback (interview_id, interviewer_id, score, recommendation,
		 strengths, weaknesses, comments, submitted_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		fb.InterviewID, fb.InterviewerID, fb.Score, fb.Recommendation,
		fb.Strengths, fb.Weaknesses, fb.Comments, fb.SubmittedAt,
	)
	if err != nil {
		return err
	}
	fb.ID, _ = result.LastInsertId()
	return nil
}

func (r *InterviewFeedbackRepository) FindByInterviewID(interviewID int64) (*recmodel.InterviewFeedback, error) {
	fb := &recmodel.InterviewFeedback{}
	err := database.DB.Get(fb,
		`SELECT id, interview_id, interviewer_id, score, recommendation,
		 strengths, weaknesses, comments, submitted_at, created_at, updated_at
		 FROM interview_feedback WHERE interview_id = ?`, interviewID,
	)
	if err != nil {
		return nil, err
	}
	return fb, nil
}

func (r *InterviewFeedbackRepository) Update(fb *recmodel.InterviewFeedback) error {
	_, err := database.DB.Exec(
		`UPDATE interview_feedback SET score=?, recommendation=?, strengths=?, weaknesses=?,
		 comments=?, submitted_at=?, updated_at=CURRENT_TIMESTAMP WHERE id=?`,
		fb.Score, fb.Recommendation, fb.Strengths, fb.Weaknesses, fb.Comments, fb.SubmittedAt, fb.ID,
	)
	return err
}

// Report Queries
func (r *InterviewRepository) GetCandidateSummary(candidateID int64) (*recmodel.CandidateEvaluationSummary, error) {
	s := &recmodel.CandidateEvaluationSummary{}
	err := database.DB.Get(s,
		`SELECT c.id AS candidate_id, c.full_name AS candidate_name, COALESCE(p.name, '') AS position_name,
		 COUNT(DISTINCT i.id) AS total_interviews,
		 SUM(CASE WHEN ifb.score IS NOT NULL THEN 1 ELSE 0 END) AS completed_reviews,
		 COALESCE(AVG(ifb.score), 0) AS average_score,
		 COALESCE(MAX(ifb.score), 0) AS highest_score,
		 COALESCE(MIN(ifb.score), 0) AS lowest_score,
		 CASE WHEN AVG(ifb.score) >= 7 THEN 'hire'
		      WHEN AVG(ifb.score) >= 5 THEN 'maybe'
		      ELSE 'reject' END AS final_decision
		 FROM candidates c
		 LEFT JOIN job_postings jp ON jp.id = c.job_posting_id
		 LEFT JOIN positions p ON p.id = jp.position_id
		 LEFT JOIN interviews i ON i.candidate_id = c.id
		 LEFT JOIN interview_feedback ifb ON ifb.interview_id = i.id
		 WHERE c.id = ?
		 GROUP BY c.id`, candidateID,
	)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (r *InterviewRepository) GetAllCandidateSummaries() ([]recmodel.CandidateEvaluationSummary, error) {
	var summaries []recmodel.CandidateEvaluationSummary
	err := database.DB.Select(&summaries,
		`SELECT c.id AS candidate_id, c.full_name AS candidate_name, COALESCE(p.name, '') AS position_name,
		 COUNT(DISTINCT i.id) AS total_interviews,
		 SUM(CASE WHEN ifb.score IS NOT NULL THEN 1 ELSE 0 END) AS completed_reviews,
		 COALESCE(AVG(ifb.score), 0) AS average_score,
		 COALESCE(MAX(ifb.score), 0) AS highest_score,
		 COALESCE(MIN(ifb.score), 0) AS lowest_score,
		 CASE WHEN AVG(ifb.score) >= 7 THEN 'hire'
		      WHEN AVG(ifb.score) >= 5 THEN 'maybe'
		      ELSE 'reject' END AS final_decision
		 FROM candidates c
		 LEFT JOIN job_postings jp ON jp.id = c.job_posting_id
		 LEFT JOIN positions p ON p.id = jp.position_id
		 LEFT JOIN interviews i ON i.candidate_id = c.id
		 LEFT JOIN interview_feedback ifb ON ifb.interview_id = i.id
		 GROUP BY c.id ORDER BY AVG(ifb.score) DESC`,
	)
	if err != nil {
		return nil, err
	}
	return summaries, nil
}

func (r *InterviewRepository) GetInterviewerActivity() ([]recmodel.InterviewerActivity, error) {
	var activities []recmodel.InterviewerActivity
	err := database.DB.Select(&activities,
		`SELECT u.id AS interviewer_id, u.name AS interviewer_name,
		 COUNT(DISTINCT i.id) AS total_interviews,
		 SUM(CASE WHEN ifb.score IS NOT NULL THEN 1 ELSE 0 END) AS completed_reviews,
		 SUM(CASE WHEN ifb.score IS NULL AND i.id IS NOT NULL THEN 1 ELSE 0 END) AS pending_reviews,
		 COALESCE(AVG(ifb.score), 0) AS average_score
		 FROM users u
		 JOIN interviews i ON i.interviewer_id = u.id
		 LEFT JOIN interview_feedback ifb ON ifb.interview_id = i.id
		 GROUP BY u.id ORDER BY COUNT(DISTINCT i.id) DESC`,
	)
	if err != nil {
		return nil, err
	}
	return activities, nil
}

func (r *InterviewRepository) GetHiringDecisions() ([]recmodel.HiringDecision, error) {
	var decisions []recmodel.HiringDecision
	err := database.DB.Select(&decisions,
		`SELECT c.id AS candidate_id, c.full_name AS candidate_name, COALESCE(p.name, '') AS position_name, COALESCE(d.name, '') AS department_name,
		 COALESCE(AVG(ifb.score), 0) AS average_score,
		 COUNT(DISTINCT i.id) AS interview_count,
		 c.status AS decision,
		 COALESCE(c.updated_at, '') AS decision_date,
		 COALESCE((SELECT u.name FROM users u ORDER BY u.id LIMIT 1), 'System') AS decided_by
		 FROM candidates c
		 LEFT JOIN job_postings jp ON jp.id = c.job_posting_id
		 LEFT JOIN positions p ON p.id = jp.position_id
		 LEFT JOIN departments d ON d.id = p.department_id
		 LEFT JOIN interviews i ON i.candidate_id = c.id
		 LEFT JOIN interview_feedback ifb ON ifb.interview_id = i.id
		 GROUP BY c.id ORDER BY AVG(ifb.score) DESC`,
	)
	if err != nil {
		return nil, err
	}
	return decisions, nil
}
