-- EMS schema for Turso/libSQL (SQLite-compatible).
--
-- Managed by goose (see backend/database/migrations.go). The same file can be
-- imported manually -- the goose annotations are plain SQL comments:
--
--   sqlite3 ./storage/db < backend/database/migrations/turso/000001_init.sql
--
-- Every statement uses IF NOT EXISTS so a database created before goose can be
-- adopted without conflicts. The same schema runs unchanged against local
-- libSQL files and remote Turso databases.

-- +goose Up

CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS roles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS permissions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS user_roles (
    user_id INTEGER NOT NULL,
    role_id INTEGER NOT NULL,
    PRIMARY KEY (user_id, role_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS role_permissions (
    role_id INTEGER NOT NULL,
    permission_id INTEGER NOT NULL,
    PRIMARY KEY (role_id, permission_id),
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
    FOREIGN KEY (permission_id) REFERENCES permissions(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS organizations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS departments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    organization_id INTEGER NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS positions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    department_id INTEGER NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (department_id) REFERENCES departments(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS employees (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER,
    employee_code VARCHAR(50) NOT NULL UNIQUE,
    department_id INTEGER,
    position_id INTEGER,
    manager_id INTEGER,
    hire_date DATE,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL,
    FOREIGN KEY (department_id) REFERENCES departments(id) ON DELETE SET NULL,
    FOREIGN KEY (position_id) REFERENCES positions(id) ON DELETE SET NULL,
    FOREIGN KEY (manager_id) REFERENCES employees(id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS job_postings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    position_id INTEGER NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    requirements TEXT,
    skills_required TEXT,
    salary_min DECIMAL(12,2),
    salary_max DECIMAL(12,2),
    employment_type VARCHAR(50) NOT NULL DEFAULT 'full-time',
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    posted_at DATETIME,
    closed_at DATETIME,
    created_by INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (position_id) REFERENCES positions(id) ON DELETE CASCADE,
    FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS candidates (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    full_name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    phone VARCHAR(50),
    job_posting_id INTEGER,
    skills TEXT NOT NULL,
    cv_path TEXT NOT NULL,
    cv_name VARCHAR(255) NOT NULL,
    cv_mime_type VARCHAR(100) NOT NULL,
    cv_size INTEGER NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'new',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (job_posting_id) REFERENCES job_postings(id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS candidate_assessments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    candidate_id INTEGER NOT NULL,
    evaluator_id INTEGER NOT NULL,
    skill_score DECIMAL(5,2) NOT NULL,
    notes TEXT,
    recommendation VARCHAR(20) NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (candidate_id) REFERENCES candidates(id) ON DELETE CASCADE,
    FOREIGN KEY (evaluator_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS skills (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(255) NOT NULL UNIQUE,
    category VARCHAR(100),
    description TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS candidate_skills (
    candidate_id INTEGER NOT NULL,
    skill_id INTEGER NOT NULL,
    proficiency VARCHAR(20) NOT NULL DEFAULT 'intermediate',
    years_experience INTEGER,
    PRIMARY KEY (candidate_id, skill_id),
    FOREIGN KEY (candidate_id) REFERENCES candidates(id) ON DELETE CASCADE,
    FOREIGN KEY (skill_id) REFERENCES skills(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS job_posting_skills (
    job_posting_id INTEGER NOT NULL,
    skill_id INTEGER NOT NULL,
    is_required BOOLEAN NOT NULL DEFAULT TRUE,
    min_proficiency VARCHAR(20) NOT NULL DEFAULT 'intermediate',
    PRIMARY KEY (job_posting_id, skill_id),
    FOREIGN KEY (job_posting_id) REFERENCES job_postings(id) ON DELETE CASCADE,
    FOREIGN KEY (skill_id) REFERENCES skills(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS interview_rounds (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    job_posting_id INTEGER NOT NULL,
    name VARCHAR(255) NOT NULL,
    round_number INTEGER NOT NULL,
    round_type VARCHAR(50) NOT NULL DEFAULT 'technical',
    description TEXT,
    max_score DECIMAL(5,2) NOT NULL DEFAULT 10.0,
    pass_score DECIMAL(5,2) NOT NULL DEFAULT 6.0,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (job_posting_id) REFERENCES job_postings(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS interviews (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    candidate_id INTEGER NOT NULL,
    interview_round_id INTEGER NOT NULL,
    interviewer_id INTEGER NOT NULL,
    scheduled_at DATETIME NOT NULL,
    duration_minutes INTEGER NOT NULL DEFAULT 60,
    location VARCHAR(500),
    meeting_link VARCHAR(500),
    status VARCHAR(20) NOT NULL DEFAULT 'scheduled',
    notes TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (candidate_id) REFERENCES candidates(id) ON DELETE CASCADE,
    FOREIGN KEY (interview_round_id) REFERENCES interview_rounds(id) ON DELETE CASCADE,
    FOREIGN KEY (interviewer_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS interview_feedback (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    interview_id INTEGER NOT NULL UNIQUE,
    interviewer_id INTEGER NOT NULL,
    score DECIMAL(5,2),
    recommendation VARCHAR(20) NOT NULL DEFAULT 'pending',
    strengths TEXT,
    weaknesses TEXT,
    comments TEXT,
    submitted_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (interview_id) REFERENCES interviews(id) ON DELETE CASCADE,
    FOREIGN KEY (interviewer_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS evaluation_periods (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    created_by INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS rating_scales (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS rating_scale_options (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    rating_scale_id INTEGER NOT NULL,
    label VARCHAR(255) NOT NULL,
    value DECIMAL(5,2) NOT NULL,
    description TEXT,
    sort_order INTEGER NOT NULL DEFAULT 0,
    FOREIGN KEY (rating_scale_id) REFERENCES rating_scales(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS evaluation_templates (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_by INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS evaluation_sections (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    evaluation_template_id INTEGER NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (evaluation_template_id) REFERENCES evaluation_templates(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS evaluation_questions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    evaluation_section_id INTEGER NOT NULL,
    question TEXT NOT NULL,
    description TEXT,
    question_type VARCHAR(50) NOT NULL DEFAULT 'rating',
    is_required BOOLEAN NOT NULL DEFAULT TRUE,
    weight DECIMAL(5,2) NOT NULL DEFAULT 1.0,
    sort_order INTEGER NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (evaluation_section_id) REFERENCES evaluation_sections(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS evaluation_assignments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    evaluation_period_id INTEGER NOT NULL,
    evaluation_template_id INTEGER NOT NULL,
    evaluator_id INTEGER NOT NULL,
    evaluatee_id INTEGER NOT NULL,
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

CREATE TABLE IF NOT EXISTS evaluation_responses (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    evaluation_assignment_id INTEGER NOT NULL,
    evaluation_question_id INTEGER NOT NULL,
    rating_value DECIMAL(5,2),
    text_value TEXT,
    numeric_value DECIMAL(10,2),
    boolean_value BOOLEAN,
    selected_option_id INTEGER,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (evaluation_assignment_id) REFERENCES evaluation_assignments(id) ON DELETE CASCADE,
    FOREIGN KEY (evaluation_question_id) REFERENCES evaluation_questions(id) ON DELETE CASCADE
);

-- evaluation_responses.Upsert relies on ON CONFLICT(evaluation_assignment_id,
-- evaluation_question_id), which SQLite rejects unless that pair is unique.
-- The table shipped without the constraint, so drop duplicate rows first so a
-- database with old data can still take the index.
DELETE FROM evaluation_responses WHERE id NOT IN (
    SELECT MAX(id) FROM evaluation_responses
    GROUP BY evaluation_assignment_id, evaluation_question_id
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_evaluation_responses_assignment_question
    ON evaluation_responses(evaluation_assignment_id, evaluation_question_id);

CREATE TABLE IF NOT EXISTS evaluation_results (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    evaluation_assignment_id INTEGER NOT NULL UNIQUE,
    total_score DECIMAL(10,2) NOT NULL,
    percentage DECIMAL(5,2) NOT NULL,
    grade VARCHAR(10),
    calculated_at DATETIME NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (evaluation_assignment_id) REFERENCES evaluation_assignments(id) ON DELETE CASCADE
);

-- A result calculated while responses could not be saved is zero with
-- nothing behind it. No answers means no score, so clear those rows rather
-- than report a verdict nobody gave.
DELETE FROM evaluation_results
WHERE NOT EXISTS (
    SELECT 1 FROM evaluation_responses r
    WHERE r.evaluation_assignment_id = evaluation_results.evaluation_assignment_id
);

CREATE TABLE IF NOT EXISTS evaluation_comments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    evaluation_assignment_id INTEGER NOT NULL,
    evaluation_question_id INTEGER,
    user_id INTEGER NOT NULL,
    comment TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (evaluation_assignment_id) REFERENCES evaluation_assignments(id) ON DELETE CASCADE,
    FOREIGN KEY (evaluation_question_id) REFERENCES evaluation_questions(id) ON DELETE SET NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS evaluation_approvals (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    evaluation_assignment_id INTEGER NOT NULL,
    approver_id INTEGER NOT NULL,
    level INTEGER NOT NULL DEFAULT 1,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    comment TEXT,
    approved_at DATETIME,
    rejected_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (evaluation_assignment_id) REFERENCES evaluation_assignments(id) ON DELETE CASCADE,
    FOREIGN KEY (approver_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS audit_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    action VARCHAR(100) NOT NULL,
    entity_type VARCHAR(100) NOT NULL,
    entity_id INTEGER NOT NULL,
    old_values TEXT,
    new_values TEXT,
    ip_address VARCHAR(45),
    user_agent TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- +goose Down

DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS evaluation_approvals;
DROP TABLE IF EXISTS evaluation_comments;
DROP TABLE IF EXISTS evaluation_results;
DROP TABLE IF EXISTS evaluation_responses;
DROP TABLE IF EXISTS evaluation_assignments;
DROP TABLE IF EXISTS evaluation_questions;
DROP TABLE IF EXISTS evaluation_sections;
DROP TABLE IF EXISTS evaluation_templates;
DROP TABLE IF EXISTS rating_scale_options;
DROP TABLE IF EXISTS rating_scales;
DROP TABLE IF EXISTS evaluation_periods;
DROP TABLE IF EXISTS interview_feedback;
DROP TABLE IF EXISTS interviews;
DROP TABLE IF EXISTS interview_rounds;
DROP TABLE IF EXISTS job_posting_skills;
DROP TABLE IF EXISTS candidate_skills;
DROP TABLE IF EXISTS skills;
DROP TABLE IF EXISTS candidate_assessments;
DROP TABLE IF EXISTS candidates;
DROP TABLE IF EXISTS job_postings;
DROP TABLE IF EXISTS employees;
DROP TABLE IF EXISTS positions;
DROP TABLE IF EXISTS departments;
DROP TABLE IF EXISTS organizations;
DROP TABLE IF EXISTS role_permissions;
DROP TABLE IF EXISTS user_roles;
DROP TABLE IF EXISTS permissions;
DROP TABLE IF EXISTS roles;
DROP TABLE IF EXISTS users;
