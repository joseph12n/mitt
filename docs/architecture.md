# Architecture

Intent: hexagonal. The core domain (catalog, tables, orders, tab,
expenses) stays independent of storage, transport, and UI. Adapters
(SQLite, LAN HTTP API, native shell) plug in around it.

## Core vs adapters

| Layer    | Holds                          | Depends on          |
|----------|--------------------------------|---------------------|
| Core     | Domain types, tab rules, reports | Nothing (stdlib only) |
| Adapters | SQLite WAL store, LAN API, shell | Core interfaces     |

Dependency direction points inward: adapters know the core, the core
knows no adapter.

## Platform notes

- Targets: Windows 10+ and any modern Linux distro.
- Pure Go with no CGO by default, so `GOOS=windows` and `GOOS=linux`
  cross-compiles work from either host.
- SQLite WAL storage arrives in T3 via a pure-Go driver; no CGO
  dependency is introduced.
- LAN API (T4) serves mobile clients on the local network; pairing
  keeps onboarding simple without cloud accounts.

## UI seam

- The `ui.DashboardService` port (`internal/ui`) is the only contract UI
  shells program against: one `Snapshot` returning open tables, today's
  sales, catalog size, and out-of-stock names.
- The stdlib text dashboard (`internal/tui`, `text/tabwriter`, `-ui
  snapshot`) proves the seam with zero dependencies: it renders OPEN
  TABLES / TODAY / CATALOG from the port and exits.
- The graphical shell decision (Fyne vs Wails) stays open and plugs
  behind `ui.DashboardService` — no store or API imports in shell code.
- Known gap: today's sales read 0 until a store day-sales query lands
  (see the TODO on `ui.DashboardData`); the store surface stays unchanged
  until that storage task.
