# mitt — bar POS hub (PC + mobile)

mitt is an open-source point of sale for small bars: a PC hub owns the
catalog, open tabs per table, the automatic tab total, day sales, and
simple expenses. An Android client (later) takes orders over the LAN;
the PC stays the source of truth.

## Quick path

1. Load the catalog (`GET /api/products`).
2. Open a tab per table (`POST /api/tabs`), add items as orders arrive.
3. Close the tab at payment (`POST /api/tabs/{id}/close`) — total is automatic.
4. Log day expenses (`POST /api/expenses`), check the dashboard at close.

## Users

| User | Goal | Primary surface |
|------|------|-----------------|
| Waiter (mozo) | Take an order in seconds, never lose a table's tab | Mobile client (later); PC as fallback |
| Manager (encargado) | Know today's sales, top products, low stock; close the day | PC dashboard |

## Core flows

| Flow | Steps | Source of truth |
|------|-------|-----------------|
| Catalog | Create / rename / reprice / toggle availability | `POST/PATCH /api/products` |
| Table → order → tab | Open tab per table → add items (price snapshotted, same-product lines merged) → close at payment | `POST /api/tabs`, `POST /api/tabs/{id}/items`, `POST /api/tabs/{id}/close` |
| Day close | Open tabs + closed sales of the day + expenses → bought-vs-sold report | Dashboard reads tabs, sales, expenses |
| Simple expenses | Log `{description, qty, cost_cents}` any time during the day | `POST /api/expenses` |

## Non-goals (phase 2, explicitly out)

- Split bill per person; split tips.
- Electronic invoicing (factura electronica).
- Suppliers / purchase orders; cloud sync / accounts / complex roles.
- Table writer endpoint (seeding `tables_tbl` is out of scope for T4).

## Platform facts

| Fact | Decision |
|------|----------|
| Language / targets | Pure Go, no CGO by default; Windows 10+ and any modern Linux |
| Storage | Local SQLite in WAL mode; money in integer cents, expense qty may be fractional |
| LAN API | `http://<pc-lan-ip>:8080`, JSON, `Authorization: Bearer <pairing-token>` (`-token` flag, `MITT_TOKEN` env, or random at startup); `GET /api/health` is the only unauthenticated route |
| Offline | Mobile is offline-first (later); PC works with no network at all |
| Architecture | Hexagonal: domain core (catalog, tables, orders, tab, expenses) depends on nothing; SQLite, LAN API, native shell are adapters — see `docs/architecture.md` |
| License / repo | Open source, adjustable; no sensitive data in repo; this task touches repo `mitt` only, never `mott` |

## Checklist

- [ ] A new waiter can take a first order without training beyond the 3-tap flow.
- [ ] A manager can close the day from the dashboard alone.
- [ ] Every flow works with the LAN cable unplugged (PC local) and syncs to mobile over LAN when present.
- [ ] No phase-2 feature leaks into MVP scope without an explicit scope change.

## Next step

Shared visual language lives in `docs/design-tokens.md`; PC dashboard
structure lives in `docs/dashboard.md`.
