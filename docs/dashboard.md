# PC dashboard IA (Operate mode)

One screen answers "how is the bar right now": open tables, today's
money, what sells, what runs out. Scanable from 2 meters in a dark bar.
Mode: Operate — task completion and scanability beat expression.

## Quick path

Layout top-to-bottom, left-to-right: open tables → day sales → top
products → low-stock alerts. Big numerals for money, color + label for
every status, empty/error state per section. No UI code in this doc —
structure and data contracts only (T6 builds the shell).

## Sections

| # | Section | Purpose | Data source |
|---|---------|---------|-------------|
| 1 | Open tables | See every occupied table + running total at a glance | `GET /api/tabs/open` (oldest first); per-tab items already in response |
| 2 | Day sales | Today's closed revenue + ticket count | Closed sales of the day (close endpoint: `POST /api/tabs/{id}/close`); day-close report (TBD, T6+) |
| 3 | Top products | What sells most today | Aggregate of today's closed-tab items by `product_id` |
| 4 | Low-stock alerts | What is unavailable / running out before the rush | `GET /api/products?available=false` + manual low-stock flag (TBD) |

## Section contracts

### 1. Open tables

- Tile per open tab: table name (`title`), running total (`display` + `mono-numeric`), elapsed time (`caption`).
- Occupied = `surface-raised` + `accent` edge + "Occupied · $ total" label; free tables show `text-muted` "Free".
- Tab open > 90 min adds a `warn` "Open HH:MM" badge.
- Empty: "No open tables — open one from Tables." Error: "Could not load tabs. Retry." + retry button.

### 2. Day sales

- One `display` numeral: today's total; below it ticket count + average ticket (`body`).
- Empty (no sales yet): "$ 0 — first close appears here." Error: same retry pattern as §1.

### 3. Top products

- Ranked list (max 5): name, units sold, revenue (`mono-numeric`).
- Empty: "No sales yet today." Not an error — never show a red state for zero sales.

### 4. Low-stock alerts

- List of unavailable products with `danger` + "Out of stock" chip; low (not out) uses `warn` + "Low" chip.
- Empty (all good): "All products available." with `success` label — reassure, don't blank.
- Error: "Could not load catalog. Retry."

## Readability rules

- Big numerals: totals in `display`, never smaller; tabular figures so digits don't jitter on update.
- Color + label, never color-only (see state tokens in `docs/design-tokens.md`).
- Refresh: poll open tabs every 15–30 s; sales/products/alerts on close-tab or catalog event. Show "Updated HH:MM:SS" (`caption`, `text-muted`).
- Night-bar-first: dark theme default; no pure white surfaces; no thin gray-on-dark body text below AA.

## Mobile order flow sketch (stays compaginado)

Max 3 taps per order; same tokens, same state labels as PC.

| Tap | Screen | Content |
|-----|--------|---------|
| 1 | Tables | Grid of tables: occupied (`accent` edge + total) vs free (`text-muted`); big 48dp targets |
| 2 | Products | Available products with price (`mono-numeric`); out-of-stock rows disabled + "Out of stock" chip |
| 3 | Confirm | Order summary (lines + total in `display`), one Confirm button; success shows `success` "Sent" |

Shared rules: same hex, same type scale, same state labels; mobile uses
48dp targets where PC uses 32px dense; totals always `mono-numeric`.

## Checklist

- [ ] Each section names its API source and its empty + error state.
- [ ] No section relies on color alone.
- [ ] Money is always big + tabular; timestamps always `caption` + muted.
- [ ] Mobile flow stays ≤ 3 taps and reuses PC tokens/states verbatim.

## Next step

T6 native shell implements this IA against the tokens; new endpoints
(day sales aggregate, low-stock flag) get specced there, not here.
