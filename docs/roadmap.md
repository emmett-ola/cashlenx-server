# CashLenX Roadmap

This roadmap tracks backend work by versioned milestones. During the `v0.x` phase, user-facing product functionality takes priority over infrastructure enhancements unless the enhancement directly unblocks user-facing work.

## Current Direction

- Active branch line: `dev/v0.8.0`
- Active API path version: `/api/v0`
- Current roadmap milestone: `v0.8.0` implementation complete; final verification and branch promotion remain
- Next enhancement milestone: `v0.9.0` performance and caching
- Integration status: MongoDB and MySQL runtime smoke coverage exists, and the planned `v0.8.0` migration, preflight/progress, rollback, and branch-trigger work is implemented.

## Next Execution Order

1. [Completed] Reconcile `docker/mongodb/init-mongo.js`, `migrations/001_add_indexes.js`, and `service/manage_service/indexes.go` with the current user + type + parent + name category uniqueness model; remove obsolete cash-flow `flow_type` indexes.
2. [Completed] Add a MySQL migration runner and applied-version table, with explicit detection of partially applied or out-of-order migrations.
3. [Completed] Add backup/restore preflight validation and progress reporting to both admin and user CLI flows.
4. [Completed] Define and implement compensating rollback for failed migrations and destructive admin restores, retaining dirty migration state when compensation cannot complete.
5. [Completed] Add `dev/**` to `.github/workflows/smoke.yml` triggers so the managed MongoDB smoke workflow follows the active branch policy.

After final `v0.8.0` verification, continue with the remaining `v0.9.0` work: optional recent-query caching, benchmarks, and a Redis decision.

## Versioning Policy

- Use minor version increments for milestone-sized work during active development.
- Keep API routes under `/api/v0` until the project is ready for a stable API release.
- Introduce `/api/v1` with the first stable release line, such as `v1.0.0`.
- Keep `model/version.go`, OpenAPI `info.version`, this roadmap, and release notes synchronized when version values intentionally change.

## Priority Policy

- Finish and validate user-facing backend capabilities before enhancement milestones.
- Enhancement work includes observability, performance/caching, migration tooling, cloud hardening, and release automation.
- Enhancement work may move earlier only when it is required to unblock user-facing functionality or safe delivery.

## Tags

- `#api` API endpoints and contracts
- `#security` authentication, authorization, secrets, rate limits
- `#docs` documentation and developer experience
- `#devops` CI/CD, build, release
- `#observability` logs, metrics, tracing, pprof
- `#data` migrations, backups, schemas
- `#performance` pagination, caching, efficiency
- `#dx` configuration and local setup
- `#flutter` backend support needed by the Flutter client

## Core User-Facing Scope

Core completion means these backend surfaces are usable and documented enough for the Flutter client and normal API consumers:

- Authentication, registration, login, logout, and token refresh #api #security
- User profile, account deletion, password change, password reset, and email-change flows #api #security
- Cash flow CRUD, date queries, range queries, summaries, pagination, and filtering #api
- Category CRUD, lookup, children, and tree flows #api
- Statistics summaries, breakdowns, trends, top expenses, dashboard, and chart APIs #api
- Import/export, user backup/restore, and admin database dump/restore #api #data
- Admin user and database management #api #security

## Completed Milestones

### v0.1.0 - Base CLI, REST API, and Persistence

- [x] CLI commands and base REST API #api #dx
- [x] Cash flow and category CRUD endpoints #api
- [x] Persistence abstraction with MongoDB and MySQL backends #data
- [x] Import/export MVP and CLI commands #api #data
- [x] Docker Compose baseline for local/self-hosted deployments #devops
- [x] CORS, request logging, health, and version endpoints #api #observability
- [x] Basic docs and unit tests in validation/cache/errors #docs

### v0.2.0 - API Contract, Tooling, and Developer Experience

- [x] OpenAPI coverage for core endpoints #api #docs
- [x] OpenAPI schema validation middleware #api
- [x] Pagination and filtering for listing endpoints #api #performance
- [x] Interactive scripts and cross-platform helper scripts #dx #devops
- [x] `.env` loading for CLI/API consistency #dx
- [x] Docker build/startup improvements and health checks #devops
- [x] HTML docs artifact generation from OpenAPI #docs #devops
- [x] Consistent response wrapper and centralized error mapping #api
- [x] Backup/restore statistics reporting and CI baseline #data #devops

### v0.3.0 - User Management, Authentication, and Isolation

- [x] JWT authentication and local user model #security #api
- [x] User registration and login #security #api
- [x] Per-user data isolation across storage backends #security #data
- [x] Admin user management endpoints #api #security
- [x] Route organization for `/open`, `/admin`, and authenticated user APIs #api
- [x] Flow type validation and enum use #api
- [x] Backup/restore with user data support #data

### Completed Out Of Order - Statistics, Import/Export, and API Versioning

These capabilities exist in code today but were previously recorded under later/confusing version headings. They should be verified during `v0.4.0` cleanup and folded into the canonical milestone history before stable release.

- [x] User-scoped statistics summary, breakdown, trend, top expense, dashboard, and chart APIs #api
- [x] Multi-format export, including CSV, Excel, and PDF #api #data
- [x] Import with category auto-creation and user association #api #data
- [x] User-scoped exports/imports with ownership checks #security #data
- [x] Batch insert and bulk data processing mapper support #data #performance
- [x] Global API path versioning through `/api/v0` and `API_VERSION` #api #dx
- [x] Auth flow refinements around login/refresh and multipart validation fixes #api #security

## Completed Milestone

### v0.4.0 - Roadmap and Product-Scope Cleanup

- [x] Align `README.md` with current implemented features #docs
- [x] Align legacy command references with `go run main.go open start -p 8080` #docs #dx
- [x] Reconcile `model/version.go`, OpenAPI `info.version`, roadmap status, and branch naming #docs #devops
- [x] Compare `docs/openapi.yaml` with `controller/server.go` and list missing or inaccurate API contracts #api #docs
- [x] Review OpenAPI request/response schemas for field names, enum casing, auth/token behavior, and statistic/chart endpoints #api #docs
- [x] Confirm `/open/auth/logout` and `/auth/tokens` route placement and document or adjust the intended auth semantics #api #security
- [x] Verify the core user-facing scope against the Flutter client needs and record incomplete or inconsistent behavior #flutter #api
- [x] Produce the final `v0.5.0` feature-completion checklist #docs

Cleanup notes:

- `README.md` now describes implemented auth, account, cash, category, statistics, import/export, admin, Docker, and OpenAPI capabilities.
- Current local server command is documented as `go run main.go open start -p 8080`.
- `model/version.go` and OpenAPI `info.version` now report `0.8.0` while API routes remain under `/api/v0`.
- OpenAPI now covers the registered auth/token/password-reset, category tree, and statistic/dashboard/chart routes from `controller/server.go`.
- Cash/category enum examples use lowercase `income` and `expense`, matching `model/constants.go`.
- `/open/auth/logout` is mounted under `/open` as a public idempotent endpoint; it returns OK without credentials and revokes sessions only when a valid refresh token or bearer access token is supplied. `/auth/tokens` is documented as an authenticated token-management endpoint.
- Flutter-facing backend validation for `v0.4.0` found no new code task beyond completing and verifying the documented API surface; deeper client smoke testing belongs in active beta readiness work.

## Completed Milestone

### v0.5.0 - Core User-Facing Feature Completion

- [x] Reconcile token expiration defaults: access tokens use `JWT_EXPIRATION_MINUTES` / `auth.jwt.expiration_minutes` with a 30-minute default, and refresh tokens use `REFRESH_TOKEN_EXPIRATION_DAYS` / `auth.refresh_token.expiration_days` with a 14-day default #security #api
- [x] Decide beta support stance: MongoDB is the supported beta deployment backend, while MySQL should remain build-compatible and avoid known mapper gaps in touched code #data #security
- [x] Normalize response and error behavior across auth, account, cash flow, category, statistic, import/export, and admin APIs #api
- [x] Add practical targeted tests for corrected user-facing flows, prioritizing paths touched by fixes #api
- [x] Align CI Go version and branch coverage with `go.mod` and the active `dev/*` branch line before beta tagging #devops
- [x] Update `docs/openapi.yaml`, `README.md`, `AGENTS.md`, and implementation docs whenever a user-facing API contract changes #docs

## Completed Milestone

### v0.6.0 - Beta Readiness and Service Testability

- [x] Move category, cash flow, and statistic services to constructor-based mapper dependency injection while preserving package-level compatibility functions #api #dx
- [x] Keep service unit tests database-free with in-memory mapper fakes for category, cash flow, statistic, user, and verification flows #api #dx
- [x] Enforce admin role lifecycle: only startup bootstrap creates admin accounts, generic creation creates users, role updates are rejected, and admin deletion is blocked #security #api
- [x] Align displayed implementation version and active branch metadata to `0.6.0` while API routes remain under `/api/v0` #docs #devops
- [x] Run MongoDB-backed managed beta smoke checks against `/api/v0` for admin bootstrap, login, token refresh, logout, profile, password change, cash flow, category, statistics, export, user backup, and admin APIs #api #security
- [x] Add a managed smoke workflow that starts disposable MongoDB and the local API server before running `scripts/smoke-api.sh --managed` #api #devops

## Carried Follow-Up Work

These items were left behind from beta-readiness work and should be resolved deliberately, but they do not block tracking `dev/v0.8.0` as the active branch line.

- [x] Manually verify configured SMTP delivery; do not require repeated provider-level smoke testing for routine backend changes #api #security
- [x] Cover public registration and password-reset request/confirmation procedures with deterministic tests that replace email delivery and persistence side effects #api #security
- [x] Run disposable MongoDB-backed Flutter-client smoke checks against `/api/v0` for login, registration, logout, token refresh, profile, password reset, cash flow, category, statistics, import/export, backup, and admin flows #flutter #api
- [x] Resolve and verify currently known beta API smoke gaps: accepted gender values, compact cash-flow dates, period-specific statistic dates, and MySQL category mapper selection #flutter #api #data
- [x] Confirm the MongoDB default path from a fresh disposable container through `scripts/smoke-api.sh --managed` and the managed smoke workflow #data #devops
- [x] Keep MySQL build-compatible and verify the full Flutter/API flow against disposable MySQL 8 #data #flutter

## Carried Enhancement Milestone

### v0.7.0 - Observability

- [x] Request ID propagation and structured request/error logging #observability
- [ ] `/metrics` endpoint with Prometheus counters/histograms #observability #devops
- [ ] Enable `pprof` in development #observability

## Completed Milestone

### v0.8.0 - Migration Tooling

- [x] Maintain numbered MongoDB and MySQL migration assets under `migrations/` #data
- [x] Introduce a MySQL migration runner with checksums, dirty-state detection, existing-schema baselining, and applied-version tracking #data #devops
- [x] Reconcile MongoDB index definitions, validate them at startup, and invoke index management safely #data
- [x] Extend admin/user backup and restore CLI flows with progress reporting and explicit destination/schema/relationship preflight validation #data #devops
- [x] Run MongoDB API integration smoke checks against a fresh disposable container #data #devops
- [x] Add disposable MySQL runtime and numbered-migration integration coverage, separate from normal unit tests #data #devops
- [x] Add compensating rollback for failed MySQL migrations and destructive admin restore, with fail-stop dirty state when migration rollback cannot complete #data #security

## Later Enhancement Milestones

### v0.9.0 - Performance and Caching

- [x] Provide the in-memory category cache and invalidate it on category writes for both mapper backends #performance
- [x] Extend cache coverage to exact user-scoped category lookups, return defensive copies, and replace broad cache clears with targeted invalidation where safe #performance
- [ ] Optional read-through cache for recent queries #performance
- [ ] Benchmarks for summaries and mapper queries #performance #devops
- [ ] Consider Redis for category caching #performance
- [x] Implement parameterized category-name lookup methods for MongoDB and MySQL mappers #performance #security
- [x] Review MySQL mapper query construction, parameterize dynamic ID lists, and correct user-scope column names in deletion queries #security

### v0.10.0 - Cloud and Self-Hosted Hardening

- [ ] Docker Compose profiles for single-tenant and multi-tenant deployments #devops
- [ ] Helm chart draft for cloud deployments, if needed #devops
- [ ] Secure defaults for production CORS, rate limits, and secrets #security #devops

### v1.0.0 - Stable Release Readiness

- [ ] GitHub Actions release pipeline with tagged binaries and images #devops
- [ ] Module caching and reproducible builds #devops
- [ ] Start `CHANGELOG.md` and sync displayed version with release tags #docs
- [ ] Decide and publish stable `/api/v1` compatibility policy #api #docs

## Known Drift To Resolve

- Keep `README.md`, `docs/openapi.yaml`, `model/version.go`, and this roadmap synchronized when route contracts or milestone versions change.
- SMTP delivery has been manually verified by the project owner. Registration and password-reset procedure tests should remain deterministic and must not send real email.
- `scripts/smoke-api.sh --managed` provides a MongoDB-backed beta API smoke flow for the non-SMTP core API surface by starting disposable MongoDB and a local API server.
- CI now uses GitHub Actions for build/test/release automation, Codecov for coverage reporting, and DeepSource for code analysis; managed live API smoke testing is available through `.github/workflows/smoke.yml`.
- The main CI and managed MongoDB smoke workflows include the active `dev/**` branch line.
- `service/manage_service/indexes.go` now reconciles current indexes during server startup. Migration version tracking remains separate work.
- `docker/mongodb/init-mongo-demo.js` is a legacy single-user fixture and is not compatible with current ownership, audit, category-type, or BSON date fields; use managed API smoke data instead.
- `../cashlenx-app/scripts/smoke-api.ps1 -Database mysql` runs the Flutter/API contract against disposable MySQL; `scripts/smoke-mysql-migrations.ps1` validates the numbered SQL sequence independently.

## Notes

- Keep this roadmap synchronized with `AGENTS.md` when durable planning policy changes.
- Keep enhancement milestones visible, but do not advance them ahead of incomplete user-facing functions unless they unblock those functions.
- Use the Flutter client as an important validation source for the core user-facing API surface.
