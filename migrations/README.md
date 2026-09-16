# Database Migrations

This directory contains database migration scripts for CashLenX.

## MongoDB runner

MongoDB startup loads the ordered `*.js` assets, verifies each immutable
SHA-256 checksum, and records applied state in the application database's
`schema_migrations` collection. The JavaScript files are the durable migration
identity; matching native Go handlers execute them without requiring `mongosh`
inside the API image.

The first startup of an installation without a ledger runs every known
idempotent migration against the existing database and records the result. This
is the baseline path for both fresh and existing installations. Existing user
data is not deleted or replaced. An incompatible index or data condition leaves
the current migration dirty and blocks startup.

Adding a lower version after a higher version was applied, renaming or changing
an applied asset, an unknown ledger version, and a dirty record all fail closed.
Repeated startup verifies history and performs no migration work.

After startup, operators can run the read-only verification command with the
same environment configuration as the API:

```bash
cashlenx migration verify
```

The command exits non-zero for pending, dirty, unknown, reordered, renamed, or
checksum-mismatched history. It never repairs or edits migration state.

Validate fresh, existing, repeat, checksum-change, reordered, handler, and
failed/dirty behavior against disposable MongoDB 7 with:

```powershell
powershell -ExecutionPolicy Bypass -File test/scripts/mongodb-migrations-smoke.ps1
```

The application-level JSON backup/restore flow owns user data, not schema
history. A deployment backup of the selected MongoDB database or volume must
include `schema_migrations`. Recovery restores a verified database backup, or
repairs the incompatible data/index condition under an approved migration plan;
manual ledger edits are not a routine recovery mechanism.

## MySQL validation

On Windows, apply every numbered SQL migration to disposable MySQL 8 and
verify the expected tables with:

```powershell
powershell -ExecutionPolicy Bypass -File test/scripts/mysql-migrations-smoke.ps1
```

This validates the migration sequence independently. The application runner
tracks versions, filenames, checksums, dirty state, and timestamps in the
MySQL `schema_migrations` table.

The runner itself has a build-tagged integration test for clean application,
compensating rollback, and out-of-order history rejection. Run it against a
disposable MySQL database with:

```bash
MYSQL_TEST_DSN='user:password@tcp(localhost:3306)/cashlenx?parseTime=true' \
  go test -tags=integration -run TestMySQLMigrationRunnerIntegration -v ./migrations
```

At startup, an existing complete pre-runner schema is baselined through version
`011`; a partial schema is rejected. Empty schemas apply all SQL migrations in
order. Failed migrations remain dirty and require explicit repair or restore.

Migrations that provide a matching `.down.sql` file are compensated
automatically when an up statement fails. If any down statement fails, the
dirty row is retained and startup remains blocked for explicit repair.

The Docker bootstrap files under `docker/` are fresh-install snapshots, not an
applied migration history. Changing them does not update an existing database.

## Available Assets

### MongoDB: `001`, `010`, and `016`

The ordered assets reconcile legacy cash-flow/category indexes, verification
code indexes, user-configuration uniqueness, and budget indexes. Handlers are
idempotent so they can safely baseline an existing compatible installation.

## Migration Guidelines

1. **Always backup** your database before running migrations
2. **Test migrations** on a copy of production data first
3. **Run migrations** during low-traffic periods
4. **Monitor performance** after migration
5. **Have a rollback plan** ready

### MySQL: `002` through `012`

SQL migrations `002` through `011` create the base development schema, and
`012` reconciles active category uniqueness with type, parent, and soft-delete behavior.
Always apply them in filename order. Migrations `008` through `010` are retained
as compatibility markers from the earlier development sequence; the canonical
table definitions already contain their final fields.

The disposable validation script applies upgrade `*.sql` files and explicitly
excludes `*.down.sql`; the JavaScript assets are MongoDB-specific.

## Rollback

The application runner automatically compensates a failed migration when that
migration has a matching `.down.sql` asset. If compensation is unavailable or
fails, the migration stays dirty and startup remains blocked. Recreate
disposable development databases when validating the sequence, and take a
verified backup before any manual migration of persistent data.
