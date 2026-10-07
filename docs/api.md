# LAN HTTP API

Base URL: `http://<pc-lan-ip>:8080`. Every request except `GET /api/health`,
`GET /api/branding`, and `GET /api/branding/logo` needs
`Authorization: Bearer <pairing-token>`. The token comes from the
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
| POST | /api/tabs | yes | `{table_id}` must reference a table created first via `POST /api/tables`; idempotent per open table | 201 tab (200 when already open), 422 empty label or unknown_table |
| GET | /api/tables | yes | — | 200 `{tables:[{id, label, occupied}]}` with occupied derived from open tabs |
| POST | /api/tables | yes | `{label}` trimmed, 1..40 chars | 201 table, 400 bad JSON, 422 domain error |
| DELETE | /api/tables/{id} | yes | — | 204, 404 unknown id, 409 table_occupied while an open tab exists |
| GET | /api/tabs/open | yes | — | 200 `{tabs}` oldest first |
| POST | /api/tabs/{id}/items | yes | `{product_id, qty>0}`; price snapshotted, same-product lines merged | 200 tab, 404 unknown tab/product, 422 unavailable or invalid qty |
| POST | /api/tabs/{id}/close | yes | — | 200 sale `{id, table_id, items, total_cents, closed_at}`, 404 unknown tab, 422 empty tab |
| GET | /api/expenses | yes | — | 200 `{expenses}` oldest first |
| POST | /api/expenses | yes | `{description, qty>0, cost_cents>=0}` | 201 expense, 400 bad JSON, 422 domain error |
| GET | /api/suppliers | yes | — | 200 `{suppliers}` ordered by name |
| POST | /api/suppliers | yes | `{name 1..80 chars trimmed, phone? ≤40 chars, note? ≤200 chars}` | 201 supplier, 400 bad JSON, 422 domain error |
| PATCH | /api/suppliers/{id} | yes | `{name?, phone?, note?}` merged onto stored | 200 supplier, 404 unknown id, 422 domain error |
| DELETE | /api/suppliers/{id} | yes | — | 204, 404 unknown id |
| GET | /api/sales | yes | —, `?limit=` optional (default 50, capped at 500) | 200 `{sales}` newest first, same `{id, table_id, items, total_cents, closed_at}` shape as closing a tab, 400 bad limit |
| GET | /api/sales/today | yes | — | 200 `{date YYYY-MM-DD, count, total_cents}` |
| GET | /api/pairing | yes | — | 200 `{url, pairing_code}`, 503 no LAN address |
| GET | /api/branding | no | — | 200 `{shop_name, primary, accent, background, has_logo, updated_at}` |
| GET | /api/branding/logo | no | — | 200 raw logo bytes (`Content-Type` = stored mime, immutable + ETag), 404 no logo |
| PATCH | /api/branding | yes | `{shop_name?, primary?, accent?, background? (#rrggbb), logo_data_url? ("data:<mime>;base64,..." or null to clear)}` merged onto stored | 200 branding, 400 bad JSON, 413 logo over 512KiB, 422 domain error |

Notes:

- The hub owns the tables: `POST /api/tabs` with an unknown `table_id`
  returns `422 unknown_table`. Clients must create the table first via
  `POST /api/tables`; previously any non-empty string was accepted. Table
  `occupied` is derived from open tabs at read time, never stored.
- Closing an already-closed tab returns 404 (only open tabs are addressable).
- `422` always carries the domain validation message for debugging.
- Unknown paths return JSON `404`; wrong-method use returns the stdlib `405`.

Human dashboard: `GET /` serves the embedded staff page with no auth and
no JSON envelope. It is not part of this JSON API; every route above keeps
its pairing-token contract unchanged.
