// Panel: service summary from LIVE hub data. Sales history feeds the
// charts, the today summary feeds the KPIs, open tabs feed the in-progress
// total. Money is integer cents; hours come from closed_at timestamps.
import { Area, AreaChart, Bar, BarChart, CartesianGrid, Cell, ResponsiveContainer, Tooltip, XAxis, YAxis } from 'recharts'
import { ArrowUpRight, TrendingUp, ReceiptText, Armchair, Wallet } from 'lucide-react'
import { labelOf, money, type Expense, type Product, type Sale, type Supplier, type Tab, type Table as HubTable, type TodaySummary } from '../api'
import { Header, Pill } from './ui'

const tip = { borderRadius: 12, border: '1px solid var(--color-line)', background: 'var(--color-surface)', color: 'var(--color-ink)' }

export interface PanelData {
  products: Product[]
  tables: HubTable[]
  tabs: Tab[]
  sales: Sale[]
  today: TodaySummary | null
  expenses: Expense[]
  suppliers: Supplier[]
  name: string
}

const hourOf = (iso: string) => {
  const d = new Date(iso)
  return Number.isNaN(d.getTime()) ? -1 : d.getHours()
}
const fmtHour = (iso: string) => {
  const d = new Date(iso)
  return Number.isNaN(d.getTime()) ? '—' : d.toLocaleTimeString('es-AR', { hour: '2-digit', minute: '2-digit' })
}
// Same local-day check the staff sees on the wall clock (hub days are UTC;
// the Panel groups by the bar's own calendar day).
const isToday = (iso: string) => {
  const d = new Date(iso)
  const n = new Date()
  return !Number.isNaN(d.getTime()) && d.getFullYear() === n.getFullYear() && d.getMonth() === n.getMonth() && d.getDate() === n.getDate()
}

export default function Panel({ d, go }: { d: PanelData; go: (v: string) => void }) {
  const price = (id: string) => d.products.find((p) => p.id === id)?.price_cents ?? 0
  const openTotal = d.tabs.reduce((a, t) => a + (t.items || []).reduce((x, l) => x + l.qty * price(l.product_id), 0), 0)
  const sold = d.today?.total_cents ?? 0
  const orderCount = d.today?.count ?? 0
  const avg = orderCount ? Math.round(sold / orderCount) : 0
  // Expense rows carry qty x unit cost; the Panel totals the full line.
  const spent = d.expenses.filter((e) => isToday(e.date)).reduce((a, e) => a + Math.round(e.qty * e.cost_cents), 0)

  const hours = Array.from({ length: 14 }, (_, i) => {
    const h = i + 10
    return { h: `${h}h`, v: d.sales.filter((s) => hourOf(s.closed_at) === h).reduce((a, s) => a + s.total_cents, 0) }
  })
  const byTable = Object.entries(d.sales.reduce<Record<string, { total: number; n: number }>>((m, s) => {
    const label = labelOf(d.tables, s.table_id)
    m[label] = { total: (m[label]?.total ?? 0) + s.total_cents, n: (m[label]?.n ?? 0) + 1 }
    return m
  }, {})).map(([table, v]) => ({ table, ...v })).sort((a, b) => b.total - a.total)
  const prods: Record<string, number> = {}
  d.sales.forEach((s) => (s.items || []).forEach((i) => (prods[i.name] = (prods[i.name] ?? 0) + i.qty)))
  const top = Object.entries(prods).sort((a, b) => b[1] - a[1]).slice(0, 5)
  const maxTop = top[0]?.[1] ?? 1

  const kpis = [
    { label: 'Ventas cobradas', value: money(sold), icon: TrendingUp, note: 'acumulado de hoy', hero: true },
    { label: 'Órdenes', value: String(orderCount), icon: ReceiptText, note: 'cerradas hoy' },
    { label: 'Ticket promedio', value: money(avg), icon: Wallet, note: 'por orden' },
    { label: 'En mesas abiertas', value: money(openTotal), icon: Armchair, note: `${d.tabs.length} mesas en curso` },
  ]

  return (
    <>
      <Header title="Panel" sub={`Resumen del servicio de hoy · ${d.name}`}
        right={<button className="btn btn-primary" onClick={() => go('conexion')}>Conectar app móvil <ArrowUpRight size={16} /></button>} />

      <section className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
        {kpis.map((k) => (
          <div key={k.label} className={`rounded-[1.25rem] p-5 border ${k.hero ? 'bg-brand text-on-brand border-transparent' : 'card'}`}>
            <div className="flex items-center justify-between">
              <span className={`text-sm font-medium ${k.hero ? 'text-on-brand/80' : 'text-mute'}`}>{k.label}</span>
              <k.icon size={18} className={k.hero ? 'text-on-brand/80' : 'text-mute'} />
            </div>
            <div className="num text-2xl xl:text-[1.65rem] font-medium mt-4 tracking-tight">{k.value}</div>
            <div className={`text-xs mt-1 ${k.hero ? 'text-on-brand/70' : 'text-mute'}`}>{k.note}</div>
          </div>
        ))}
      </section>

      <section className="grid gap-4 mt-4 xl:grid-cols-[1.7fr_1fr]">
        <div className="card p-6">
          <div className="flex items-center justify-between mb-4">
            <h2 className="font-display font-bold text-xl">Ventas por hora</h2>
            <Pill tone="mute">Hoy</Pill>
          </div>
          <div className="h-64">
            <ResponsiveContainer>
              <AreaChart data={hours} margin={{ left: -10, right: 4, top: 8 }}>
                <defs><linearGradient id="g" x1="0" y1="0" x2="0" y2="1"><stop offset="0%" stopColor="var(--color-brand)" stopOpacity={0.25} /><stop offset="100%" stopColor="var(--color-brand)" stopOpacity={0} /></linearGradient></defs>
                <CartesianGrid vertical={false} stroke="var(--color-line)" />
                <XAxis dataKey="h" tickLine={false} axisLine={false} tick={{ fontSize: 12, fill: 'var(--color-mute)' }} />
                <YAxis tickLine={false} axisLine={false} tick={{ fontSize: 12, fill: 'var(--color-mute)' }} tickFormatter={(v: number) => `${v / 100}`} />
                <Tooltip formatter={(v) => money(Number(v))} contentStyle={tip} />
                <Area type="monotone" dataKey="v" name="Ventas" stroke="var(--color-brand)" strokeWidth={2.5} fill="url(#g)" />
              </AreaChart>
            </ResponsiveContainer>
          </div>
        </div>

        <div className="card p-6">
          <h2 className="font-display font-bold text-xl mb-4">Más pedidos</h2>
          {top.length === 0 ? (
            <p className="text-sm text-mute">Todavía no hay ventas registradas.</p>
          ) : (
            <ul className="space-y-4">
              {top.map(([n, q], i) => (
                <li key={n}>
                  <div className="flex justify-between text-sm mb-1.5"><span className="font-semibold">{i + 1}. {n}</span><span className="num text-mute">{q} u.</span></div>
                  <div className="h-2 rounded-full bg-sunk"><div className="h-full rounded-full bg-brand" style={{ width: `${(q / maxTop) * 100}%` }} /></div>
                </li>
              ))}
            </ul>
          )}
        </div>
      </section>

      <section className="grid gap-4 mt-4 xl:grid-cols-[1fr_1.3fr]">
        <div className="card p-6">
          <div className="flex items-center justify-between mb-1">
            <h2 className="font-display font-bold text-xl">Compras por mesa</h2>
            <span className="text-xs text-mute">{byTable.length} mesas</span>
          </div>
          <p className="text-sm text-mute mb-3">Total cobrado de cada mesa</p>
          <div className="h-64">
            <ResponsiveContainer>
              <BarChart data={byTable} layout="vertical" margin={{ left: 10, right: 10 }}>
                <XAxis type="number" hide />
                <YAxis type="category" dataKey="table" tickLine={false} axisLine={false} width={80} tick={{ fontSize: 12, fill: 'var(--color-ink)' }} />
                <Tooltip cursor={{ fill: 'transparent' }} formatter={(v) => money(Number(v))} contentStyle={tip} />
                <Bar dataKey="total" name="Total" radius={8} barSize={16}>
                  {byTable.map((_, i) => <Cell key={i} fill={i === 0 ? 'var(--color-sun)' : 'var(--color-brand)'} />)}
                </Bar>
              </BarChart>
            </ResponsiveContainer>
          </div>
        </div>

        <div className="card p-6 overflow-x-auto">
          <h2 className="font-display font-bold text-xl mb-4">Órdenes recientes</h2>
          {d.sales.length === 0 ? (
            <p className="text-sm text-mute">Aún no hay ventas. Cierra la primera cuenta desde Mesas.</p>
          ) : (
            <table className="w-full text-sm min-w-[480px]">
              <thead><tr className="text-left text-mute text-xs uppercase tracking-wide"><th className="pb-3 font-semibold">Mesa</th><th className="pb-3 font-semibold">Hora</th><th className="pb-3 font-semibold">Ítems</th><th className="pb-3 font-semibold text-right">Total</th></tr></thead>
              <tbody>
                {d.sales.slice(0, 7).map((s) => (
                  <tr key={s.id} className="border-t border-line">
                    <td className="py-3 font-semibold">{labelOf(d.tables, s.table_id)}</td>
                    <td className="py-3 num text-mute">{fmtHour(s.closed_at)}</td>
                    <td className="py-3 num text-mute">{(s.items || []).reduce((a, i) => a + i.qty, 0)}</td>
                    <td className="py-3 num text-right">{money(s.total_cents)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
          <div className="mt-3 text-xs text-mute">Gastos de hoy: <span className="num">{money(spent)}</span> · Neto: <span className="num text-brand font-medium">{money(sold - spent)}</span></div>
        </div>
      </section>

      <section className="card p-6 mt-4">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div>
            <h2 className="font-display font-bold text-xl">Proveedores</h2>
            <p className="text-sm text-mute">{d.suppliers.length === 0 ? 'Aún no hay proveedores registrados.' : `${d.suppliers.length} registrados · ${d.suppliers.slice(0, 3).map((s) => s.name).join(', ')}${d.suppliers.length > 3 ? '…' : ''}`}</p>
          </div>
          <button className="btn btn-ghost" onClick={() => go('proveedores')}>Ver proveedores <ArrowUpRight size={16} /></button>
        </div>
      </section>
    </>
  )
}
