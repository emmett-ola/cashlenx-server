# Contributing to CashLenX

Thank you for helping improve CashLenX. The project is split across repositories
with explicit ownership boundaries so that each area remains independently
buildable and reviewable.

## Choose the Owning Repository

| Repository | Responsibility |
| --- | --- |
| [cashlenx-app](https://github.com/emmett-ola/cashlenx-app) | Flutter client and user experience. |
| [cashlenx-server](https://github.com/emmett-ola/cashlenx-server) | Go API, CLI, authentication, finance services, and persistence. |
| [cashlenx-design](https://github.com/emmett-ola/cashlenx-design) | Visual and interaction reference. |
| [cashlenx-website](https://github.com/emmett-ola/cashlenx-website) | Public product and developer-information website. |
| [cashlenx-spec](https://github.com/emmett-ola/cashlenx-spec) | Product/system facts, governance, workflow, decisions, and delivery evidence. |

Open an issue or pull request in the repository that owns the behavior. Describe
cross-repository impact explicitly instead of coupling repositories at build
time or runtime.

## Contribution Workflow

1. Read the repository README and nearest `AGENTS.md` when present.
2. State the user or maintainer outcome and the observable done condition.
3. Confirm current behavior from code, contracts, tests, and documentation.
4. Make the smallest complete change and preserve compatibility unless the
   change intentionally defines a new boundary.
5. Add focused validation for the behavior and its first relevant failure,
   authorization, or repeat path.
6. Update OpenAPI, user documentation, or CashLenX Spec facts when the external
   contract or verified behavior changes.
7. Open a pull request with the scope, validation evidence, known limits, and
   any migration, security, or cross-repository implications.

Do not include credentials, private environment files, personal data, production
logs, or unsanitized telemetry in issues, commits, tests, or pull requests.

## Pull Request Expectations

A reviewable pull request should include:

- the problem and intended outcome;
- the owning repository and any affected repositories;
- tests or other reproducible validation;
- API, data, migration, security, and deployment impact;
- screenshots for visible UI changes;
- known limitations or intentionally deferred work.

AI-assisted contributions are welcome. The contributor remains responsible for
understanding the change, reviewing generated material, protecting sensitive
data, and providing evidence that the result is correct.

## License

Unless explicitly identified as third-party material under different terms,
contributions are submitted under the repository's [MIT License](LICENSE).
