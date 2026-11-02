-- HR-side recruitment: interview attribution, interviewer eligibility and
-- weighted scoring.
--
-- The recruitment module is operated by HR. Interviewers only need to be
-- schedulable employees; they are not expected to log in and submit feedback
-- themselves, so every feedback row records *who conducted* the interview
-- (interviewer_id, kept from the interview) separately from *who typed it in*
-- (recorded_by_user_id).
--
-- These statements are additive. SQLite/libSQL cannot drop or re-type columns
-- in place, so Down is intentionally a no-op.

-- +goose Up

-- Whether an active worker may be picked as an interviewer. Existing workers
-- default to eligible; HR can switch individuals off later.
ALTER TABLE employees ADD COLUMN can_interview INTEGER NOT NULL DEFAULT 1;

-- Relative importance of a round when averaging a candidate's interview
-- scores. 1.0 keeps the previous behaviour.
ALTER TABLE interview_rounds ADD COLUMN weight DECIMAL(5,2) NOT NULL DEFAULT 1.0;

-- The HR user who entered the score. NULL on rows written before this
-- migration (backfilled below from the old interviewer_id).
ALTER TABLE interview_feedback ADD COLUMN recorded_by_user_id INTEGER REFERENCES users(id);

UPDATE interview_feedback
SET recorded_by_user_id = interviewer_id
WHERE recorded_by_user_id IS NULL;

-- +goose Down

SELECT 1;
