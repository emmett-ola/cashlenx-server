# Optional Runtime Dependencies

MongoDB and MySQL are independent Compose projects for local and operator-managed environments. The CashLenX server build and startup scripts do not start, stop, or remove them.

From the server repository root, use the dependency required by `DB_TYPE`:

```bash
docker compose --env-file .env -f docker/dependencies/compose.mongodb.yml up -d --wait
docker compose --env-file .env -f docker/dependencies/compose.mysql.yml up -d --wait
```

Inspect or stop a dependency with the same `--env-file` and `-f` arguments followed by `ps`, `stop`, or `down`. `down` preserves its named data volume unless `--volumes` is explicitly supplied.
