-- Interview panels: an interview can be conducted by more than one person.
--
-- interviews.interviewer_id stays as the lead interviewer. It keeps existing
-- feedback attribution working and gives older queries a sensible fallback,
-- while this table records every member of the panel.
--
-- Additive only: SQLite/libSQL cannot remove columns safely, so Down is a
-- no-op.

-- +goose Up

CREATE TABLE IF NOT EXISTS interview_interviewers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    interview_id INTEGER NOT NULL,
    interviewer_id INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (interview_id, interviewer_id),
    FOREIGN KEY (interview_id) REFERENCES interviews(id) ON DELETE CASCADE,
    FOREIGN KEY (interviewer_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_interview_interviewers_interview
    ON interview_interviewers (interview_id);

-- Every existing interview keeps its original interviewer as the first panel
-- member, so listings and interviewer reports continue to show it.
INSERT OR IGNORE INTO interview_interviewers (interview_id, interviewer_id)
SELECT id, interviewer_id FROM interviews;

-- +goose Down

SELECT 1;
