// Proveedores: live CRUD against GET/POST/PATCH/DELETE /api/suppliers.
// The hub stores a short record (name, phone, note) with no per-supplier
// catalog or purchase ledger, so the view edits exactly those fields.
import { useEffect, useState } from 'react'
import { Plus, Trash2, Phone, Truck } from 'lucide-react'
import { api, type Supplier } from '../api'
import { Header, Pill } from './ui'

interface ProveedoresProps {
  suppliers: Supplier[]
  refresh: () => void
  notify: (msg: string) => void
  fail: (err: unknown) => void
}

export default function Proveedores({ suppliers, refresh, notify, fail }: ProveedoresProps) {
  const [sel, setSel] = useState(suppliers[0]?.id ?? '')
  const [f, setF] = useState({ name: '', phone: '', note: '' })
  const [edit, setEdit] = useState({ name: '', phone: '', note: '' })
  const sup = suppliers.find((x) => x.id === sel)

  useEffect(() => {
    if (sup) setEdit({ name: sup.name, phone: sup.phone, note: sup.note })
  }, [sup?.id]) // eslint-disable-line react-hooks/exhaustive-deps

  useEffect(() => {
    if (suppliers.length > 0 && !suppliers.some((x) => x.id === sel)) setSel(suppliers[0].id)
  }, [suppliers, sel])

  const addSup = async () => {
    const name = f.name.trim()
    if (!name) {
      notify('Escriba el nombre del proveedor.')
      return
    }
    try {
      const created = await api<Supplier>('POST', '/api/suppliers', { name, phone: f.phone, note: f.note })
      setSel(created.id)
      setF({ name: '', phone: '', note: '' })
      refresh()
    } catch (err) {
      fail(err)
    }
  }

  const saveSup = async () => {
    if (!sup) return
    const name = edit.name.trim()
    if (!name) {
      notify('El nombre no puede quedar vacío.')
      return
    }
    try {
      await api('PATCH', `/api/suppliers/${encodeURIComponent(sup.id)}`, { name, phone: edit.phone, note: edit.note })
      refresh()
    } catch (err) {
      fail(err)
    }
  }

  const delSup = async () => {
    if (!sup) return
    if (!confirm(`Eliminar proveedor ${sup.name}?`)) return
    try {
      await api('DELETE', `/api/suppliers/${encodeURIComponent(sup.id)}`)
      setSel('')
      refresh()
    } catch (err) {
      fail(err)
    }
  }

  const dirty = sup ? edit.name.trim() !== sup.name || edit.phone !== sup.phone || edit.note !== sup.note : false

  return (
    <>
      <Header title="Proveedores" sub={`${suppliers.length} proveedores registrados`} />

      <div className="card p-4 flex flex-wrap gap-3 mb-6">
        <input className="field flex-1 min-w-44" placeholder="Nombre del proveedor" value={f.name} onChange={(e) => setF({ ...f, name: e.target.value })} />
        <input className="field w-44" placeholder="Teléfono" value={f.phone} onChange={(e) => setF({ ...f, phone: e.target.value })} />
        <input className="field flex-1 min-w-44" placeholder="Nota (rubro, horario…)" value={f.note} onChange={(e) => setF({ ...f, note: e.target.value })} />
        <button className="btn btn-primary" disabled={!f.name.trim()} onClick={addSup}><Plus size={16} />Agregar proveedor</button>
      </div>

      <div className="grid gap-4 xl:grid-cols-[340px_1fr]">
        <div className="space-y-3">
          {suppliers.map((x) => (
            <button key={x.id} onClick={() => setSel(x.id)} className={`card w-full text-left p-4 transition ${sel === x.id ? 'ring-2 ring-brand' : 'hover:bg-sunk'}`}>
              <div className="flex items-center gap-3">
                <div className="size-10 rounded-xl bg-brand-soft text-brand grid place-items-center"><Truck size={18} /></div>
                <div className="flex-1 min-w-0"><div className="font-semibold truncate">{x.name}</div><div className="text-xs text-mute">{x.phone || 'Sin teléfono'}</div></div>
              </div>
              {x.note && <div className="mt-2 text-sm text-mute truncate">{x.note}</div>}
            </button>
          ))}
          {suppliers.length === 0 && <div className="card p-8 text-center text-mute">Agrega tu primer proveedor.</div>}
        </div>

        {sup ? (
          <div className="card p-6">
            <div className="flex flex-wrap items-start gap-3">
              <div>
                <h2 className="font-display font-bold text-2xl">{sup.name}</h2>
                <div className="flex items-center gap-3 text-sm text-mute mt-1">
                  {sup.phone ? <span className="flex items-center gap-1"><Phone size={13} />{sup.phone}</span> : <Pill tone="mute">Sin teléfono</Pill>}
                </div>
              </div>
              <button className="btn btn-danger ml-auto" onClick={delSup}><Trash2 size={15} />Eliminar</button>
            </div>
            <div className="grid gap-3 mt-6">
              <label className="block text-sm font-semibold">Nombre
                <input className="field w-full mt-2" value={edit.name} maxLength={80} onChange={(e) => setEdit({ ...edit, name: e.target.value })} />
              </label>
              <label className="block text-sm font-semibold">Teléfono
                <input className="field w-full mt-2" value={edit.phone} maxLength={40} onChange={(e) => setEdit({ ...edit, phone: e.target.value })} />
              </label>
              <label className="block text-sm font-semibold">Nota
                <textarea className="field w-full mt-2 h-auto! py-2 min-h-20" value={edit.note} maxLength={200} rows={3} onChange={(e) => setEdit({ ...edit, note: e.target.value })} />
              </label>
            </div>
            <div className="mt-4">
              <button className="btn btn-primary" disabled={!dirty || !edit.name.trim()} onClick={saveSup}>Guardar cambios</button>
            </div>
          </div>
        ) : <div className="card p-10 text-center text-mute">Selecciona o agrega un proveedor.</div>}
      </div>
    </>
  )
}
