# Shared design tokens (PC + mobile)

One token set feeds every future theme (Fyne/web/Kotlin). Copy 1:1;
never restyle per platform. Dark is the default (night-bar-first);
light exists for daylight admin.

## Quick path

1. Take the dark column as default theme values.
2. Apply typography + spacing + radius + touch tables verbatim.
3. Use state tokens (never bare color) for availability and table status.

## Colors

| Role | Dark (default) | Light | Usage |
|------|---------------|-------|-------|
| `bg` | `#121417` | `#F5F3EC` | App background |
| `surface` | `#1C1F24` | `#FFFFFF` | Cards, panels, table tiles |
| `surface-raised` | `#262B33` | `#EDEAE0` | Hover, selected tile, modal |
| `text` | `#F2EFE6` | `#1A1C1E` | Primary text |
| `text-muted` | `#9AA1AD` | `#62676F` | Secondary text, timestamps |
| `accent` | `#E8A33D` | `#B97A1A` | Primary actions, active table highlight |
| `success` | `#4CAF6D` | `#2E7D46` | Paid / closed, in-stock |
| `warn` | `#E0B13E` | `#9A6F14` | Low stock, tab open long |
| `danger` | `#E05C5C` | `#B33737` | Errors, out of stock, delete |
| `on-accent` | `#1A1206` | `#FFFFFF` | Text on accent fills |

Rules: text on `bg`/`surface` must pass WCAG AA (4.5:1); status is
always color + label, never color-only.

## Typography

| Token | Size / weight | Usage |
|-------|--------------|-------|
| `display` | 32 / 700 | Day total, tab total |
| `title` | 20 / 600 | Section headers, table names |
| `body` | 15 / 400 | Lists, forms, descriptions |
| `caption` | 12 / 400, `text-muted` | Timestamps, hints |
| `mono-numeric` | tabular figures, 400–700 matching context | Every price, qty, and total |

Rules: all money uses `mono-numeric` with tabular figures so columns
align; totals are `display`, never `body`; currency formatting is
`$ 1.250,00`-style locale output of integer cents, no float math.

## Spacing, radius, touch

| Token | Value |
|-------|-------|
| Spacing scale | 4pt: `4 / 8 / 12 / 16 / 24 / 32` |
| `radius-sm` | 4px (chips, badges) |
| `radius-md` | 8px (buttons, inputs, tiles) |
| `radius-lg` | 12px (cards, modals) |
| Mobile touch minimum | 48dp target, 8dp gap |
| PC dense minimum | 32px target, 4px gap |

Rules: mobile order actions always meet 48dp; PC dashboard may go dense
(32px) but never below; destructive actions need a gap, never adjacency.

## States

| State | Token combo | Label shown |
|-------|------------|-------------|
| Product available | `text` + normal tile | name + price |
| Product out of stock | `text-muted` + `danger` badge | "Out of stock" chip, disabled add |
| Table free | `surface`, `text-muted` label | "Free" |
| Table occupied | `surface-raised` edge in `accent` | "Occupied · $ total" (mono-numeric) |
| Tab open long (>90 min) | `warn` badge | "Open HH:MM" |
| Paid / closed | `success` label | "Paid" |

## Checklist

- [ ] Theme copies tokens verbatim — no per-platform hex drift.
- [ ] Every status pairs color with a text label.
- [ ] Every price/total uses `mono-numeric` tabular figures.
- [ ] Touch targets meet 48dp mobile / 32px PC minimums.

## Next step

Dashboard layout and section contracts: `docs/dashboard.md`.
