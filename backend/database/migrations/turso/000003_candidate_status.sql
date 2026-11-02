-- Candidate pipeline: an auditable status history.
--
-- HR moves applicants through new -> screening -> interviewing -> hired /
-- rejected. The current status stays on candidates.status so every existing
-- query keeps working; this table records each transition, who made it and
-- why (used for rejection reasons).
--
-- Additive only: SQLite/libSQL cannot remove columns, so Down is a no-op.

-- +goose Up

CREATE TABLE IF NOT EXISTS candidate_status_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    candidate_id INTEGER NOT NULL,
    from_status VARCHAR(20) NOT NULL,
    to_status VARCHAR(20) NOT NULL,
    reason TEXT,
    changed_by_user_id INTEGER REFERENCES users(id),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (candidate_id) REFERENCES candidates(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_candidate_status_history_candidate
    ON candidate_status_history (candidate_id, created_at);

-- +goose Down

SELECT 1;
