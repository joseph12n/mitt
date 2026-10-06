# mitt

PC hub for a small bar: catalog, tables, per-table tab, day dashboard,
simple expenses, and a LAN API for offline-first mobile clients.

No secrets are stored in this repo. Local data stays in SQLite files,
which are git-ignored.

## Quick path

Requirements: Go 1.22+ (pure Go, no CGO needed for this baseline).

```bash
go run ./cmd/mitt
```

Build for Linux:

```bash
make build-linux
```

Build for Windows (from Linux or Windows):

```bash
make build-windows
```

Run checks:

```bash
make vet
make test
```

## Project layout

| Path                | Purpose                                  |
|---------------------|------------------------------------------|
| `cmd/mitt/`         | PC hub entrypoint (wiring only).         |
| `docs/architecture.md` | Hexagonal intent and platform notes.  |
| `odd/tasks/`        | Feature plans and task tracking.         |

Core domain and adapters land in later tasks (T2+); this baseline
only sets up the module, entrypoint, and docs.

## Next step

See `docs/architecture.md` for the hexagonal intent and T2 domain skeleton.
