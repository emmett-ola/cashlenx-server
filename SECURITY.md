# Security Policy

CashLenX is currently a pre-production `v0.x` project. Security reports are
welcome throughout development, especially when they affect authentication,
authorization, user-data isolation, secrets, API validation, database
migrations, backup/restore, imports/exports, container boundaries, or
dependencies.

## Reporting a Vulnerability

Use the repository's private vulnerability-reporting or Security Advisory
feature when it is available. Include:

- the affected repository, branch, commit, route, or component;
- the impact and realistic attack conditions;
- minimal reproduction steps or a proof of concept;
- suggested mitigations, if known.

Do not open a public issue containing exploit details, credentials, personal
data, private URLs, or production logs. If private reporting is unavailable,
open a minimal issue asking the maintainer to establish a private channel and
omit all sensitive details.

## Handling Reports

The maintainer will validate the report, identify affected repositories, and
coordinate a fix and disclosure appropriate to the project's pre-production
state. Reports may require coordinated changes to the server, app, contracts,
deployment configuration, and specification evidence.

Only supported project code and maintained dependencies are in scope.
Third-party services and infrastructure should be reported to their respective
owners unless the vulnerability is caused by CashLenX integration behavior.
