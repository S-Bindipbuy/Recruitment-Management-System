-- ============================================================
-- EMS – Demo data: "Schedule Interview from a candidate row"
-- ============================================================
--
-- Creates a small but complete recruiting setup so the candidate-row
-- scheduling flow can be demonstrated end to end:
--
--   • 3 job postings (Backend, Frontend, Product)
--   • 4 active interviewers with employee records (grouped by department)
--   • 1 active evaluation period covering the next 30 days
--   • 1 evaluation template with 3 sections and 5 questions
--   • rounds that use that template, plus one round with NO template
--     (to demonstrate the optional template picker in the dialog)
--   • 5 candidates still in the pipeline, ready to be scheduled
--
-- Idempotent: every insert is guarded, so it is safe to run repeatedly.
--
-- Usage:
--   sqlite3 ./storage/db < backend/database/seed.sql
--   sqlite3 ./storage/db < backend/database/seed_demo_scheduling.sql
-- ============================================================

-- ------------------------------------------------------------
-- Departments and positions
-- ------------------------------------------------------------
INSERT OR IGNORE INTO departments (organization_id, name, description, status)
SELECT COALESCE((SELECT id FROM organizations ORDER BY id LIMIT 1), 1),
       'Engineering', 'Demo engineering department', 'active'
WHERE NOT EXISTS (SELECT 1 FROM departments WHERE name = 'Engineering');

INSERT OR IGNORE INTO departments (organization_id, name, description, status)
SELECT COALESCE((SELECT id FROM organizations ORDER BY id LIMIT 1), 1),
       'Product', 'Demo product department', 'active'
WHERE NOT EXISTS (SELECT 1 FROM departments WHERE name = 'Product');

INSERT OR IGNORE INTO positions (department_id, name, description, status)
SELECT d.id, 'Backend Engineer', 'Builds backend services', 'active'
FROM departments d
WHERE d.name = 'Engineering'
  AND NOT EXISTS (SELECT 1 FROM positions p WHERE p.name = 'Backend Engineer');

INSERT OR IGNORE INTO positions (department_id, name, description, status)
SELECT d.id, 'Frontend Engineer', 'Builds web interfaces', 'active'
FROM departments d
WHERE d.name = 'Engineering'
  AND NOT EXISTS (SELECT 1 FROM positions p WHERE p.name = 'Frontend Engineer');

INSERT OR IGNORE INTO positions (department_id, name, description, status)
SELECT d.id, 'Product Manager', 'Owns product direction', 'active'
FROM departments d
WHERE d.name = 'Product'
  AND NOT EXISTS (SELECT 1 FROM positions p WHERE p.name = 'Product Manager');

-- ------------------------------------------------------------
-- Interviewers (active users + active employee records)
-- ------------------------------------------------------------
-- Password is the same development hash as admin123.
INSERT OR IGNORE INTO users (name, email, password, status)
SELECT 'Alex Chen', 'alex.chen@example.com',
       '$2a$10$y7G6U1K..UovJGi8JvVi.uA9DUfHzkQk1Ovjz3VB8PudYzyXE7xme', 'active'
WHERE NOT EXISTS (SELECT 1 FROM users WHERE email = 'alex.chen@example.com');

INSERT OR IGNORE INTO users (name, email, password, status)
SELECT 'Taylor Morgan', 'taylor.morgan@example.com',
       '$2a$10$y7G6U1K..UovJGi8JvVi.uA9DUfHzkQk1Ovjz3VB8PudYzyXE7xme', 'active'
WHERE NOT EXISTS (SELECT 1 FROM users WHERE email = 'taylor.morgan@example.com');

INSERT OR IGNORE INTO users (name, email, password, status)
SELECT 'Jordan Kim', 'jordan.kim@example.com',
       '$2a$10$y7G6U1K..UovJGi8JvVi.uA9DUfHzkQk1Ovjz3VB8PudYzyXE7xme', 'active'
WHERE NOT EXISTS (SELECT 1 FROM users WHERE email = 'jordan.kim@example.com');

INSERT OR IGNORE INTO users (name, email, password, status)
SELECT 'Sam Rivera', 'sam.rivera@example.com',
       '$2a$10$y7G6U1K..UovJGi8JvVi.uA9DUfHzkQk1Ovjz3VB8PudYzyXE7xme', 'active'
WHERE NOT EXISTS (SELECT 1 FROM users WHERE email = 'sam.rivera@example.com');

INSERT OR IGNORE INTO user_roles (user_id, role_id)
SELECT u.id, r.id
FROM users u JOIN roles r ON r.name = 'interviewer'
WHERE u.email = 'alex.chen@example.com';

INSERT OR IGNORE INTO user_roles (user_id, role_id)
SELECT u.id, r.id
FROM users u JOIN roles r ON r.name = 'interviewer'
WHERE u.email = 'taylor.morgan@example.com';

INSERT OR IGNORE INTO user_roles (user_id, role_id)
SELECT u.id, r.id
FROM users u JOIN roles r ON r.name = 'interviewer'
WHERE u.email = 'jordan.kim@example.com';

INSERT OR IGNORE INTO user_roles (user_id, role_id)
SELECT u.id, r.id
FROM users u JOIN roles r ON r.name = 'interviewer'
WHERE u.email = 'sam.rivera@example.com';

-- Employee records make the users selectable as panel interviewers.
INSERT OR IGNORE INTO employees (user_id, employee_code, department_id, position_id, hire_date, status, can_interview)
SELECT u.id, 'DEMO-ALEX', d.id, p.id, date('now', '-120 days'), 'active', 1
FROM users u
JOIN departments d ON d.name = 'Engineering'
JOIN positions p ON p.name = 'Backend Engineer'
WHERE u.email = 'alex.chen@example.com'
  AND NOT EXISTS (SELECT 1 FROM employees e WHERE e.user_id = u.id);

INSERT OR IGNORE INTO employees (user_id, employee_code, department_id, position_id, hire_date, status, can_interview)
SELECT u.id, 'DEMO-TAYLOR', d.id, p.id, date('now', '-90 days'), 'active', 1
FROM users u
JOIN departments d ON d.name = 'Engineering'
JOIN positions p ON p.name = 'Frontend Engineer'
WHERE u.email = 'taylor.morgan@example.com'
  AND NOT EXISTS (SELECT 1 FROM employees e WHERE e.user_id = u.id);

INSERT OR IGNORE INTO employees (user_id, employee_code, department_id, position_id, hire_date, status, can_interview)
SELECT u.id, 'DEMO-JORDAN', d.id, p.id, date('now', '-60 days'), 'active', 1
FROM users u
JOIN departments d ON d.name = 'Product'
JOIN positions p ON p.name = 'Product Manager'
WHERE u.email = 'jordan.kim@example.com'
  AND NOT EXISTS (SELECT 1 FROM employees e WHERE e.user_id = u.id);

INSERT OR IGNORE INTO employees (user_id, employee_code, department_id, position_id, hire_date, status, can_interview)
SELECT u.id, 'DEMO-SAM', d.id, p.id, date('now', '-30 days'), 'active', 1
FROM users u
JOIN departments d ON d.name = 'Engineering'
JOIN positions p ON p.name = 'Backend Engineer'
WHERE u.email = 'sam.rivera@example.com'
  AND NOT EXISTS (SELECT 1 FROM employees e WHERE e.user_id = u.id);

-- ------------------------------------------------------------
-- Job postings
-- ------------------------------------------------------------
INSERT OR IGNORE INTO job_postings (position_id, title, description, requirements, skills_required, employment_type, status, posted_at, created_by)
SELECT p.id, 'Backend Engineer',
       'Build and scale the services behind the product.',
       'Strong Go and SQL; REST API design.',
       'Go, SQL, REST',
       'full-time', 'published', CURRENT_TIMESTAMP,
       (SELECT id FROM users WHERE email = 'admin@example.com' LIMIT 1)
FROM positions p
WHERE p.name = 'Backend Engineer'
  AND NOT EXISTS (SELECT 1 FROM job_postings jp WHERE jp.title = 'Backend Engineer');

INSERT OR IGNORE INTO job_postings (position_id, title, description, requirements, skills_required, employment_type, status, posted_at, created_by)
SELECT p.id, 'Frontend Engineer',
       'Build the interfaces candidates and employees use every day.',
       'TypeScript, HTML, CSS and a modern framework.',
       'TypeScript, React, CSS',
       'full-time', 'published', CURRENT_TIMESTAMP,
       (SELECT id FROM users WHERE email = 'admin@example.com' LIMIT 1)
FROM positions p
WHERE p.name = 'Frontend Engineer'
  AND NOT EXISTS (SELECT 1 FROM job_postings jp WHERE jp.title = 'Frontend Engineer');

INSERT OR IGNORE INTO job_postings (position_id, title, description, requirements, skills_required, employment_type, status, posted_at, created_by)
SELECT p.id, 'Product Manager',
       'Own the roadmap and work with engineering and design.',
       'Discovery, analytics and clear communication.',
       'Roadmapping, Analytics, Communication',
       'full-time', 'published', CURRENT_TIMESTAMP,
       (SELECT id FROM users WHERE email = 'admin@example.com' LIMIT 1)
FROM positions p
WHERE p.name = 'Product Manager'
  AND NOT EXISTS (SELECT 1 FROM job_postings jp WHERE jp.title = 'Product Manager');

-- ------------------------------------------------------------
-- Active evaluation period (covers the demo interview dates)
-- ------------------------------------------------------------
INSERT OR IGNORE INTO evaluation_periods (name, description, start_date, end_date, status, created_by)
SELECT 'Demo Interview Period',
       'Active period for demonstrating candidate-row interview scheduling.',
       date('now', '-7 days'), date('now', '+30 days'), 'active',
       (SELECT id FROM users WHERE email = 'admin@example.com' LIMIT 1)
WHERE NOT EXISTS (SELECT 1 FROM evaluation_periods WHERE name = 'Demo Interview Period');

-- ------------------------------------------------------------
-- Evaluation template, sections and questions
-- ------------------------------------------------------------
INSERT OR IGNORE INTO evaluation_templates (name, description, status, created_by)
SELECT 'Demo Technical Interview',
       'Standard interview scorecard used by the demo rounds.',
       'active',
       (SELECT id FROM users WHERE email = 'admin@example.com' LIMIT 1)
WHERE NOT EXISTS (SELECT 1 FROM evaluation_templates WHERE name = 'Demo Technical Interview');

INSERT OR IGNORE INTO evaluation_sections (evaluation_template_id, name, description, sort_order)
SELECT t.id, 'Technical Skills', 'Problem solving and technical depth', 1
FROM evaluation_templates t
WHERE t.name = 'Demo Technical Interview'
  AND NOT EXISTS (SELECT 1 FROM evaluation_sections s WHERE s.evaluation_template_id = t.id AND s.name = 'Technical Skills');

INSERT OR IGNORE INTO evaluation_sections (evaluation_template_id, name, description, sort_order)
SELECT t.id, 'Communication', 'Clarity and collaboration', 2
FROM evaluation_templates t
WHERE t.name = 'Demo Technical Interview'
  AND NOT EXISTS (SELECT 1 FROM evaluation_sections s WHERE s.evaluation_template_id = t.id AND s.name = 'Communication');

INSERT OR IGNORE INTO evaluation_sections (evaluation_template_id, name, description, sort_order)
SELECT t.id, 'Culture Fit', 'Values and teamwork', 3
FROM evaluation_templates t
WHERE t.name = 'Demo Technical Interview'
  AND NOT EXISTS (SELECT 1 FROM evaluation_sections s WHERE s.evaluation_template_id = t.id AND s.name = 'Culture Fit');

INSERT OR IGNORE INTO evaluation_questions (evaluation_section_id, question, question_type, is_required, weight, sort_order, status)
SELECT s.id, 'Ability to solve algorithmic problems', 'rating', 1, 1, 1, 'active'
FROM evaluation_sections s JOIN evaluation_templates t ON t.id = s.evaluation_template_id
WHERE t.name = 'Demo Technical Interview' AND s.name = 'Technical Skills'
  AND NOT EXISTS (SELECT 1 FROM evaluation_questions q WHERE q.evaluation_section_id = s.id AND q.question = 'Ability to solve algorithmic problems');

INSERT OR IGNORE INTO evaluation_questions (evaluation_section_id, question, question_type, is_required, weight, sort_order, status)
SELECT s.id, 'Knowledge of relevant technologies', 'rating', 1, 1, 2, 'active'
FROM evaluation_sections s JOIN evaluation_templates t ON t.id = s.evaluation_template_id
WHERE t.name = 'Demo Technical Interview' AND s.name = 'Technical Skills'
  AND NOT EXISTS (SELECT 1 FROM evaluation_questions q WHERE q.evaluation_section_id = s.id AND q.question = 'Knowledge of relevant technologies');

INSERT OR IGNORE INTO evaluation_questions (evaluation_section_id, question, question_type, is_required, weight, sort_order, status)
SELECT s.id, 'Explains ideas clearly', 'rating', 1, 1, 1, 'active'
FROM evaluation_sections s JOIN evaluation_templates t ON t.id = s.evaluation_template_id
WHERE t.name = 'Demo Technical Interview' AND s.name = 'Communication'
  AND NOT EXISTS (SELECT 1 FROM evaluation_questions q WHERE q.evaluation_section_id = s.id AND q.question = 'Explains ideas clearly');

INSERT OR IGNORE INTO evaluation_questions (evaluation_section_id, question, question_type, is_required, weight, sort_order, status)
SELECT s.id, 'Asks thoughtful questions', 'rating', 1, 1, 2, 'active'
FROM evaluation_sections s JOIN evaluation_templates t ON t.id = s.evaluation_template_id
WHERE t.name = 'Demo Technical Interview' AND s.name = 'Communication'
  AND NOT EXISTS (SELECT 1 FROM evaluation_questions q WHERE q.evaluation_section_id = s.id AND q.question = 'Asks thoughtful questions');

INSERT OR IGNORE INTO evaluation_questions (evaluation_section_id, question, question_type, is_required, weight, sort_order, status)
SELECT s.id, 'Alignment with team values', 'rating', 1, 1, 1, 'active'
FROM evaluation_sections s JOIN evaluation_templates t ON t.id = s.evaluation_template_id
WHERE t.name = 'Demo Technical Interview' AND s.name = 'Culture Fit'
  AND NOT EXISTS (SELECT 1 FROM evaluation_questions q WHERE q.evaluation_section_id = s.id AND q.question = 'Alignment with team values');

-- ------------------------------------------------------------
-- Interview rounds
-- ------------------------------------------------------------
-- Backend Engineer: two rounds, both using the demo template.
INSERT OR IGNORE INTO interview_rounds (job_posting_id, name, round_number, round_type, description, max_score, pass_score, status, weight, evaluation_template_id)
SELECT jp.id, 'Technical Screen', 1, 'technical', 'Initial technical assessment', 10, 6, 'active', 1,
       (SELECT id FROM evaluation_templates WHERE name = 'Demo Technical Interview')
FROM job_postings jp
WHERE jp.title = 'Backend Engineer'
  AND NOT EXISTS (SELECT 1 FROM interview_rounds r WHERE r.job_posting_id = jp.id AND r.name = 'Technical Screen');

INSERT OR IGNORE INTO interview_rounds (job_posting_id, name, round_number, round_type, description, max_score, pass_score, status, weight, evaluation_template_id)
SELECT jp.id, 'System Design', 2, 'technical', 'System design discussion', 10, 6, 'active', 1,
       (SELECT id FROM evaluation_templates WHERE name = 'Demo Technical Interview')
FROM job_postings jp
WHERE jp.title = 'Backend Engineer'
  AND NOT EXISTS (SELECT 1 FROM interview_rounds r WHERE r.job_posting_id = jp.id AND r.name = 'System Design');

-- Frontend Engineer: one round with the demo template.
INSERT OR IGNORE INTO interview_rounds (job_posting_id, name, round_number, round_type, description, max_score, pass_score, status, weight, evaluation_template_id)
SELECT jp.id, 'Frontend Technical', 1, 'technical', 'Frontend coding interview', 10, 6, 'active', 1,
       (SELECT id FROM evaluation_templates WHERE name = 'Demo Technical Interview')
FROM job_postings jp
WHERE jp.title = 'Frontend Engineer'
  AND NOT EXISTS (SELECT 1 FROM interview_rounds r WHERE r.job_posting_id = jp.id AND r.name = 'Frontend Technical');

-- Product Manager: deliberately has NO template, so the scheduler must use
-- the optional "Evaluation Template" picker in the dialog.
INSERT OR IGNORE INTO interview_rounds (job_posting_id, name, round_number, round_type, description, max_score, pass_score, status, weight, evaluation_template_id)
SELECT jp.id, 'Behavioral', 1, 'behavioral', 'Behavioral interview - pick a template when scheduling', 10, 6, 'active', 1, NULL
FROM job_postings jp
WHERE jp.title = 'Product Manager'
  AND NOT EXISTS (SELECT 1 FROM interview_rounds r WHERE r.job_posting_id = jp.id AND r.name = 'Behavioral');

-- Tie the demo template's sections to the rounds that use it, so each round
-- asks exactly the scorecard sections it was configured with.
INSERT OR IGNORE INTO interview_round_sections (interview_round_id, evaluation_section_id)
SELECT r.id, s.id
FROM interview_rounds r
JOIN evaluation_templates t ON t.id = r.evaluation_template_id
JOIN evaluation_sections s ON s.evaluation_template_id = t.id
WHERE t.name = 'Demo Technical Interview';

-- ------------------------------------------------------------
-- Candidates still in the pipeline
-- ------------------------------------------------------------
INSERT OR IGNORE INTO candidates (full_name, email, phone, job_posting_id, skills, cv_path, cv_name, cv_mime_type, cv_size, status)
SELECT 'Jamie Park', 'jamie.park@example.com', '+1-555-0101', jp.id, 'Go, PostgreSQL, Docker', '', 'jamie-park-cv.pdf', 'application/pdf', 0, 'screening'
FROM job_postings jp
WHERE jp.title = 'Backend Engineer'
  AND NOT EXISTS (SELECT 1 FROM candidates c WHERE c.email = 'jamie.park@example.com');

INSERT OR IGNORE INTO candidates (full_name, email, phone, job_posting_id, skills, cv_path, cv_name, cv_mime_type, cv_size, status)
SELECT 'Morgan Ellis', 'morgan.ellis@example.com', '+1-555-0102', jp.id, 'Go, gRPC, Redis', '', 'morgan-ellis-cv.pdf', 'application/pdf', 0, 'new'
FROM job_postings jp
WHERE jp.title = 'Backend Engineer'
  AND NOT EXISTS (SELECT 1 FROM candidates c WHERE c.email = 'morgan.ellis@example.com');

INSERT OR IGNORE INTO candidates (full_name, email, phone, job_posting_id, skills, cv_path, cv_name, cv_mime_type, cv_size, status)
SELECT 'Cameron Lee', 'cameron.lee@example.com', '+1-555-0201', jp.id, 'TypeScript, React, Tailwind', '', 'cameron-lee-cv.pdf', 'application/pdf', 0, 'screening'
FROM job_postings jp
WHERE jp.title = 'Frontend Engineer'
  AND NOT EXISTS (SELECT 1 FROM candidates c WHERE c.email = 'cameron.lee@example.com');

INSERT OR IGNORE INTO candidates (full_name, email, phone, job_posting_id, skills, cv_path, cv_name, cv_mime_type, cv_size, status)
SELECT 'Drew Johnson', 'drew.johnson@example.com', '+1-555-0301', jp.id, 'Product strategy, SQL, Analytics', '', 'drew-johnson-cv.pdf', 'application/pdf', 0, 'new'
FROM job_postings jp
WHERE jp.title = 'Product Manager'
  AND NOT EXISTS (SELECT 1 FROM candidates c WHERE c.email = 'drew.johnson@example.com');

INSERT OR IGNORE INTO candidates (full_name, email, phone, job_posting_id, skills, cv_path, cv_name, cv_mime_type, cv_size, status)
SELECT 'Riley Chen', 'riley.chen@example.com', '+1-555-0302', jp.id, 'Agile, User research, Roadmapping', '', 'riley-chen-cv.pdf', 'application/pdf', 0, 'screening'
FROM job_postings jp
WHERE jp.title = 'Product Manager'
  AND NOT EXISTS (SELECT 1 FROM candidates c WHERE c.email = 'riley.chen@example.com');
