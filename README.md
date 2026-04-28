# CashLenX Server

CashLenX Server is a Go backend for personal finance tracking. It provides a Cobra CLI and a Gorilla Mux REST API for authentication, user accounts, cash flows, categories, statistics, import/export, and admin database management.

The project is still in active `v0.x` development. The current API path is `/api/v0`; stable `/api/v1` compatibility is planned for a later stable release.

## Features

- Local registration/login with JWT access tokens and persisted refresh tokens
- User profile, password change, password reset, email-change request/confirm, and account deletion flows
- Per-user cash flow CRUD, date/range queries, summaries, pagination, and filtering
- Per-user category CRUD, name lookup, children lookup, and tree APIs
- Statistics, dashboard, and chart endpoints for summaries, breakdowns, trends, top expenses, income/expense charts, category distribution, monthly comparison, and spending heatmaps
- User import/export and backup/restore flows
- Admin user management and full database backup/restore
- MongoDB and MySQL persistence implementations
- Docker Compose profiles for local MongoDB, MySQL, and backend startup
- OpenAPI contract in `docs/openapi.yaml`

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
├── scripts/                 # Start and docs helper scripts
├── service/                 # Business logic
├── util/                    # Config, logging, DB, email, date, ID, HTTP helpers
├── validation/              # Validation helpers and tests
└── main.go                  # Entry point
```

## Quick Start

### 1. Configure

```bash
cp .env.sample .env
```

MongoDB is the default development database, but users may choose MongoDB or MySQL during bootstrap by setting the DB-related values in `.env`.

### 2. Start a Database

```bash
# MongoDB
docker compose --profile mongodb up -d mongodb

# MySQL
docker compose --profile mysql up -d mysql
```

### 3. Run the API Server

```bash
go run main.go open start -p 8080
```

The local base URL is:

```text
http://localhost:8080/api/v0
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
- `GET /api/v0/statistic/dashboard/{period}/{date}`
- `GET /api/v0/statistic/chart/income-expense/{period}/{date}`

See `docs/openapi.yaml` for the current API contract and `docs/api.md` for additional API notes.

## Documentation

- `AGENTS.md` - shared working guide for coding agents
- `docs/cli.md` - CLI command reference
- `docs/api.md` - REST API notes
- `docs/roadmap.md` - versioned roadmap and task tracking
- `docs/openapi.yaml` - OpenAPI specification

## Build and Test

```bash
go build -o cashlenx main.go
go test ./...
```

Test coverage is still uneven while the project is under development. Prefer targeted tests for touched areas and report manual verification clearly when no meaningful tests exist.

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

## License

See `LICENSE` for details.
