# Container Build Contract

`images.env` is the tracked authority for base-image references. References are
pinned by digest so clean and warm builds use the same inputs. Update one pin in
an isolated change, build and verify the candidate image, and roll back by
reverting that change.

The root build context is allowlisted by `.dockerignore`. Environment files,
credentials, Git state, logs, test output, database data, and unrelated files
are never sent to the builder.

Run `scripts/build.sh` to derive the source version and full revision, create a
deterministic commit timestamp, build from `go.sum`, and verify the executable,
OpenAPI contract, default-category data, OCI labels, embedded version output,
and prohibited file absence.
