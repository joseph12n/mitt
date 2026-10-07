// mitt web shell: Figma master layout wired to the LIVE hub. App owns one
// refresh that pulls every collection in parallel; views receive slices plus
// the refresh callback and mutate through the LAN API. Branding authority is
// GET /api/branding (no local brand keys); localStorage keeps only the
// pairing token (+base) and the light/dark theme. Spanish UI, English code.
import { useCallback, useEffect, useRef, useState } from 'react'
import { LayoutDashboard, Armchair, BookOpen, Receipt, QrCode, Palette, RefreshCw, Truck, Moon, Sun } from 'lucide-react'
import {
  applyBranding,
  baseUrl,
  checkToken,
  DEFAULT_BRANDING,
  fetchPairing,
  fetchPublicBranding,
  fetchToday,
  listExpenses,
  listOpenTabs,
  listProducts,
  listSales,
  listSuppliers,
  listTables,
  logoUrl,
  storedToken,
  type BrandingDTO,
  type Expense,
  type Pairing,
  type Product,
  type Sale,
  type Supplier,
  type Tab,
  type Table as HubTable,
  type TodaySummary,
} from './api'
import Panel from './views/Panel'
import Mesas from './views/Mesas'
import Catalogo from './views/Catalogo'
import Proveedores from './views/Proveedores'
import Gastos from './views/Gastos'
import Conexion, { type PillState } from './views/Conexion'
import Personalizar from './views/Personalizar'

const nav = [
  { id: 'panel', label: 'Panel', icon: LayoutDashboard },
  { id: 'mesas', label: 'Mesas', icon: Armchair },
  { id: 'catalogo', label: 'Catálogo', icon: BookOpen },
  { id: 'proveedores', label: 'Proveedores', icon: Truck },
  { id: 'gastos', label: 'Gastos', icon: Receipt },
  { id: 'conexion', label: 'Conexión', icon: QrCode },
  { id: 'personalizar', label: 'Personalizar', icon: Palette },
] as const
type View = (typeof nav)[number]['id']

const pillLabel: Record<PillState, string> = {
  empty: 'SIN TOKEN',
  checking: 'VERIFICANDO…',
  ok: 'CONECTADO',
  bad: 'INVÁLIDO',
}

// Legacy seed-era keys: the hub owns branding now, so drop them once.
const LEGACY_BRAND_KEYS = ['mitt:name', 'mitt:accent', 'mitt:accent2', 'mitt:logo']

export default function App() {
  const [view, setView] = useState<View>('panel')
  const [products, setProducts] = useState<Product[]>([])
  const [tables, setTables] = useState<HubTable[]>([])
  const [tabs, setTabs] = useState<Tab[]>([])
  const [expenses, setExpenses] = useState<Expense[]>([])
  const [suppliers, setSuppliers] = useState<Supplier[]>([])
  const [sales, setSales] = useState<Sale[]>([])
  const [today, setToday] = useState<TodaySummary | null>(null)
  const [pairing, setPairing] = useState<Pairing | null>(null)
  const [branding, setBranding] = useState<BrandingDTO>(DEFAULT_BRANDING)
  const [logoSrc, setLogoSrc] = useState<string | null>(null)
  const [pill, setPill] = useState<PillState>(storedToken() ? 'checking' : 'empty')
  const [toast, setToast] = useState('')
  // Master defaults light; the toggle persists as today under mitt_theme.
  const [dark, setDarkState] = useState(false)
  const toastTimer = useRef<number | null>(null)

  const showToast = useCallback((msg: string) => {
    setToast(msg)
    if (toastTimer.current !== null) window.clearTimeout(toastTimer.current)
    toastTimer.current = window.setTimeout(() => setToast(''), 4000)
  }, [])

  const fail = useCallback(
    (err: unknown) => {
      showToast(err instanceof Error ? err.message : 'Ocurrió un error. Intente de nuevo.')
    },
    [showToast],
  )

  const setDark = useCallback((d: boolean) => {
    setDarkState(d)
    try {
      localStorage.setItem('mitt_theme', d ? 'dark' : 'light')
    } catch {
      // Private mode: the session keeps the default, nothing breaks.
    }
    document.documentElement.classList.toggle('dark', d)
  }, [])

  const paintPill = useCallback(async () => {
    const token = storedToken()
    if (!token) {
      setPill('empty')
      return
    }
    setPill('checking')
    const probed = token
    const ok = await checkToken()
    if (storedToken() !== probed) return // user saved a new token while probing
    setPill(ok ? 'ok' : 'bad')
  }, [])

  const refresh = useCallback(async () => {
    if (!storedToken()) {
      setPill('empty')
      return
    }
    paintPill()
    try {
      const [tabsRes, prodsRes, expsRes, tablesRes, supsRes, salesRes, todayRes, pairRes] = await Promise.all([
        listOpenTabs(),
        listProducts(),
        listExpenses(),
        listTables(),
        listSuppliers(),
        listSales(),
        fetchToday(),
        fetchPairing().catch(() => null),
      ])
      setTabs(tabsRes)
      setProducts(prodsRes)
      setExpenses(expsRes)
      setTables(tablesRes)
      setSuppliers(supsRes)
      setSales(salesRes)
      setToday(todayRes)
      setPairing(pairRes)
    } catch (err) {
      fail(err)
    }
  }, [paintPill, fail])

  // Theme + legacy cleanup + public identity paint (login-less, offline-safe).
  useEffect(() => {
    let stored: string | null = null
    try {
      stored = localStorage.getItem('mitt_theme')
      LEGACY_BRAND_KEYS.forEach((k) => localStorage.removeItem(k))
    } catch {
      // Private mode: defaults still paint.
    }
    const isDark = stored === 'dark'
    setDarkState(isDark)
    document.documentElement.classList.toggle('dark', isDark)
    let live = true
    fetchPublicBranding(baseUrl()).then((b) => {
      if (!live || !b) return
      setBranding(b)
      applyBranding(b)
      setLogoSrc(b.has_logo ? logoUrl(baseUrl(), b.updated_at) : null)
    })
    return () => {
      live = false
    }
  }, [])

  useEffect(() => {
    paintPill()
    if (storedToken()) refresh()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const handleBrandingSaved = useCallback(
    (next: BrandingDTO) => {
      setBranding(next)
      applyBranding(next)
      setLogoSrc(next.has_logo ? logoUrl(baseUrl(), next.updated_at) : null)
    },
    [],
  )

  // Token saved from Conexión: re-verify, reload identity + collections.
  const handleConnected = useCallback(() => {
    paintPill()
    refresh()
    fetchPublicBranding(baseUrl()).then((b) => {
      if (!b) return
      setBranding(b)
      applyBranding(b)
      setLogoSrc(b.has_logo ? logoUrl(baseUrl(), b.updated_at) : null)
    })
  }, [paintPill, refresh])

  const openCount = tabs.length
  const name = branding.shop_name.trim() || 'mitt'

  return (
    <div className="min-h-screen lg:flex">
      <aside className="lg:sticky lg:top-0 lg:h-screen lg:w-64 shrink-0 flex lg:flex-col gap-2 p-3 lg:p-5 border-b lg:border-b-0 lg:border-r border-line bg-paper overflow-x-auto">
        <div className="hidden lg:flex items-center gap-3 mb-8 px-2">
          <div className="size-10 rounded-xl bg-brand text-on-brand grid place-items-center font-display font-extrabold text-lg overflow-hidden shrink-0">
            {logoSrc ? <img src={logoSrc} alt={`Logo de ${name}`} className="size-full object-cover" onError={() => setLogoSrc(null)} /> : name.slice(0, 1).toUpperCase()}
          </div>
          <div>
            <div className="font-display font-bold text-lg leading-none">{name}</div>
            <div className="text-xs text-mute mt-1">Servidor local</div>
          </div>
        </div>
        <nav className="flex lg:flex-col gap-1 flex-1" aria-label="Secciones">
          {nav.map(({ id, label, icon: Icon }) => (
            <button key={id} onClick={() => setView(id)}
              className={`flex items-center gap-3 px-3 h-11 rounded-xl text-sm font-semibold transition whitespace-nowrap ${view === id ? 'bg-ink text-paper' : 'text-mute hover:bg-sunk hover:text-ink'}`}>
              <Icon size={18} />{label}
              {id === 'mesas' && openCount > 0 && (
                <span className={`ml-auto num text-xs rounded-full px-2 py-0.5 ${view === id ? 'bg-paper/20' : 'bg-brand-soft text-brand'}`}>{openCount}</span>
              )}
            </button>
          ))}
        </nav>
        <div className="hidden lg:block space-y-3">
          <div className="flex items-center gap-2 px-2 text-xs font-bold tracking-wide text-brand" role="status">
            <span className={`size-2 rounded-full ${pill === 'ok' ? 'bg-brand live' : 'bg-berry'}`} />{pillLabel[pill]}
          </div>
          <button className="btn btn-ghost w-full" onClick={refresh}><RefreshCw size={16} />Refrescar</button>
          <button className="btn btn-ghost w-full" onClick={() => setDark(!dark)}>{dark ? <Sun size={16} /> : <Moon size={16} />}{dark ? 'Modo claro' : 'Modo oscuro'}</button>
        </div>
      </aside>
      <main className="flex-1 min-w-0 p-5 lg:p-10 max-w-[1400px]">
        {view === 'panel' && (
          <Panel d={{ products, tables, tabs, sales, today, expenses, suppliers, name }} go={(v) => setView(v as View)} />
        )}
        {view === 'mesas' && <Mesas tables={tables} tabs={tabs} products={products} refresh={refresh} notify={showToast} fail={fail} />}
        {view === 'catalogo' && <Catalogo products={products} refresh={refresh} notify={showToast} fail={fail} />}
        {view === 'proveedores' && <Proveedores suppliers={suppliers} refresh={refresh} notify={showToast} fail={fail} />}
        {view === 'gastos' && <Gastos expenses={expenses} refresh={refresh} notify={showToast} fail={fail} />}
        {view === 'conexion' && (
          <Conexion pairing={pairing} tables={tables} products={products} pill={pill} onConnected={handleConnected} />
        )}
        {view === 'personalizar' && (
          <Personalizar
            branding={branding}
            logoSrc={logoSrc}
            hasToken={storedToken() !== ''}
            dark={dark}
            setDark={setDark}
            onSaved={handleBrandingSaved}
            notify={showToast}
            fail={fail}
          />
        )}
      </main>

      {toast && (
        <div role="alert" className="fixed bottom-4 left-1/2 max-w-[90vw] -translate-x-1/2 rounded-xl bg-ink text-paper px-5 py-3 text-sm font-semibold">
          {toast}
        </div>
      )}
    </div>
  )
}
