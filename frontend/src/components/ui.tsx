// Shared form/button/card vocabulary so every section paints identically.
// Single source: sidebar, content sections, and Personalizar import these
// instead of redefining their own controls (Operate: consistent affordances).
export const inputCls =
  'rounded-lg border border-mitt-raised bg-mitt-raised px-4 py-2 min-h-12 text-mitt-text placeholder:text-mitt-muted focus:outline-2 focus:outline-mitt-accent';
export const btnPrimary =
  'mitt-press rounded-lg bg-mitt-accent px-4 py-2 min-h-12 font-semibold text-mitt-on-accent hover:brightness-110 disabled:opacity-50';
export const btnGhost =
  'mitt-press rounded-lg border border-mitt-muted bg-mitt-raised px-4 py-2 min-h-12 text-mitt-text hover:brightness-125 disabled:opacity-50';
export const btnDanger =
  'mitt-press rounded-lg bg-mitt-danger px-4 py-2 min-h-12 font-semibold text-mitt-text hover:brightness-110 disabled:opacity-50';
export const cardCls = 'rounded-xl border border-mitt-raised bg-mitt-surface p-4';
export const sectionTitle = 'text-[20px] font-semibold text-mitt-text';
