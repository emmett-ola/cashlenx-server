# CashLenX Roadmap (Versioned Todo)

## Versioning Policy
- Current version: `v0.3.0`
- Increase minor by `+0.1` for each completed big feature
- Optionally use patch `+0.0.1` for bugfix releases

## Tags
- `#api` API endpoints and contracts
- `#security` authentication, authorization, secrets, rate limits
- `#docs` documentation and developer experience
- `#devops` CI/CD, build, release
- `#observability` logs, metrics, tracing, pprof
- `#data` migrations, backups, schemas
- `#performance` pagination, caching, efficiency
- `#dx` configuration and local setup

## v0.1.0
- [x] CLI commands and base REST API #api #dx
- [x] Cash flow and category CRUD endpoints #api
- [x] Persistence abstraction; MySQL and MongoDB backends #data
- [x] Import/export MVP (Excel) and CLI commands #api #data #cli
- [x] Docker Compose baseline for self-hosted deployments #devops #cloud
- [x] Middleware: CORS and request logging #observability
- [x] Health and version endpoints #api
- [x] Basic docs: README, API, CLI, deployment #docs
- [x] Unit tests in validation, cache, and errors #api

## v0.2.0 — Big Feature: API Contract & OpenAPI
- [x] Finalize OpenAPI coverage for core endpoints #api #docs
- [x] Auto-validate requests/responses against schema with middleware #api
- [x] Add pagination and filtering to listing endpoints #api #performance
- [x] Enhance interactive scripts for all commands #cli #dx
- [x] Add cross-platform support with PowerShell scripts #devops #dx
- [x] Improve code documentation and translate comments to English #docs
- [x] Update dependencies and align toolchain #devops #dx
- [x] Fix CLI query empty response by adding .env file support with godotenv #dx #api
- [x] Ensure consistent env var handling between CLI and API #dx
- [x] Fix slow server startup by implementing multi-stage Docker build #devops
- [x] Update Docker configuration to use health API for container readiness #devops #observability
- [x] Publish HTML docs artifact from OpenAPI #docs #devops
- [x] Introduce consistent response wrapper `{code,data,message,meta,extra,errors}` #api
- [x] Centralize error types and mapping #api
- [x] Increase unit test coverage in services and mappers #api
- [x] Implement ADMIN_TOKEN verification for sensitive endpoints #security #api
- [x] Add truncate database functionality #data #security
- [x] Enhance backup/restore operations with statistics reporting #data #observability
- [x] Set up CI (build, test, lint, Docker image) #devops

## v0.3.0 — Big Feature: User Management & Authentication (Completed)
- [x] Optional JWT auth middleware protecting mutating endpoints #security #api
- [x] Minimal user model and env toggles (`AUTH_ENABLED`, `JWT_SECRET`) #security #dx
- [x] Role-less single-user default; document flows #docs
- [x] OIDC login support with local user records #security #api
- [x] Per-user data isolation across storage backends #security #data
- [x] Admin endpoints for user lifecycle (disable/export/delete) #api #security
- [x] Use enums for flow types and validate values #api

## v0.4.0 — Big Feature: Observability
- [ ] Request IDs propagation and structured logging #observability
- [ ] `/metrics` endpoint with Prometheus counters/histograms #observability #devops
- [ ] Enable `pprof` in development #observability

## v0.5.0 — Big Feature: Import/Export Refinements
- [x] CSV format alongside Excel; unify parsers #api #data
- [x] Bulk import validation with partial success reporting #api
- [x] Schema versioning for exports #data #docs
- [x] User-scoped exports/imports with ownership checks #security #data
- [x] Add batchInsert() method to mappers for efficient bulk data processing #data #performance
- [x] Add DeleteAllCashFlows() and DeleteAllCategories() methods to mapper interfaces #data
- [x] Implement data backup and restoration functionality #data #devops

## v0.6.0 — Big Feature: Migration Tooling
- [ ] Introduce MySQL migration tooling and track schema changes #data #devops
- [ ] Validate MongoDB indexes at startup and apply scripts #data
- [ ] Backup/restore CLI with progress and validation #data #devops
- [ ] Integration tests via Docker Compose for MongoDB/MySQL #data #devops
- [ ] Add rollback functionality for failed database operations #data #security

## v0.7.0 — Big Feature: Statistics & Insights (Completed)
- [x] Summary endpoints per user, per category, per period #api #stats
- [x] Trend and aggregation endpoints (weekly/monthly) #api #stats
- [x] Export stats reports (CSV/Excel) #api #stats #data
- [x] Add aggregation methods to mappers (CountCashFlowsByType, GetEarliest/latest dates) #data #stats

## v0.8.0 — Big Feature: Performance & Caching
- [ ] Extend category cache; add invalidation on writes #performance
- [ ] Optional read-through cache for recent queries #performance
- [ ] Benchmarks for summaries and mapper queries #performance #devops
- [ ] Use Redis for category caching #performance
- [ ] Implement mapper for fetching category name efficiently #performance
- [ ] Fix SQL injection vulnerability in MySQL mappers #security

## v2.1.0 — Big Feature: Statistic Module (Completed)
- [x] Three-layer architecture for statistics (Controller -> Service -> Mapper) #api #stats
- [x] Multi-format export (Excel, CSV, PDF) with binary file download #api #data
- [x] Import functionality with category auto-creation and user association #api #data
- [x] Financial summaries (daily/monthly/yearly) with user isolation #api #stats
- [x] Category breakdown analysis with percentage calculation #api #stats
- [x] Spending trends analysis (time-series data) #api #stats
- [x] Top expenses ranking #api #stats
- [x] Dashboard visualization endpoints for charts (income-expense, distribution, heatmap) #api #stats
- [x] CLI commands for all statistic features with user context #cli
- [x] Strict user data isolation across all statistic operations #security

## v2.0.0 — Big Feature: User Isolation & Admin (Completed)
- [x] User authentication and authorization with JWT #security
- [x] User data isolation for cash flows and categories #security
- [x] Reorganized routes into /open and /admin #api
- [x] Admin user management endpoints #api #security
- [x] Backup/restore with user data support #data

## v0.9.0 — Big Feature: Cloud & Self-Hosted
- [ ] Docker Compose profiles for single-tenant and multi-tenant #devops #cloud
- [ ] Helm chart draft for cloud deployments (optional) #devops #cloud
- [ ] Secure defaults for production (CORS, rate limits, secrets) #security #devops

## v1.0.0 — Big Feature: DevOps & Releases
- [ ] GitHub Actions: release pipeline with tagged binaries and images #devops
- [ ] Module caching; reproducible builds #devops
- [ ] Start `CHANGELOG.md` and sync displayed version with code #docs #devops

## Version Sources
- `model/version.go:4` defines the canonical version constant
- `cmd/version.go:12` prints version in CLI `cashlenx-server version`
- `controller/server.go:91` returns version from `GET /api/open/version`

## Notes
- Each completed big feature increases the minor version by `+0.1`
- When a version’s todo list is fully checked, tag and release that version
- Keep roadmap synchronized with CLI `version` output and `CHANGELOG.md`
