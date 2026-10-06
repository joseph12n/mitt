# LAN HTTP API

Base URL: `http://<pc-lan-ip>:8080`. Every request except `GET /api/health`
needs `Authorization: Bearer <pairing-token>`. The token comes from the
`-token` flag, the `MITT_TOKEN` env var, or a random value printed once at
startup. Every response (including errors) is `application/json` with the
envelope `{"error":{"code","message"}}` on failures.

Money is in integer cents; expense `qty` may be fractional.

| Method | Path | Auth | Body | Responses |
|---|---|---|---|---|
| GET | /api/health | no | — | 200 `{status}` |
| GET | /api/products | yes | —, `?available=true\|false` optional | 200 `{products}`, 400 bad filter |
| POST | /api/products | yes | `{name, price_cents>=0, available?}` (available defaults true) | 201 product, 400 bad JSON, 422 domain error |
| PATCH | /api/products/{id} | yes | `{name?, price_cents?, available?}` | 200 product, 404 unknown id, 422 domain error |
| POST | /api/tabs | yes | `{table_id}` non-empty; idempotent per open table | 201 tab (200 when already open), 422 empty label |
| GET | /api/tabs/open | yes | — | 200 `{tabs}` oldest first |
| POST | /api/tabs/{id}/items | yes | `{product_id, qty>0}`; price snapshotted, same-product lines merged | 200 tab, 404 unknown tab/product, 422 unavailable or invalid qty |
| POST | /api/tabs/{id}/close | yes | — | 200 sale `{id, table_id, items, total_cents, closed_at}`, 404 unknown tab, 422 empty tab |
| GET | /api/expenses | yes | — | 200 `{expenses}` oldest first |
| POST | /api/expenses | yes | `{description, qty>0, cost_cents>=0}` | 201 expense, 400 bad JSON, 422 domain error |

Notes:

- Closing an already-closed tab returns 404 (only open tabs are addressable).
- `422` always carries the domain validation message for debugging.
- Unknown paths return JSON `404`; wrong-method use returns the stdlib `405`.
