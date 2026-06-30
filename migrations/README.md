# Database Migrations

This directory contains database migration scripts for CashLenX.

## Validation

On Windows, apply every numbered SQL migration to disposable MySQL 8 and
verify the expected tables with:

```powershell
powershell -ExecutionPolicy Bypass -File scripts/smoke-mysql-migrations.ps1
```

This validates the migration sequence but does not track applied versions.
Production migration version tracking remains roadmap work.

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

### MySQL: `002` through `011`

SQL migrations `002` through `011` create the current development schema.
Always apply them in filename order. Migrations `008` through `010` are retained
as compatibility markers from the earlier development sequence; the canonical
table definitions already contain their final fields.

The disposable validation script applies only `*.sql`; the JavaScript assets
are MongoDB-specific. There is not yet an application migration runner or an
applied-version table.

## Rollback

Automated rollback is not implemented. Recreate disposable development
databases when validating the sequence, and take a verified backup before any
manual migration of persistent data.
