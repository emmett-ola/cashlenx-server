# CashLenX API Notes

**Version**: 0.9.0
**Last Updated**: 2026-06-30

This document is a human-readable companion to `docs/openapi.yaml`. The OpenAPI file is the detailed API contract and is used by schema validation when enabled.

## Base URL

```text
http://localhost:8080/api/v0
```

The API path version defaults to `/api/v0` and is configurable through `API_VERSION`.

Operational endpoints are mounted outside the versioned API:

- `GET /metrics` exposes Prometheus API request counters and duration histograms plus Go runtime/process metrics.
- `/debug/pprof/*` is available only when `ENV=dev`.

Operational endpoints intentionally bypass JWT and OpenAPI validation. Production deployments should restrict `/metrics` to trusted monitoring networks; pprof is not registered outside development.

## Response Shape

Most JSON endpoints respond through the shared response wrapper from `util.ComposeJSONResponse`:

```json
{
  "code": "OK",
  "message": "",
  "data": {},
  "meta": {},
  "extra": {},
  "errors": []
}
```

File download endpoints may return binary content instead of the JSON wrapper.

## Authentication

Most non-open routes require:

```text
Authorization: Bearer <access_token>
```

Route groups:

- `/open/*` is public by convention. `/open/auth/logout` is idempotent: it returns OK without credentials, revokes one session when a valid `refresh_token` is provided, and revokes all sessions when a valid bearer access token is provided without `refresh_token`.
- `/auth/tokens` is authenticated token-management API.
- `/admin/*` requires authenticated admin role.
- `/user/*`, `/cash/*`, `/category/*`, `/budget/*`, and `/statistic/*` are authenticated user-scoped APIs.

## Implemented Route Surface

System and auth:

- `GET /open/health`
- `GET /open/version`
- `POST /open/auth/login`
- `POST /open/auth/register`
- `POST /open/verification/code`
- `POST /open/verification/verify`
- `POST /open/auth/logout`
- `GET /auth/tokens`
- `POST /open/auth/reset-password`
- `POST /open/auth/reset-password/confirm`

User and admin:

- `GET /user/profile`
- `PUT /user/profile`
- `GET /user/configuration`
- `POST /user/configuration`
- `PUT /user/configuration`
- `PUT /user/password`
- `POST /user/email/change`
- `POST /user/email/confirm`
- `DELETE /user/account`
- `GET /user/database/backup`
- `POST /user/database/restore`
- `POST /admin/user`
- `GET /admin/user`
- `GET /admin/user/{id}`
- `PUT /admin/user/{id}`
- `DELETE /admin/user/{id}`
- `GET /admin/database/backup`
- `POST /admin/database/restore`

Cash flow:

- `POST /cash/expense`
- `POST /cash/income`
- `GET /cash`
- `GET /cash/range`
- `GET /cash/date/{date}`
- `DELETE /cash/date/{date}`
- `GET /cash/{id}`
- `PUT /cash/{id}`
- `DELETE /cash/{id}`
- `GET /cash/summary/total`
- `GET /cash/summary/daily/{date}`
- `GET /cash/summary/monthly/{month}`
- `GET /cash/summary/yearly/{year}`

Category:

- `POST /category`
- `GET /category`
- `GET /category/name/{name}`
- `GET /category/{parent_id}/children`
- `GET /category/tree`
- `GET /category/{id}`
- `PUT /category/{id}`
- `DELETE /category/{id}`

Budget:

- `POST /budget`
- `GET /budget?period=YYYY-MM`
- `GET /budget/{id}`
- `PUT /budget/{id}`
- `DELETE /budget/{id}`

Budgets accept only user-owned expense categories and are unique per active
user/category/month scope. Responses derive spent, remaining, and progress from
the same user's cash flows rather than storing duplicate ledger totals.

Statistic, dashboard, chart, and import/export:

- `GET /statistic/export`
- `POST /statistic/import`
- `GET /statistic/summary/daily/{date}`
- `GET /statistic/summary/monthly/{month}`
- `GET /statistic/summary/yearly/{year}`
- `GET /statistic/breakdown/daily/{date}`
- `GET /statistic/breakdown/monthly/{month}`
- `GET /statistic/breakdown/yearly/{year}`
- `GET /statistic/trends/daily/{date}`
- `GET /statistic/trends/monthly/{month}`
- `GET /statistic/trends/yearly/{year}`
- `GET /statistic/top/daily/{date}`
- `GET /statistic/top/monthly/{month}`
- `GET /statistic/top/yearly/{year}`
- `GET /statistic/dashboard/{period}/{date}`
- `GET /statistic/chart/income-expense/{period}/{date}`
- `GET /statistic/chart/category-distribution/{period}/{date}`
- `GET /statistic/chart/monthly-comparison/{year}`
- `GET /statistic/chart/spending-heatmap/{year}`

## Request Examples

Register:

```bash
curl -X POST http://localhost:8080/api/v0/open/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"test123","email":"test@example.com","verification_token":"<token>"}'
```

Login:

```bash
curl -X POST http://localhost:8080/api/v0/open/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"test123"}'
```

Refresh token:

```bash
curl -X POST http://localhost:8080/api/v0/open/auth/login \
  -H "Content-Type: application/json" \
  -d '{"refresh_token":"<refresh_token>"}'
```

Create expense:

```bash
curl -X POST http://localhost:8080/api/v0/cash/expense \
  -H "Authorization: Bearer <access_token>" \
  -H "Content-Type: application/json" \
  -d '{"belongs_date":"20260101","category_name":"Food","amount":45.5,"description":"Lunch"}'
```

Get dashboard:

```bash
curl -H "Authorization: Bearer <access_token>" \
  http://localhost:8080/api/v0/statistic/dashboard/monthly/202601
```

## Notes For Maintainers

- Keep `docs/openapi.yaml` synchronized with `controller/server.go` when changing routes.
- Keep enum examples lowercase for cash/category types: `income` and `expense`.
- Configured SMTP delivery has been manually verified. Keep automated registration and password-reset tests provider-free by replacing email delivery.
- If this document and code disagree, trust code first, then update this document and OpenAPI together.
