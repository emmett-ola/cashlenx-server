# Container Build Contract

`images.env` is the tracked authority for base-image references. References are
pinned by digest so clean and warm builds use the same inputs. Update one pin in
an isolated change, build and verify the candidate image, and roll back by
reverting that change.
The same file pins Go 1.23.12. The Dockerfile verifies the compiler identity and
uses a read-only module graph before building; CI uses the identical toolchain.

The root build context is allowlisted by `.dockerignore`. Environment files,
credentials, Git state, logs, test output, database data, and unrelated files
are never sent to the builder.

Run `scripts/build.sh` to derive the source version and full revision, create a
deterministic commit timestamp, build from `go.sum`, and verify the executable,
OpenAPI contract, default-category data, OCI labels, embedded version output,
and prohibited file absence.

The repository-local lifecycle helper supports Docker Compose v2 and nerdctl
2.2+, validates the selected frontend before mutation, and derives the Server
image reference from `SERVER_IMAGE_NAME` plus `SERVER_IMAGE_TAG` without
`config --images`. The same helper serves the independent MongoDB and MySQL
projects.
