-- Interview panels, per-interviewer feedback and interview permissions.
--
-- One interview can be scored by every member of its panel. The original
-- interview_feedback table allowed only one row per interview (UNIQUE
-- interview_id), which made a multi-interviewer panel impossible. Rebuild it
-- with UNIQUE (interview_id, interviewer_id) so each panel member keeps their
-- own score, recommendation and per-question answers.
--
-- Additive permissions:
--   interview.view    - see schedules assigned to you
--   interview.assign  - schedule interviews and choose the panel
--   interview.score   - submit your own score as a panel member
--   interview.approve - see every schedule and the combined candidate ranking
--
-- recruitment.manage continues to imply all four. evaluation.assign also
-- grants a read-only view of every schedule (enforced in the page/API code).
--
-- The table rebuild must run outside goose's transaction so foreign_keys can
-- be disabled while the old table is dropped.

-- +goose NO TRANSACTION
-- +goose Up

PRAGMA foreign_keys=OFF;

CREATE TABLE interview_feedback_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    interview_id INTEGER NOT NULL,
    interviewer_id INTEGER NOT NULL,
    score DECIMAL(5,2),
    recommendation VARCHAR(20) NOT NULL DEFAULT 'pending',
    strengths TEXT,
    weaknesses TEXT,
    comments TEXT,
    submitted_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    recorded_by_user_id INTEGER REFERENCES users(id),
    UNIQUE (interview_id, interviewer_id),
    FOREIGN KEY (interview_id) REFERENCES interviews(id) ON DELETE CASCADE,
    FOREIGN KEY (interviewer_id) REFERENCES users(id) ON DELETE CASCADE
);

INSERT INTO interview_feedback_new (
    id, interview_id, interviewer_id, score, recommendation, strengths,
    weaknesses, comments, submitted_at, created_at, updated_at, recorded_by_user_id
)
SELECT id, interview_id, interviewer_id, score, recommendation, strengths,
       weaknesses, comments, submitted_at, created_at, updated_at,
       COALESCE(recorded_by_user_id, interviewer_id)
FROM interview_feedback;

DROP TABLE interview_feedback;
ALTER TABLE interview_feedback_new RENAME TO interview_feedback;

CREATE INDEX IF NOT EXISTS idx_interview_feedback_interviewer
    ON interview_feedback (interview_id, interviewer_id);
CREATE INDEX IF NOT EXISTS idx_interview_feedback_scored
    ON interview_feedback (score);

PRAGMA foreign_keys=ON;

-- Lookup indexes used by the scoped schedule views and ranking aggregation.
CREATE INDEX IF NOT EXISTS idx_interview_interviewers_interviewer
    ON interview_interviewers (interviewer_id);
CREATE INDEX IF NOT EXISTS idx_interviews_candidate
    ON interviews (candidate_id);
CREATE INDEX IF NOT EXISTS idx_interviews_round
    ON interviews (interview_round_id);
CREATE INDEX IF NOT EXISTS idx_evaluation_assignments_evaluator_candidate
    ON evaluation_assignments (evaluator_id, candidate_id);

-- Permission catalog. SeedDefaultRBAC also inserts these on boot; this keeps
-- existing installations in step before the seed runs.
INSERT OR IGNORE INTO permissions (name, description) VALUES
    ('interview.view', 'View interview schedules (own assignments unless combined with a wider interview permission)'),
    ('interview.assign', 'Schedule interviews and assign interviewer panels'),
    ('interview.score', 'Score interviews as an assigned interviewer'),
    ('interview.approve', 'Review all interview schedules and the final candidate ranking');

-- Default role grants for existing installations. Custom roles can be adjusted
-- from Administration > Roles.
INSERT OR IGNORE INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p
WHERE LOWER(r.name) = 'admin'
  AND p.name IN ('interview.view', 'interview.assign', 'interview.score', 'interview.approve');

INSERT OR IGNORE INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p
WHERE LOWER(r.name) = 'hrm'
  AND p.name IN ('interview.view', 'interview.assign', 'interview.score', 'interview.approve');

INSERT OR IGNORE INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p
WHERE LOWER(r.name) = 'manager'
  AND p.name IN ('interview.view', 'interview.assign', 'interview.score', 'interview.approve');

INSERT OR IGNORE INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p
WHERE LOWER(r.name) = 'employee'
  AND p.name IN ('interview.view', 'interview.score');

-- +goose Down

SELECT 1;
