-- Interview rounds use evaluation templates as their question bank.
--
-- A round belongs to one template and a subset of that template's sections.
-- Splitting a large template across rounds means a screening round asks only
-- its own questions, a technical round only its own, and so on. Questions are
-- answered per interview, so the existing interview_feedback score continues
-- to feed the weighted round average.
--
-- Additive only: SQLite/libSQL cannot remove columns safely, so Down is a
-- no-op.

-- +goose Up

ALTER TABLE interview_rounds
    ADD COLUMN evaluation_template_id INTEGER REFERENCES evaluation_templates(id) ON DELETE SET NULL;

CREATE TABLE IF NOT EXISTS interview_round_sections (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    interview_round_id INTEGER NOT NULL,
    evaluation_section_id INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (interview_round_id, evaluation_section_id),
    FOREIGN KEY (interview_round_id) REFERENCES interview_rounds(id) ON DELETE CASCADE,
    FOREIGN KEY (evaluation_section_id) REFERENCES evaluation_sections(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_interview_round_sections_round
    ON interview_round_sections (interview_round_id);

CREATE TABLE IF NOT EXISTS interview_feedback_answers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    feedback_id INTEGER NOT NULL,
    evaluation_question_id INTEGER NOT NULL,
    rating_value DECIMAL(5,2),
    text_value TEXT,
    numeric_value DECIMAL(10,2),
    boolean_value BOOLEAN,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (feedback_id, evaluation_question_id),
    FOREIGN KEY (feedback_id) REFERENCES interview_feedback(id) ON DELETE CASCADE,
    FOREIGN KEY (evaluation_question_id) REFERENCES evaluation_questions(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_interview_feedback_answers_feedback
    ON interview_feedback_answers (feedback_id);

-- +goose Down

SELECT 1;
