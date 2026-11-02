-- Keep the source candidate linked to employees created by the hiring flow so
-- HR can display their name even when they do not have a login account.
-- +goose Up

ALTER TABLE employees ADD COLUMN candidate_id INTEGER REFERENCES candidates(id) ON DELETE SET NULL;

-- Earlier hires encode the candidate id in their generated employee code.
UPDATE employees
SET candidate_id = CAST(SUBSTR(employee_code, 5) AS INTEGER)
WHERE employee_code GLOB 'EMP-[0-9]*'
  AND EXISTS (
      SELECT 1 FROM candidates c
      WHERE c.id = CAST(SUBSTR(employees.employee_code, 5) AS INTEGER)
  );

CREATE INDEX IF NOT EXISTS idx_employees_candidate_id ON employees(candidate_id);

-- +goose Down

SELECT 1;
