# Documentation Map

Use this page to choose the authoritative document for a task.

| Document | Purpose | Source of truth |
| --- | --- | --- |
| [`openapi.yaml`](openapi.yaml) | Machine-readable HTTP contract | Registered API routes and request/response schemas |
| [`api.md`](api.md) | Human-readable API behavior | Auth semantics, operational endpoints, and API usage notes |
| [`cli.md`](cli.md) | Cobra command reference | Command hierarchy, flags, and examples |
| [`roadmap.md`](roadmap.md) | Active and future work | Current milestone audit, execution order, and exit criteria |
| [`milestones.md`](milestones.md) | Completed history | Delivered milestone capabilities and durable decisions |
| [`performance.md`](performance.md) | Performance baseline | Reproducible benchmarks and cache decisions |
| [`../AGENTS.md`](../AGENTS.md) | Repository working guide | Architecture, workflow, testing, security, and known drift |

## Update Rules

- External API changes require corresponding controller, OpenAPI, and API-note updates.
- CLI shape changes require `cli.md` updates.
- Move a milestone from `roadmap.md` to `milestones.md` only after implementation and verification are complete.
- Keep implementation version values aligned across `model/version.go`, `openapi.yaml`, and user-facing docs.
- Keep transient session progress out of committed documentation.

Database-specific operational documentation remains next to its assets under
`migrations/`, `docker/mongodb/`, and `docker/mysql/`.
