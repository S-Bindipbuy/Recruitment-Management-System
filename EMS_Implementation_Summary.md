# EMS Implementation Summary

Developer-facing summary of what is implemented and where the code lives.
Schema details live in [`Evaluation_Management_System_Project_Documentation.md`](Evaluation_Management_System_Project_Documentation.md)
(§7) and the full endpoint reference in [`API_DOCUMENTATION.md`](API_DOCUMENTATION.md).
The authoritative schema source is `backend/database/migrations/*/000001_init.sql`.

---

## Module Status

| Module | Status | Key code |
|---|---|---|
| Auth (JWT + RBAC) | done | `app/middleware/auth.go`, `app/handlers/auth.go`, `app/handlers/pages/core/authz.go`, `database/seed.go` |
| Users / Roles / Permissions | done | `app/handlers/user/user.go`, `app/handlers/pages/user/pages_user.go`, `pages_role.go`, `pages_permission.go` |
| Organization (orgs/departments/positions) | done | `app/handlers/organization/`, `app/handlers/pages/organization/pages_organization.go` |
| Employees | done | `app/handlers/employee/employee.go`, `app/handlers/pages/organization/pages_employee.go` |
| Candidates + CV upload + hire | done | `app/handlers/candidate/candidate.go`, `app/handlers/pages/candidate/pages_candidate.go` |
| Skills (library, candidate/posting skills) | schema + repos done, **no routes wired** | `app/repositories/recruitment/skill.go` (`SkillRepository`), `app/services/recruitment/skill.go` |
| Job postings (+ wizard) | done | `app/handlers/recruitment/job_posting.go`, `app/handlers/pages/recruitment/pages_recruitment.go`, `pages_job_wizard.go` |
| Interview rounds / interviews / feedback | done | `app/handlers/recruitment/interview_round.go`, `interview.go`, `interview_feedback.go` |
| Recruitment reports | done (queries join a missing `candidates.position_id`, see below) | `app/handlers/recruitment/report.go` |
| Evaluation (periods, templates, sections, questions, assignments) | done | `app/handlers/evaluation/`, `app/handlers/pages/evaluation/pages_evaluation.go` |
| Cycle wizard (bulk assignment launcher) | done | `app/handlers/pages/cycle/`, rules in `app/services/cycles/` |
| Evaluation fill / scoring / comments | done | `app/services/evaluation/` (`assignment.go`, `response.go`, `result.go`, `comment.go`) |
| Approvals | done | `app/handlers/pages/approval/pages_approval.go` |
| Reports (evaluation) | done | `app/handlers/pages/report/pages_report.go` |
| Rating scales | schema only — **no handler, no route** | `frontend/templates/pages/rating_scales.html` is an orphaned template |
| Public job board / apply | done | `app/handlers/public/` |
| Audit logs | done | `app/handlers/audit/audit.go` |

Known gaps, preserved deliberately:

- **Skills API/pages are not registered.** `SkillRepository` and the skills
  handlers exist, but `routes/` registers no `/skills` endpoints; skills are
  managed through candidate/posting data today.
- **Rating scales lost their handlers in the package split.** Only the schema
  (`rating_scales`, `rating_scale_options`) and the unused template file
  `frontend/templates/pages/rating_scales.html` are left; no Go code and no
  route reference them.
- **Recruitment reports are broken by schema drift.** The report queries join
  `candidates.position_id`, a column that has never existed in the schema
  (`candidates` has `job_posting_id`), so `/api/reports/*` and
  `/pages/recruitment-reports` return an error. The SQL is kept as-is (aliased
  for sqlx) in `app/repositories/recruitment/interview_feedback.go` until the
  reports are redesigned.
- **`FindPendingForApprover` ignores its argument** and returns every pending
  approval — historical behavior kept intentionally.

---

## Recruitment Feature Set

1. **Job & position management** — create/edit/publish/close/reopen postings linked
   to positions/departments; salary range, employment type; multi-step posting wizard.
2. **Skills library** — centralized skills with categories; candidate proficiency and
   posting requirements (schema + repositories ready; UI/API not wired).
3. **Interview scheduling** — candidates matched to rounds; interviewers chosen from
   active workers; date/time, duration, location, meeting link; status tracking.
4. **Interview rounds** — HR/Technical/Final per posting, numbering, max/pass scores.
5. **Interview feedback** — 0–10 score, recommendation, strengths/weaknesses/comments
   (one feedback row per interview, enforced by a unique index).
6. **Reporting** — candidate evaluation summary, interviewer activity, hiring decisions
   (currently erroring — see known gaps).

---

## Where Things Live

The backend is split by **feature area**: the same area name appears in every
layer, so a feature lives in one package per layer instead of one big file per
layer.

```text
backend/
├── main.go                      # boot: connect → goose migrations → seeds → Fiber
├── config/config.go             # env config (SERVER_PORT, DB_PATH, JWT_SECRET, …)
├── database/
│   ├── connection.go            # *sqlx.DB, DSN per dialect (WAL, busy timeout)
│   ├── migrations.go            # goose: embed + dialect + Up
│   ├── migrations/
│   │   ├── sqlite/000001_init.sql
│   │   └── postgres/000001_init.sql
│   ├── seed.go                  # SeedDefaultRBAC, SeedAdminUser (dev admin)
│   └── seed.sql                 # portable manual seed
├── routes/                      # /api/* registration, one file per area
│   ├── routes.go                # SetupRoutes: api group + pages group
│   ├── auth.go  user.go  employee.go  candidate.go  organization.go
│   ├── dashboard.go  evaluation.go  audit.go  recruitment.go
│   ├── public.go                # /jobs, /apply (no auth)
│   ├── component.go             # /api/components/:name
│   └── pages/                   # every /pages/* route, one file per area
│       ├── routes.go            # Register(): the whole server-rendered surface
│       ├── dashboard.go  organization.go  employees.go  candidates.go
│       ├── evaluations.go  evaluation_fill.go  cycle.go  approvals.go
│       ├── reports.go  users.go  roles.go  permissions.go  recruitment.go
├── app/
│   ├── handlers/                # REST handlers, one package per area
│   │   ├── auth.go  component.go
│   │   ├── audit/  candidate/  dashboard/  employee/  evaluation/
│   │   ├── organization/  recruitment/  user/  public/
│   │   └── pages/               # Page handlers for SPA fragments, one package per area
│   │       ├── core/            # shared plumbing: templates, forms, authz
│   │       ├── approval/  candidate/  cycle/  dashboard/  evaluation/
│   │       └── organization/  recruitment/  report/  user/
│   ├── services/                # business logic, one package per area
│   │   ├── util.go
│   │   ├── candidate/  cycles/  evaluation/  organization/  recruitment/  user/
│   ├── repositories/            # sqlx data access, one package per area
│   │   ├── repo.go              # shared helpers (InClause, ActiveRecordExists)
│   │   ├── audit/  candidate/  evaluation/  organization/  recruitment/
│   │   ├── stats/  user/
│   ├── models/                  # structs with db tags, one package per area
│   │   ├── audit.go  stats.go
│   │   ├── candidate/  evaluation/  organization/  recruitment/  user/
│   ├── middleware/auth.go       # JWT verification
│   ├── enums/                   # status/permission constants
│   └── httperr/error.go         # JSON error handler

frontend/                        # static assets + HTML templates
├── static/                      # css, js, components (sidebar, header, auth)
└── templates/
    ├── public_layout.html       # layout of the public pages
    ├── content_*.html           # public pages (jobs, job detail, apply)
    └── pages/*.html             # one template per page, parsed on its own

flake.nix                        # nix build (vet + tests) / nix develop
```

Each `services/<area>` and `repositories/<area>` package keeps its singletons
in an `instances.go` file (`PeriodServiceInstance`, `PeriodRepo`, …); handlers
use those instances and nothing is built per request.

Route registration mirrors the handler packages: `routes/*.go` for `/api/*`,
`routes/pages/*.go` for `/pages/*`. `routes/pages/routes.go` is the single
place where every page area is registered.

---

## Data Access (sqlx)

- `database.DB` is a `*sqlx.DB`; repositories use `Get`/`Select` (scan by column
  name) and `Exec`/`Beginx` for writes.
- Models carry `db:"…"` tags; SQL expressions are aliased to the destination name.
- Nullable columns map to pointer fields (`*int64`, `*time.Time`, `*float64`,
  `*bool`); plain `string`/`float64` destinations are `COALESCE`d in SQL only
  where NULLs can occur.

## Migrations (goose)

- Versioned SQL files per dialect under `backend/database/migrations/`;
  applied automatically at boot (`goose.Up`), versions tracked in
  `goose_db_version`.
- Schema is fully `IF NOT EXISTS`, so pre-goose databases are adopted in place;
  the init migration also deduplicates `evaluation_responses` (enabling the
  unique index behind `Upsert`) and clears stale zero-score results.
- Add a migration: new numbered file in **both** dialect directories — never
  edit an applied one. See the project documentation §4 for the SQLite →
  PostgreSQL translation rules.

---

## Verification

```bash
go vet ./... && go test ./...   # vet clean; go test compiles only - no test files
nix build                       # builds backend with vet + tests in the sandbox
```

The test suites that used to guard this code (migration/seed, evaluation
scoring, template render check, golden pages, route table) were removed during
the package split, so **the repository currently has no `_test.go` files**.
`go vet ./...` and `nix build` are the checks the project is verified with.
