# CashLenX Server

Go backend providing a CLI and REST API for daily expense tracking. Users can record income/expense cash flows with amount, category, date, and an optional remark. The server is designed for both self-hosted and cloud deployments and integrates with an external web UI.

## Features
- Income/expense tracking with amount, category, date, remark
- Category-based organization
- Date range queries and summaries
- Import/export for backup and migration (Excel; CSV planned)
- Pluggable storage via abstract persistence interface (MongoDB, MySQL)
- Structured logging and validation
- CLI tooling (Cobra) and REST API (Gorilla Mux)
- Docker Compose for local/self-hosted setups

## Planned
- User management with per-user data isolation
- OIDC authentication with local user records
- Statistics endpoints for insights and reporting
- OpenAPI specification and generated docs

## Project Structure
```
cashlenx-server/
├── cmd/                 # CLI commands (Cobra)
├── controller/          # HTTP controllers
├── service/             # Business logic
├── mapper/              # Database mappers
├── model/               # Data models
├── middleware/          # HTTP middleware
├── util/                # Utilities
├── docker/              # Database init assets
├── docs/                # Server documentation
└── main.go              # Entry point
```

## Quick Start

### 1) Database
```bash
cd cashlenx-server
docker compose up -d mongodb
# or
docker compose --profile mysql up -d mysql
```

### 2) Configure
```bash
cd cashlenx-server
cp .env.sample .env
export $(cat .env | xargs)
```

### 3) Run
```bash
# API server
go run main.go server start -p 8080

# CLI examples
go run main.go cash expense -c "Food" -a 45.50 -d "Lunch"
go run main.go cash income -c "Salary" -a 5000
go run main.go cash summary -f 2024-01-01 -t 2024-01-31
```

## REST API
Base URL: `http://localhost:8080/api/v0` (configurable version)

- `POST /api/v0/cash/expense`
- `POST /api/v0/cash/income`
- `GET /api/v0/cash/{id}`
- `GET /api/v0/cash/date/{date}`
- `DELETE /api/v0/cash/{id}`
- `DELETE /api/v0/cash/date/{date}`
- `GET /api/v0/open/health`
- `GET /api/v0/open/version`

See `docs/api.md` for detailed endpoints.

## Documentation
- `docs/cli.md` — CLI command reference
- `docs/api.md` — REST API reference
- `docs/roadmap.md` — Versioned roadmap and task tracking
- `docs/openapi.yaml` — OpenAPI Specification

## Build and Test
```bash
go build -o cashlenx main.go
go test .
```

## Technology
- Go 1.21+
- Cobra, Gorilla Mux, Zap, Excelize
- MongoDB and MySQL drivers

## License
See `LICENSE` for details.
