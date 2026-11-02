-- Per-round evaluations are linked to the interview that created them.
--
-- Scheduling an interview now creates one pending evaluation per panel
-- interviewer, using the round's own template and sections. The link lets the
-- evaluation form ask only that round's questions, lets feedback and the
-- evaluation stay in sync, and keeps these round assignments out of the
-- candidate-level approval workflow.
--
-- Additive only: SQLite/libSQL cannot drop columns safely, so Down is a no-op.

-- +goose Up

ALTER TABLE evaluation_assignments
    ADD COLUMN interview_id INTEGER REFERENCES interviews(id) ON DELETE CASCADE;

ALTER TABLE evaluation_assignments
    ADD COLUMN interview_round_id INTEGER REFERENCES interview_rounds(id) ON DELETE SET NULL;

-- One evaluation per interviewer per interview. NULL interview_id keeps the
-- existing candidate-level assignments unaffected by the uniqueness rule.
CREATE UNIQUE INDEX IF NOT EXISTS idx_evaluation_assignments_interview_evaluator
    ON evaluation_assignments (interview_id, evaluator_id)
    WHERE interview_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_evaluation_assignments_interview_round
    ON evaluation_assignments (interview_round_id);

-- +goose Down

SELECT 1;
