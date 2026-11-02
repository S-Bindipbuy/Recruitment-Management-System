package recruitment

import (
	recmodel "ems/app/models/recruitment"
	"ems/database"
)

type InterviewFeedbackRepository struct{}

func (r *InterviewFeedbackRepository) Create(fb *recmodel.InterviewFeedback) error {
	result, err := database.DB.Exec(
		`INSERT INTO interview_feedback (interview_id, interviewer_id, recorded_by_user_id, score,
		 recommendation, strengths, weaknesses, comments, submitted_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		fb.InterviewID, fb.InterviewerID, fb.RecordedByID, fb.Score, fb.Recommendation,
		fb.Strengths, fb.Weaknesses, fb.Comments, fb.SubmittedAt,
	)
	if err != nil {
		return err
	}
	fb.ID, _ = result.LastInsertId()
	return nil
}

const interviewFeedbackColumns = `f.id, f.interview_id, f.interviewer_id,
	COALESCE(iu.name, '') AS interviewer_name,
	COALESCE(f.recorded_by_user_id, 0) AS recorded_by_user_id,
	COALESCE(u.name, '') AS recorded_by_name,
	f.score, COALESCE(f.recommendation, '') AS recommendation,
	COALESCE(f.strengths, '') AS strengths,
	COALESCE(f.weaknesses, '') AS weaknesses,
	COALESCE(f.comments, '') AS comments,
	f.submitted_at, f.created_at, f.updated_at`

const interviewFeedbackJoins = `FROM interview_feedback f
	LEFT JOIN users iu ON iu.id = f.interviewer_id
	LEFT JOIN users u ON u.id = f.recorded_by_user_id`

// FindByInterviewID returns the first review on an interview. Older callers
// only knew one reviewer per interview; new code should use
// FindByInterviewAndInterviewer or ListByInterviewID.
func (r *InterviewFeedbackRepository) FindByInterviewID(interviewID int64) (*recmodel.InterviewFeedback, error) {
	fb := &recmodel.InterviewFeedback{}
	err := database.DB.Get(fb,
		`SELECT `+interviewFeedbackColumns+` `+interviewFeedbackJoins+`
		 WHERE f.interview_id = ? ORDER BY f.id LIMIT 1`, interviewID,
	)
	if err != nil {
		return nil, err
	}
	return fb, nil
}

// FindByID returns one feedback row.
func (r *InterviewFeedbackRepository) FindByID(id int64) (*recmodel.InterviewFeedback, error) {
	fb := &recmodel.InterviewFeedback{}
	err := database.DB.Get(fb,
		`SELECT `+interviewFeedbackColumns+` `+interviewFeedbackJoins+`
		 WHERE f.id = ?`, id,
	)
	if err != nil {
		return nil, err
	}
	return fb, nil
}

// FindByInterviewAndInterviewer returns one panel member's review. Each
// interviewer keeps their own score, recommendation and answers.
func (r *InterviewFeedbackRepository) FindByInterviewAndInterviewer(interviewID, interviewerID int64) (*recmodel.InterviewFeedback, error) {
	fb := &recmodel.InterviewFeedback{}
	err := database.DB.Get(fb,
		`SELECT `+interviewFeedbackColumns+` `+interviewFeedbackJoins+`
		 WHERE f.interview_id = ? AND f.interviewer_id = ?`, interviewID, interviewerID,
	)
	if err != nil {
		return nil, err
	}
	return fb, nil
}

// ListByInterviewID returns every panel member's review. The final approver
// reads this to compare interviewer scores before ranking candidates.
func (r *InterviewFeedbackRepository) ListByInterviewID(interviewID int64) ([]recmodel.InterviewFeedback, error) {
	var feedback []recmodel.InterviewFeedback
	err := database.DB.Select(&feedback,
		`SELECT `+interviewFeedbackColumns+` `+interviewFeedbackJoins+`
		 WHERE f.interview_id = ? ORDER BY f.id`, interviewID,
	)
	if err != nil {
		return nil, err
	}
	return feedback, nil
}

// Update changes the content of an existing review. recorded_by_user_id keeps
// the person who first entered it; re-edits do not steal attribution.
func (r *InterviewFeedbackRepository) Update(fb *recmodel.InterviewFeedback) error {
	_, err := database.DB.Exec(
		`UPDATE interview_feedback SET interviewer_id=?, score=?, recommendation=?,
		 strengths=?, weaknesses=?, comments=?, submitted_at=?,
		 updated_at=CURRENT_TIMESTAMP WHERE id=?`,
		fb.InterviewerID, fb.Score, fb.Recommendation, fb.Strengths, fb.Weaknesses,
		fb.Comments, fb.SubmittedAt, fb.ID,
	)
	return err
}

// UpsertAnswer stores one question answer. Re-saving the same interview
// updates the row instead of adding a duplicate, so editing feedback is
// idempotent.
func (r *InterviewFeedbackRepository) UpsertAnswer(answer *recmodel.InterviewFeedbackAnswer) error {
	_, err := database.DB.Exec(
		`INSERT INTO interview_feedback_answers (feedback_id, evaluation_question_id, rating_value,
		 text_value, numeric_value, boolean_value)
		 VALUES (?, ?, ?, ?, ?, ?)
		 ON CONFLICT(feedback_id, evaluation_question_id) DO UPDATE SET
		 rating_value = excluded.rating_value,
		 text_value = excluded.text_value,
		 numeric_value = excluded.numeric_value,
		 boolean_value = excluded.boolean_value,
		 updated_at = CURRENT_TIMESTAMP`,
		answer.FeedbackID, answer.EvaluationQuestionID, answer.RatingValue,
		answer.TextValue, answer.NumericValue, answer.BooleanValue,
	)
	return err
}

func (r *InterviewFeedbackRepository) FindAnswersByFeedbackID(feedbackID int64) ([]recmodel.InterviewFeedbackAnswer, error) {
	var answers []recmodel.InterviewFeedbackAnswer
	err := database.DB.Select(&answers,
		`SELECT id, feedback_id, evaluation_question_id, rating_value,
		 COALESCE(text_value, '') AS text_value, numeric_value, boolean_value,
		 created_at, updated_at
		 FROM interview_feedback_answers WHERE feedback_id = ?`,
		feedbackID,
	)
	if err != nil {
		return nil, err
	}
	return answers, nil
}

// candidateSummaryQuery scores a candidate on a weighted average of the
// feedback that exists. The interview score is the evaluation: once every
// active round has been scored the combined score decides hire/maybe/reject.
// There is no separate candidate-level evaluation to wait for; the final
// approver accepts whichever candidate they choose. Weights live on
// interview_rounds so a technical round can outweigh an HR screen.
const candidateSummaryQuery = `
	SELECT s.*,
	       CASE
	           WHEN s.candidate_status = 'hired' THEN 'hire'
	           WHEN s.candidate_status = 'rejected' THEN 'reject'
	           WHEN s.completed_reviews = 0 THEN 'pending'
	           WHEN s.total_rounds > 0 AND s.completed_rounds < s.total_rounds THEN 'in_progress'
	           WHEN s.average_score >= 7 THEN 'hire'
	           WHEN s.average_score >= 5 THEN 'maybe'
	           ELSE 'reject'
	       END AS final_decision
	FROM (
	    SELECT c.id AS candidate_id,
	           c.full_name AS candidate_name,
	           c.status AS candidate_status,
	           COALESCE(p.name, '') AS position_name,
	           COALESCE(c.job_posting_id, 0) AS job_posting_id,
	           COALESCE(jp.title, '') AS posting_title,
	           COUNT(DISTINCT i.id) AS total_interviews,
	           COALESCE(MAX(CASE WHEN f.score IS NOT NULL THEN i.id END), 0) AS latest_interview_id,
	           COUNT(DISTINCT CASE WHEN f.score IS NOT NULL THEN i.id END) AS completed_reviews,
	           COUNT(DISTINCT CASE WHEN f.score IS NOT NULL THEN f.interviewer_id END) AS interviewer_count,
	           COUNT(DISTINCT CASE WHEN f.score IS NOT NULL THEN i.interview_round_id END) AS completed_rounds,
	           (SELECT COUNT(*) FROM interview_rounds r2
	             WHERE r2.job_posting_id = c.job_posting_id AND r2.status = 'active') AS total_rounds,
	           COALESCE(
	               1.0 * SUM(CASE WHEN f.score IS NOT NULL THEN f.score * r.weight END) /
	               NULLIF(SUM(CASE WHEN f.score IS NOT NULL THEN r.weight END), 0),
	               0) AS average_score,
	           COALESCE(SUM(CASE WHEN f.score IS NOT NULL THEN f.score ELSE 0 END), 0) AS total_score,
	           COALESCE(MAX(f.score), 0) AS highest_score,
	           COALESCE(MIN(f.score), 0) AS lowest_score
	    FROM candidates c
	    LEFT JOIN job_postings jp ON jp.id = c.job_posting_id
	    LEFT JOIN positions p ON p.id = jp.position_id
	    LEFT JOIN interviews i ON i.candidate_id = c.id
	    LEFT JOIN interview_rounds r ON r.id = i.interview_round_id
	    LEFT JOIN interview_feedback f ON f.interview_id = i.id
	    GROUP BY c.id
	) s`

// Report Queries
func (r *InterviewRepository) GetCandidateSummary(candidateID int64) (*recmodel.CandidateEvaluationSummary, error) {
	s := &recmodel.CandidateEvaluationSummary{}
	err := database.DB.Get(s, candidateSummaryQuery+` WHERE s.candidate_id = ?`, candidateID)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (r *InterviewRepository) GetAllCandidateSummaries() ([]recmodel.CandidateEvaluationSummary, error) {
	var summaries []recmodel.CandidateEvaluationSummary
	err := database.DB.Select(&summaries,
		candidateSummaryQuery+` ORDER BY s.average_score DESC, s.completed_reviews DESC`,
	)
	if err != nil {
		return nil, err
	}

	// Rank candidates inside their own posting as soon as at least one panel
	// member has scored them. The review counts stay on the row, so a partially
	// reviewed candidate is still visibly distinguishable from a finished one.
	ranked := make(map[int64]int, len(summaries))
	for i := range summaries {
		s := &summaries[i]
		if s.CompletedReviews == 0 {
			continue
		}
		ranked[s.JobPostingID]++
		s.Rank = ranked[s.JobPostingID]
	}
	return summaries, nil
}

func (r *InterviewRepository) GetInterviewerActivity(jobPostingID int64) ([]recmodel.InterviewerActivity, error) {
	var activities []recmodel.InterviewerActivity
	err := database.DB.Select(&activities,
		`SELECT u.id AS interviewer_id, u.name AS interviewer_name,
		 COUNT(DISTINCT i.id) AS total_interviews,
		 SUM(CASE WHEN f.score IS NOT NULL THEN 1 ELSE 0 END) AS completed_reviews,
		 SUM(CASE WHEN f.score IS NULL AND i.id IS NOT NULL THEN 1 ELSE 0 END) AS pending_reviews,
		 COALESCE(AVG(f.score), 0) AS average_score
		 FROM users u
		 JOIN interview_interviewers ii ON ii.interviewer_id = u.id
		 JOIN interviews i ON i.id = ii.interview_id
		 JOIN candidates c ON c.id = i.candidate_id
		 LEFT JOIN interview_feedback f ON f.interview_id = i.id AND f.interviewer_id = u.id
		 WHERE ? = 0 OR c.job_posting_id = ?
		 GROUP BY u.id ORDER BY COUNT(DISTINCT i.id) DESC`,
		jobPostingID, jobPostingID,
	)
	if err != nil {
		return nil, err
	}
	return activities, nil
}

func (r *InterviewRepository) GetHiringDecisions() ([]recmodel.HiringDecision, error) {
	var decisions []recmodel.HiringDecision
	err := database.DB.Select(&decisions,
		`SELECT c.id AS candidate_id, c.full_name AS candidate_name,
		 COALESCE(c.job_posting_id, 0) AS job_posting_id,
		 COALESCE(p.name, '') AS position_name, COALESCE(d.name, '') AS department_name,
		 COALESCE(
		     1.0 * SUM(CASE WHEN f.score IS NOT NULL THEN f.score * r.weight END) /
		     NULLIF(SUM(CASE WHEN f.score IS NOT NULL THEN r.weight END), 0),
		     0) AS average_score,
		 COUNT(DISTINCT i.id) AS interview_count,
		 COALESCE(MAX(CASE WHEN f.score IS NOT NULL THEN i.id END), 0) AS latest_interview_id,
		 c.status AS decision,
		 COALESCE(c.updated_at, '') AS decision_date,
		 COALESCE((
		     SELECT u.name FROM interview_feedback f2
		     JOIN interviews i2 ON i2.id = f2.interview_id
		     JOIN users u ON u.id = COALESCE(f2.recorded_by_user_id, f2.interviewer_id)
		     WHERE i2.candidate_id = c.id AND f2.score IS NOT NULL
		     ORDER BY f2.updated_at DESC LIMIT 1
		 ), 'HR') AS decided_by
		 FROM candidates c
		 LEFT JOIN job_postings jp ON jp.id = c.job_posting_id
		 LEFT JOIN positions p ON p.id = jp.position_id
		 LEFT JOIN departments d ON d.id = p.department_id
		 LEFT JOIN interviews i ON i.candidate_id = c.id
		 LEFT JOIN interview_rounds r ON r.id = i.interview_round_id
		 LEFT JOIN interview_feedback f ON f.interview_id = i.id
		 GROUP BY c.id ORDER BY average_score DESC`,
	)
	if err != nil {
		return nil, err
	}
	return decisions, nil
}
