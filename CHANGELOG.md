# Changelog

All notable changes to CashLenX Server are recorded here. Entries describe
released contracts and operational behavior separately from deployment state.

## [Unreleased]

### Changed

- Added exact candidate image references to package metadata and made API
  container start fail instead of pulling a missing configured image.
- Unified the API, MongoDB, and MySQL lifecycle entry points behind a
  repository-local Docker Compose and nerdctl 2.2 portability layer with
  pre-mutation validation, deterministic configured image identity, value-safe
  start output, and cold-database readiness handling.

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
