# Backend API Documentation

Base URL: `/api` · Pages URL: `/pages`

Two interfaces:

- **`/api/*`** — REST, JSON in / JSON out, for programmatic clients.
- **`/pages/*`** — server-rendered HTML fragments for the vanilla JS SPA: form-encoded in,
  HTML out. Not intended for REST clients.

Authentication: `Authorization: Bearer <JWT>` on everything except the public
routes listed at the end (`/api/auth/login`, `/api/auth/register`,
`/api/components/:name`, `/jobs*`).

### Where the code lives

Routes are registered per area, and each area has its own handler package:

| Area | Route file | Handler package |
|---|---|---|
| Auth | `backend/routes/auth.go` | `app/handlers/auth.go` |
| Users | `backend/routes/user.go` | `app/handlers/user/` |
| Employees | `backend/routes/employee.go` | `app/handlers/employee/` |
| Candidates | `backend/routes/candidate.go` | `app/handlers/candidate/` |
| Organizations / departments / positions | `backend/routes/organization.go` | `app/handlers/organization/` |
| Dashboard | `backend/routes/dashboard.go` | `app/handlers/dashboard/` |
| Evaluation (periods, templates, sections, questions, assignments, responses, results, comments, approvals) | `backend/routes/evaluation.go` | `app/handlers/evaluation/` |
| Audit logs | `backend/routes/audit.go` | `app/handlers/audit/` |
| Recruitment (job postings, rounds, interviews, feedback, reports) | `backend/routes/recruitment.go` | `app/handlers/recruitment/` |
| Public job board / apply | `backend/routes/public.go` | `app/handlers/public/` |
| Shared components | `backend/routes/component.go` | `app/handlers/component.go` |
| Pages — dashboard | `backend/routes/pages/dashboard.go` | `app/handlers/pages/dashboard/` |
| Pages — organization & employees | `backend/routes/pages/organization.go`, `employees.go` | `app/handlers/pages/organization/` |
| Pages — candidates | `backend/routes/pages/candidates.go` | `app/handlers/pages/candidate/` |
| Pages — evaluations, periods, templates | `backend/routes/pages/evaluations.go` | `app/handlers/pages/evaluation/` |
| Pages — evaluation fill | `backend/routes/pages/evaluation_fill.go` | `app/handlers/pages/evaluation/` |
| Pages — cycle wizard | `backend/routes/pages/cycle.go` | `app/handlers/pages/cycle/` |
| Pages — approvals | `backend/routes/pages/approvals.go` | `app/handlers/pages/approval/` |
| Pages — reports | `backend/routes/pages/reports.go` | `app/handlers/pages/report/` |
| Pages — users, roles, permissions | `backend/routes/pages/users.go`, `roles.go`, `permissions.go` | `app/handlers/pages/user/` |
| Pages — recruitment | `backend/routes/pages/recruitment.go` | `app/handlers/pages/recruitment/` |

`backend/routes/routes.go` wires the `/api` group and the `/pages` group;
`backend/routes/pages/routes.go` registers every page area in one place.

---

# API Endpoints (JSON)

## Authentication

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/auth/login` | User login |
| POST | `/api/auth/register` | User registration |
| GET | `/api/auth/me` | Get current user. Returns user details plus `is_admin` (true when the user holds a role named "admin") |

Note: a role whose name is `admin` (case-insensitive) is treated as a superuser and is
the only role that can manage users, roles and permissions.

## Dashboard

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/dashboard/stats` | Get dashboard statistics |

## Users

Admin-only endpoints (require a user holding the `admin` role).

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/users` | List users (+ their current role IDs) |
| GET | `/api/users/:id` | Get user by ID |
| PUT | `/api/users/:id` | Update user (name, email, status, role) |
| DELETE | `/api/users/:id` | Delete user (a user may not delete themselves) |

## Employees

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/employees` | List employees |
| GET | `/api/employees/:id` | Get employee by ID |
| POST | `/api/employees` | Create employee |
| PUT | `/api/employees/:id` | Update employee |
| DELETE | `/api/employees/:id` | Delete employee |

## Organizations

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/organizations` | List organizations |
| POST | `/api/organizations` | Create organization |
| PUT | `/api/organizations/:id` | Update organization |
| DELETE | `/api/organizations/:id` | Delete organization |

## Departments

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/departments` | List departments |
| POST | `/api/departments` | Create department |
| PUT | `/api/departments/:id` | Update department |
| DELETE | `/api/departments/:id` | Delete department |

## Positions

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/positions` | List positions |
| POST | `/api/positions` | Create position |
| PUT | `/api/positions/:id` | Update position |
| DELETE | `/api/positions/:id` | Delete position |

## Job Postings

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/job-postings` | List job postings (with position/department names) |
| GET | `/api/job-postings/:id` | Get job posting by ID |
| POST | `/api/job-postings` | Create job posting |
| PUT | `/api/job-postings/:id` | Update job posting |
| DELETE | `/api/job-postings/:id` | Delete job posting |
| POST | `/api/job-postings/:id/publish` | Publish job posting |
| POST | `/api/job-postings/:id/close` | Close job posting |

## Interview Rounds

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/job-postings/:posting_id/rounds` | List rounds for a posting |
| POST | `/api/interview-rounds` | Create interview round |
| DELETE | `/api/interview-rounds/:id` | Delete interview round |

## Interviews

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/interviews` | List interviews (candidate/round/interviewer names) |
| GET | `/api/interviews/:id` | Get interview by ID |
| POST | `/api/interviews` | Schedule interview |
| PUT | `/api/interviews/:id/status` | Update interview status |
| DELETE | `/api/interviews/:id` | Delete interview |
| GET | `/api/interviews/candidate/:candidate_id` | List interviews by candidate |

## Interview Feedback

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/interviews/:interview_id/feedback` | Get feedback for an interview |
| POST | `/api/interviews/:interview_id/feedback` | Submit feedback |
| PUT | `/api/interview-feedback/:id` | Update feedback |

## Reports

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/reports/candidate-summary/:candidate_id` | Candidate evaluation summary |
| GET | `/api/reports/candidate-summaries` | All candidate summaries |
| GET | `/api/reports/interviewer-activity` | Interviewer activity report |
| GET | `/api/reports/hiring-decisions` | Hiring decisions report |

> Note: the candidate report queries join `candidates.position_id`, a column the
> current schema does not have (candidates carry `job_posting_id`); those
> endpoints return an error until the reports are reworked. The SQL lives in
> `backend/app/repositories/recruitment/interview_feedback.go`.

## Candidates

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/candidates` | List candidates (with posting titles) |
| POST | `/api/candidates` | Create candidate (multipart: details + CV file) |
| GET | `/api/candidates/:id/cv` | Download candidate CV |

Hiring (accepting a candidate as a worker) is done from the pages UI:
`POST /pages/candidates/:id/hire`.

## Evaluation Periods

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/evaluation-periods` | List periods |
| GET | `/api/evaluation-periods/:id` | Get period by ID |
| POST | `/api/evaluation-periods` | Create period |
| PUT | `/api/evaluation-periods/:id` | Update period |
| DELETE | `/api/evaluation-periods/:id` | Delete period |

## Evaluation Templates

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/evaluation-templates` | List templates |
| GET | `/api/evaluation-templates/:id` | Get template by ID |
| POST | `/api/evaluation-templates` | Create template |
| PUT | `/api/evaluation-templates/:id` | Update template |
| DELETE | `/api/evaluation-templates/:id` | Delete template |
| GET | `/api/evaluation-templates/:template_id/sections` | List sections by template |

## Evaluation Sections

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/evaluation-sections` | Create section |
| PUT | `/api/evaluation-sections/:id` | Update section |
| DELETE | `/api/evaluation-sections/:id` | Delete section |
| GET | `/api/evaluation-sections/:section_id/questions` | List questions by section |

## Evaluation Questions

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/evaluation-questions` | Create question |
| PUT | `/api/evaluation-questions/:id` | Update question |
| DELETE | `/api/evaluation-questions/:id` | Delete question |

## Evaluation Assignments

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/evaluation-assignments` | List assignments |
| GET | `/api/evaluation-assignments/:id` | Get assignment by ID |
| POST | `/api/evaluation-assignments` | Create assignment |
| POST | `/api/evaluation-assignments/:id/start` | Start evaluation |
| POST | `/api/evaluation-assignments/:id/submit` | Submit evaluation |

## Evaluation Responses

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/evaluation-responses` | Save response (upsert per assignment+question) |
| GET | `/api/evaluation-assignments/:assignment_id/responses` | Get responses for assignment |

## Evaluation Results

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/evaluation-assignments/:assignment_id/calculate` | Calculate result |
| GET | `/api/evaluation-assignments/:assignment_id/result` | Get result |

## Evaluation Comments

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/evaluation-comments` | Create comment |
| GET | `/api/evaluation-assignments/:assignment_id/comments` | List comments for assignment |

## Evaluation Approvals

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/evaluation-approvals/pending` | List pending approvals |
| POST | `/api/evaluation-approvals` | Create approval request |
| POST | `/api/evaluation-approvals/:id/approve` | Approve evaluation |
| POST | `/api/evaluation-approvals/:id/reject` | Reject evaluation |

## Audit Logs

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/audit-logs` | List audit logs (newest first, paginated) |

## Components

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/components/:name` | Get a shared frontend component (no auth required) |

The endpoint is public and serves only the allowlisted components `auth`,
`header` and `sidebar`; any other name returns 404 even if a file exists.

---

# Pages API (SPA Fragments)

Server-rendered HTML fragments under `/pages`, used by the frontend. All routes
require authentication. Requests are `application/x-www-form-urlencoded`;
successful responses are HTML fragments (`text/html`) loaded dynamically by the vanilla JS SPA.
**Errors are always JSON** (`{"error": "..."}`) with a non-2xx status; the
frontend turns them into a toast or an in-page message.

## Dashboard

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/pages` · `/pages/dashboard` | Dashboard with stats for the current user |

## Employees

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/pages/employees` | Employees page (worker candidates, departments, positions) |
| POST | `/pages/employees` | Create employee from a selected user |

Validates that the user is active, has no worker record yet, and the department/position exist and are active.

## Organizations / Departments / Positions

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/pages/organizations` | Organizations, departments, positions overview |
| POST | `/pages/organizations` | Create organization |
| PUT | `/pages/organizations/:id` | Update organization |
| DELETE | `/pages/organizations/:id` | Delete organization |
| GET | `/pages/departments` | Departments section |
| POST | `/pages/departments` | Create department in an existing organization |
| PUT | `/pages/departments/:id` | Update department |
| DELETE | `/pages/departments/:id` | Delete department |
| GET | `/pages/positions` | Positions section |
| POST | `/pages/positions` | Create position in an existing department |
| PUT | `/pages/positions/:id` | Update position |
| DELETE | `/pages/positions/:id` | Delete position |

## Users, Roles & Permissions

Admin-only pages (require a user holding the `admin` role). Non-admins do not see
these links in the sidebar.

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/pages/users` | Users page |
| POST | `/pages/users` | Create user |
| PUT | `/pages/users/:id` | Update user (name, email, status, role) |
| DELETE | `/pages/users/:id` | Delete user |

The user edit dialog pre-selects the user's current role; the role selector is
only shown to admins.

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/pages/roles` | Roles page |
| POST | `/pages/roles` | Create role |
| PUT | `/pages/roles/:id` | Update role (name, description) |
| DELETE | `/pages/roles/:id` | Delete role |
| POST | `/pages/roles/:id/permissions` | Set role permissions (`permission_id[]`) |

Role permissions are picked per **feature area** (General, Organization,
Recruitment, Evaluation, Analytics, Administration) using an **access level**
dropdown: No access / View only / Manage / Full control / Custom. Expanding
"Customize individual permissions" reveals the raw permission checkboxes for
fine-tuning. The Evaluation area offers dedicated levels: **Evaluate (fill my
assigned)**, **Assign and manage**, and **Full control**.

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/pages/permissions` | Permissions page |
| POST | `/pages/permissions` | Create permission |
| DELETE | `/pages/permissions/:id` | Delete permission |

## Candidates

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/pages/candidates` | Candidates page |
| POST | `/pages/candidates` | Create candidate with CV upload |
| POST | `/pages/candidates/:id/assign-eval` | Assign evaluation to candidate |
| POST | `/pages/candidates/:id/hire` | Accept candidate: creates active user + worker record |

Hiring requires a rating of 1–5 and a temporary password (min 8 characters);
the new account gets the position from the candidate's job posting.

## Evaluations

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/pages/evaluations` | Periods, templates, and the current user's assignments |
| POST | `/pages/evaluations/assignments` | Create assignment |
| POST | `/pages/evaluations/:id/start` | Start assignment |

Assignment validation: period/template must exist, template must be active, evaluator ≠ evaluatee, and both must be selected (active) workers.

An assignment can be started, filled and submitted by either:
- the assigned evaluator holding the `evaluation.evaluate` permission, or
- any user holding `evaluation.manage` or `evaluation.assign` (admins/managers can drive any assignment).

## Evaluation Fill

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/pages/evaluation-fill/:id` | Fill form (open assignments) or read-only result + comments |
| POST | `/pages/evaluation-fill/:id/start` | Start from the fill view |
| POST | `/pages/evaluation-fill/:id` | Save all answers and submit; recalculates the result |
| POST | `/pages/evaluation-fill/:id/comments` | Add comment, returns updated comment list fragment |

## Approvals

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/pages/approvals` | Pending approvals for the current user + request form data |
| POST | `/pages/approvals` | Request approval for an assignment from a selected worker |
| POST | `/pages/approvals/:id/approve` | Approve (only the assigned approver) |
| POST | `/pages/approvals/:id/reject` | Reject (only the assigned approver); sends the evaluation back to pending |

Approval flow: submitting an evaluation auto-creates a level-1 approval request for the evaluator.
Approving completes the assignment; rejecting resets it to pending so the evaluator can revise and resubmit,
which creates a fresh approval request.

## Evaluation Periods

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/pages/periods` | Periods page |
| POST | `/pages/periods` | Create evaluation period |
| PUT | `/pages/periods/:id` | Update period |
| DELETE | `/pages/periods/:id` | Delete period |
| GET | `/pages/periods/:id/history` | Period history (results, approvals, averages) |

## Evaluation Templates (builder)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/pages/templates-eval` | Templates page |
| GET | `/pages/templates-eval/:id` | Template detail (sections + questions) |
| POST | `/pages/templates` | Create template |
| PUT | `/pages/templates/:id` | Update template |
| DELETE | `/pages/templates/:id` | Delete template |
| POST | `/pages/templates-eval/:id/sections` | Add section to template |
| POST | `/pages/sections/:id/questions` | Add question to section |
| DELETE | `/pages/sections/:id` | Delete section |
| DELETE | `/pages/questions/:id` | Delete question |

## Cycle Wizard (bulk assignment launcher)

One decision per screen; answers are carried forward as hidden fields.
The order is period → template → scope → review.

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/pages/cycle` | Wizard start (step 1: pick or create a period) |
| POST | `/pages/cycle/period` | Choose a period, or create one (`new_period_name`, `new_period_start`, `new_period_end`) |
| POST | `/pages/cycle/template` | Choose the template (inactive templates and templates with no questions are refused) |
| POST | `/pages/cycle/scope` | Choose scope (who gets evaluated) and the evaluator |
| POST | `/pages/cycle/assignment/drop` | Take a pending/started candidate evaluation back out of the period |
| POST | `/pages/cycle/launch` | Create all assignments |
| POST | `/pages/cycle/back/period` · `back/template` · `back/scope` | Navigate back |

## Reports

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/pages/reports` | Results per reviewed assignment + average score per department |
| GET | `/pages/recruitment-reports` | Recruitment analytics (candidate summaries, interviewer activity, hiring decisions) |

## Job Postings

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/pages/job-postings` | Job postings page |
| POST | `/pages/job-postings` | Create job posting |
| POST | `/pages/job-postings/:id/publish` | Publish job posting |
| POST | `/pages/job-postings/:id/close` | Close job posting |
| POST | `/pages/job-postings/:id/reopen` | Reopen a closed job posting |
| POST | `/pages/job-postings/:id/delete` | Delete job posting |
| GET | `/pages/job-postings/:id/candidates` | View candidates for a specific posting |

## Job Posting Wizard

Same shape as the cycle wizard: one decision per screen.

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/pages/job-wizard` | Wizard start |
| POST | `/pages/job-wizard/basics` | Step: basics |
| POST | `/pages/job-wizard/details` | Step: details |
| POST | `/pages/job-wizard/terms` | Step: terms |
| POST | `/pages/job-wizard/launch` | Create + publish |
| POST | `/pages/job-wizard/back/basics` · `back/details` · `back/terms` | Navigate back |

## Interviews

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/pages/interviews` | Interview scheduling page |
| POST | `/pages/interviews` | Schedule interview |
| POST | `/pages/interviews/:id/status` | Update interview status |
| POST | `/pages/interviews/:id/delete` | Delete interview |
| GET | `/pages/interviews/:id/feedback` | Interview feedback form |
| POST | `/pages/interviews/:id/feedback` | Submit interview feedback |

---

## Notes

- All endpoints except `/api/auth/login`, `/api/auth/register`,
  `/api/components/:name`, and `/jobs*` require authentication.
- Protected endpoints require a valid JWT in the `Authorization` header
  (`Bearer <token>`).
- `/api/*` endpoints: JSON in/out (`Content-Type: application/json`).
- `/pages/*` endpoints: form-encoded in, HTML out on success, JSON on error —
  meant for SPA fragments, not REST clients.
- `/jobs*` endpoints: public, no authentication required.
- Worker selection rule: evaluators, evaluatees, approvers, and new employees must always be chosen from
  active workers via dropdowns; arbitrary or hardcoded IDs are rejected server-side.
- Approve/reject are restricted to the approval's assigned approver and only work while it is still pending.
- Sidebar navigation is filtered by permission: links with `data-admin` (Users/Roles/Permissions) are hidden
  from non-admins, and other links require their matching permission level to be visible.
- Not registered, although the code exists: the skills endpoints (`app/handlers/recruitment/skill.go`)
  and the rating-scale pages (no handler and no route; only the schema and an unused
  `frontend/templates/pages/rating_scales.html` are left).

---

## Public Routes (No Auth Required)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/jobs` | List published job postings |
| GET | `/jobs/:id` | View job posting details |
| GET | `/apply` | Application form (query param: `posting_id`) |
| POST | `/apply` | Submit application (multipart: name, email, phone, skills, CV file) |
