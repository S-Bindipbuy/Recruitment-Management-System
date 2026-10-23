# Evaluation Management System (EMS)

## 1. Project Overview

**Evaluation Management System (EMS)** is a full-stack system for creating, assigning, completing, reviewing, approving, and reporting evaluations, plus the recruitment workflow around it (job postings, candidates, interviews, hiring).

### Main Stack

- **Frontend:** Server-rendered HTML templates + vanilla JavaScript SPA (custom router/navigation), hand-written CSS design system (`frontend/static/css/styles.css`, Rosé Pine Dawn palette) — no Node.js, Tailwind, or shadcn/ui build step
- **Backend:** Go + Fiber
- **Database:** Turso/libSQL (local `file:` database, remote Turso database, or embedded replica)
- **Data access:** [sqlx](https://github.com/jmoiron/sqlx) (`Get`/`Select` struct scanning) over `database/sql`
- **Migrations:** [goose](https://github.com/pressly/goose) (versioned SQL files, per-dialect directories)
- **Authentication:** JWT (Bearer), manual RBAC (roles + permissions)
- **API:** REST (`/api/*`, JSON) + server-rendered HTML fragments for SPA (`/pages/*`)
- **Build/packaging:** Nix flake (`nix build`, `nix develop`)

---

## 2. System Modules

- Authentication & Authorization
- Users & Roles
- Organization / Departments / Positions
- Employees
- Candidates
- Skills
- Job Postings (wizard + cycle launchers)
- Interview Rounds
- Interviews
- Interview Feedback
- Evaluation Periods
- Evaluation Templates
- Evaluation Sections
- Evaluation Questions
- Evaluation Assignments
- Evaluation Responses
- Evaluation Results
- Evaluation Comments
- Evaluation Approvals
- Reports
- Recruitment Reports
- Audit Logs

---

## 3. Architecture

### Request flow

```text
HTTP request
  └── Fiber router (routes/)
        ├── public routes (/jobs, /apply, /api/auth/login, /api/components/:name)
        ├── middleware.Protected()  — JWT from Authorization: Bearer …
        └── handler (app/handlers/)  — parse input, call service
              └── service (app/services/)  — business rules
                    └── repository (app/repositories/)  — sqlx queries
                          └── database.DB (*sqlx.DB)
                                └── Turso / libSQL
```

`/api/*` endpoints speak JSON; `/pages/*` endpoints accept form-encoded input and
return HTML fragments for the SPA. Errors are the exception on both: the app-wide
error handler (`app/httperr`) always answers with JSON (`{"error": "..."}`), and
the frontend turns it into a toast or an in-page message.

### Package layout

| Path | Role |
|---|---|
| `backend/main.go` | Boot: config → `database.Connect` → `RunMigrations` → seeds → Fiber + graceful shutdown |
| `backend/config/` | Environment configuration (see §5) |
| `backend/database/` | Connection, goose migrations, seeds (`seed.go`, `seed.sql`) |
| `backend/routes/` | REST route registration (`/api/*`), one file per area |
| `backend/routes/pages/` | Every `/pages/*` route, one file per area |
| `backend/app/handlers/` | REST handlers, one package per area |
| `backend/app/handlers/pages/` | Page handlers for SPA fragments, one package per area |
| `backend/app/handlers/pages/core/` | Shared page plumbing: template loading, form parsing, permission checks |
| `backend/app/services/` | Business logic, one package per area |
| `backend/app/repositories/` | Data access (sqlx), one package per area |
| `backend/app/models/` | Structs with `db` tags, one package per area |
| `backend/app/middleware/` | JWT auth |
| `backend/app/enums/` | Status/permission constants |
| `backend/app/httperr/` | JSON error handler wired into Fiber |
| `frontend/` | Static assets + HTML templates |

### One package per area

The tree is split by **feature area**, not by layer: the same area name is
reused in every layer, so a feature is found in one place per layer instead of
being scattered over files named after it.

| Area | Packages |
|---|---|
| audit | `handlers/audit`, `repositories/audit` |
| candidate | `handlers/candidate`, `handlers/pages/candidate`, `models/candidate`, `repositories/candidate`, `services/candidate` |
| cycles | `services/cycles` (wizard rules), `handlers/pages/cycle` (wizard HTTP) |
| dashboard | `handlers/dashboard`, `handlers/pages/dashboard` |
| employee | `handlers/employee`, `models/organization`, `repositories/organization`, `services/organization` |
| evaluation | `handlers/evaluation`, `handlers/pages/evaluation`, `models/evaluation`, `repositories/evaluation`, `services/evaluation` |
| organization | `handlers/organization`, `handlers/pages/organization`, `models/organization`, `repositories/organization`, `services/organization` |
| public | `handlers/public` (job board + application form) |
| recruitment | `handlers/recruitment`, `handlers/pages/recruitment`, `models/recruitment`, `repositories/recruitment`, `services/recruitment` |
| user | `handlers/user`, `handlers/pages/user`, `models/user`, `repositories/user`, `services/user` |

Conventions inside an area:

- **`instances.go`** holds the singletons of a services/repositories package
  (`PeriodServiceInstance`, `PeriodRepo`, …). Handlers use those instances;
  nothing is constructed per request.
- **One file per resource** where an area is big enough to need it
  (`evaluation/period.go`, `evaluation/template.go`, `evaluation/result.go`, …);
  a small area keeps a single file.
- **`handlers/pages/core`** is the only shared plumbing page handlers import:
  `RenderPage`/`RenderTemplate`, form parsing (`FormInt64`, `ReadScopeForm`),
  permission checks (`RequirePermission`, `RequireCanFillEvaluation`) and the
  worker lookups the templates are written against.
- **Route registration mirrors the handler packages**: `routes/*.go` for
  `/api/*`, `routes/pages/*.go` for `/pages/*`. `routes/pages/routes.go` is the
  single place where every page area is registered.

### Templates

Every file under `frontend/templates/pages/` is parsed on its own into its own
template set, so a syntax error in one page answers 500 for that page only, and
the error names the file that broke. The public pages
(`frontend/templates/content_*.html`) are parsed together with
`public_layout.html`, which they are wrapped in.

### Data-access conventions (sqlx)

- Repositories use `database.DB.Get` (one row) and `database.DB.Select` (many rows)
  with structs scanned **by column name**; models carry `db:"column"` tags.
- SQL expressions must be aliased to the destination name
  (`COALESCE(jp.title, '') AS posting_title`, `COUNT(i.id) AS total_interviews`).
- `COALESCE` is used only where the value can be NULL **and** the destination is a
  plain `string`/`float64` — nullable fields are modeled as pointers (`*int64`,
  `*time.Time`, `*float64`) so NULL scans without help.
- Embedded structs (e.g. `CandidateListItem` → `Candidate`) flatten automatically.
- Writes use `Exec` with `?` placeholders; multi-statement work runs in `Beginx` transactions.

---

## 4. Database & Migrations

### Tooling

Schema changes are managed by **goose**. The SQL lives in a single directory
because libSQL is SQLite-compatible:

```text
backend/database/migrations/
└── turso/000001_init.sql     # schema for local libSQL and remote Turso
```

Each file is ordinary SQL with goose annotations as comments:

```sql
-- +goose Up
CREATE TABLE IF NOT EXISTS ...;
-- +goose Down
DROP TABLE IF EXISTS ...;
```

(No `StatementBegin`/`StatementEnd` markers are needed: statements are
`;`-terminated and comments inside statements are avoided. File-level docs go
before `-- +goose Up`.)

### Boot behavior

`main.go` calls `database.RunMigrations()` on every start:

1. goose uses its `turso` querier (SQLite-compatible).
2. `goose.Up` applies pending migrations from `migrations/turso`; versions are
   tracked in the `goose_db_version` table, so a no-op run is cheap.
3. Failure to migrate is fatal (`log.Fatalf`) — the app never runs against a wrong schema.

Everything in `000001_init.sql` uses `IF NOT EXISTS`, so a database created
before goose (by the old hand-rolled migration) is **adopted in place**: the
statements no-op on existing tables and goose simply records version 1. The
migration also repairs legacy data — duplicate `evaluation_responses` rows are
deleted so the unique index (required by `Upsert`) can be created, and stale
zero-score `evaluation_results` without responses are cleared.

### Adding a migration

1. Create `backend/database/migrations/turso/0000N2_<what>.sql`.
2. Annotate with `-- +goose Up` and `-- +goose Down`; make every statement
   idempotent where possible so partial states converge.
3. Never edit an already-applied migration — add a new one instead.
4. Use SQLite-compatible types (`INTEGER PRIMARY KEY AUTOINCREMENT`, `DATETIME`,
   `INTEGER`, `TEXT`). The same SQL runs locally and against Turso Cloud.
5. Run the app once (`./result/bin/backend`) and check that the migration
   applied cleanly — there is no migration test suite left in the repo
   (see §6).

### Local, remote and replica connections

`database/connection.go` picks a mode from `TURSO_DATABASE_URL`:

- `file:./storage/db` — local libSQL/SQLite file; WAL and a 5s busy timeout are
  applied on connect.
- `libsql://<db>-<org>.turso.io` — remote Turso; `TURSO_AUTH_TOKEN` is appended
  as the `authToken` parameter.
- `http://127.0.0.1:8080` — a local `turso dev` server.
- Remote URL + `TURSO_REPLICA_PATH` — embedded replica: reads are local, writes
  go to the remote primary. `TURSO_SYNC_INTERVAL` (seconds) enables periodic
  background sync; `0` keeps only the initial sync.

All modes use the same migrations and SQL. The native library supplied by
`go-libsql` is vendored; after `go mod vendor` run
`backend/scripts/vendor-libsql.sh` (vendor omits the driver's `lib/` directory).

### Seeds

- **`database/seed.sql`** — portable, SQLite-compatible import (roles, permissions,
  defaults). Usable manually: `sqlite3 storage/db < backend/database/seed.sql`.
- **`SeedDefaultRBAC()`** — runs at boot: creates the permission catalog
  (12 permissions), roles (`admin`, manager, etc.), grants the admin role every
  permission.
- **`SeedAdminUser()`** — runs at boot: ensures the development admin
  exists (`admin@example.com` / `admin123`, role `admin`, every permission).
  Set `APP_ENV=production` to disable it.

---

## 5. Configuration

Environment variables (defaults from `backend/config/config.go`; `APP_ENV`
and `ENV_FILE` are read straight from the environment where they are used):

| Variable | Default | Purpose |
|---|---|---|
| `SERVER_PORT` | `3000` | HTTP port |
| `TURSO_DATABASE_URL` | **required** | libSQL/Turso URL: local `file:` path, `libsql://…` remote, or `http://…` for `turso dev` |
| `TURSO_AUTH_TOKEN` | — | Auth token for remote Turso databases (not needed for local files/`turso dev`) |
| `TURSO_REPLICA_PATH` | — | When set with a remote URL, opens an embedded replica at this local path |
| `TURSO_SYNC_INTERVAL` | `0` | Replica background sync interval in seconds (`0` = initial sync only) |
| `APP_ENV` | — | `production` disables the dev-admin seeder (read in `database/seed.go`) |
| `JWT_SECRET` | **required** | JWT signing key; no insecure default |
| `JWT_EXPIRATION_HOURS` | `24` | Token lifetime |
| `ALLOWED_ORIGINS` | `*` | CORS |
| `FRONTEND_PATH` | **required** | Static assets root; the server refuses to start when unset |
| `UPLOAD_DIR` | `storage/uploads/cvs` | CV upload directory |
| `REDIS_ADDR` / `REDIS_PASSWORD` / `REDIS_DB` | `localhost:6379` / empty / `0` | Reserved for cache/queue |
| `ENV_FILE` | — | Explicit path to the `.env` file to load |

`config.Load()` loads the `.env` file from the current working directory (or
`ENV_FILE` when set). `.env` is a local-development convenience; in production,
provide configuration through the service environment instead (systemd
`EnvironmentFile`, Docker `env_file`, or the Nix module). `.env` is git-ignored.

`TURSO_DATABASE_URL`, `JWT_SECRET` and `FRONTEND_PATH` have no defaults: if any
of them is missing or empty, `config.Load()` logs a clear error and exits before
the HTTP server starts. Copy `.env-example` to `.env` to get started.

Local databases use WAL with a 5s busy timeout. Remote connections append
`TURSO_AUTH_TOKEN` as `authToken`; an `authToken` already present in
`TURSO_DATABASE_URL` takes precedence when the env var is empty.

---

## 6. Development

### Build

```bash
nix build              # backend (vet + tests run in the sandbox) + packaged frontend assets
nix develop            # dev shell with the ems-compile helper
ems-compile            # go build -C backend -o main  (moves the binary to ./main)
```

The Nix build uses `buildGoModule` with `vendorHash = null` → the committed
`backend/vendor/` directory is the source of truth. After adding or upgrading
dependencies run:

```bash
go mod tidy && go mod vendor
```

### Test

```bash
go vet ./...           # the check that still says something
go test ./...          # compiles every package; no test files are registered
```

**There are no `_test.go` files in the repository today.** The migration suite
(`database/migrate_test.go`), the evaluation scoring suite
(`app/services/evaluation_scoring_test.go`), the template render check
(`app/handlers/tpl_render_check_test.go`), the golden-page tests
(`app/handlers/pages/golden_test.go`, `template_isolation_test.go` and their
`testdata/golden/` fixtures) and the route table test (`routes/http_test.go`)
were removed during the package split. `go vet ./...` and `nix build` are what
the project is currently checked with.

### Run

```bash
./result/bin/backend                     # or a locally built binary
TURSO_DATABASE_URL=file:/path/to/db SERVER_PORT=3111 ./result/bin/backend
```

On boot the server applies migrations, seeds RBAC + dev admin, and listens.
Shutdown on SIGINT/SIGTERM drains in-flight requests.

---

# 7. Database Schema

Portable type names below; the SQLite-compatible specifics live in
`backend/database/migrations/turso/000001_init.sql` (authoritative source).

## 7.1 Authentication & Authorization

### `users`

| Column | Type | Notes |
|---|---|---|
| id | bigint | PK |
| name | varchar | |
| email | varchar | Unique |
| password | varchar | bcrypt hash |
| status | varchar | active/inactive |
| created_at | timestamp | |
| updated_at | timestamp | |

### `roles`

| Column | Type | Notes |
|---|---|---|
| id | bigint | PK |
| name | varchar | Unique |
| description | text | Nullable |
| created_at | timestamp | |
| updated_at | timestamp | |

### `permissions`

| Column | Type | Notes |
|---|---|---|
| id | bigint | PK |
| name | varchar | Unique |
| description | text | Nullable |
| created_at | timestamp | |
| updated_at | timestamp | |

### `user_roles`

- `user_id` → `users.id`
- `role_id` → `roles.id`

### `role_permissions`

- `role_id` → `roles.id`
- `permission_id` → `permissions.id`

> **Role model:** a role named `admin` (case-insensitive) is treated as a superuser and is the
> only role that can manage users, roles and permissions. In the UI, role permissions are assigned
> per feature area (General, Organization, Recruitment, Evaluation, Analytics, Administration) using
> an access level (No access / View only / Manage / Full control / Custom) mapped to the raw
> permission rows below; the Evaluation area offers dedicated "Evaluate" and "Assign and manage" levels.
> An evaluation assignment can be started/filled/submitted by its assigned evaluator (requires
> `evaluation.evaluate`) or by any user holding `evaluation.manage`/`evaluation.assign`.

---

## 7.2 Organization

### `organizations`

| Column | Type | Notes |
|---|---|---|
| id | bigint | PK |
| name | varchar | |
| description | text | Nullable |
| status | varchar | |
| created_at | timestamp | |
| updated_at | timestamp | |

### `departments`

| Column | Type | Notes |
|---|---|---|
| id | bigint | PK |
| organization_id | bigint | FK organizations |
| name | varchar | |
| description | text | Nullable |
| status | varchar | |
| created_at | timestamp | |
| updated_at | timestamp | |

### `positions`

| Column | Type | Notes |
|---|---|---|
| id | bigint | PK |
| department_id | bigint | FK departments |
| name | varchar | |
| description | text | Nullable |
| status | varchar | |
| created_at | timestamp | |
| updated_at | timestamp | |

### `employees`

| Column | Type | Notes |
|---|---|---|
| id | bigint | PK |
| user_id | bigint | FK users |
| employee_code | varchar | Unique (`EMP-…`) |
| department_id | bigint | FK departments, Nullable |
| position_id | bigint | FK positions, Nullable |
| manager_id | bigint | Self FK employees, Nullable |
| hire_date | date | Nullable |
| status | varchar | |
| created_at | timestamp | |
| updated_at | timestamp | |

---

## 7.3 Recruitment

### `job_postings`

| Column | Type | Notes |
|---|---|---|
| id | bigint | PK |
| position_id | bigint | FK positions |
| title | varchar | |
| description | text | |
| requirements | text | |
| skills_required | text | Nullable (legacy free-text field) |
| salary_min / salary_max | decimal | Nullable |
| employment_type | varchar | default `full-time` |
| status | varchar | draft/published/closed |
| posted_at / closed_at | timestamp | Nullable |
| created_by | bigint | FK users |
| created_at / updated_at | timestamp | |

### `candidates`

| Column | Type | Notes |
|---|---|---|
| id | bigint | PK |
| full_name | varchar | |
| email | varchar | |
| phone | varchar | Nullable |
| job_posting_id | bigint | FK job_postings, Nullable |
| skills | text | free-text skills |
| cv_path / cv_name / cv_mime_type / cv_size | — | uploaded CV |
| status | varchar | new/…/hired |
| created_at / updated_at | timestamp | |

### `candidate_assessments`

| Column | Type | Notes |
|---|---|---|
| id | bigint | PK |
| candidate_id | bigint | FK candidates |
| evaluator_id | bigint | FK users |
| skill_score | decimal | |
| notes | text | |
| recommendation | varchar | e.g. `accepted` on hire |
| created_at | timestamp | |

### `skills`

| Column | Type | Notes |
|---|---|---|
| id | bigint | PK |
| name | varchar | Unique |
| category | varchar | Nullable |
| description | text | |
| created_at | timestamp | |

### `candidate_skills`

- `candidate_id` → `candidates.id`, `skill_id` → `skills.id` (composite PK)
- `proficiency` varchar (beginner/intermediate/…), `years_experience` integer

### `job_posting_skills`

- `job_posting_id` → `job_postings.id`, `skill_id` → `skills.id` (composite PK)
- `is_required` boolean, `min_proficiency` varchar

### `interview_rounds`

| Column | Type | Notes |
|---|---|---|
| id | bigint | PK |
| job_posting_id | bigint | FK job_postings |
| name | varchar | e.g. HR / Technical / Final |
| round_number | integer | ordering |
| round_type | varchar | default `technical` |
| description | text | |
| max_score / pass_score | decimal | default 10.0 / 6.0 |
| status | varchar | |
| created_at / updated_at | timestamp | |

### `interviews`

| Column | Type | Notes |
|---|---|---|
| id | bigint | PK |
| candidate_id | bigint | FK candidates |
| interview_round_id | bigint | FK interview_rounds |
| interviewer_id | bigint | FK users |
| scheduled_at | timestamp | |
| duration_minutes | integer | default 60 |
| location / meeting_link | varchar | Nullable |
| status | varchar | scheduled/completed/cancelled |
| notes | text | |
| created_at / updated_at | timestamp | |

### `interview_feedback`

| Column | Type | Notes |
|---|---|---|
| id | bigint | PK |
| interview_id | bigint | FK interviews, Unique (one feedback per interview) |
| interviewer_id | bigint | FK users |
| score | decimal | 0–10, Nullable |
| recommendation | varchar | pending/strong-hire/… |
| strengths / weaknesses / comments | text | |
| submitted_at | timestamp | Nullable |
| created_at / updated_at | timestamp | |

---

## 7.4 Evaluation

### `evaluation_periods`

| Column | Type | Notes |
|---|---|---|
| id | bigint | PK |
| name | varchar | |
| description | text | Nullable |
| start_date | date | |
| end_date | date | |
| status | varchar | draft/open/closed/archived |
| created_by | bigint | FK users |
| created_at | timestamp | |
| updated_at | timestamp | |

### `evaluation_templates`

| Column | Type | Notes |
|---|---|---|
| id | bigint | PK |
| name | varchar | |
| description | text | Nullable |
| status | varchar | |
| created_by | bigint | FK users |
| created_at | timestamp | |
| updated_at | timestamp | |

### `evaluation_sections`

| Column | Type | Notes |
|---|---|---|
| id | bigint | PK |
| evaluation_template_id | bigint | FK evaluation_templates |
| name | varchar | |
| description | text | Nullable |
| sort_order | integer | |
| created_at | timestamp | |
| updated_at | timestamp | |

### `evaluation_questions`

| Column | Type | Notes |
|---|---|---|
| id | bigint | PK |
| evaluation_section_id | bigint | FK evaluation_sections |
| question | text | |
| description | text | Nullable |
| question_type | varchar | rating/text/number/boolean/option |
| is_required | boolean | |
| weight | decimal | scoring weight |
| sort_order | integer | |
| status | varchar | |
| created_at | timestamp | |
| updated_at | timestamp | |

### `rating_scales` / `rating_scale_options`

- `rating_scales`: id PK, name, description, timestamps.
- `rating_scale_options`: id PK, `rating_scale_id` → `rating_scales.id`,
  label, value (decimal), description, sort_order.

### `evaluation_assignments`

Connects the **evaluator**, **evaluatee**, **period**, and **template**.

| Column | Type | Notes |
|---|---|---|
| id | bigint | PK |
| evaluation_period_id | bigint | FK |
| evaluation_template_id | bigint | FK |
| evaluator_id | bigint | FK users |
| evaluatee_id | bigint | FK employees/users |
| candidate_id | bigint | FK candidates, Nullable — set when evaluating a candidate instead of `evaluatee_id` |
| due_date | date | Nullable |
| status | varchar | pending/started/submitted/completed/overdue |
| started_at | timestamp | Nullable |
| submitted_at | timestamp | Nullable |
| completed_at | timestamp | Nullable |
| created_at | timestamp | |
| updated_at | timestamp | |

### `evaluation_responses`

| Column | Type | Notes |
|---|---|---|
| id | bigint | PK |
| evaluation_assignment_id | bigint | FK |
| evaluation_question_id | bigint | FK |
| rating_value | decimal | Nullable |
| text_value | text | Nullable |
| numeric_value | decimal | Nullable |
| boolean_value | boolean | Nullable |
| selected_option_id | bigint | Nullable |
| created_at | timestamp | |
| updated_at | timestamp | |

> **Unique index** `idx_evaluation_responses_assignment_question` on
> `(evaluation_assignment_id, evaluation_question_id)` — required by
> `ResponseRepo.Upsert` (`ON CONFLICT … DO UPDATE`). The migration creates it
> after deleting any duplicate rows left by older data.

### `evaluation_results`

| Column | Type | Notes |
|---|---|---|
| id | bigint | PK |
| evaluation_assignment_id | bigint | FK, Unique |
| total_score | decimal | |
| percentage | decimal | |
| grade | varchar | Nullable |
| calculated_at | timestamp | |
| created_at | timestamp | |
| updated_at | timestamp | |

### `evaluation_comments`

| Column | Type | Notes |
|---|---|---|
| id | bigint | PK |
| evaluation_assignment_id | bigint | FK |
| evaluation_question_id | bigint | Nullable |
| user_id | bigint | FK |
| comment | text | |
| created_at | timestamp | |
| updated_at | timestamp | |

### `evaluation_approvals`

| Column | Type | Notes |
|---|---|---|
| id | bigint | PK |
| evaluation_assignment_id | bigint | FK |
| approver_id | bigint | FK users |
| level | integer | |
| status | varchar | pending/approved/rejected |
| comment | text | Nullable |
| approved_at | timestamp | Nullable |
| rejected_at | timestamp | Nullable |
| created_at | timestamp | |
| updated_at | timestamp | |

---

## 7.5 Audit

### `audit_logs`

| Column | Type | Notes |
|---|---|---|
| id | bigint | PK |
| user_id | bigint | FK |
| action | varchar | |
| entity_type | varchar | |
| entity_id | bigint | |
| old_values | json | Nullable |
| new_values | json | Nullable |
| ip_address | varchar | Nullable |
| user_agent | text | Nullable |
| created_at | timestamp | |

---

# 8. Main Relationships

```text
Organization
 └── Departments
      └── Positions
           ├── Employees (self-ref: manager_id)
           └── Job Postings
                ├── Job Posting Skills → Skills
                └── Interview Rounds
                     └── Interviews → Candidates
                          └── Interview Feedback
                                (Candidates also have Candidate Skills → Skills)

Evaluation Period
 └── Evaluation Assignments

Evaluation Template
 └── Sections
      └── Questions

Evaluation Assignment
 ├── Evaluator (users)
 ├── Evaluatee (employees)
 ├── Candidate (optional)
 ├── Responses → Questions        [unique per assignment+question]
 ├── Comments
 ├── Result                       [unique per assignment]
 └── Approvals

Users ── user_roles ── Roles ── role_permissions ── Permissions
Users ── audit_logs
```
