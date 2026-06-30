# Database Migrations

This directory contains database migration scripts for CashLenX.

## Validation

On Windows, apply every numbered SQL migration to disposable MySQL 8 and
verify the expected tables with:

```powershell
powershell -ExecutionPolicy Bypass -File scripts/smoke-mysql-migrations.ps1
```

This validates the migration sequence independently. The application runner
tracks versions, filenames, checksums, dirty state, and timestamps in the
MySQL `schema_migrations` table.

At startup, an existing complete pre-runner schema is baselined through version
`011`; a partial schema is rejected. Empty schemas apply all SQL migrations in
order. Failed migrations remain dirty and require explicit repair or restore.

Migrations that provide a matching `.down.sql` file are compensated
automatically when an up statement fails. If any down statement fails, the
dirty row is retained and startup remains blocked for explicit repair.

The Docker bootstrap files under `docker/` are fresh-install snapshots, not an
applied migration history. Changing them does not update an existing database.

## Available Assets

### MongoDB: `001_add_indexes.js`

This script reconciles legacy index names with current multi-user cash-flow and
category query patterns. It removes obsolete `flow_type` indexes and replaces
the broad category-name constraint with a partial unique index scoped by user,
type, parent, and active records.

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

The disposable validation script applies only `*.sql`; the JavaScript assets
are MongoDB-specific.

## Rollback

Automated rollback is not implemented. Recreate disposable development
databases when validating the sequence, and take a verified backup before any
manual migration of persistent data.
