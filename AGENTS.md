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
- Gin is still a direct dependency, but current HTTP routing is Gorilla Mux; Gin only appears in legacy response helpers under `util/http_util.go`
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
- We are currently developing on the `dev/v0.6.0` branch line
- The branch line maps to the feature/version batch currently in progress
- Once planned work for `v0.6.0` is complete, it is intended to be merged/promoted to `main`, then development moves to the next branch line such as `dev/v0.7.0`
- API versioning stays under `/api/v0` during active development
- A stable release is expected to introduce `/api/v1` alongside a stable application version such as `v1.0.0`
- User-facing feature/function completion takes priority over enhancement work such as observability, performance, migration tooling, cloud hardening, and release automation
- Enhancement work should move earlier only when it directly unblocks user-facing functionality or safe delivery
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
- User profile query/update with nickname, gender, active status, user configuration profile query/create/update, password change, email change request/confirm, and account deletion
- Password reset request/confirm using verification codes
- Purpose-scoped email verification code send/verify flow for sign-up, password reset, and email change
- Login can also refresh tokens by passing `refresh_token` to `POST /open/auth/login`
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
  - `summary`, `breakdown`, `trends`, `top`, `dashboard`, `chart`, `export`, `import`

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
- `POST /api/{version}/open/verification/code`
- `POST /api/{version}/open/verification/verify`
- `POST /api/{version}/open/auth/reset-password`
- `POST /api/{version}/open/auth/reset-password/confirm`

Note: `logout` lives under `/open/auth/logout` and is intentionally public/idempotent. It returns OK without credentials and only performs revocation when a valid `refresh_token` or bearer access token is provided.

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
  - `GET /api/{version}/user/configuration`
  - `POST /api/{version}/user/configuration`
  - `PUT /api/{version}/user/configuration`
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
- `operation_confirm_code_mapper`

Both MongoDB and MySQL implementations are expected for production-facing features. When adding persistence features, update both backends unless the change is explicitly database-specific and documented.

Beta deployments are MongoDB-first, but touched persistence code should remain build-compatible across MongoDB and MySQL. Verification-code persistence is implemented for both mapper backends.

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

Admin role lifecycle:

- Admin users are only bootstrapped by `user_service.InitAdminUser()` during server startup when no admin account exists
- Registration and admin/user creation paths must always create `user` role accounts, even if a request body includes `role: admin`
- User update paths must not promote or demote roles; preserve admin role changes for explicit future design work rather than accepting them through generic user management
- User deletion must reject admin accounts

## Auth and Verification

### Auth components

- `auth/auth.go` exposes the configured auth service singleton
- `auth/service/auth_service.go` wraps provider behavior
- `auth/provider/local_auth.go` implements JWT handling and middleware behavior

### Token-related persistence

- Refresh tokens are stored through `mapper/refresh_token_mapper`
- Verification codes are stored through `mapper/operation_confirm_code_mapper`

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
- `auth.jwt.expiration_minutes`
- `auth.refresh_token.expiration_days`
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
- `verification.code.send_interval_seconds`
- `smtp.enabled`
- `smtp.host`
- `smtp.port`
- `smtp.username`
- `smtp.password`
- `smtp.from_address`
- `smtp.from_name`
- `smtp.max_retries`
- `smtp.retry_interval`
- `smtp.rate_limit.daily_per_ip`
- `smtp.rate_limit.daily_per_email`

Important nuance:

- Use `db.mongodb.url` and `db.mysql.url` for database connection strings; legacy `mongodb.uri` and `mysql.uri` keys are intentionally not registered.
- SMTP keys are wired from `.env`, but email flows still need provider-level smoke testing before beta.

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

`middleware.Logging` adds or preserves `X-Request-ID`, stores it in request context, echoes it in the response header, and includes it in structured request logs. `util.ComposeErrorResponse` logs API errors centrally with request ID, status, method/path, caller location, and user ID when present; 4xx responses log at warn level and 5xx responses log at error level.

Auth middleware skips all `/api/{version}/open/*` routes. `/open/auth/logout` handles optional token validation in its controller so the `/open` prefix remains consistently public.

OpenAPI schema validation loads `docs/openapi.yaml` at package init when enabled. Keep route paths in that spec aligned with `controller/server.go`; validation is bypassed automatically if the spec cannot be loaded or parsed.

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
- `scripts/smoke-api.sh`
- `scripts/ci-test.sh`

### CI

GitHub Actions workflow: `.github/workflows/ci.yml`
Codecov configuration: `codecov.yml`
DeepSource configuration: `.deepsource.toml`

Current CI behavior:

- builds the repo
- runs `scripts/ci-test.sh`, which executes `go test -v -race -covermode=atomic -coverprofile=coverage.out ./...`
- uploads coverage to Codecov from GitHub Actions when provider credentials/network are available
- uses DeepSource for repository code analysis
- builds Docker image
- generates Swagger UI HTML docs from `docs/openapi.yaml`

Important: CI unit tests do not run the live MongoDB-backed `scripts/smoke-api.sh`; that remains an environment smoke check against a running server.

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

- `cache/category_cache_test.go`
- `errors/errors_test.go`
- `middleware/cors_test.go`
- `validation/validators_test.go`
- `middleware/schema_validation_test.go`
- `util/date_util_test.go`
- `util/http_util_test.go`
- `service/cash_flow_service/*_test.go`
- `service/category_service/*_test.go`
- `service/manage_service/*_test.go`
- `service/refresh_token_service/refresh_token_test.go`
- `service/statistic_service/*_test.go`
- `service/user_service/*_test.go`
- `service/verification_service/*_test.go`

Before relying on a refactor, check whether the affected path is covered. In many areas, manual verification is still necessary.

Unit test guidance:

- Unit tests should be deterministic and should not create, delete, or mutate real files or database data
- Prefer in-memory fakes/stubs and package-local dependency seams for mapper, email, token, time, randomness, and other side effects
- If code structure blocks practical unit testing, refactor the source code to expose a small testable seam while preserving business logic, API contracts, persistence behavior, and error semantics
- Keep live database and filesystem verification in integration/smoke checks, not unit tests
- Service-layer code should use constructor-based dependency injection for mappers; package-level functions may remain for compatibility, but should delegate to default service instances
- Service unit tests should instantiate service structs with in-memory fake mapper implementations instead of overriding mapper globals
- Mapper integration tests should be isolated from normal unit tests, use explicit integration naming/build tags or scripts, and run only against disposable test databases
- API integration tests should exercise controller-to-database behavior separately from `go test ./...`, for example through the smoke script against Docker-backed services

CLI statistic import/export note: `cmd/statistic_cmd/export.go` and `cmd/statistic_cmd/import.go` accept `--user`, but currently fall back to `user_service.GetDefaultAdminUserId()` when it is omitted. Treat this as a development convenience, not a finished multi-user CLI auth story.

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

# Run API smoke flow against a running local server
BASE_URL=http://localhost:8080/api/v0 scripts/smoke-api.sh
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

- [ ] Keep `README.md`, `docs/openapi.yaml`, `docs/roadmap.md`, and `model/version.go` synchronized when the active milestone or API contract changes
- [x] Treat `/open/auth/logout` as a public idempotent compatibility path; it only revokes sessions when a valid token is supplied
- [ ] Treat `/auth/tokens` as authenticated token-management API; keep OpenAPI/docs explicit about its auth expectation
- [ ] Smoke test SMTP-backed password reset and email-change flows with a real provider before beta
- [ ] Decide on the future provider strategy for email delivery, likely a third-party provider such as Mailgun, and document the intended integration approach
- [ ] Replace statistic CLI import/export default-admin fallback with an explicit user/auth model before treating those commands as production-ready multi-user workflows
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
