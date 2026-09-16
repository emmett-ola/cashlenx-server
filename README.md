# CashLenX Server

CashLenX Server is a Go backend for personal finance tracking. It provides a Cobra CLI and a Gorilla Mux REST API for authentication, user accounts, cash flows, categories, monthly budgets, statistics, import/export, and admin database management.

The product remains pre-release while the accepted stable API contract is implemented under canonical `/api/v1`. The frozen `/api/v0` alias protects previously shipped clients during the transition.

## CashLenX Project

CashLenX is developed as a set of independently buildable repositories with
explicit ownership boundaries:

| Repository | Responsibility |
| --- | --- |
| [cashlenx-app](https://github.com/emmett-ola/cashlenx-app) | Cross-platform Flutter client and user experience. |
| [cashlenx-server](https://github.com/emmett-ola/cashlenx-server) | Go REST API, Cobra CLI, authentication, finance services, and MongoDB/MySQL persistence. |
| [cashlenx-design](https://github.com/emmett-ola/cashlenx-design) | Figma-exported React/Vite visual and interaction reference. |
| [cashlenx-website](https://github.com/emmett-ola/cashlenx-website) | Public product and developer-information website. |
| [cashlenx-spec](https://github.com/emmett-ola/cashlenx-spec) | Product and system facts, delivery workflow, decisions, and retained evidence. |

This repository owns the backend, command-line interface, public API contract,
and persistence behavior. Cross-repository contracts are coordinated through
OpenAPI and the CashLenX Spec workflow. Runtime repositories remain
independently buildable and do not depend on the spec or design reference at
build time or runtime.

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

Local and Docker database URIs in `.env.example` reuse the username, password,
port, and database name defined above them. Direct local routes use the
published host port; Docker routes use the configured dependency container name
and internal port on the shared network. Docker Compose and the Server dotenv
loader expand those references, so each atomic value has one configuration
owner.

Every assignment in `.env.example` is active. Optional capabilities use explicit
lowercase boolean switches such as `SMTP_ENABLED=false`; database dependency
selection remains the explicit script invocation rather than an enable flag.
`TIMEZONE` is the single application and container timezone source. Use `UTC`
or a region-based IANA name such as `Asia/Shanghai`. Every start entry point
rejects fixed offsets, ambiguous abbreviations, and POSIX-sign `Etc/GMT` forms;
API startup additionally rejects names absent from the Go timezone database.

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
waits on an in-container readiness probe, and `stop.sh` removes its container
while retaining the image and named volume. Every start attaches to the
explicitly configured `DOCKER_NETWORK_NAME`; a stop removes that network only when no CashLenX
container remains attached. Root Server
scripts do not start, stop, or remove dependencies, regardless of `DB_TYPE`.

The named volumes remain the default. Set `MONGO_DATA_PATH` or
`MYSQL_DATA_PATH` to an absolute host path to use a bind mount instead. Relative
paths, filesystem roots, and parent traversal are rejected by dependency start;
stop never removes either a bind path or a named volume. Configure
`MONGO_DATA_VOLUME_NAME` or `MYSQL_DATA_VOLUME_NAME` when a different Docker
named-volume identity is preferred and leave the matching data path empty.

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
container from that existing image and waits on the API health endpoint from
inside the container. It does not require Compose `up --wait` or
Compose-managed health status, which keeps the lifecycle compatible with
nerdctl Compose.
`stop.sh` removes the API container while preserving the image, bind-mounted
logs, database projects, and database volumes. It removes the shared external
network only when the network has no connected containers.

All five Server-owned entry-point groups support Docker Compose v2 and nerdctl
2.2 or newer. `CONTAINER_FRONTEND` accepts `auto`, `docker`, or `nerdctl`; auto
mode detects the implementation reported by the selected command, including a
command named `docker` that wraps nerdctl. Set `CONTAINER_CLI` in the invoking
shell for a nonstandard executable path. Runtime availability and Compose
configuration are validated before build, pull, network creation, start, or
stop. Server image identity is derived directly from validated
`SERVER_IMAGE_NAME` and `SERVER_IMAGE_TAG` values, not from
`compose config --images`. Readiness uses portable engine inspection and exec
commands with a default 600-second limit that can be overridden through
`CONTAINER_READY_TIMEOUT_SECONDS`. Start commands suppress frontend command
traces so nerdctl cannot print configured credentials in informational output.
Runtime start also uses `--pull never`; candidate deployment must preload and
verify the exact image identity before replacing the API container.

Use another repository-local configuration consistently with
`ENV_FILE=.env.testing scripts/build.sh`, `scripts/start.sh`, and
`scripts/stop.sh`. Missing files and paths outside this repository are rejected.
Alternatively, `.env` may be a symbolic link to `.env.local`, `.env.testing`, or
`.env.production`; the fully resolved target must remain a regular file inside
this repository. Builds may use placeholder values. Startup validates only the
selected database and enabled capabilities, reports only unsafe or missing key
names, and stops until relevant values are replaced. Disabled SMTP and unselected database
placeholders do not block API startup. Stop remains available with an existing
file even when its values are incomplete.

The server and database ports bind to `127.0.0.1` by default. `.env.example`
explicitly sets CPU, memory, PID, graceful-stop, health-check, image, log path,
and timezone values. `TIMEZONE` is passed to every Server and database container
as its standard `TZ` environment variable. Compose only passes the string
through; lifecycle validation and the Go runtime enforce the shared timezone
contract. The server image records the source revision in
the OCI `org.opencontainers.image.revision` label.
The image build does not install Alpine packages: Go embeds the IANA timezone
database, and the Compose healthcheck uses the Alpine base image's BusyBox
`wget` applet. A temporary Alpine package-index outage therefore does not block
the Server image build.

The Server Dockerfile and main Compose definitions live at `docker/Dockerfile`
and `docker/compose.yml`. `SERVER_PROJECT_NAME`, `MONGO_PROJECT_NAME`, and
`MYSQL_PROJECT_NAME` explicitly name the three independent Compose projects;
their matching container-name keys remain independent. All attach to the
absolute `DOCKER_NETWORK_NAME`, which any start script creates idempotently. The
API reaches project-managed databases through their configured container DNS
name and internal port, without a host-gateway route.

Dependency scripts support the same `ENV_FILE` selection. For example:

```bash
ENV_FILE=.env.testing scripts/dependencies/mongodb/build.sh
ENV_FILE=.env.testing scripts/dependencies/mongodb/start.sh
ENV_FILE=.env.testing scripts/dependencies/mongodb/stop.sh
```

`test/scripts/dependency-lifecycle-smoke.sh` validates Docker and nerdctl 2.2
command shapes, a `docker` wrapper around nerdctl, environment-file symlinks,
and fail-before-mutation behavior without changing real containers.

The default container name is `cashlenx-server`.

### 3. Run the API Server

```bash
go run main.go open start -p 10063
```

The local base URL is:

```text
http://127.0.0.1:10063/api/v1
```

### 4. Useful CLI Commands

```bash
go run main.go open health
go run main.go open version
go run main.go admin database backup -o backup.json
go run main.go admin database restore -i backup.json
```

## REST API Highlights

- `GET /api/v1/open/health`
- `GET /api/v1/open/version`
- `POST /api/v1/open/auth/register`
- `POST /api/v1/open/auth/login`
- `POST /api/v1/open/auth/logout`
- `GET /api/v1/auth/tokens`
- `GET /api/v1/user/profile`
- `POST /api/v1/cash/expense`
- `POST /api/v1/cash/income`
- `GET /api/v1/cash`
- `GET /api/v1/category/tree`
- `GET /api/v1/budget?period=YYYY-MM`
- `GET /api/v1/statistic/dashboard/{period}/{date}`

`/api/v0` remains a frozen compatibility alias for previously shipped clients. New integrations must use `/api/v1`. Login keeps the `username` JSON field and accepts either a username or email address.
- `GET /api/v1/statistic/chart/income-expense/{period}/{date}`

See `docs/openapi.yaml` for the current API contract and `docs/api.md` for additional API notes.

## Operational Endpoints

- `GET /metrics` exposes Prometheus request counters, duration histograms, and Go process/runtime metrics.
- `/debug/pprof/*` exposes Go profiling handlers only when `ENV=dev`.

These endpoints are intentionally outside the versioned API and the OpenAPI/JWT middleware. Restrict `/metrics` to trusted monitoring networks at the reverse proxy or firewall in deployed environments.

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
- `CONTRIBUTING.md` - public contribution workflow and pull-request evidence
- `SECURITY.md` - private vulnerability-reporting and response policy
- [Shared Governance](https://github.com/emmett-ola/cashlenx-spec/blob/main/GOVERNANCE.md)
- [Shared Delivery Workflow](https://github.com/emmett-ola/cashlenx-spec/blob/main/WORKFLOW.md)

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

This project is licensed under the [MIT License](LICENSE). Commercial use,
modification, and redistribution are permitted when the copyright and license
notices are retained.
