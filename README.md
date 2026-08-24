# CashLenX Server

CashLenX Server is a Go backend for personal finance tracking. It provides a Cobra CLI and a Gorilla Mux REST API for authentication, user accounts, cash flows, categories, monthly budgets, statistics, import/export, and admin database management.

The project is still in active `v0.x` development. The current API path is `/api/v0`; stable `/api/v1` compatibility is planned for a later stable release.

## Features

- Local registration/login with JWT access tokens and persisted refresh tokens
- User profile, password change, password reset, email-change request/confirm, and account deletion flows
- Per-user cash flow CRUD, date/range queries, summaries, pagination, and filtering
- Per-user category CRUD, name lookup, children lookup, and tree APIs
- Per-user monthly budget CRUD with ledger-derived spending for MongoDB and MySQL
- Statistics, dashboard, and chart endpoints for summaries, breakdowns, trends, top expenses, income/expense charts, category distribution, monthly comparison, and spending heatmaps
- User import/export and backup/restore flows
- Admin user management and full database backup/restore
- MongoDB and MySQL persistence implementations
- Independent Docker Compose projects for the server, MongoDB, and MySQL
- OpenAPI contract in `docs/openapi.yaml`
- Prometheus request metrics and development-only Go profiling endpoints
- Reproducible service and mapper benchmarks with a measured `v0.9.0` cache decision

## Project Structure

```text
cashlenx-server/
├── auth/                    # Auth service/provider abstraction
├── cmd/                     # CLI commands (Cobra)
├── config/                  # Runtime data files
├── controller/              # HTTP route registration and handlers
├── docker/                  # API Compose plus isolated dependency projects
├── docs/                    # API, CLI, OpenAPI, roadmap docs
├── errors/                  # Custom error types
├── mapper/                  # MongoDB/MySQL persistence mappers
├── middleware/              # Auth, admin, CORS, logging, schema validation
├── migrations/              # MongoDB/MySQL migration scripts
├── model/                   # Entities, DTOs, response types, constants
├── scripts/                 # API and dependency lifecycle entry points
├── test/scripts/            # Disposable integration smoke checks
├── service/                 # Business logic
├── util/                    # Config, logging, DB, email, date, ID, HTTP helpers
├── validation/              # Validation helpers and tests
└── main.go                  # Entry point
```

## Quick Start

### 1. Configure

```bash
cp .env.example .env
```

MongoDB remains the default and supported beta deployment database. MySQL 8 is also runnable: the full Flutter/API contract and the numbered SQL sequence are verified against disposable containers. Server startup tracks and applies MySQL migrations through `schema_migrations`.

Database URIs in `.env.example` reuse the username, password, and database name
defined above them. Docker Compose and the Server dotenv loader expand those
references, so each credential has one configuration owner.

Every assignment in `.env.example` is active. Optional capabilities use explicit
lowercase boolean switches such as `SMTP_ENABLED=false`; database dependency
selection remains the explicit script invocation rather than an enable flag.
`TIMEZONE` is the single application and container timezone source.

MongoDB bootstrap, migration, and runtime index definitions now follow the current user/type/parent/name category scope. MongoDB applied-version tracking is not implemented and remains unscheduled architecture debt; see `docker/dependencies/mongodb/README.md` and `AGENTS.md`.

### 2. Start a Database

```bash
# MongoDB
scripts/dependencies/mongodb/build.sh
scripts/dependencies/mongodb/start.sh
# Stop it later when required:
scripts/dependencies/mongodb/stop.sh

# MySQL
scripts/dependencies/mysql/build.sh
scripts/dependencies/mysql/start.sh
# Stop it later when required:
scripts/dependencies/mysql/stop.sh
```

The dependencies are separate operator-managed projects and persist data in
`cashlenx-mongodb-data` and `cashlenx-mysql-data`. Each dependency `build.sh`
pulls its configured upstream image, `start.sh` starts only that database and
waits for health, and `stop.sh` removes its container and network while retaining
the image and named volume. Root Server scripts do not start, stop, or remove
dependencies, regardless of `DB_TYPE`.

### Docker Deployment

Build the server image, start only the backend, and verify the API health
endpoint. Start the selected dependency separately first.

```bash
scripts/build.sh
scripts/start.sh
scripts/stop.sh
```

The scripts require an existing reviewed `.env`; they never create one from
development defaults and never manage dependency containers. `build.sh`
compiles the server and builds its image. `start.sh` starts or updates the API
container from that existing image and waits for the Compose healthcheck.
`stop.sh` removes the API container and its project network while preserving
the image, bind-mounted logs, database projects, and database volumes.

Use another repository-local configuration consistently with
`ENV_FILE=.env.testing scripts/build.sh`, `scripts/start.sh`, and
`scripts/stop.sh`. Missing files and paths outside this repository are rejected.
Builds may use placeholder values. Startup validates only the selected database
and enabled capabilities, reports only unsafe or missing key names, and stops
until relevant values are replaced. Disabled SMTP and unselected database
placeholders do not block API startup. Stop remains available with an existing
file even when its values are incomplete.

The server and database ports bind to `127.0.0.1` by default. `.env.example`
explicitly sets CPU, memory, PID, graceful-stop, health-check, image, log path,
and timezone values. `TIMEZONE` is passed to every Server and database container
as its standard `TZ` environment variable. The server image records the source revision in
the OCI `org.opencontainers.image.revision` label.

Dependency scripts support the same `ENV_FILE` selection. For example:

```bash
ENV_FILE=.env.testing scripts/dependencies/mongodb/build.sh
ENV_FILE=.env.testing scripts/dependencies/mongodb/start.sh
ENV_FILE=.env.testing scripts/dependencies/mongodb/stop.sh
```

The default container name is `cashlenx-server`.

### 3. Run the API Server

```bash
go run main.go open start -p 11063
```

The local base URL is:

```text
http://localhost:11063/api/v0
```

### 4. Useful CLI Commands

```bash
go run main.go open health
go run main.go open version
go run main.go admin database backup -o backup.json
go run main.go admin database restore -i backup.json
```

## REST API Highlights

- `GET /api/v0/open/health`
- `GET /api/v0/open/version`
- `POST /api/v0/open/auth/register`
- `POST /api/v0/open/auth/login`
- `POST /api/v0/open/auth/logout`
- `GET /api/v0/auth/tokens`
- `GET /api/v0/user/profile`
- `POST /api/v0/cash/expense`
- `POST /api/v0/cash/income`
- `GET /api/v0/cash`
- `GET /api/v0/category/tree`
- `GET /api/v0/budget?period=YYYY-MM`
- `GET /api/v0/statistic/dashboard/{period}/{date}`
- `GET /api/v0/statistic/chart/income-expense/{period}/{date}`

See `docs/openapi.yaml` for the current API contract and `docs/api.md` for additional API notes.

## Operational Endpoints

- `GET /metrics` exposes Prometheus request counters, duration histograms, and Go process/runtime metrics.
- `/debug/pprof/*` exposes Go profiling handlers only when `ENV=dev`.

These endpoints are intentionally outside `/api/v0` and the OpenAPI/JWT middleware. Restrict `/metrics` to trusted monitoring networks at the reverse proxy or firewall in deployed environments.

## Documentation

- `docs/README.md` - documentation map and update rules
- `AGENTS.md` - shared working guide for coding agents
- `docs/cli.md` - CLI command reference
- `docs/api.md` - REST API notes
- `docs/roadmap.md` - active/future milestone planning
- `docs/milestones.md` - completed milestone history
- `docs/performance.md` - benchmark baseline and cache decisions
- `docs/openapi.yaml` - OpenAPI specification
- `docker/dependencies/README.md` - dependency lifecycle and ownership boundary
- `docker/dependencies/mongodb/README.md` - MongoDB bootstrap status and index lifecycle
- `docker/dependencies/mysql/README.md` - MySQL bootstrap and migration validation notes

## Build and Test

```bash
go build -o cashlenx main.go
go test ./...
go test -v -race -covermode=atomic -coverprofile=coverage.out ./...
```

On Windows, validate the numbered MySQL migrations independently:

```powershell
powershell -ExecutionPolicy Bypass -File test/scripts/mysql-migrations-smoke.ps1

# Focused user-scoped budget parity against disposable databases
powershell -ExecutionPolicy Bypass -File test/scripts/budget-smoke.ps1 -Database mongodb
powershell -ExecutionPolicy Bypass -File test/scripts/budget-smoke.ps1 -Database mysql
```

Run the managed MongoDB API smoke flow with
`test/scripts/api-smoke.sh --managed`. The sibling Flutter client does not
currently ship a maintained live API harness.

Test coverage is still uneven while the project is under development. GitHub
Actions runs the full Go test suite with race detection and `coverage.out`
generation for Codecov. DeepSource handles code analysis.

## Technology

- Go `1.23.0`
- Cobra CLI
- Gorilla Mux HTTP routing
- Zap logging
- MongoDB and MySQL drivers
- JWT via `github.com/golang-jwt/jwt/v5`
- `shopspring/decimal` for money values
- `excelize` and `gofpdf` for exports
- OpenAPI validation through `kin-openapi`
- Prometheus instrumentation through `prometheus/client_golang`

## License

See `LICENSE` for details.
