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
│   ├── version
│   ├── auth
│   │   ├── login
│   │   ├── register
│   │   ├── logout
│   │   ├── reset-password
│   │   └── reset-password-confirm
│   └── verification
│       ├── code
│       └── verify
├── auth
│   └── tokens
├── admin
│   ├── user
│   │   ├── create
│   │   ├── list
│   │   ├── get
│   │   ├── update
│   │   └── delete
│   └── database
│       ├── backup
│       └── restore
├── user
│   ├── profile
│   │   ├── get
│   │   └── update
│   ├── configuration
│   │   ├── get
│   │   └── upsert
│   ├── password
│   ├── email
│   │   ├── change
│   │   └── confirm
│   ├── account
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
go run main.go open verification code --purpose signup --email user@example.com
go run main.go open verification verify --purpose signup --email user@example.com --code <code>
go run main.go open auth register --username alice --password <password> --email user@example.com --verification-token <token>
go run main.go open auth login --username alice --password <password>
go run main.go open auth login --refresh-token <refresh_token>
go run main.go open auth logout
go run main.go open auth reset-password --email-or-username user@example.com
go run main.go open auth reset-password-confirm --token <token> --password <new_password>
```

- `open start` starts the REST API server.
- `open health` calls the local health endpoint.
- `open version` prints version, build time, git commit, Go version, and OS/architecture.
- `open auth login` mirrors `POST /open/auth/login` and saves the returned access token and refresh token for later CLI commands.
- `open auth logout` revokes the saved refresh token when available and clears the local CLI session.
- `open auth` and `open verification` mirror public authentication and verification endpoints.

CLI sessions are stored in the user's config directory as `cashlenx/cli_auth.json` with file mode `0600`. Set `CASHLENX_CLI_AUTH_FILE` to override the path for tests or isolated local runs.

## Auth Commands

```bash
go run main.go auth tokens
```

- `auth tokens` lists refresh tokens for the currently logged-in CLI user, matching authenticated `/auth/tokens` behavior.

## Admin Commands

```bash
go run main.go admin user create --username bob --password <password> --email bob@example.com
go run main.go admin user list --limit 20 --offset 0
go run main.go admin user get --id <user_id>
go run main.go admin user update --id <user_id> --nickname "Bob"
go run main.go admin user delete --id <user_id>
go run main.go admin database backup -o backup.json
go run main.go admin database restore -i backup.json --force
```

- `admin user` mirrors admin user CRUD endpoints. Created users are always normal `user` role accounts.
- `admin database backup` exports a full database dump.
- `admin database restore` restores a full database dump.
- Admin commands require a saved CLI session whose JWT role is `admin`, matching `/admin/*` API access checks.

## User Commands

```bash
go run main.go user profile get --user <user_id>
go run main.go user profile update --nickname "Alice" --gender female --user <user_id>
go run main.go user configuration get --user <user_id>
go run main.go user configuration upsert --currency-code USD --display-language en --active-theme-color "#2563eb" --user <user_id>
go run main.go user password --old-password <old_password> --new-password <new_password> --user <user_id>
go run main.go user email change --new-email new@example.com --verification-token <token> --user <user_id>
go run main.go user email confirm --token <token> --password <password> --user <user_id>
go run main.go user account --force --user <user_id>
go run main.go user database backup --output user_backup.json --user <user_id>
go run main.go user database restore --input user_backup.json --user <user_id>
```

User commands are user-scoped and mirror `/user/*` endpoints. They use the saved CLI session's user ID; if `--user` is supplied it must match the logged-in user.

## Cash Commands

```bash
go run main.go cash expense -c "Food" -a 45.50 -d "Lunch"
go run main.go cash income -c "Salary" -a 5000
go run main.go cash list --limit 20 --offset 0 --type expense --category-id <category_id> --from-date 2026-01-01 --to-date 2026-01-31 --user <user_id>
go run main.go cash query --id <cash_flow_id>
go run main.go cash query --date 2026-01-01
go run main.go cash range --from 2026-01-01 --to 2026-01-31
go run main.go cash summary --period monthly --date 2026-01
go run main.go cash update --id <cash_flow_id> --amount 50
go run main.go cash delete --id <cash_flow_id>
```

Cash commands use the saved CLI session's user ID; if `--user` is supplied it must match the logged-in user. `cash list` mirrors `GET /cash` filters with `--type`, `--category-id`, `--description`, `--exact-description`, `--from-date`, `--to-date`, `--limit`, `--offset`, and `--page`. CLI date flags generally use dashed formats such as `YYYY-MM-DD`, `YYYY-MM`, or `YYYY`, depending on the command.

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

Statistic commands are user-scoped and use the saved CLI session's user ID; if `--user` is supplied it must match the logged-in user. Period flags use the service-level values `daily`, `monthly`, and `yearly`.

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
