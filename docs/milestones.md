# Completed Milestones

This file preserves completed CashLenX backend milestone history. Active and
future work is tracked in [`roadmap.md`](roadmap.md).

## v0.1.0 - Base CLI, REST API, and Persistence

- Cobra CLI and Gorilla Mux REST API
- Cash-flow and category CRUD
- MongoDB/MySQL mapper abstraction
- Initial import/export and Docker Compose support
- CORS, logging, health/version endpoints, documentation, and baseline tests

## v0.2.0 - API Contract and Developer Experience

- OpenAPI coverage and request validation
- Pagination/filtering and consistent response/error mapping
- `.env`, startup, interactive, and documentation-generation tooling
- Docker health checks, CI baseline, and backup/restore operation statistics

## v0.3.0 - Authentication, Users, and Isolation

- JWT authentication, registration, login, and persisted refresh tokens
- User/admin management and `/open`, `/admin`, and authenticated route groups
- Per-user data isolation across MongoDB and MySQL
- Flow-type validation and user-aware backup/restore

## Cross-Milestone Capabilities Completed Before v0.4.0

- User-scoped statistics, breakdown, trends, top expenses, dashboard, and charts
- CSV, Excel, and PDF export
- User-scoped import with category auto-creation
- Batch insert/bulk processing support
- Global `/api/v0` path versioning and refined login/refresh behavior

## v0.4.0 - Roadmap and Product-Scope Cleanup

- Aligned README, API, CLI, OpenAPI, version, and branch documentation
- Reconciled registered routes with OpenAPI contracts and enum casing
- Established public idempotent `/open/auth/logout` semantics
- Established authenticated `/auth/tokens` semantics
- Verified the Flutter-facing backend scope and produced the v0.5 checklist

## v0.5.0 - Core User-Facing Completion

- Reconciled access/refresh token expiration configuration
- Established MongoDB-first beta support while maintaining MySQL compatibility
- Normalized user-facing response/error behavior
- Added practical tests for corrected flows
- Aligned CI Go version and active development branch coverage

## v0.6.0 - Beta Readiness and Testability

- Constructor-based mapper injection for category, cash-flow, and statistic services
- Database-free service tests with in-memory fakes
- Enforced admin bootstrap, role, update, and deletion rules
- Added managed MongoDB API smoke workflow
- Verified registration, reset-password, SMTP configuration, Flutter MongoDB flow, and MySQL runtime parity

## v0.7.0 - Observability

- Request ID propagation and structured request/error logging
- Prometheus counters/histograms at `/metrics`
- Development-only standard Go pprof endpoints

## v0.8.0 - Migration Tooling

- Reconciled MongoDB/MySQL migration and index assets
- Added tracked MySQL migrations with checksums, baselining, dirty-state and ordering detection
- Added backup/restore preflight validation and CLI progress reporting
- Added compensating rollback for MySQL migrations and destructive admin restore
- Added disposable MongoDB API, MySQL runner, and numbered SQL verification
- Added active `dev/**` smoke-workflow triggers

## Durable Historical Decisions

- API paths remain under `/api/v0` throughout active `v0.x` development.
- MongoDB is the default development database; touched persistence behavior must preserve MySQL parity.
- SMTP delivery is manually verified, while automated email-flow tests replace provider delivery.
- CORS remains the outermost API middleware for browser preflight compatibility.
- Operational `/metrics` bypasses JWT/OpenAPI middleware; pprof is registered only in development.
