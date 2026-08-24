# MySQL Dependency

This directory owns the MySQL Compose project and bootstrap assets.

## Canonical Bootstrap

`init-mysql.sql` is mounted by Docker Compose and runs only when a new MySQL
data volume is initialized. The official image executes it in `MYSQL_DATABASE`,
which Compose derives from `DB_NAME`; the SQL does not repeat a database name.
It creates the current tables and indexes for:

- users and user configurations;
- refresh tokens and operation confirmation codes;
- user-scoped categories;
- user-scoped cash flows.

It does not insert categories or transactions. Default categories are created
per user by the application from `config/default_categories.json`.

Prepare and start MySQL 8 from the repository root with:

```bash
scripts/dependencies/mysql/build.sh
scripts/dependencies/mysql/start.sh
```

Stop its container and network while retaining `cashlenx-mysql-data` with
`scripts/dependencies/mysql/stop.sh`. These scripts never start or stop the API
container.

Leave `MYSQL_DATA_PATH` empty to use the named volume configured by
`MYSQL_DATA_VOLUME_NAME`. Set it to an absolute host path to bind that directory
to `/var/lib/mysql`. Changing the source does not copy existing records; the
previous volume or directory remains untouched and must be migrated explicitly
when its data should follow the new source.

Changing an initialization SQL file does not update an existing Docker volume.
Use the numbered migrations for deliberate schema work, or recreate a
disposable development volume.

## Numbered Migration Validation

The current MySQL development schema is also represented by the SQL migrations
under `../../../migrations/`. On Windows, apply the complete sequence to a
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

- Docker Compose: `compose.yml`
- Lifecycle scripts: `../../../scripts/dependencies/mysql/`
- Current models: `../../../model/`
- MySQL mappers: `../../../mapper/`
- Numbered migrations: `../../../migrations/`
