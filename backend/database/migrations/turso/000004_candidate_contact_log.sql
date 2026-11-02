-- Candidate contact log: what HR did when contacting an applicant outside the
-- system (phone, email, Telegram, in person) and when to follow up next.
--
-- Additive only: SQLite/libSQL cannot remove columns or tables safely, so Down
-- is a no-op.

-- +goose Up

CREATE TABLE IF NOT EXISTS candidate_contact_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    candidate_id INTEGER NOT NULL,
    contacted_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    method VARCHAR(20) NOT NULL DEFAULT 'phone',
    direction VARCHAR(10) NOT NULL DEFAULT 'outbound',
    summary TEXT NOT NULL,
    outcome VARCHAR(20),
    next_follow_up_at DATETIME,
    created_by_user_id INTEGER REFERENCES users(id),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (candidate_id) REFERENCES candidates(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_candidate_contact_log_candidate
    ON candidate_contact_log (candidate_id, contacted_at DESC);

CREATE INDEX IF NOT EXISTS idx_candidate_contact_log_follow_up
    ON candidate_contact_log (next_follow_up_at);

-- +goose Down

SELECT 1;
