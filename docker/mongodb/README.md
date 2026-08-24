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

## Index Lifecycle

`init-mongo.js`, `migrations/001_add_indexes.js`, and the runtime index manager
use the current user/type/parent/name category scope. The server invokes the
runtime manager at startup and removes known obsolete `flow_type` and overly
broad category-name indexes after current replacements are created.

## Demo Script Status

`init-mongo-demo.js` is retained as a legacy fixture only. Do not load it into
the current schema: it expects global categories, writes obsolete `flow_type`
fields, stores dates as strings, and does not set user ownership or audit
metadata. Use the managed API smoke flow to create isolated, user-scoped test
data instead.

```bash
# Starts disposable MongoDB plus the API and runs the backend smoke flow.
test/scripts/api-smoke.sh --managed
```

The sibling Flutter client currently has no maintained live API harness; use
the test environment for complete app journeys.

## Related Files

- Compose configuration: `../../compose.yml`
- Default categories: `../../config/default_categories.json`
- Current models: `../../model/`
- MongoDB mappers: `../../mapper/`
- Migration notes: `../../migrations/README.md`
- MySQL initialization guide: `../mysql/README.md`
