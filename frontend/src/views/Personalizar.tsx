// Personalizar: shop identity editor. Figma layout with the preset gallery;
// every change stages locally and lands with one merged PATCH /api/branding
// (server wins). The hub background stays untouched; only name, primary,
// accent and logo travel. Theme toggle persists locally, default light.
import { useEffect, useRef, useState } from 'react'
import { Check, ImagePlus, Moon, Sun, Trash2 } from 'lucide-react'
import { api, DEFAULT_BRANDING, HEX_RE, MAX_LOGO_BYTES, type BrandingDTO } from '../api'
import { presets } from '../presets'
import { Header } from './ui'

interface PersonalizarProps {
  branding: BrandingDTO
  logoSrc: string | null
  hasToken: boolean
  dark: boolean
  setDark: (d: boolean) => void
  onSaved: (next: BrandingDTO) => void
  notify: (msg: string) => void
  fail: (err: unknown) => void
}

const ACCEPTED_LOGO_MIME = ['image/png', 'image/jpeg', 'image/svg+xml']

export default function Personalizar({ branding, logoSrc, hasToken, dark, setDark, onSaved, notify, fail }: PersonalizarProps) {
  const [shopName, setShopName] = useState(branding.shop_name)
  const [primary, setPrimary] = useState(branding.primary)
  const [accent, setAccent] = useState(branding.accent)
  const [logoMode, setLogoMode] = useState<'keep' | 'set' | 'clear'>('keep')
  const [logoDataUrl, setLogoDataUrl] = useState('')
  const [logoError, setLogoError] = useState('')
  const [saving, setSaving] = useState(false)
  const file = useRef<HTMLInputElement>(null)

  // A save lands new live branding: rebase the form onto it.
  useEffect(() => {
    setShopName(branding.shop_name)
    setPrimary(branding.primary)
    setAccent(branding.accent)
    setLogoMode('keep')
    setLogoDataUrl('')
    setLogoError('')
  }, [branding])

  const nameOk = shopName.trim().length >= 1 && shopName.trim().length <= 60
  const colorsOk = HEX_RE.test(primary) && HEX_RE.test(accent)
  const dirty =
    shopName.trim() !== branding.shop_name ||
    primary !== branding.primary ||
    accent !== branding.accent ||
    logoMode !== 'keep'
  const canSave = hasToken && nameOk && colorsOk && dirty && !saving && (logoMode !== 'set' || logoDataUrl !== '')

  const onFile = (f: File | undefined) => {
    if (!f) return
    const svgByName = /\.svg$/i.test(f.name)
    const mimeOk = ACCEPTED_LOGO_MIME.includes(f.type) || (svgByName && (f.type === '' || f.type === 'image/svg+xml'))
    if (!mimeOk) {
      setLogoError('El logo debe ser PNG, JPEG o SVG.')
      return
    }
    if (f.size > MAX_LOGO_BYTES) {
      setLogoError('El logo no debe superar 512 KB.')
      return
    }
    setLogoError('')
    const reader = new FileReader()
    reader.onload = () => {
      let dataUrl = String(reader.result || '')
      // Nameless-svg fallback: some pickers report no MIME, so label it.
      if (dataUrl.startsWith('data:application/octet-stream;base64,')) {
        dataUrl = dataUrl.replace('data:application/octet-stream;base64,', 'data:image/svg+xml;base64,')
      }
      setLogoDataUrl(dataUrl)
      setLogoMode('set')
    }
    reader.onerror = () => setLogoError('No se pudo leer el archivo.')
    reader.readAsDataURL(f)
  }

  const save = async () => {
    if (!canSave) return
    const body: Record<string, string | null> = {}
    const name = shopName.trim()
    if (name !== branding.shop_name) body.shop_name = name
    if (primary !== branding.primary) body.primary = primary
    if (accent !== branding.accent) body.accent = accent
    if (logoMode === 'set' && logoDataUrl !== '') body.logo_data_url = logoDataUrl
    if (logoMode === 'clear') body.logo_data_url = null
    setSaving(true)
    try {
      const next = await api<BrandingDTO>('PATCH', '/api/branding', body)
      onSaved(next)
      notify('Marca guardada.')
    } catch (err) {
      fail(err)
    } finally {
      setSaving(false)
    }
  }

  // Reset restores the factory name + palette; the logo is untouched.
  const reset = async () => {
    if (!hasToken || saving) return
    if (!confirm('Restablecer nombre y colores a los valores de fábrica?')) return
    setSaving(true)
    try {
      const next = await api<BrandingDTO>('PATCH', '/api/branding', {
        shop_name: DEFAULT_BRANDING.shop_name,
        primary: DEFAULT_BRANDING.primary,
        accent: DEFAULT_BRANDING.accent,
        background: DEFAULT_BRANDING.background,
      })
      onSaved(next)
      notify('Marca restablecida.')
    } catch (err) {
      fail(err)
    } finally {
      setSaving(false)
    }
  }

  const previewLogo = logoMode === 'set' ? logoDataUrl : logoMode === 'clear' ? null : logoSrc

  return (
    <>
      <Header title="Personalizar" sub="Identidad, tema y paleta de tu establecimiento" />
      {!hasToken && (
        <div className="card p-4 mb-4">
          <p className="font-semibold">Se necesita conexión</p>
          <p className="mt-1 text-sm text-mute">Guarda un token en Conexión para personalizar la marca.</p>
        </div>
      )}
      <div className="grid gap-4 xl:grid-cols-[420px_1fr]">
        <div className="space-y-4">
          <div className="card p-6 space-y-5">
            <h2 className="font-display font-bold text-xl">Establecimiento</h2>
            <div className="flex items-center gap-4">
              <div className="size-20 rounded-2xl bg-sunk border border-dashed border-line grid place-items-center overflow-hidden text-mute">
                {previewLogo ? <img src={previewLogo} alt="Logo" className="size-full object-cover" /> : <ImagePlus size={24} />}
              </div>
              <div className="flex flex-col gap-2">
                <input ref={file} type="file" accept="image/png,image/jpeg,image/svg+xml,.svg" hidden
                  aria-label="Elegir archivo de logo"
                  onChange={(e) => { onFile(e.target.files?.[0]); e.target.value = '' }} />
                <button className="btn btn-primary h-9!" onClick={() => file.current?.click()} disabled={!hasToken}><ImagePlus size={15} />{previewLogo ? 'Cambiar logo' : 'Subir logo'}</button>
                {previewLogo && <button className="btn btn-ghost h-9!" onClick={() => { setLogoMode('clear'); setLogoDataUrl(''); setLogoError('') }}><Trash2 size={15} />Quitar</button>}
              </div>
            </div>
            {logoError && <p className="text-xs text-berry font-semibold">{logoError}</p>}
            <label className="block text-sm font-semibold">Nombre del local
              <input className="field w-full mt-2" maxLength={60} value={shopName}
                autoComplete="off" autoCorrect="off" spellCheck={false}
                onChange={(e) => setShopName(e.target.value)} />
            </label>
            {!nameOk && <p className="text-xs text-berry font-semibold">El nombre lleva de 1 a 60 caracteres.</p>}
          </div>

          <div className="card p-6">
            <h2 className="font-display font-bold text-xl mb-4">Tema</h2>
            <div className="grid grid-cols-2 gap-3">
              {([[false, 'Claro', Sun], [true, 'Oscuro', Moon]] as Array<[boolean, string, typeof Sun]>).map(([v, l, I]) => (
                <button key={l} onClick={() => setDark(v)} className={`h-12 rounded-xl border flex items-center justify-center gap-2 text-sm font-semibold transition ${dark === v ? 'border-brand bg-brand-soft text-brand' : 'border-line hover:bg-sunk'}`}><I size={16} />{l}</button>
              ))}
            </div>
            <h3 className="text-sm font-semibold mt-6 mb-3">Colores a medida</h3>
            <div className="grid grid-cols-2 gap-3">
              {([['Principal', primary, setPrimary], ['Acento', accent, setAccent]] as Array<[string, string, (c: string) => void]>).map(([l, v, set]) => (
                <label key={l} className="flex items-center gap-3 rounded-xl bg-sunk p-2 pr-3 text-sm font-medium cursor-pointer">
                  <input type="color" value={HEX_RE.test(v) ? v : '#000000'} onChange={(e) => set(e.target.value)} className="size-9 rounded-lg border-0 bg-transparent cursor-pointer" />{l}
                </label>
              ))}
            </div>
            {(!HEX_RE.test(primary) || !HEX_RE.test(accent)) && <p className="text-xs text-berry font-semibold mt-2">Usa el formato #rrggbb.</p>}
          </div>

          <div className="flex flex-wrap gap-2">
            <button className="btn btn-primary" onClick={save} disabled={!canSave}>{saving ? 'Guardando…' : 'Guardar'}</button>
            <button className="btn btn-ghost" onClick={reset} disabled={!hasToken || saving}>Restablecer</button>
          </div>
        </div>

        <div className="card p-6">
          <div className="flex items-end justify-between mb-5">
            <div><h2 className="font-display font-bold text-xl">Presets de diseño</h2><p className="text-sm text-mute">{presets.length} paletas para cada tipo de local</p></div>
          </div>
          <div className="grid gap-3 sm:grid-cols-2 2xl:grid-cols-3">
            {presets.map((p) => {
              const on = primary === p.brand && accent === p.accent
              return (
                <button key={p.id} onClick={() => { setPrimary(p.brand); setAccent(p.accent) }}
                  className={`text-left rounded-2xl border p-3 transition hover:-translate-y-0.5 ${on ? 'border-brand ring-2 ring-brand' : 'border-line hover:bg-sunk'}`}>
                  <div className="h-14 rounded-xl flex overflow-hidden">
                    <div className="flex-[3]" style={{ background: p.brand }} />
                    <div className="flex-[1.2]" style={{ background: p.accent }} />
                    <div className="flex-1" style={{ background: `color-mix(in srgb, ${p.brand} 14%, white)` }} />
                  </div>
                  <div className="flex items-center justify-between mt-3">
                    <div><div className="text-sm font-semibold">{p.name}</div><div className="text-xs text-mute">{p.tag}</div></div>
                    {on && <span className="size-6 rounded-full bg-brand text-on-brand grid place-items-center"><Check size={14} /></span>}
                  </div>
                </button>
              )
            })}
          </div>
        </div>
      </div>
    </>
  )
}
