// Catalogo: live hub catalog. The hub has no category field, so products
// render as one grid; it also has no product delete, so unavailable items
// are hidden from sale with PATCH instead of being removed.
import { useState } from 'react'
import { Plus } from 'lucide-react'
import { api, money, type Product } from '../api'
import { Header, Pill } from './ui'

interface CatalogoProps {
  products: Product[]
  refresh: () => void
  notify: (msg: string) => void
  fail: (err: unknown) => void
}

export default function Catalogo({ products, refresh, notify, fail }: CatalogoProps) {
  const [n, setN] = useState(''); const [p, setP] = useState('')
  const toggle = async (prod: Product) => {
    try {
      await api('PATCH', `/api/products/${encodeURIComponent(prod.id)}`, { available: !prod.available })
      refresh()
    } catch (err) {
      fail(err)
    }
  }
  const add = async () => {
    const name = n.trim()
    const cents = Math.round(parseFloat(p) * 100)
    if (!name || !(cents >= 0)) {
      notify('Escriba nombre y precio válido.')
      return
    }
    try {
      await api('POST', '/api/products', { name, price_cents: cents })
      setN(''); setP('')
      refresh()
    } catch (err) {
      fail(err)
    }
  }
  return (
    <>
      <Header title="Catálogo" sub={`${products.length} productos disponibles para la app móvil`} />
      <div className="card p-4 flex flex-wrap gap-3 mb-6">
        <input className="field flex-1 min-w-48" placeholder="Nombre del producto" value={n} onChange={(e) => setN(e.target.value)} />
        <input className="field w-32 num" type="number" placeholder="Precio" value={p} onChange={(e) => setP(e.target.value)} />
        <button className="btn btn-primary" disabled={!n.trim() || !(parseFloat(p) >= 0)} onClick={add}><Plus size={16} />Agregar producto</button>
      </div>
      {products.length === 0 ? (
        <div className="card p-10 text-center text-mute">Catálogo vacío. Agrega el primer producto arriba.</div>
      ) : (
        <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-3">
          {products.map((x) => (
            <div key={x.id} className="card p-4 flex items-center gap-3 group">
              <div className="size-11 rounded-xl bg-brand-soft text-brand grid place-items-center font-display font-extrabold">{x.name[0]?.toUpperCase()}</div>
              <div className="flex-1 min-w-0"><div className="font-semibold truncate">{x.name}</div><div className="num text-sm text-mute">{money(x.price_cents)}</div></div>
              <button title={x.available ? 'Marcar sin stock' : 'Marcar disponible'} onClick={() => toggle(x)}>
                {x.available ? <Pill>En la app</Pill> : <Pill tone="sun">Sin stock</Pill>}
              </button>
            </div>
          ))}
        </div>
      )}
    </>
  )
}
