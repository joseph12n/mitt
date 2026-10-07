// Typed client for the mitt LAN API (docs/api.md). Field names mirror the
// snake_case json tags in internal/api/*.go; money stays integer cents and
// every request carries the pairing token except the public dashboard page.

export interface Product {
  id: string;
  name: string;
  price_cents: number;
  available: boolean;
}

export interface TabItem {
  product_id: string;
  name: string;
  unit_price_cents: number;
  qty: number;
  line_total_cents: number;
}

export interface Tab {
  id: string;
  table_id: string;
  status: string;
  opened_at: string;
  items: TabItem[];
  total_cents: number;
}

export interface Sale {
  id: string;
  table_id: string;
  items: TabItem[];
  total_cents: number;
  closed_at: string;
}

export interface Expense {
  id: string;
  description: string;
  qty: number;
  cost_cents: number;
  date: string;
}

export interface Pairing {
  url: string;
  pairing_code: string;
}

// Hub-owned table row (GET/POST/DELETE /api/tables). Occupied is derived
// server-side from open tabs at read time, never stored.
export interface Table {
  id: string;
  label: string;
  occupied: boolean;
}

export interface ApiError {
  status: number;
  message: string;
}

// Supplier: short provider record (name 1..80, phone <=40, note <=200).
// The hub stores no catalog or purchase history per supplier.
export interface Supplier {
  id: string;
  name: string;
  phone: string;
  note: string;
}

export interface TodaySummary {
  date: string;
  count: number;
  total_cents: number;
}

// BrandingDTO mirrors the snake_case wire form in internal/api/branding.go.
// The hub is the single authority: the web UI never persists brand keys
// locally, it only paints what GET /api/branding returns.
export interface BrandingDTO {
  shop_name: string;
  primary: string;
  accent: string;
  background: string;
  has_logo: boolean;
  updated_at: string;
}

// Factory identity copies domain.DefaultBranding (night-bar-first).
export const DEFAULT_BRANDING: BrandingDTO = {
  shop_name: 'mitt',
  primary: '#1C1F24',
  accent: '#E8A33D',
  background: '#121417',
  has_logo: false,
  updated_at: '',
};

export const HEX_RE = /^#[0-9a-fA-F]{6}$/;

// Client-side logo cap mirrors domain.MaxLogoBytes (512KiB).
export const MAX_LOGO_BYTES = 512 * 1024;

// fetchPublicBranding loads the login-less identity. It never throws: an
// offline or unreachable backend keeps the built-in tokens working.
export async function fetchPublicBranding(base: string): Promise<BrandingDTO | null> {
  try {
    const res = await fetch(`${base.replace(/\/$/, '')}/api/branding`);
    if (!res.ok) return null;
    const body = (await res.json()) as Partial<BrandingDTO>;
    if (typeof body.shop_name !== 'string' || body.shop_name.trim() === '') return null;
    return {
      shop_name: body.shop_name,
      primary:
        typeof body.primary === 'string' && HEX_RE.test(body.primary)
          ? body.primary
          : DEFAULT_BRANDING.primary,
      accent:
        typeof body.accent === 'string' && HEX_RE.test(body.accent)
          ? body.accent
          : DEFAULT_BRANDING.accent,
      background:
        typeof body.background === 'string' && HEX_RE.test(body.background)
          ? body.background
          : DEFAULT_BRANDING.background,
      has_logo: body.has_logo === true,
      updated_at: typeof body.updated_at === 'string' ? body.updated_at : '',
    };
  } catch {
    return null;
  }
}

// applyBranding paints the hub identity onto the Figma theme: primary
// drives --brand-base (brand surfaces and actions), accent drives
// --sun-base (highlights). The hub background is kept server-side only;
// the theme derives its paper/sunk tones from the brand base instead.
// Invalid values are skipped so partial payloads never blank the UI.
export function applyBranding(b: BrandingDTO): void {
  const root = document.documentElement;
  if (HEX_RE.test(b.primary)) root.style.setProperty('--brand-base', b.primary);
  if (HEX_RE.test(b.accent)) root.style.setProperty('--sun-base', b.accent);
  const name = b.shop_name.trim() || DEFAULT_BRANDING.shop_name;
  document.title = `${name} · Panel del bar`;
}

// logoUrl points at the public raw logo; updated_at busts the immutable cache.
export function logoUrl(base: string, updatedAt: string): string {
  const clean = base.replace(/\/$/, '');
  return updatedAt
    ? `${clean}/api/branding/logo?ts=${encodeURIComponent(updatedAt)}`
    : `${clean}/api/branding/logo`;
}

// Live list helpers: every view reads these, never local seeds.
export async function listProducts(): Promise<Product[]> {
  const res = await api<{ products: Product[] }>('GET', '/api/products');
  return res.products || [];
}

export async function listOpenTabs(): Promise<Tab[]> {
  const res = await api<{ tabs: Tab[] }>('GET', '/api/tabs/open');
  return res.tabs || [];
}

export async function listExpenses(): Promise<Expense[]> {
  const res = await api<{ expenses: Expense[] }>('GET', '/api/expenses');
  return res.expenses || [];
}

export async function listSuppliers(): Promise<Supplier[]> {
  const res = await api<{ suppliers: Supplier[] }>('GET', '/api/suppliers');
  return res.suppliers || [];
}

// Sales history, newest first, for the Panel charts and recent orders.
export async function listSales(limit = 200): Promise<Sale[]> {
  const res = await api<{ sales: Sale[] }>('GET', `/api/sales?limit=${limit}`);
  return res.sales || [];
}

export function fetchToday(): Promise<TodaySummary> {
  return api<TodaySummary>('GET', '/api/sales/today');
}

export function fetchPairing(): Promise<Pairing> {
  return api<Pairing>('GET', '/api/pairing');
}

// Hub ids are opaque: resolve the human label for display, falling back
// to the raw id.
export function labelOf(tables: Table[], tableId: string): string {
  return tables.find((t) => t.id === tableId)?.label ?? tableId;
}

export function baseUrl(): string {
  return localStorage.getItem('mitt_base') || location.origin;
}

export function storedToken(): string {
  return localStorage.getItem('mitt_token') || '';
}

function failMessage(status: number, body: string): string {
  if (status === 401 || status === 403) return 'Token inválido. Revise la conexión.';
  if (status === 404) return 'No encontrado. Pulse REFRESCAR.';
  let code = '';
  let detail = '';
  try {
    const parsed = JSON.parse(body || '');
    if (parsed && parsed.error) {
      if (parsed.error.code) code = String(parsed.error.code);
      if (parsed.error.message) detail = String(parsed.error.message);
    }
  } catch {
    detail = '';
  }
  // The hub refuses table deletes while the bill is open; staff get the
  // reason, not a code.
  if (status === 409 && code === 'table_occupied') return 'LA MESA TIENE CUENTA ABIERTA.';
  if (status === 422) return detail ? `Dato inválido: ${detail}` : 'Dato inválido. Revise el formulario.';
  return `Ocurrió un error (código ${status}). Intente de nuevo.`;
}

export async function api<T>(method: string, path: string, data?: unknown): Promise<T> {
  const token = storedToken();
  const headers: Record<string, string> = { Authorization: `Bearer ${token}` };
  const init: RequestInit = { method, headers };
  if (data !== undefined) {
    headers['Content-Type'] = 'application/json';
    init.body = JSON.stringify(data);
  }
  const res = await fetch(baseUrl() + path, init);
  const text = await res.text();
  if (!res.ok) throw new Error(failMessage(res.status, text));
  // DELETE answers 204 with no body; every other call returns JSON.
  if (!text) return undefined as T;
  return JSON.parse(text) as T;
}

// Quiet probe: does the saved token actually work? Never throws.
export async function checkToken(): Promise<boolean> {
  try {
    const res = await fetch(baseUrl() + '/api/products', {
      headers: { Authorization: `Bearer ${storedToken()}` },
    });
    return res.ok;
  } catch {
    return false;
  }
}

// Accept a pasted "mitt://pair?url=..&token=.." link: extract token and base URL.
export function readPairing(raw: string): { token: string; base: string | null } {
  const trimmed = (raw || '').trim();
  if (!trimmed.startsWith('mitt://pair')) return { token: trimmed, base: null };
  const query = trimmed.slice(trimmed.indexOf('?') + 1);
  const out = { token: '', base: null as string | null };
  query.split('&').forEach((part) => {
    const eq = part.indexOf('=');
    const key = decodeURIComponent(eq < 0 ? part : part.slice(0, eq));
    const value = decodeURIComponent(eq < 0 ? '' : part.slice(eq + 1));
    if (key === 'token') out.token = value;
    if (key === 'url') out.base = value.replace(/\/$/, '');
  });
  return out;
}

// Money: integer cents to "$ 12.50" display (no float math on totals).
export function money(cents: number): string {
  return '$ ' + (cents / 100).toFixed(2);
}

// Hub-owned tables: the hub is the single source of truth, the web UI only
// lists, creates, and deletes rows through these helpers.
export async function listTables(): Promise<Table[]> {
  const res = await api<{ tables: Table[] }>('GET', '/api/tables');
  return res.tables || [];
}

export function createTable(label: string): Promise<Table> {
  return api<Table>('POST', '/api/tables', { label });
}

export async function deleteTable(id: string): Promise<void> {
  await api<unknown>('DELETE', `/api/tables/${encodeURIComponent(id)}`);
}
