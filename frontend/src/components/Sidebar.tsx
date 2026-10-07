// Sticky app sidebar: brand mark, section nav, connection mini-pill,
// global refresh, and the light/dark toggle. Under 720px the shell CSS
// reflows this same node into a top bar, so there is only one nav to keep.
import type { BrandingDTO } from './branding';
import { btnGhost } from './ui';

export type SectionId = 'panel' | 'mesas' | 'catalogo' | 'gastos' | 'conexion' | 'personalizar';

export const SECTIONS: ReadonlyArray<{ id: SectionId; label: string }> = [
  { id: 'panel', label: 'Panel' },
  { id: 'mesas', label: 'Mesas' },
  { id: 'catalogo', label: 'Catálogo' },
  { id: 'gastos', label: 'Gastos' },
  { id: 'conexion', label: 'Conexión' },
  { id: 'personalizar', label: 'Personalizar' },
];

interface SidebarProps {
  branding: BrandingDTO;
  logoSrc: string | null;
  onLogoError: () => void;
  section: SectionId;
  onSelect: (s: SectionId) => void;
  pillText: string;
  pillOk: boolean;
  theme: 'light' | 'dark';
  onToggleTheme: () => void;
  onRefresh: () => void;
}

export default function Sidebar({
  branding,
  logoSrc,
  onLogoError,
  section,
  onSelect,
  pillText,
  pillOk,
  theme,
  onToggleTheme,
  onRefresh,
}: SidebarProps) {
  const name = branding.shop_name.trim() || 'mitt';
  const initial = name.slice(0, 1).toUpperCase();
  return (
    <aside className="mitt-sidebar" aria-label="Navegación principal">
      <div className="mitt-brand">
        {logoSrc ? (
          <img src={logoSrc} alt={`${name} logo`} onError={onLogoError} />
        ) : (
          <span className="mitt-brand-fallback" aria-hidden="true">
            {initial}
          </span>
        )}
        <span className="mitt-brand-name" title={name}>
          {name}
        </span>
      </div>
      <nav className="mitt-nav" aria-label="Secciones">
        {SECTIONS.map((s) => (
          <button
            key={s.id}
            type="button"
            className="mitt-navbtn"
            aria-current={section === s.id ? 'page' : undefined}
            onClick={() => onSelect(s.id)}
          >
            {s.label}
          </button>
        ))}
      </nav>
      <div className="mitt-side-foot">
        <p className="mitt-pill" role="status">
          <span className={`mitt-dot${pillOk ? ' ok' : ''}`} aria-hidden="true" />
          {pillText}
        </p>
        <button type="button" className={btnGhost} onClick={onRefresh}>
          REFRESCAR
        </button>
        <button
          type="button"
          className={btnGhost}
          onClick={onToggleTheme}
          aria-pressed={theme === 'light'}
          title={theme === 'dark' ? 'Cambiar a modo claro' : 'Cambiar a modo oscuro'}
        >
          {theme === 'dark' ? 'MODO CLARO' : 'MODO OSCURO'}
        </button>
      </div>
    </aside>
  );
}
