-- ============================================================
-- EMS – Demo interview seed (Turso/libSQL)
--
-- Optional development data. It creates one applicant who is still
-- in the pipeline (not hired) with a completed interview and feedback,
-- so the Candidate Ranking table has a row to show.
--
-- Idempotent: safe to run repeatedly. It does nothing if the demo
-- candidate already has the interview and feedback.
--
--   sqlite3 ./storage/db < backend/database/seed_interviews.sql
-- ============================================================

-- Applicant still under evaluation. Attach them to the first published
-- job posting and reuse an uploaded CV when one exists.
INSERT INTO candidates (
    full_name, email, phone, job_posting_id, skills,
    cv_path, cv_name, cv_mime_type, cv_size, status
)
SELECT
    'Demo Candidate',
    'demo.candidate@example.com',
    '',
    jp.id,
    'Go, communication',
    COALESCE((SELECT cv_path FROM candidates WHERE cv_path <> '' ORDER BY id LIMIT 1), ''),
    COALESCE((SELECT cv_name FROM candidates WHERE cv_name <> '' ORDER BY id LIMIT 1), 'demo-cv.pdf'),
    COALESCE((SELECT cv_mime_type FROM candidates WHERE cv_mime_type <> '' ORDER BY id LIMIT 1), 'application/pdf'),
    COALESCE((SELECT cv_size FROM candidates WHERE cv_size > 0 ORDER BY id LIMIT 1), 0),
    'interviewing'
FROM job_postings jp
WHERE jp.status = 'published'
  AND NOT EXISTS (
      SELECT 1 FROM candidates WHERE email = 'demo.candidate@example.com'
  )
ORDER BY jp.id
LIMIT 1;

-- Completed interview on the first active round of the posting. The lowest
-- active user sits on the panel so the score has an author.
INSERT INTO interviews (
    candidate_id, interview_round_id, interviewer_id, scheduled_at,
    duration_minutes, location, status, notes
)
SELECT
    c.id,
    r.id,
    u.id,
    CURRENT_TIMESTAMP,
    60,
    'Demo room',
    'completed',
    'Seeded demo interview'
FROM candidates c
JOIN interview_rounds r
    ON r.job_posting_id = c.job_posting_id AND r.status = 'active'
JOIN users u ON u.status = 'active'
WHERE c.email = 'demo.candidate@example.com'
  AND NOT EXISTS (
      SELECT 1 FROM interviews i
      WHERE i.candidate_id = c.id AND i.interview_round_id = r.id
  )
ORDER BY r.round_number, u.id
LIMIT 1;

INSERT OR IGNORE INTO interview_interviewers (interview_id, interviewer_id)
SELECT i.id, i.interviewer_id
FROM interviews i
JOIN candidates c ON c.id = i.candidate_id
WHERE c.email = 'demo.candidate@example.com';

-- Use 80% of the round's maximum so the seeded score is always in range.
INSERT OR IGNORE INTO interview_feedback (
    interview_id, interviewer_id, score, recommendation,
    strengths, weaknesses, comments, submitted_at, recorded_by_user_id
)
SELECT
    i.id,
    i.interviewer_id,
    ROUND(r.max_score * 0.8, 2),
    'hire',
    'Clear communication',
    'Needs more product depth',
    'Seeded demo feedback',
    CURRENT_TIMESTAMP,
    i.interviewer_id
FROM interviews i
JOIN candidates c ON c.id = i.candidate_id
JOIN interview_rounds r ON r.id = i.interview_round_id
WHERE c.email = 'demo.candidate@example.com';
