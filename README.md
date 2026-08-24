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
├── docker/                  # Database initialization assets
├── docs/                    # API, CLI, OpenAPI, roadmap docs
├── errors/                  # Custom error types
├── mapper/                  # MongoDB/MySQL persistence mappers
├── middleware/              # Auth, admin, CORS, logging, schema validation
├── migrations/              # MongoDB/MySQL migration scripts
├── model/                   # Entities, DTOs, response types, constants
├── scripts/                 # Container build/start/stop entry points
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

MongoDB bootstrap, migration, and runtime index definitions now follow the current user/type/parent/name category scope. MongoDB applied-version tracking is not implemented and remains unscheduled architecture debt; see `docker/mongodb/README.md` and `AGENTS.md`.

### 2. Start a Database

```bash
# MongoDB
docker compose --env-file .env -f docker/dependencies/compose.mongodb.yml up -d --wait

# MySQL
docker compose --env-file .env -f docker/dependencies/compose.mysql.yml up -d --wait
```

The dependencies are separate operator-managed projects and persist data in
`cashlenx-mongodb-data` and `cashlenx-mysql-data`. Server build/start scripts do
not start, stop, or remove them.

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
Builds may use placeholder values, but startup reports only the unsafe or
missing key names and stops until they are replaced. Stop remains available
with an existing file even when its values are incomplete.

The server and database ports bind to `127.0.0.1` by default. `.env.example`
defines configurable CPU, memory, PID, graceful-stop, health-check, image, log
path, and timezone settings. The server image records the source revision in
the OCI `org.opencontainers.image.revision` label.

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
- `docker/mongodb/README.md` - MongoDB bootstrap status and known index drift
- `docker/mysql/README.md` - MySQL bootstrap and migration validation notes

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
