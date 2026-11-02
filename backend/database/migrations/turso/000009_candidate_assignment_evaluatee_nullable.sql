-- Candidate evaluations no longer have an employee evaluatee.
--
-- evaluation_assignments.evaluatee_id was created NOT NULL because the table
-- originally only held worker-to-worker evaluations. Candidate assignments
-- have no employee on the evaluatee side, so they insert 0, which fails the
-- employees(id) foreign key as soon as foreign_keys is enabled. SQLite cannot
-- drop the constraint in place, so rebuild the table with a nullable column and
-- convert the old 0 placeholders to NULL.
--
-- The table rebuild must run outside goose's transaction so foreign_keys can be
-- disabled while the old table is dropped.

-- +goose NO TRANSACTION
-- +goose Up

PRAGMA foreign_keys = OFF;

BEGIN;

CREATE TABLE evaluation_assignments_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    evaluation_period_id INTEGER NOT NULL,
    evaluation_template_id INTEGER NOT NULL,
    evaluator_id INTEGER NOT NULL,
    evaluatee_id INTEGER,
    candidate_id INTEGER REFERENCES candidates(id) ON DELETE SET NULL,
    due_date DATE,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    started_at DATETIME,
    submitted_at DATETIME,
    completed_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (evaluation_period_id) REFERENCES evaluation_periods(id) ON DELETE CASCADE,
    FOREIGN KEY (evaluation_template_id) REFERENCES evaluation_templates(id) ON DELETE CASCADE,
    FOREIGN KEY (evaluator_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (evaluatee_id) REFERENCES employees(id) ON DELETE CASCADE
);

INSERT INTO evaluation_assignments_new (
    id, evaluation_period_id, evaluation_template_id, evaluator_id,
    evaluatee_id, candidate_id, due_date, status, started_at, submitted_at,
    completed_at, created_at, updated_at
)
SELECT
    id, evaluation_period_id, evaluation_template_id, evaluator_id,
    NULLIF(evaluatee_id, 0), candidate_id, due_date, status, started_at,
    submitted_at, completed_at, created_at, updated_at
FROM evaluation_assignments;

DROP TABLE evaluation_assignments;
ALTER TABLE evaluation_assignments_new RENAME TO evaluation_assignments;

CREATE INDEX IF NOT EXISTS idx_evaluation_assignments_evaluator_candidate
    ON evaluation_assignments (evaluator_id, candidate_id);

COMMIT;

PRAGMA foreign_keys = ON;

-- +goose Down

SELECT 1;
