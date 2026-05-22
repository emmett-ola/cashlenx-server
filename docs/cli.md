# CashLenX CLI Reference

**Version**: 0.6.0
**Last Updated**: 2026-05-22

The CashLenX CLI is implemented with Cobra and starts from `main.go -> cmd.Execute()`. The current executable name is `cashlenx` when built, or `go run main.go` during local development.

## Quick Start

```bash
# Start the API server
go run main.go open start -p 8080

# Check server health
go run main.go open health

# Show version information
go run main.go open version
```

## Command Tree

```text
cashlenx
├── open
│   ├── start
│   ├── health
│   └── version
├── admin
│   └── database
│       ├── backup
│       └── restore
├── cash
│   ├── expense
│   ├── income
│   ├── list
│   ├── query
│   ├── range
│   ├── summary
│   ├── update
│   └── delete
├── category
│   ├── create
│   ├── list
│   ├── query
│   ├── tree
│   ├── update
│   └── delete
└── statistic
    ├── summary
    ├── breakdown
    ├── trends
    ├── top
    ├── dashboard
    ├── chart
    │   ├── income-expense
    │   ├── category-distribution
    │   ├── monthly-comparison
    │   └── spending-heatmap
    ├── export
    └── import
```

## Open Commands

```bash
go run main.go open start -p 8080
go run main.go open health
go run main.go open version
```

- `open start` starts the REST API server.
- `open health` calls the local health endpoint.
- `open version` prints version, build time, git commit, Go version, and OS/architecture.

## Admin Commands

```bash
go run main.go admin database backup -o backup.json
go run main.go admin database restore -i backup.json --force
```

- `admin database backup` exports a full database dump.
- `admin database restore` restores a full database dump.
- Use admin commands carefully because they operate beyond a single user's normal finance data surface.

## Cash Commands

```bash
go run main.go cash expense -c "Food" -a 45.50 -d "Lunch"
go run main.go cash income -c "Salary" -a 5000
go run main.go cash list --limit 20 --offset 0 --type expense
go run main.go cash query --id <cash_flow_id>
go run main.go cash query --date 2026-01-01
go run main.go cash range --from 2026-01-01 --to 2026-01-31
go run main.go cash summary --period monthly --date 2026-01
go run main.go cash update --id <cash_flow_id> --amount 50
go run main.go cash delete --id <cash_flow_id>
```

Cash command flags are code-defined in `cmd/cash_flow_cmd/`. CLI date flags generally use dashed formats such as `YYYY-MM-DD`, `YYYY-MM`, or `YYYY`, depending on the command.

## Category Commands

```bash
go run main.go category create -n "Food" -t expense
go run main.go category list --type expense
go run main.go category query --id <category_id>
go run main.go category query --name "Food"
go run main.go category tree --type expense --deep 3
go run main.go category update --id <category_id> --name "Dining"
go run main.go category delete --id <category_id>
```

Category commands support `income` and `expense` category types. The category root command also supports a persistent `--user` flag for user-scoped operations.

## Statistic Commands

```bash
go run main.go statistic summary --period monthly --date 202601 --user <user_id>
go run main.go statistic breakdown --period monthly --date 2026-01 --user <user_id>
go run main.go statistic trends --period yearly --date 2026 --user <user_id>
go run main.go statistic top --period monthly --date 2026-01 --number 10 --user <user_id>
go run main.go statistic dashboard --period monthly --date 2026-01 --user <user_id>
go run main.go statistic chart income-expense --period monthly --date 2026-01 --user <user_id>
go run main.go statistic chart category-distribution --period monthly --date 2026-01 --type expense --user <user_id>
go run main.go statistic chart monthly-comparison --year 2026 --user <user_id>
go run main.go statistic chart spending-heatmap --year 2026 --user <user_id>
go run main.go statistic export --from 20260101 --to 20260131 --output export.xlsx --user <user_id>
go run main.go statistic import --input export.xlsx --user <user_id>
```

Statistic commands are user-scoped. Period flags use the service-level values `daily`, `monthly`, and `yearly`.

## Build

```bash
go build -o cashlenx main.go
```

Build with explicit version metadata:

```bash
go build -ldflags "\
  -X github.com/macar-x/cashlenx-server/cmd/open_cmd.Version=0.6.0 \
  -X github.com/macar-x/cashlenx-server/cmd/open_cmd.BuildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ) \
  -X github.com/macar-x/cashlenx-server/cmd/open_cmd.GitCommit=$(git rev-parse --short HEAD)" \
  -o cashlenx main.go
```

## Notes

- The API server command is `open start`, not `server start`.
- Default local development uses MongoDB, but the project still supports MongoDB and MySQL configuration.
- API route documentation lives in `docs/openapi.yaml`; this file covers CLI behavior only.
- If this document and command help disagree, trust the code in `cmd/` first and update this document in the same change set.
