# CashLenX Roadmap

This roadmap tracks backend work by versioned milestones. During the `v0.x` phase, user-facing product functionality takes priority over infrastructure enhancements unless the enhancement directly unblocks user-facing work.

## Current Direction

- Active branch line: `dev/v0.4.0`
- Active API path version: `/api/v0`
- Current roadmap milestone: `v0.4.0` roadmap and product-scope cleanup
- Next feature milestone: `v0.5.0` core user-facing feature completion

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

## Active Milestone

### v0.4.0 - Roadmap and Product-Scope Cleanup

- [ ] Align `README.md` with current implemented features #docs
- [ ] Align legacy command references with `go run main.go open start -p 8080` #docs #dx
- [ ] Reconcile `model/version.go`, OpenAPI `info.version`, roadmap status, and branch naming #docs #devops
- [ ] Compare `docs/openapi.yaml` with `controller/server.go` and list missing or inaccurate API contracts #api #docs
- [ ] Review OpenAPI request/response schemas for field names, enum casing, auth/token behavior, and statistic/chart endpoints #api #docs
- [ ] Confirm `/open/auth/logout` and `/auth/tokens` route placement and document or adjust the intended auth semantics #api #security
- [ ] Verify the core user-facing scope against the Flutter client needs and record incomplete or inconsistent behavior #flutter #api
- [ ] Produce the final `v0.5.0` feature-completion checklist #docs

## Next Feature Milestone

### v0.5.0 - Core User-Facing Feature Completion

- [ ] Fix API behavior gaps discovered in `v0.4.0` #api
- [ ] Fix documentation/API contract gaps required by the Flutter client #flutter #docs
- [ ] Complete or clearly disable unfinished email-dependent flows #api #security
- [ ] Ensure auth, account, cash flow, category, statistic, import/export, and admin APIs have consistent responses and error behavior #api
- [ ] Add practical targeted tests for corrected user-facing flows #api
- [ ] Keep MongoDB as the default development path while preserving MySQL compatibility for touched persistence behavior #data

## Later Enhancement Milestones

### v0.6.0 - Observability

- [ ] Request ID propagation and structured request logging #observability
- [ ] `/metrics` endpoint with Prometheus counters/histograms #observability #devops
- [ ] Enable `pprof` in development #observability

### v0.7.0 - Migration Tooling

- [ ] Introduce MySQL migration tooling and track schema changes #data #devops
- [ ] Validate MongoDB indexes at startup and apply scripts #data
- [ ] Backup/restore CLI with progress and validation #data #devops
- [ ] Integration tests via Docker Compose for MongoDB/MySQL #data #devops
- [ ] Add rollback functionality for failed database operations #data #security

### v0.8.0 - Performance and Caching

- [ ] Extend category cache and add invalidation on writes #performance
- [ ] Optional read-through cache for recent queries #performance
- [ ] Benchmarks for summaries and mapper queries #performance #devops
- [ ] Consider Redis for category caching #performance
- [ ] Implement efficient category-name fetch mapper support #performance
- [ ] Review and fix SQL injection risks in MySQL mappers #security

### v0.9.0 - Cloud and Self-Hosted Hardening

- [ ] Docker Compose profiles for single-tenant and multi-tenant deployments #devops
- [ ] Helm chart draft for cloud deployments, if needed #devops
- [ ] Secure defaults for production CORS, rate limits, and secrets #security #devops

### v1.0.0 - Stable Release Readiness

- [ ] GitHub Actions release pipeline with tagged binaries and images #devops
- [ ] Module caching and reproducible builds #devops
- [ ] Start `CHANGELOG.md` and sync displayed version with release tags #docs
- [ ] Decide and publish stable `/api/v1` compatibility policy #api #docs

## Known Drift To Resolve

- `README.md` is behind current implementation.
- `docs/openapi.yaml` does not cover every registered route.
- Version sources currently still report `0.3.0` while active development is on `dev/v0.4.0`.
- Earlier roadmap entries used confusing `v2.x` headings for work that belongs in the `v0.x` pre-stable history.
- SMTP configuration is documented but not fully wired through runtime config; email flows should be verified before being considered complete.
- CI still runs a narrow test subset and should expand as core user-facing behavior stabilizes.

## Notes

- Keep this roadmap synchronized with `AGENTS.md` when durable planning policy changes.
- Keep enhancement milestones visible, but do not advance them ahead of incomplete user-facing functions unless they unblock those functions.
- Use the Flutter client as an important validation source for the core user-facing API surface.
