# Runtime Dependencies

MongoDB and MySQL are separate Compose projects for local and operator-managed
environments. Each dependency owns its Compose file, initialization assets,
documentation, container, and named data volume under its own folder. All
projects use explicit project/container names and attach to the absolute shared
network selected by `DOCKER_NETWORK_NAME`.

The root `scripts/build.sh`, `scripts/start.sh`, and `scripts/stop.sh` manage only
the API container. They never inspect `DB_TYPE` to start or stop a database.
Select and manage the required dependency explicitly:

```bash
scripts/dependencies/mongodb/build.sh
scripts/dependencies/mongodb/start.sh
scripts/dependencies/mongodb/stop.sh

scripts/dependencies/mysql/build.sh
scripts/dependencies/mysql/start.sh
scripts/dependencies/mysql/stop.sh
```

Dependency `build.sh` pulls the configured upstream image, `start.sh` starts
only that dependency and waits for its healthcheck, and `stop.sh` removes its
container without removing the image or named data volume. Any start creates
the shared network when absent; stop removes it only after the final attached
container is gone.
Each project uses its existing named volume by default. An absolute
`MONGO_DATA_PATH` or `MYSQL_DATA_PATH` selects a host bind mount instead, while
the matching `*_DATA_VOLUME_NAME` changes the Docker volume identity when the
path is empty. Stop does not remove either storage form.
No database enable flag is used: script selection owns lifecycle, while the
selected environment file supplies that dependency's settings. `TIMEZONE` is
shared by the API and both database container projects. Dependency starts accept
`UTC` or the region-based IANA name shape and reject fixed offsets,
abbreviations, and `Etc/GMT` forms before creating a container.

All dependency scripts use `.env` by default and accept the same repository-local
selection interface as the API scripts:

```bash
ENV_FILE=.env.testing scripts/dependencies/mongodb/build.sh
ENV_FILE=.env.testing scripts/dependencies/mongodb/start.sh
ENV_FILE=.env.testing scripts/dependencies/mongodb/stop.sh
```

Compose remains available directly for diagnostics through
`mongodb/compose.yml` and `mysql/compose.yml`, but the scripts are the supported
lifecycle entry points.
