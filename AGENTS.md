# AGENTS.md

This file is the working guide for coding agents in this repository. It is based on the current codebase, not just the earlier project summary, and should be kept aligned with the implementation as the project evolves.

## Project Snapshot

CashLenX Server is a Go backend that exposes:

- A Cobra-based CLI via `main.go -> cmd.Execute()`
- A Gorilla Mux REST API started by `open start`
- Multi-user finance features with JWT auth and admin/user route separation
- Dual persistence support through mapper implementations for MongoDB and MySQL
- Reporting/export flows for JSON backup, CSV/XLSX/PDF export, and dashboard/statistics endpoints
- Verification flows for password reset and email change

## Current Tech Stack

- Go `1.23.0` in `go.mod`
- Cobra for CLI
- Gorilla Mux for HTTP routing
- Zap-based logging utilities in `util/log_util.go`
- MongoDB driver and MySQL driver
- JWT via `github.com/golang-jwt/jwt/v5`
- `shopspring/decimal` for money values
- `excelize` and `gofpdf` for exports
- OpenAPI validation through `kin-openapi`

## Team Workflow Defaults

These are collaboration defaults provided by the project owner and should be treated as working norms unless a task says otherwise.

- Default development database is MongoDB
- Users may still choose MongoDB or MySQL when bootstrapping their environment
- This is an under-development project, so use practical baseline validation effort rather than assuming strict release-grade test gates
- We are currently developing on the `dev/v0.4.0` branch line
- The branch line maps to the feature/version batch currently in progress
- Once planned work for `v0.4.0` is complete, it is intended to be merged/promoted to `main`, then development moves to the next branch line such as `dev/v0.5.0`
- API versioning stays under `/api/v0` during active development
- A stable release is expected to introduce `/api/v1` alongside a stable application version such as `v1.0.0`
- Unless the user explicitly says not to, make a commit after each completed request/change set

## Working Agreement

- Keep a running record of known mismatches, implementation drift, and TODO-grade architectural inconsistencies in this guide or nearby project docs when useful
- Treat those notes as collaboration context, not just criticism; they help future sessions confirm and fix issues deliberately
- Development conventions can be enriched incrementally as collaboration continues

## Collaboration Notes

- Treat this file as shared working memory for future agents
- Prefer adding concise, durable guidance over session-specific narration
- Good additions include architectural decisions, workflow conventions, verified repo realities, and durable TODO/drift notes
- Avoid adding noisy progress chatter that will age poorly
- When a repeated pitfall shows up more than once, add a short note here so later sessions do not rediscover it

## Style And Progress Notes

- Favor pragmatic progress over premature process
- Preserve backward compatibility where practical because some docs and helpers still reflect older project phases
- When docs and code disagree, trust the code first, then record the drift and fix it deliberately
- Keep cross-backend behavior in mind even though MongoDB is the default dev path
- If a task changes project conventions, update this file in the same change set when appropriate

## Safety Rules For This File

- Never place secrets, credentials, tokens, API keys, passwords, private URLs, private emails, or internal-only operational data in `AGENTS.md`
- Keep examples generic and sanitize any environment variable examples if they could be mistaken for real credentials
- `AGENTS.md` is intended to be safe to commit unless the user explicitly asks for local-only notes
- If sensitive operational notes are ever needed, store them outside git-tracked docs and do not copy them into this file

## Entry Points

- `main.go` calls `cmd.Execute()`
- `cmd/root.go` wires the CLI and initializes the MongoDB pool in `PersistentPreRun`
- `cmd/open_cmd/start.go` starts the HTTP server by calling `controller.StartServer(port)`
- `controller/server.go` registers versioned API routes under `/api/{version}`

## What Is Actually Implemented

The repository is farther along than some older docs imply. These features are present in code:

- User registration and login
- JWT access tokens plus persisted refresh tokens
- Admin bootstrap user initialization on server startup
- User profile update, password change, email change request/confirm, and account deletion
- Password reset request/confirm using verification codes
- Cash flow CRUD, date/range queries, summaries, pagination/filtering
- Category CRUD plus tree/children/name lookup
- Statistics summary, breakdown, trends, top expenses, dashboard, and chart endpoints
- Admin database backup/restore
- User-scoped export/import backup flows
- SMTP email utility for verification-related email delivery
- Snowflake ID generator initialization for distributed IDs

## Repository Layout

Top-level directories currently in active use:

```text
cashlenx-server/
├── auth/                    # Auth provider abstraction and service wrapper
├── cache/                   # Cache helpers
├── cmd/                     # Cobra commands
├── config/                  # Runtime data files (for example default categories)
├── controller/              # HTTP handlers and route registration
├── docker/                  # DB initialization assets
├── docs/                    # API, CLI, roadmap, OpenAPI spec
├── errors/                  # Custom error types
├── mapper/                  # DB-specific persistence layer
├── middleware/              # Auth, admin, CORS, logging, schema validation
├── migrations/              # MongoDB/MySQL migration scripts
├── model/                   # Entities, DTOs, response types, constants
├── scripts/                 # Start and docs helper scripts
├── service/                 # Business logic
├── test/                    # Sparse test/log area
├── util/                    # Config, logging, DB helpers, email, dates, IDs, HTTP helpers
└── validation/              # Validation helpers and tests
```

## CLI Structure

The CLI structure in code is:

- `open`
  - `start`
  - `health`
  - `version`
- `admin`
  - `database`
    - `backup`
    - `restore`
- `cash`
  - `expense`, `income`, `list`, `query`, `range`, `summary`, `update`, `delete`
- `category`
  - `create`, `list`, `query`, `tree`, `update`, `delete`
- `statistic`
  - `summary`, `breakdown`, `trends`, `top`, `export`, `import`

Important: the server start command is currently `go run main.go open start -p 8080`, not `server start`.

## HTTP Routing

`controller/server.go` builds routes under `apiPrefix := "/api/" + apiVersion`, defaulting to `/api/v0`.

### Public routes

- `GET /api/{version}/open/health`
- `GET /api/{version}/open/version`
- `POST /api/{version}/open/auth/login`
- `POST /api/{version}/open/auth/register`
- `POST /api/{version}/open/auth/logout`
- `GET /api/{version}/auth/tokens`
- `POST /api/{version}/open/auth/reset-password`
- `POST /api/{version}/open/auth/reset-password/confirm`

Note: `logout` lives under `/open/auth/logout` even though it is intended to be protected. Treat route grouping carefully when changing auth behavior.

### Admin routes

Mounted on a subrouter with `middleware.Admin`:

- `POST /api/{version}/admin/user`
- `GET /api/{version}/admin/user`
- `GET /api/{version}/admin/user/{id}`
- `PUT /api/{version}/admin/user/{id}`
- `DELETE /api/{version}/admin/user/{id}`
- `GET /api/{version}/admin/database/backup`
- `POST /api/{version}/admin/database/restore`

### Authenticated user routes

- Profile/account:
  - `GET /api/{version}/user/profile`
  - `PUT /api/{version}/user/profile`
  - `PUT /api/{version}/user/password`
  - `POST /api/{version}/user/email/change`
  - `POST /api/{version}/user/email/confirm`
  - `DELETE /api/{version}/user/account`
- User backup/restore:
  - `GET /api/{version}/user/database/backup`
  - `POST /api/{version}/user/database/restore`
- Cash:
  - `POST /api/{version}/cash/expense`
  - `POST /api/{version}/cash/income`
  - `GET /api/{version}/cash`
  - `GET /api/{version}/cash/range`
  - `GET /api/{version}/cash/date/{date}`
  - `DELETE /api/{version}/cash/date/{date}`
  - `GET /api/{version}/cash/{id}`
  - `PUT /api/{version}/cash/{id}`
  - `DELETE /api/{version}/cash/{id}`
  - Summary endpoints for daily/monthly/yearly
- Category:
  - `POST /api/{version}/category`
  - `GET /api/{version}/category`
  - `GET /api/{version}/category/name/{name}`
  - `GET /api/{version}/category/{parent_id}/children`
  - `GET /api/{version}/category/tree`
  - `GET /api/{version}/category/{id}`
  - `PUT /api/{version}/category/{id}`
  - `DELETE /api/{version}/category/{id}`
- Statistic:
  - export/import
  - summary/breakdown/trends/top
  - dashboard and chart endpoints

## Architecture Notes

### Layers

The core request path is still:

```text
HTTP -> Controller -> Service -> Mapper -> Database
```

But there are a few important supporting systems that deserve equal attention:

- `middleware/` enforces request-level auth/admin/schema behavior
- `auth/` owns token creation and auth middleware delegation
- `util/` contains runtime configuration, logging, ID generation, date/time helpers, HTTP response helpers, email, and DB connection helpers
- `config/default_categories.json` seeds built-in categories

### Mapper pattern

Each domain mapper exposes an interface and a package-level `INSTANCE` selected by DB type.

Mapper packages currently include:

- `cash_flow_mapper`
- `category_mapper`
- `user_mapper`
- `refresh_token_mapper`
- `verification_code_mapper`

Both MongoDB and MySQL implementations are expected for production-facing features. When adding persistence features, update both backends unless the change is explicitly database-specific and documented.

### Service packages

Current service packages:

- `cash_flow_service`
- `category_service`
- `db_service`
- `manage_service`
- `refresh_token_service`
- `statistic_service`
- `user_service`
- `verification_service`

Notable implementation split:

- `manage_service` is spread across `backup.go`, `restore.go`, `reset.go`, `indexes.go`, `stats.go`, `init.go`
- `user_service` is split by operation such as `create.go`, `update.go`, `delete.go`, `email_change.go`, `reset_password.go`, `init_admin.go`

## Security-Critical Conventions

### User data isolation

This is still the highest-risk area and must be preserved across layers.

1. Controllers must derive the authenticated user from request context.
2. Services must accept user context where data is user-scoped.
3. Mappers must enforce user filters for user-owned data.

Do not add direct mapper calls from controllers that skip user scoping.

### Soft delete

The codebase uses soft-delete semantics in core entity handling:

- set `is_delete = true`
- set delete metadata
- keep records for audit/history

When modifying queries, verify whether the helper already appends `is_delete = false`. Several DB helpers do this automatically.

### Admin checks

Admin routes use `middleware.Admin`, which expects `role` to be present in request context from the auth layer. If you touch auth middleware or token claims, verify this still works end-to-end.

## Auth and Verification

### Auth components

- `auth/auth.go` exposes the configured auth service singleton
- `auth/service/auth_service.go` wraps provider behavior
- `auth/provider/local_auth.go` implements JWT handling and middleware behavior

### Token-related persistence

- Refresh tokens are stored through `mapper/refresh_token_mapper`
- Verification codes are stored through `mapper/verification_code_mapper`

### Verification flows

`service/verification_service/verification_service.go` supports:

- `email_change`
- `password_reset`

It revokes prior active codes per user/operation, stores expiry, and marks codes used after completion.

## Models and IDs

Key model files:

- `model/base_entity.go`
- `model/cash_flow_entity.go`
- `model/category_entity.go`
- `model/user_entity.go`
- `model/refresh_token.go`
- `model/operation_confirm_code.go`
- `model/response.go`
- `model/constants.go`
- `model/version.go`

Notes:

- Money logic uses decimals to avoid float precision problems
- The app initializes a Snowflake generator at server startup using `snowflake.worker_id`
- Some entities still also rely on Mongo-style ObjectID validation and conversion helpers, so be careful when changing ID behavior

## Configuration

Runtime config is primarily loaded in `util/config_util.go` from `.env` plus environment variables.

Important keys currently loaded there:

- `env`
- `logger.file`
- `logger.level`
- `db.name`
- `db.type`
- `db.mongodb.url`
- `db.mysql.url`
- `api.schema.validation`
- `auth.jwt.secret`
- `auth.jwt.expiration_hours`
- `auth.registration.enabled`
- `admin.username`
- `admin.password`
- `cors.origins`
- `server.port`
- `server.host`
- `timezone`
- `api.version`
- `snowflake.worker_id`
- `default_categories.path`
- `verification.code.expire_minutes`

Important nuance:

- `.env.sample` also documents SMTP settings such as `SMTP_HOST`, `SMTP_PORT`, and friends.
- `util/email/smtp_util.go` reads `smtp.*` keys from `util.GetConfigByKey(...)`.
- `util/config_util.go` does not currently populate those SMTP keys.

If email delivery is expected to work in all environments, this gap should be resolved before relying on SMTP-driven flows.
For now, email should be treated as non-primary/in-progress infrastructure rather than a core guaranteed workflow.

## Database Utilities

`util/database/` contains shared DB helpers.

- `mongodb_util.go` manages the MongoDB client pool and collection helpers
- `mysql_util.go` handles MySQL lifecycle
- `database_util.go` stores default DB names/table constants and shared state

Current connection behavior to be aware of:

- MongoDB pool is initialized eagerly from the Cobra root command when `db.type == mongodb`
- Server startup also indirectly relies on DB access for admin-user initialization
- Many legacy DB helper functions still use package-global connection state and may `log.Fatal` or `panic` on unexpected failures

Be cautious when refactoring DB helpers; some newer code and older compatibility code are both still in play.

## Middleware Stack

Current wrapping in `controller/server.go` is:

```go
handler := middleware.CORS(
    middleware.Logging(
        middleware.Auth(
            middleware.SchemaValidation(r),
        ),
    ),
)
```

Middleware files:

- `middleware/auth.go`
- `middleware/cors.go`
- `middleware/logging.go`
- `middleware/schema_validation.go`

CORS must stay outermost so browser `OPTIONS` preflight requests are answered before auth or OpenAPI schema validation can reject them. This is required for Flutter web and other browser clients.

In `dev` and `test`, loopback browser origins such as `http://localhost:55500` are allowed even when an older exact-port `CORS_ORIGINS` value exists. In production, configure explicit origins through `CORS_ORIGINS`.

There is no `middleware.AdminAuth()` helper in the current code; admin routing uses `middleware.Admin`.

## Validation

Validation lives in `validation/validators.go`.

Current validators include:

- date and date range
- gender
- amount
- ID
- category name
- description
- flow type
- required field
- password

When adding new input DTOs, extend validation here rather than scattering ad hoc checks through controllers.

## Docs, Scripts, and Automation

### Docs

- `docs/api.md`
- `docs/cli.md`
- `docs/openapi.yaml`
- `docs/roadmap.md`

The docs are useful, but code should win when they disagree.

### Scripts

Repo scripts include:

- `scripts/start.ps1`, `scripts/start.sh`
- `scripts/interactive.ps1`, `scripts/interactive.sh`
- `scripts/generate-docs.ps1`, `scripts/generate-docs.sh`

### CI

GitHub Actions workflow: `.github/workflows/ci.yml`

Current CI behavior:

- builds the repo
- runs a narrow test subset: `go test -v ./errors ./validation`
- optionally runs `golangci-lint` if installed
- builds Docker image
- generates Swagger UI HTML docs from `docs/openapi.yaml`

Important: CI is not currently exercising the full application or the DB-backed service packages.
Also note: `go.mod` declares Go `1.23.0`, but `.github/workflows/ci.yml` currently runs Go `1.21`.

## Migrations and Data Setup

Migration assets include:

- MongoDB index script `migrations/001_add_indexes.js`
- MySQL schema creation scripts `002` through `009`
- `config/default_categories.json` for category seeding
- Docker init scripts under `docker/mongodb/` and `docker/mysql/`

When changing persistence shape, update:

1. Mapper code
2. Migration scripts
3. Docker init assets if needed
4. Backup/restore flows if exported shape changes

## Testing Reality

There are some package tests, but test coverage is uneven.

Observed test locations include:

- `validation/validators_test.go`
- `middleware/schema_validation_test.go`
- `util/date_util_test.go`
- `util/http_util_test.go`
- `service/cash_flow_service/*_test.go`
- `service/manage_service/*_test.go`

Before relying on a refactor, check whether the affected path is covered. In many areas, manual verification is still necessary.

## Development Commands

Use these as the code-accurate defaults:

```bash
# Start MongoDB
docker compose --profile mongodb up -d mongodb

# Start MySQL
docker compose --profile mysql up -d mysql

# Start backend in Docker
docker compose --profile backend up -d server

# Run API server locally
go run main.go open start -p 8080

# Check health locally
go run main.go open health

# Show version
go run main.go open version

# Admin backup
go run main.go admin database backup -o backup.json

# Admin restore
go run main.go admin database restore -i backup.json

# Build
go build -o cashlenx main.go

# Run all tests
go test ./...
```

## Guidelines For Future Changes

When adding or changing a feature:

1. Identify whether it is CLI-only, API-only, or shared.
2. Update controller, service, and mapper layers consistently.
3. Preserve user isolation and soft-delete behavior.
4. Update both MongoDB and MySQL implementations.
5. Update OpenAPI/docs if the external contract changes.
6. Update backup/import/export flows if entity shape changes.
7. Add or extend tests in the nearest existing package.

## Project Drift / Known Issues TODO

Use this section as a lightweight backlog of mismatches between implementation, docs, tooling, and intended architecture. Keep it factual and safe to commit.

- [ ] Align `README.md` with current implementation status; it still describes some implemented capabilities as planned
- [ ] Align legacy command references with the real CLI entrypoint `go run main.go open start -p 8080`
- [ ] Decide whether the roadmap version source should move from `v0.3.0` to the current active branch line `dev/v0.4.0`, or whether the roadmap is intentionally lagging behind active development
- [ ] Sync `model/version.go`, `docs/roadmap.md`, OpenAPI `info.version`, and future release/versioning rules so there is one clear canonical version story
- [ ] Review OpenAPI coverage versus actual registered routes in `controller/server.go`; the spec appears incomplete for statistics/chart routes and some auth/token behavior
- [ ] Review OpenAPI enum/value accuracy, for example cash flow type casing and request/response field naming, against the actual API behavior
- [ ] Confirm whether `/open/auth/logout` and `/auth/tokens` route placement is intentional or should be reorganized for clearer auth semantics
- [ ] Wire SMTP settings from `.env` into `util/config_util.go` if email flows are meant to become usable, or explicitly mark them disabled in runtime behavior until then
- [ ] Decide on the future provider strategy for email delivery, likely a third-party provider such as Mailgun, and document the intended integration approach
- [ ] Align CI Go version with `go.mod` so local development and automation target the same toolchain
- [ ] Expand CI and/or local verification to cover more than `./errors` and `./validation`, especially DB-backed service paths as the project matures
- [ ] Review legacy DB helper behavior that still uses package-global state plus `panic`/`log.Fatal`, and gradually normalize error handling
- [ ] Confirm whether MongoDB-only eager initialization in the Cobra root command is still the intended default lifecycle, or if DB initialization should be made more explicit and symmetric across backends
- [ ] Keep `/docs/roadmap.md` synchronized with the actual working branch/version plan as collaboration decisions evolve

## Testing Expectation Right Now

Given the current project stage:

- Prefer sensible baseline verification over heavyweight process
- Run targeted tests when they exist and are relevant
- If there are no meaningful tests for the touched area, rely on build-level or focused manual verification and report that clearly
- Do not invent strict release-process expectations that the project has not adopted yet

## If You Need To Orient Quickly

Start here for most work:

- CLI wiring: `cmd/root.go`
- Server routing: `controller/server.go`
- Auth behavior: `auth/provider/local_auth.go`
- User flows: `controller/user_controller/` and `service/user_service/`
- Cash flows: `controller/cash_flow_controller/` and `service/cash_flow_service/`
- Statistics/export: `controller/statistic_controller/` and `service/statistic_service/`
- Backup/restore: `controller/manage_controller/` and `service/manage_service/`
- Config/runtime helpers: `util/config_util.go`, `util/http_util.go`, `util/log_util.go`
