-- Employees created by the accept-worker flow do not get a login, so
-- employees.user_id has to be nullable. The original table was created with a
-- NOT NULL constraint; SQLite cannot drop that in place, so rebuild the table
-- and copy the existing rows across.
--
-- Other tables reference employees (evaluation_assignments.evaluatee_id) and
-- employees references itself through manager_id, so foreign key checks are
-- disabled while the table is swapped. The pragma cannot be changed inside a
-- transaction, hence NO TRANSACTION and the explicit BEGIN/COMMIT below.

-- +goose NO TRANSACTION

-- +goose Up

PRAGMA foreign_keys = OFF;

BEGIN;

CREATE TABLE employees_new (
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
    can_interview INTEGER NOT NULL DEFAULT 1,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL,
    FOREIGN KEY (department_id) REFERENCES departments(id) ON DELETE SET NULL,
    FOREIGN KEY (position_id) REFERENCES positions(id) ON DELETE SET NULL,
    FOREIGN KEY (manager_id) REFERENCES employees(id) ON DELETE SET NULL
);

INSERT INTO employees_new (
    id, user_id, employee_code, department_id, position_id, manager_id,
    hire_date, status, created_at, updated_at, can_interview
)
SELECT
    id, user_id, employee_code, department_id, position_id, manager_id,
    hire_date, status, created_at, updated_at, COALESCE(can_interview, 1)
FROM employees;

DROP TABLE employees;

ALTER TABLE employees_new RENAME TO employees;

COMMIT;

PRAGMA foreign_keys = ON;

-- +goose Down

SELECT 1;
