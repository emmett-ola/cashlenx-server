# MongoDB Dependency

This directory owns the MongoDB Compose project and bootstrap assets. MongoDB is
the default development and beta deployment backend.

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

Prepare and start MongoDB from the repository root with:

```bash
scripts/dependencies/mongodb/build.sh
scripts/dependencies/mongodb/start.sh
```

Stop its container while retaining `cashlenx-mongodb-data` with
`scripts/dependencies/mongodb/stop.sh`. These scripts never start or stop the
API container. Start attaches MongoDB to `DOCKER_NETWORK_NAME`; stop removes the
shared network only if no other container remains attached.

Leave `MONGO_DATA_PATH` empty to use the named volume configured by
`MONGO_DATA_VOLUME_NAME`. Set it to an absolute host path to bind that directory
to `/data/db`. Changing the source does not copy existing records; the previous
volume or directory remains untouched and must be migrated explicitly when its
data should follow the new source.

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

- Compose configuration: `compose.yml`
- Lifecycle scripts: `../../../scripts/dependencies/mongodb/`
- Default categories: `../../../config/default_categories.json`
- Current models: `../../../model/`
- MongoDB mappers: `../../../mapper/`
- Migration notes: `../../../migrations/README.md`
- MySQL initialization guide: `../mysql/README.md`
