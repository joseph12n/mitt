// Personalizar: white-label editor for the shop identity (name, palette,
// logo). Saves one merged PATCH with only the changed fields, then the
// shell repaints live. Same token rules as every other mutation.
import { useEffect, useRef, useState } from 'react';
import { api } from '../api';
import { DEFAULT_BRANDING, HEX_RE, MAX_LOGO_BYTES, type BrandingDTO } from './branding';
import { btnDanger, btnGhost, btnPrimary, cardCls, inputCls, sectionTitle } from './ui';

const ACCEPTED_LOGO_MIME = ['image/png', 'image/jpeg', 'image/svg+xml'];

interface PersonalizarProps {
  branding: BrandingDTO;
  logoSrc: string | null;
  hasToken: boolean;
  onSaved: (next: BrandingDTO) => void;
  notify: (msg: string) => void;
  fail: (err: unknown) => void;
}

function ColorRow({
  label,
  value,
  fallback,
  onChange,
}: {
  label: string;
  value: string;
  fallback: string;
  onChange: (v: string) => void;
}) {
  const valid = HEX_RE.test(value);
  return (
    <div className="mitt-color-row">
      <span className="mitt-color-label">{label}</span>
      <input
        type="color"
        className="mitt-swatch"
        aria-label={`${label} selector de color`}
        value={valid ? value : fallback}
        onChange={(e) => onChange(e.target.value)}
      />
      <input
        className={`${inputCls} money w-28`}
        type="text"
        aria-label={`${label} valor hexadecimal`}
        value={value}
        maxLength={7}
        autoComplete="off"
        autoCorrect="off"
        autoCapitalize="off"
        spellCheck={false}
        placeholder="#rrggbb"
        onChange={(e) => onChange(e.target.value.trim())}
      />
      {!valid && <span className="mitt-field-error">Use el formato #rrggbb.</span>}
    </div>
  );
}

export default function Personalizar({
  branding,
  logoSrc,
  hasToken,
  onSaved,
  notify,
  fail,
}: PersonalizarProps) {
  const [shopName, setShopName] = useState(branding.shop_name);
  const [primary, setPrimary] = useState(branding.primary);
  const [accent, setAccent] = useState(branding.accent);
  const [background, setBackground] = useState(branding.background);
  const [logoMode, setLogoMode] = useState<'keep' | 'set' | 'clear'>('keep');
  const [logoDataUrl, setLogoDataUrl] = useState('');
  const [logoError, setLogoError] = useState('');
  const [saving, setSaving] = useState(false);
  const fileRef = useRef<HTMLInputElement>(null);

  // A save (or reset) lands new live branding: rebase the form onto it.
  useEffect(() => {
    setShopName(branding.shop_name);
    setPrimary(branding.primary);
    setAccent(branding.accent);
    setBackground(branding.background);
    setLogoMode('keep');
    setLogoDataUrl('');
    setLogoError('');
  }, [branding]);

  const nameOk = shopName.trim().length >= 1 && shopName.trim().length <= 60;
  const colorsOk = HEX_RE.test(primary) && HEX_RE.test(accent) && HEX_RE.test(background);
  const dirty =
    shopName.trim() !== branding.shop_name ||
    primary !== branding.primary ||
    accent !== branding.accent ||
    background !== branding.background ||
    logoMode !== 'keep';
  const canSave =
    hasToken && nameOk && colorsOk && dirty && !saving && (logoMode !== 'set' || logoDataUrl !== '');

  const pickFile = () => fileRef.current?.click();

  const onFile = (file: File | undefined) => {
    if (!file) return;
    const svgByName = /\.svg$/i.test(file.name);
    const mimeOk =
      ACCEPTED_LOGO_MIME.includes(file.type) || (svgByName && (file.type === '' || file.type === 'image/svg+xml'));
    if (!mimeOk) {
      setLogoError('El logo debe ser PNG, JPEG o SVG.');
      return;
    }
    if (file.size > MAX_LOGO_BYTES) {
      setLogoError('El logo no debe superar 512 KB.');
      return;
    }
    setLogoError('');
    const reader = new FileReader();
    reader.onload = () => {
      let url = String(reader.result || '');
      // Nameless-svg fallback: some pickers report no MIME, so label it.
      if (url.startsWith('data:application/octet-stream;base64,')) {
        url = url.replace('data:application/octet-stream;base64,', 'data:image/svg+xml;base64,');
      }
      setLogoDataUrl(url);
      setLogoMode('set');
    };
    reader.onerror = () => setLogoError('No se pudo leer el archivo.');
    reader.readAsDataURL(file);
  };

  const save = async () => {
    if (!canSave) return;
    const body: Record<string, string | null> = {};
    const name = shopName.trim();
    if (name !== branding.shop_name) body.shop_name = name;
    if (primary !== branding.primary) body.primary = primary;
    if (accent !== branding.accent) body.accent = accent;
    if (background !== branding.background) body.background = background;
    if (logoMode === 'set' && logoDataUrl !== '') body.logo_data_url = logoDataUrl;
    if (logoMode === 'clear') body.logo_data_url = null;
    setSaving(true);
    try {
      const next = await api<BrandingDTO>('PATCH', '/api/branding', body);
      onSaved(next);
      notify('Marca guardada.');
    } catch (err) {
      fail(err);
    } finally {
      setSaving(false);
    }
  };

  // Reset restores the factory name + palette only; the logo is untouched.
  const reset = async () => {
    if (!hasToken || saving) return;
    if (!confirm('Restablecer nombre y colores a los valores de fábrica?')) return;
    setSaving(true);
    try {
      const next = await api<BrandingDTO>('PATCH', '/api/branding', {
        shop_name: DEFAULT_BRANDING.shop_name,
        primary: DEFAULT_BRANDING.primary,
        accent: DEFAULT_BRANDING.accent,
        background: DEFAULT_BRANDING.background,
      });
      onSaved(next);
      notify('Marca restablecida.');
    } catch (err) {
      fail(err);
    } finally {
      setSaving(false);
    }
  };

  const previewLogo = logoMode === 'set' ? logoDataUrl : logoMode === 'clear' ? null : logoSrc;
  const previewName = shopName.trim() || '—';

  return (
    <section aria-label="Personalizar">
      <h2 className={`${sectionTitle} mb-2 mt-2`}>Personalizar</h2>
      {!hasToken && (
        <div className={`${cardCls} mb-2`}>
          <p className="font-semibold text-mitt-text">Se necesita conexión</p>
          <p className="mt-1 text-sm text-mitt-muted">
            Guarde un token en Conexión para personalizar la marca.
          </p>
        </div>
      )}
      <div className={`${cardCls} flex flex-col gap-4`}>
        <div className="flex flex-col gap-2">
          <label className="text-xs text-mitt-muted" htmlFor="shop-name-input">
            Nombre del local
          </label>
          <input
            id="shop-name-input"
            className={`${inputCls} w-full max-w-md`}
            type="text"
            value={shopName}
            maxLength={60}
            autoComplete="off"
            autoCorrect="off"
            spellCheck={false}
            onChange={(e) => setShopName(e.target.value)}
          />
          {!nameOk && (
            <span className="mitt-field-error">El nombre lleva de 1 a 60 caracteres.</span>
          )}
        </div>

        <div className="flex flex-col gap-3" aria-label="Paleta">
          <ColorRow label="Primario" value={primary} fallback={branding.primary} onChange={setPrimary} />
          <ColorRow label="Acento" value={accent} fallback={branding.accent} onChange={setAccent} />
          <ColorRow label="Fondo" value={background} fallback={branding.background} onChange={setBackground} />
        </div>

        <div className="flex flex-col gap-2" aria-label="Logo">
          <span className="text-xs text-mitt-muted">Logo (PNG, JPEG o SVG, hasta 512 KB)</span>
          <input
            ref={fileRef}
            type="file"
            className="hidden"
            accept="image/png,image/jpeg,image/svg+xml,.svg"
            aria-label="Elegir archivo de logo"
            onChange={(e) => {
              onFile(e.target.files?.[0]);
              e.target.value = '';
            }}
          />
          <div className="flex flex-wrap gap-2">
            <button type="button" className={btnGhost} onClick={pickFile} disabled={!hasToken}>
              ELEGIR LOGO
            </button>
            <button
              type="button"
              className={btnDanger}
              disabled={!hasToken || (!logoSrc && logoMode !== 'set')}
              onClick={() => {
                setLogoMode('clear');
                setLogoDataUrl('');
                setLogoError('');
              }}
            >
              QUITAR LOGO
            </button>
          </div>
          {logoError && <span className="mitt-field-error">{logoError}</span>}
          {logoMode === 'clear' && (
            <span className="text-xs text-mitt-muted">El logo se quitará al guardar.</span>
          )}
        </div>

        <div aria-label="Vista previa">
          <p className="mb-2 text-xs text-mitt-muted">Vista previa</p>
          <div
            className="mitt-preview"
            style={{ background: HEX_RE.test(background) ? background : undefined }}
          >
            {previewLogo ? (
              <img src={previewLogo} alt="Vista previa del logo" />
            ) : (
              <span className="mitt-preview-nologo">Sin logo</span>
            )}
            <strong>{previewName}</strong>
            <span className="mitt-preview-swatches" aria-hidden="true">
              {[primary, accent, background].map((c, i) => (
                <span key={i} style={{ background: HEX_RE.test(c) ? c : 'transparent' }} />
              ))}
            </span>
            <span
              className="mitt-preview-cta"
              style={{ background: HEX_RE.test(accent) ? accent : undefined }}
            >
              EJEMPLO
            </span>
          </div>
        </div>

        <div className="flex flex-wrap gap-2">
          <button type="button" className={btnPrimary} onClick={save} disabled={!canSave}>
            {saving ? 'GUARDANDO…' : 'GUARDAR'}
          </button>
          <button type="button" className={btnGhost} onClick={reset} disabled={!hasToken || saving}>
            RESTABLECER
          </button>
        </div>
      </div>
    </section>
  );
}
