// White-label branding helpers shared by the shell and Personalizar.
// Wire shape mirrors the snake_case DTO in internal/api/branding.go; the
// defaults copy domain.DefaultBranding (docs/design-tokens.md dark column).
export interface BrandingDTO {
  shop_name: string;
  primary: string;
  accent: string;
  background: string;
  has_logo: boolean;
  updated_at: string;
}

// Factory identity: night-bar-first tokens, no logo.
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
// offline or unreachable backend keeps the built-in dark tokens working.
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

// applyBranding paints the shop identity over the active light/dark theme:
// background -> page, primary -> surfaces, accent -> primary actions.
// Invalid values are skipped so partial payloads never blank the UI.
export function applyBranding(b: BrandingDTO): void {
  const root = document.documentElement;
  if (HEX_RE.test(b.primary)) root.style.setProperty('--mitt-surface', b.primary);
  if (HEX_RE.test(b.accent)) root.style.setProperty('--mitt-accent', b.accent);
  if (HEX_RE.test(b.background)) root.style.setProperty('--mitt-bg', b.background);
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
