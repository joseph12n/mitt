// Mesas: hub-owned tables plus their open tabs (the bill is the tab, the
// table row only carries the label). Every action hits the LAN API and then
// refreshes; the hub derives occupied itself, so there is no local flag.
import { useState } from 'react'
import { Plus, X } from 'lucide-react'
import { api, createTable, deleteTable, labelOf, money, type Product, type Tab, type Table as HubTable } from '../api'
import { Header, Pill } from './ui'

interface MesasProps {
  tables: HubTable[]
  tabs: Tab[]
  products: Product[]
  refresh: () => void
  notify: (msg: string) => void
  fail: (err: unknown) => void
}

export default function Mesas({ tables, tabs, products, refresh, notify, fail }: MesasProps) {
  const [name, setName] = useState('')
  const [free, setFree] = useState('')
  const [pick, setPick] = useState<Record<string, { p: string; q: number }>>({})
  const tabOf = (tableId: string) => tabs.find((t) => t.table_id === tableId)
  const freeTables = tables.filter((t) => !t.occupied)
  const openTables = tables.filter((t) => t.occupied)

  const openTab = async () => {
    if (!free) {
      notify('Elija una mesa libre.')
      return
    }
    try {
      await api('POST', '/api/tabs', { table_id: free })
      setFree('')
      refresh()
    } catch (err) {
      fail(err)
    }
  }

  const create = async () => {
    const label = name.trim()
    if (!label) {
      notify('Escriba el nombre de la mesa.')
      return
    }
    try {
      await createTable(label)
      setName('')
      refresh()
    } catch (err) {
      fail(err)
    }
  }

  const remove = async (t: HubTable) => {
    if (t.occupied) {
      notify('LA MESA TIENE CUENTA ABIERTA.')
      return
    }
    if (!confirm(`Eliminar mesa ${t.label}?`)) return
    try {
      await deleteTable(t.id)
      refresh()
    } catch (err) {
      fail(err)
    }
  }

  const add = async (tab: Tab) => {
    const c = pick[tab.id] ?? { p: '', q: 1 }
    if (!c.p || !(c.q > 0)) {
      notify('Elija un producto y una cantidad válida.')
      return
    }
    try {
      await api('POST', `/api/tabs/${encodeURIComponent(tab.id)}/items`, { product_id: c.p, qty: c.q })
      refresh()
    } catch (err) {
      fail(err)
    }
  }

  const close = async (tab: Tab) => {
    if (!confirm(`Cerrar mesa ${labelOf(tables, tab.table_id)} por ${money(tab.total_cents)}?`)) return
    try {
      const sale = await api<{ table_id: string; total_cents: number }>('POST', `/api/tabs/${encodeURIComponent(tab.id)}/close`)
      notify(`Mesa ${labelOf(tables, sale.table_id)} cobrada: ${money(sale.total_cents)}.`)
      refresh()
    } catch (err) {
      fail(err)
    }
  }

  return (
    <>
      <Header title="Mesas" sub={`${openTables.length} abiertas · ${money(tabs.reduce((a, t) => a + t.total_cents, 0))} en curso`} />

      <div className="card p-4 flex flex-wrap gap-3 items-center mb-4">
        <select className="field w-48" value={free} onChange={(e) => setFree(e.target.value)}>
          <option value="">Mesa libre…</option>
          {freeTables.map((t) => <option key={t.id} value={t.id}>{t.label}</option>)}
        </select>
        <button className="btn btn-primary" disabled={!free} onClick={openTab}>Abrir mesa</button>
        <div className="hidden sm:block w-px h-8 bg-line mx-2" />
        <input className="field w-48" placeholder="Nombre de nueva mesa" value={name} onChange={(e) => setName(e.target.value)} />
        <button className="btn btn-ghost" disabled={!name.trim()} onClick={create}><Plus size={16} />Agregar mesa</button>
        <span className="text-xs text-mute">{freeTables.length ? `${freeTables.length} libres` : 'Sin mesas libres.'}</span>
      </div>

      <div className="card p-4 mb-6">
        <div className="text-xs font-bold uppercase tracking-wide text-mute mb-3">Todas las mesas</div>
        <div className="flex flex-wrap gap-2">
          {tables.map((t) => (
            <div key={t.id} className="flex items-center gap-2 rounded-xl bg-sunk pl-3 pr-1.5 h-10 text-sm">
              <span className={`size-2 rounded-full ${t.occupied ? 'bg-berry' : 'bg-brand'}`} />
              <b>{t.label}</b>
              <span className="text-mute text-xs">{t.occupied ? 'Ocupada' : 'Libre'}</span>
              <button disabled={t.occupied} title={t.occupied ? 'Con cuenta abierta' : 'Eliminar'} onClick={() => remove(t)}
                className="size-7 rounded-lg grid place-items-center text-mute hover:bg-surface hover:text-berry disabled:opacity-30 disabled:hover:bg-transparent disabled:hover:text-mute"><X size={14} /></button>
            </div>
          ))}
          {tables.length === 0 && <span className="text-sm text-mute">Sin mesas registradas.</span>}
        </div>
      </div>

      <div className="grid gap-4 lg:grid-cols-2">
        {openTables.map((t) => {
          const tab = tabOf(t.id)
          if (!tab) return null
          const c = pick[tab.id] ?? { p: '', q: 1 }
          return (
            <article key={t.id} className="card p-5 flex flex-col">
              <div className="flex items-center gap-3 flex-wrap">
                <h3 className="font-display font-bold text-xl">{t.label}</h3>
                <Pill tone="sun">Ocupada</Pill>
              </div>
              <ul className="my-4 text-sm flex-1 min-h-12">
                {(tab.items || []).length === 0 && <li className="text-mute">Sin consumos.</li>}
                {(tab.items || []).map((l) => (
                  <li key={l.product_id} className="flex justify-between py-2 border-b border-line last:border-0">
                    <span><span className="num text-mute mr-2">{l.qty}×</span>{l.name}</span>
                    <span className="num">{money(l.line_total_cents)}</span>
                  </li>
                ))}
              </ul>
              <div className="num text-4xl font-medium tracking-tight mb-4">{money(tab.total_cents)}</div>
              <div className="flex flex-wrap gap-2">
                <select className="field flex-1 min-w-36" value={c.p} onChange={(e) => setPick({ ...pick, [tab.id]: { ...c, p: e.target.value } })}>
                  <option value="">Producto…</option>
                  {products.map((p) => <option key={p.id} value={p.id} disabled={!p.available}>{p.name}</option>)}
                </select>
                <input type="number" min={1} className="field w-20 num" value={c.q} onChange={(e) => setPick({ ...pick, [tab.id]: { ...c, q: Math.max(1, +e.target.value || 1) } })} />
                <button className="btn btn-primary" disabled={!c.p} onClick={() => add(tab)}>Agregar</button>
                <button className="btn btn-danger" onClick={() => close(tab)}>Cerrar</button>
              </div>
            </article>
          )
        })}
        {openTables.length === 0 && <div className="card p-10 text-center text-mute lg:col-span-2">No hay mesas abiertas. Abre una desde el selector de arriba.</div>}
      </div>
    </>
  )
}
