# MongoDB Initialization

This directory contains the MongoDB bootstrap assets used by `compose.yml`.
MongoDB is the default development and beta deployment backend.

## Canonical Bootstrap

`init-mongo.js` is mounted under `/docker-entrypoint-initdb.d/` and runs only
when MongoDB initializes a new data directory. It currently:

- selects the database from `MONGO_INITDB_DATABASE`, which Compose derives from
  `DB_NAME`;
- creates the `users`, `refresh_tokens`, `operation_confirm_codes`,
  `cash_flows`, and `categories` collections;
- creates bootstrap indexes;
- leaves user and transaction data empty.

Default categories are not global database seed data. The application creates
a user-scoped copy from `config/default_categories.json` whenever a user is
created. The bootstrap admin is created by the server on first startup.

Start a fresh MongoDB container with:

```bash
docker compose --profile mongodb up -d mongodb
```

Changing `init-mongo.js` does not update an existing data directory. Recreate
only disposable development data, or apply a reviewed migration to persistent
data.

## Current Schema Shape

Cash flows are user-scoped and store `belongs_user_id`, `category_id`,
`belongs_date` as a BSON datetime, `amount`, `description`, `remark`, and the
soft-delete/audit fields from `BaseEntity`. Flow type is derived from the
referenced category; `flow_type` is no longer stored on current cash-flow
documents.

Categories are user-scoped and store `belongs_user_id`, `parent_id`, `name`,
`type`, `remark`, and soft-delete/audit fields. The service permits a category
name to repeat when its type or parent differs, so the intended uniqueness key
is user + type + parent + name.

## Known Bootstrap Drift

The current `init-mongo.js` still creates obsolete `flow_type` indexes and a
unique `(belongs_user_id, name)` category index. That category index is stricter
than the service contract and can reject otherwise valid categories. Reconcile
the bootstrap indexes with the mapper queries and service uniqueness rules
before treating them as production migration definitions.

`migrations/001_add_indexes.js` has related legacy index definitions and must
also not be applied as-is. Index lifecycle reconciliation is tracked in the
`v0.8.0` roadmap.

## Demo Script Status

`init-mongo-demo.js` is retained as a legacy fixture only. Do not load it into
the current schema: it expects global categories, writes obsolete `flow_type`
fields, stores dates as strings, and does not set user ownership or audit
metadata. Use the managed API smoke flow or the sibling Flutter smoke flow to
create isolated, user-scoped test data instead.

```bash
# Starts disposable MongoDB plus the API and runs the backend smoke flow.
scripts/smoke-api.sh --managed
```

From `../cashlenx-app`, `scripts/smoke-api.ps1` runs the broader Flutter/API
contract against disposable infrastructure.

## Related Files

- Compose configuration: `../../compose.yml`
- Default categories: `../../config/default_categories.json`
- Current models: `../../model/`
- MongoDB mappers: `../../mapper/`
- Migration notes: `../../migrations/README.md`
- MySQL initialization guide: `../mysql/README.md`
