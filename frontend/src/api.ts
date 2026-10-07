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

export interface ApiError {
  status: number;
  message: string;
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
  let detail = '';
  try {
    const parsed = JSON.parse(body || '');
    if (parsed && parsed.error && parsed.error.message) detail = String(parsed.error.message);
  } catch {
    detail = '';
  }
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
