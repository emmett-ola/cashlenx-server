# Changelog

All notable changes to CashLenX Server are recorded here. Entries describe
released contracts and operational behavior separately from deployment state.

## [Unreleased]

## [1.0.3] - 2026-10-09

### Changed

- Aligned the Server runtime and OpenAPI version with the coordinated CashLenX
  1.0.3 brand release. This release introduces no Server behavior, API, schema,
  migration, or database change.

- Upgraded CI and the Server toolchain to Go 1.27.1, refreshed the locked
  module graph within the existing public module APIs, and advanced the
  minimal runtime to digest-pinned Alpine 3.23.
- Added exact candidate image references to package metadata and made API
  container start fail instead of pulling a missing configured image.
- Unified the API, MongoDB, and MySQL lifecycle entry points behind a
  repository-local Docker Compose and nerdctl 2.2 portability layer with
  pre-mutation validation, deterministic configured image identity, value-safe
  start output, and cold-database readiness handling.
- Added API, MongoDB, and MySQL status/doctor/log entry points, selected-database
  state reporting, effective image verification, and bounded observable stop
  outcomes including forced and repeated stops.
- Pinned MongoDB 7.0.43 and MySQL 8.0.46 by immutable digest, added exact image
  verification, rejected unsafe MongoDB shared filesystems before
  initialization, and made MongoDB readiness wait for the final daemon.

## [1.0.0-rc.1] - 2026-09-16

### Added

- Stable `/api/v1` routing with a frozen `/api/v0` compatibility alias.
- Username-or-normalized-email authentication and normalized unique email
  management.
- Durable MongoDB migration identity, encrypted database backup, and disposable
  restore-drill tooling.
- Traceable container packaging with semantic version, exact source revision,
  input-set digest, image identity, and SHA-256 artifact sidecars.

### Security

- Production startup validation, exact HTTPS CORS allowlists, request-rate
  controls, protected metrics, and secret-safe container contexts.
