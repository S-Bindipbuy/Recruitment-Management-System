-- ============================================================
-- EMS – Development Seed Script (Turso/libSQL)
--
-- Portable: import directly after the init migration:
--   sqlite3 ./storage/db < backend/database/migrations/turso/000001_init.sql
--   sqlite3 ./storage/db < backend/database/seed.sql
--
-- Idempotent: safe to run repeatedly (INSERT OR IGNORE / NOT EXISTS).
-- Mirrors the Go seeder (backend/database/seed.go) and includes the
-- development administrator:
--
--   email:    admin@example.com
--   password: admin123
-- ============================================================

-- ------------------------------------------------------------
-- Permissions
-- ------------------------------------------------------------
INSERT OR IGNORE INTO permissions (name, description) VALUES
    ('audit.view', 'View audit logs'),
    ('dashboard.view', 'View the dashboard'),
    ('report.view', 'View reports'),
    ('user.manage', 'Manage users and their roles'),
    ('role.manage', 'Manage roles and role permissions'),
    ('employee.manage', 'Manage employees'),
    ('organization.manage', 'Manage organizations, departments and positions'),
    ('recruitment.manage', 'Manage recruitment (job postings, interviews, candidates)'),
    ('evaluation.manage', 'Manage evaluation periods, templates, sections and questions'),
    ('evaluation.assign', 'Assign workers to evaluations'),
    ('evaluation.evaluate', 'Start, fill and submit assigned evaluations'),
    ('evaluation.approve', 'Create and decide evaluation approval requests');

-- ------------------------------------------------------------
-- Roles
-- ------------------------------------------------------------
INSERT OR IGNORE INTO roles (name, description) VALUES
    ('admin', 'Full system access'),
    ('hrm', 'HR management: employees, evaluations, reports'),
    ('manager', 'Manages workers and evaluates reports to them'),
    ('employee', 'Fills in the evaluations assigned to them');

-- ------------------------------------------------------------
-- Role permissions
-- ------------------------------------------------------------

-- admin: every permission in the catalog
INSERT OR IGNORE INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p
WHERE r.name = 'admin';

-- hrm
INSERT OR IGNORE INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p
WHERE r.name = 'hrm'
  AND p.name IN (
    'dashboard.view',
    'report.view',
    'employee.manage',
    'organization.manage',
    'recruitment.manage',
    'evaluation.manage',
    'evaluation.assign',
    'evaluation.evaluate',
    'evaluation.approve',
    'audit.view'
  );

-- manager
INSERT OR IGNORE INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p
WHERE r.name = 'manager'
  AND p.name IN (
    'dashboard.view',
    'report.view',
    'evaluation.assign',
    'evaluation.evaluate',
    'evaluation.approve'
  );

-- employee
INSERT OR IGNORE INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p
WHERE r.name = 'employee'
  AND p.name IN (
    'dashboard.view',
    'evaluation.evaluate'
  );

-- ------------------------------------------------------------
-- Development administrator
-- ------------------------------------------------------------
INSERT OR IGNORE INTO users (name, email, password, status) VALUES
    ('Administrator', 'admin@example.com', '$2a$10$y7G6U1K..UovJGi8JvVi.uA9DUfHzkQk1Ovjz3VB8PudYzyXE7xme', 'active');

INSERT OR IGNORE INTO user_roles (user_id, role_id)
SELECT u.id, r.id
FROM users u
JOIN roles r
WHERE u.email = 'admin@example.com'
  AND r.name = 'admin';
