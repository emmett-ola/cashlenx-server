# CashLenX Roadmap

This file tracks active and future work. Completed milestone history lives in
[`milestones.md`](milestones.md), and the documentation map lives in
[`README.md`](README.md).

## Current Status

- Active branch line: `dev/v0.8.0`
- Active API path: `/api/v0`
- Current implementation version: `0.8.0`
- `v0.8.0` implementation and verification: complete
- Remaining `v0.8.0` action: push/promote the branch, then create the `dev/v0.9.0` line
- Next milestone: `v0.9.0` performance and caching

The Go suite, MongoDB API smoke flow, MySQL migration runner, and independent
numbered SQL sequence have passed against disposable Docker environments.

## v0.9.0 Audit

### Completed foundations

- [x] In-memory category cache used by MongoDB and MySQL mappers
- [x] Exact user/type/parent/name cache keys for user-scoped category lookup
- [x] Defensive entity copies and targeted ID/user invalidation
- [x] Parameterized category-name queries for both mapper backends
- [x] Parameterized dynamic MySQL ID lists and corrected user-scoped deletion columns
- [x] Redis decision for this milestone: do not add Redis; reconsider it when multi-instance deployment requires shared cache coherence

### Remaining work

- [ ] Remove the unused global name-only category-cache index and its compatibility-only tests
- [ ] Add deterministic benchmarks for cash-flow summaries and statistic summary/dashboard calculations
- [ ] Add disposable integration benchmarks for MongoDB and MySQL filtered/date-range mapper queries
- [ ] Capture benchmark baselines before adding another cache layer
- [ ] Decide from measurements whether a recent-query cache provides a material benefit
- [ ] If justified, implement a bounded TTL read-through cache with user-scoped keys, defensive copies, and explicit invalidation on cash/category writes, imports, restores, and account deletion

There is currently no recent-query cache and no `Benchmark...` coverage in the
repository. Statistic and dashboard endpoints repeatedly read user/date-range
cash flows, so those paths are candidates, not preselected cache targets.

## v0.9.0 Execution Order

1. Remove the unused name-only category-cache index so all retained cache keys are ownership-safe.
2. Add service benchmarks using deterministic in-memory mapper fakes.
3. Add build-tagged mapper benchmarks against disposable MongoDB and MySQL databases.
4. Record baseline results and identify actual bottlenecks.
5. Implement recent-query caching only when the measured improvement justifies invalidation complexity.
6. Re-run race tests, both database benchmark suites, and the MongoDB/MySQL smoke flows.

## v0.9.0 Guardrails

- Cache keys must include the authenticated user and every query dimension.
- Cached entities and response objects must never expose mutable internal state.
- Every write path affecting cached results must have a tested invalidation path.
- Cache size and TTL must be bounded and configurable if recent-query caching is added.
- MongoDB and MySQL behavior must remain equivalent.
- Do not add Redis merely to complete a checklist; revisit it with horizontal scaling or cross-process invalidation requirements.

## v0.9.0 Exit Criteria

- Benchmark baselines are committed with reproducible commands and fixture sizes.
- The recent-query cache is either implemented with measured benefit or explicitly rejected with benchmark evidence.
- Category cache no longer maintains unused unscoped lookup state.
- Race tests and relevant disposable database checks pass.
- `README.md`, `AGENTS.md`, and milestone documentation reflect the final decision.

## Future Milestones

### v0.10.0 - Cloud and Self-Hosted Hardening

- [ ] Docker Compose profiles for single-tenant and multi-tenant deployments
- [ ] Helm chart draft for cloud deployments, if needed
- [ ] Secure defaults for production CORS, rate limits, secrets, and operational endpoints
- [ ] Revisit Redis or another shared cache only if multi-instance deployment is adopted

### v1.0.0 - Stable Release Readiness

- [ ] GitHub Actions release pipeline with tagged binaries and images
- [ ] Module caching and reproducible builds
- [ ] Start `CHANGELOG.md` and synchronize displayed versions with release tags
- [ ] Decide and publish the stable `/api/v1` compatibility policy

## Planning Policy

- User-facing functionality takes priority over infrastructure enhancements during `v0.x` development.
- Keep API routes under `/api/v0` until the first stable API release.
- Keep `model/version.go`, OpenAPI `info.version`, this roadmap, and release notes synchronized.
- Treat code as authoritative when documentation drifts, then correct the documentation deliberately.
- Track repository-wide architectural debt in `AGENTS.md`; keep this roadmap focused on milestone work.
