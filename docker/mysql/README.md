# MySQL Initialization Files

## Canonical Bootstrap

`init-mysql.sql` is mounted by Docker Compose and runs only when a new MySQL
data volume is initialized. It creates the current tables and indexes for:

- users and user configurations;
- refresh tokens and operation confirmation codes;
- user-scoped categories;
- user-scoped cash flows.

It does not insert categories or transactions. Default categories are created
per user by the application from `config/default_categories.json`.

Start a fresh MySQL 8 container with:

```bash
docker compose --profile mysql up -d mysql
```

Changing an initialization SQL file does not update an existing Docker volume.
Use the numbered migrations for deliberate schema work, or recreate a
disposable development volume.

## Numbered Migration Validation

The current MySQL development schema is also represented by the SQL migrations
under `../../migrations/`. On Windows, apply the complete sequence to a
disposable MySQL 8 container with:

```powershell
powershell -ExecutionPolicy Bypass -File test/scripts/mysql-migrations-smoke.ps1
```

This verifies the files in order. Server startup also runs the embedded SQL
sequence and records checksums and dirty/applied state in `schema_migrations`.

## Legacy Files

- `init-mysql-schema.sql` is retained for historical comparison only.
- `init-mysql-demo.sql` targets an older schema and must not be loaded into the
  current database. Use the API smoke flow to create isolated test data.

The current cash-flow schema has no `flow_type` column. Income and expense are
derived from the linked category's `type`, and all category/cash-flow queries
must remain scoped by `belongs_user_id`.

## Integration Smoke Test

Run the focused server budget contract against MySQL:

```powershell
powershell -ExecutionPolicy Bypass -File test/scripts/budget-smoke.ps1 -Database mysql
```

The script uses disposable infrastructure and cleans up its test data.

## Related Files

- Docker Compose: `../../compose.yml`
- Current models: `../../model/`
- MySQL mappers: `../../mapper/`
- Numbered migrations: `../../migrations/`
