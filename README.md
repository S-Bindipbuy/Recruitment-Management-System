# EMS — Evaluation Management System

A full-stack system for managing employees, recruitment, and performance
evaluations. It covers the whole evaluation lifecycle — creating templates,
running periods and cycles, assigning evaluations, collecting responses,
approvals, and reporting — plus the recruitment workflow around it (job
postings, candidates, interviews, hiring).

The backend serves both a JSON REST API (`/api/*`) and server-rendered HTML
fragments (`/pages/*`) used by the vanilla JavaScript SPA. There is a public job board at `/jobs`.

## Features

- Authentication with JWT and role/permission-based access control
- Organizations, departments, positions, and employees
- Job postings, candidates, CV uploads, interview rounds, and feedback
- Evaluation periods, templates, sections, questions, and rating scales
- Evaluation assignments, responses, results, comments, and approvals
- Dashboards, reports, recruitment reports, and audit logs
- Public job listings and application form

## Tech stack

- **Backend:** Go 1.26 + [Fiber v2](https://gofiber.io/)
- **Database:** [Turso/libSQL](https://turso.tech/) via `go-libsql` — works with a
  local `file:` database, a remote Turso database, or a local embedded replica
- **Data access:** [`sqlx`](https://github.com/jmoiron/sqlx) over `database/sql`
- **Migrations:** [goose](https://github.com/pressly/goose) (SQL files, applied on boot)
- **Auth:** JWT (`golang-jwt/jwt/v5`), manual RBAC
- **Frontend:** Server-rendered HTML templates + vanilla JavaScript SPA
  (custom router/navigation), hand-written CSS
  (`frontend/static/css/styles.css`, Rosé Pine Dawn palette)
- **Icons:** Font Awesome (CDN)
- **Build/packaging:** Nix flake (`nix build`, `nix develop`)

No Node.js, Tailwind, or shadcn/ui build step is used.

## Project structure

```text
backend/
  main.go                     # boot: config → DB → migrations/seeds → Fiber
  config/                     # environment configuration
  database/                   # connection, migrations, seeds
  routes/                     # /api/* routes
  routes/pages/               # /pages/* routes
  app/handlers/               # REST handlers (one package per area)
  app/handlers/pages/         # Page handlers for SPA fragments
  app/services/               # business logic
  app/repositories/           # sqlx queries
  app/models/                 # db-tagged structs
  app/middleware/             # JWT auth
  app/enums/                  # status/permission constants
frontend/
  templates/                  # page templates and public layouts
  components/                 # shared HTML fragments
  static/                     # CSS, JavaScript, and icons
```

Each feature area (candidate, evaluation, organization, recruitment, user, …)
uses the same package name across handlers, services, repositories, and models.

## Getting started

### Prerequisites

- Go 1.26+, or Nix with flakes enabled
- CGO enabled (required by the libSQL driver); the Nix dev shell/build handles this
- No external database is required for the default local `file:` setup

### 1. Configure the environment

```bash
cp .env-example .env
```

Set a JWT secret in `.env`:

```bash
openssl rand -hex 32
```

These variables are **required**; the server refuses to start if any is missing
or empty:

| Variable | Description |
|---|---|
| `TURSO_DATABASE_URL` | libSQL/Turso URL: `file:./storage/db`, `libsql://<db>-<org>.turso.io`, or `http://127.0.0.1:8080` for `turso dev` |
| `JWT_SECRET` | Secret used to sign JWT tokens |
| `FRONTEND_PATH` | Frontend root, e.g. `./frontend` |

Optional variables (defaults in parentheses): `TURSO_AUTH_TOKEN` (empty),
`TURSO_REPLICA_PATH` (empty), `TURSO_SYNC_INTERVAL` (`0`), `SERVER_PORT`
(`3000`), `JWT_EXPIRATION_HOURS` (`24`), `ALLOWED_ORIGINS` (`*`), `UPLOAD_DIR`
(`storage/uploads/cvs`), `REDIS_ADDR` (`localhost:6379`), `REDIS_PASSWORD`
(empty), `REDIS_DB` (`0`), `APP_ENV` (development).

See [`.env-example`](.env-example) for the full list.

### 2. Run the server

With Nix:

```bash
nix develop
ems-compile        # builds backend and moves the binary to ./main
./main
```

Or directly with Go (run from the project root so `.env` and relative paths resolve):

```bash
go build -C backend -o ../main .
./main
```

The server listens on `http://localhost:3000` by default:

- App: <http://localhost:3000>
- Public job board: <http://localhost:3000/jobs>

On boot it creates/updates the libSQL/Turso schema, seeds RBAC data, and
(outside production) creates a development admin account:

```text
email:    admin@example.com
password: admin123
```

Change the password after the first login. Set `APP_ENV=production` to disable
the development admin seeder.

## Build with Nix

```bash
nix build
```

The package contains:

- `result/bin/backend` — server binary
- `result/share/frontend` — frontend assets

Run it with the required configuration, for example:

```bash
FRONTEND_PATH="$PWD/result/share/frontend" \
JWT_SECRET="$(openssl rand -hex 32)" \
TURSO_DATABASE_URL="file:$PWD/storage/db" \
./result/bin/backend
```

## Database (Turso / libSQL)

`TURSO_DATABASE_URL` selects the connection mode:

```bash
# Local libSQL file (no account or network required)
TURSO_DATABASE_URL=file:./storage/db

# Remote Turso database
TURSO_DATABASE_URL=libsql://my-db-my-org.turso.io
TURSO_AUTH_TOKEN="$(turso db tokens create my-db)"

# Local turso dev server
TURSO_DATABASE_URL=http://127.0.0.1:8080
```

To keep a local replica that syncs from a remote primary:

```bash
TURSO_DATABASE_URL=libsql://my-db-my-org.turso.io
TURSO_AUTH_TOKEN=...
TURSO_REPLICA_PATH=./storage/replica.db
TURSO_SYNC_INTERVAL=60   # seconds; 0 = initial sync only
```

Migrations run automatically against whichever database is configured. When
dependencies are re-vendored, run `backend/scripts/vendor-libsql.sh` after
`go mod vendor` — `go mod vendor` omits the driver's native library directory.

## Tests and checks

```bash
go vet -C backend ./...
go build -C backend ./...
```

## Documentation

- [Evaluation Management System Project Documentation](Evaluation_Management_System_Project_Documentation.md) — architecture, configuration, and database schema
- [API Documentation](API_DOCUMENTATION.md) — REST and page routes
- [How to Use](HOW_TO_USE.txt) — end-user guide
- [Implementation Summary](EMS_Implementation_Summary.md) — current implementation status

## License

[MIT](LICENSE)
