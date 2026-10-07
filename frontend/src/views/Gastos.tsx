// Gastos: live list plus create. The hub exposes GET + POST /api/expenses
// only, so rows have no delete control; each line totals qty x unit cost.
import { useState } from 'react'
import { Plus } from 'lucide-react'
import { api, money, type Expense } from '../api'
import { Header } from './ui'

interface GastosProps {
  expenses: Expense[]
  refresh: () => void
  notify: (msg: string) => void
  fail: (err: unknown) => void
}

const fmtDate = (iso: string) => {
  const d = new Date(iso)
  return Number.isNaN(d.getTime()) ? '—' : d.toLocaleString('es-AR', { day: '2-digit', month: '2-digit', hour: '2-digit', minute: '2-digit' })
}

export default function Gastos({ expenses, refresh, notify, fail }: GastosProps) {
  const [c, setC] = useState(''); const [q, setQ] = useState(''); const [a, setA] = useState('')
  const total = expenses.reduce((x, e) => x + Math.round(e.qty * e.cost_cents), 0)
  const add = async () => {
    const description = c.trim()
    const qty = parseFloat(q)
    const cents = Math.round(parseFloat(a) * 100)
    if (!description || !(qty > 0) || !(cents >= 0)) {
      notify('Complete descripción, cantidad y costo.')
      return
    }
    try {
      await api('POST', '/api/expenses', { description, qty, cost_cents: cents })
      setC(''); setQ(''); setA('')
      refresh()
    } catch (err) {
      fail(err)
    }
  }
  return (
    <>
      <Header title="Gastos" sub="Egresos registrados del servicio" right={<div className="num text-3xl font-medium">{money(total)}</div>} />
      <div className="card p-4 flex flex-wrap gap-3 mb-6">
        <input className="field flex-1 min-w-48" placeholder="Concepto" value={c} onChange={(e) => setC(e.target.value)} />
        <input className="field w-28 num" type="number" placeholder="Cant." value={q} onChange={(e) => setQ(e.target.value)} />
        <input className="field w-40 num" type="number" placeholder="Costo unitario" value={a} onChange={(e) => setA(e.target.value)} />
        <button className="btn btn-primary" disabled={!c.trim() || !(parseFloat(q) > 0) || !(parseFloat(a) >= 0)} onClick={add}><Plus size={16} />Registrar gasto</button>
      </div>
      <div className="card divide-y divide-line">
        {expenses.length === 0 && <div className="px-5 py-8 text-center text-mute text-sm">Sin gastos registrados.</div>}
        {expenses.map((e) => (
          <div key={e.id} className="flex items-center gap-4 px-5 py-4">
            <div className="flex-1"><div className="font-semibold">{e.description}</div><div className="text-xs text-mute num">{fmtDate(e.date)} · {e.qty} × {money(e.cost_cents)}</div></div>
            <div className="num">{money(Math.round(e.qty * e.cost_cents))}</div>
          </div>
        ))}
      </div>
    </>
  )
}
