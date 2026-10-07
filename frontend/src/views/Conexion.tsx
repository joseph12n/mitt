// Conexion: mobile pairing plus honest API diagnostics. The QR encodes the
// real pairing_code from GET /api/pairing; the endpoint table lists the REAL
// hub routes (docs/api.md) and the Verificar button probes every safe GET
// with the saved token so staff see what actually answers.
import { useState } from 'react'
import { QRCodeSVG } from 'qrcode.react'
import { Copy, Check, Smartphone, Server, ScanLine } from 'lucide-react'
import { baseUrl, readPairing, storedToken, type Pairing, type Product, type Table as HubTable } from '../api'
import { Header, Pill } from './ui'

export type PillState = 'empty' | 'checking' | 'ok' | 'bad'

interface ConexionProps {
  pairing: Pairing | null
  tables: HubTable[]
  products: Product[]
  pill: PillState
  onConnected: () => void
}

// Every route the hub actually serves. live=true marks the safe GETs the
// Verificar button probes; mutations are documented but never fired.
const endpoints: Array<{ m: string; p: string; live: boolean }> = [
  { m: 'GET', p: '/api/health', live: true },
  { m: 'GET', p: '/api/products', live: true },
  { m: 'POST', p: '/api/products', live: false },
  { m: 'PATCH', p: '/api/products/{id}', live: false },
  { m: 'GET', p: '/api/tables', live: true },
  { m: 'POST', p: '/api/tables', live: false },
  { m: 'DELETE', p: '/api/tables/{id}', live: false },
  { m: 'POST', p: '/api/tabs', live: false },
  { m: 'GET', p: '/api/tabs/open', live: true },
  { m: 'POST', p: '/api/tabs/{id}/items', live: false },
  { m: 'POST', p: '/api/tabs/{id}/close', live: false },
  { m: 'GET', p: '/api/expenses', live: true },
  { m: 'POST', p: '/api/expenses', live: false },
  { m: 'GET', p: '/api/suppliers', live: true },
  { m: 'POST', p: '/api/suppliers', live: false },
  { m: 'PATCH', p: '/api/suppliers/{id}', live: false },
  { m: 'DELETE', p: '/api/suppliers/{id}', live: false },
  { m: 'GET', p: '/api/sales', live: true },
  { m: 'GET', p: '/api/sales/today', live: true },
  { m: 'GET', p: '/api/pairing', live: true },
  { m: 'GET', p: '/api/branding', live: true },
  { m: 'GET', p: '/api/branding/logo', live: true },
  { m: 'PATCH', p: '/api/branding', live: false },
]

export default function Conexion({ pairing, tables, products, pill, onConnected }: ConexionProps) {
  const [tokenInput, setTokenInput] = useState(storedToken())
  const [copied, setCopied] = useState(false)
  const [probe, setProbe] = useState<Record<string, string>>({})
  const [probing, setProbing] = useState(false)
  const url = pairing?.url ?? baseUrl()
  const copy = () => { navigator.clipboard?.writeText(url); setCopied(true); setTimeout(() => setCopied(false), 1500) }

  const saveToken = () => {
    const parsed = readPairing(tokenInput)
    const token = (parsed.token || '').trim()
    if (parsed.base) localStorage.setItem('mitt_base', parsed.base)
    localStorage.setItem('mitt_token', token)
    setTokenInput(token || tokenInput)
    onConnected()
  }

  const verify = async () => {
    setProbing(true)
    const out: Record<string, string> = {}
    const token = storedToken()
    await Promise.all(endpoints.filter((e) => e.live).map(async (e) => {
      const t0 = performance.now()
      try {
        const headers: Record<string, string> = {}
        // Health and branding are public; the rest need the pairing token.
        if (e.p !== '/api/health' && e.p !== '/api/branding' && e.p !== '/api/branding/logo') {
          headers.Authorization = `Bearer ${token}`
        }
        const res = await fetch(baseUrl() + e.p, { headers })
        out[e.p] = res.ok ? `${res.status} · ${Math.round(performance.now() - t0)} ms` : `${res.status} · error`
      } catch {
        out[e.p] = 'sin red'
      }
    }))
    setProbe(out)
    setProbing(false)
  }

  const steps = [
    'Conecta el celular a la misma red Wi-Fi que este PC.',
    'Abre la app de pedidos y toca “Escanear código”.',
    'Apunta la cámara al QR. Mesas y productos se descargan solos.',
  ]

  return (
    <>
      <Header title="Conexión" sub="Enlaza la app móvil con el servidor de este PC"
        right={pill === 'ok'
          ? <Pill><span className="size-2 rounded-full bg-brand live" />Servidor activo</Pill>
          : <Pill tone="mute">Sin token válido</Pill>} />
      <div className="grid gap-4 xl:grid-cols-[420px_1fr]">
        <div className="card p-8 flex flex-col items-center text-center">
          {pairing ? (
            <div className="rounded-3xl bg-white p-5 border border-line shadow-[0_20px_50px_-20px_rgba(0,0,0,.25)]">
              <QRCodeSVG value={pairing.pairing_code} size={220} level="M" fgColor="#16201b" />
            </div>
          ) : (
            <div className="rounded-3xl bg-sunk p-8 text-sm text-mute max-w-60">Guarda un token abajo para ver el código de emparejamiento.</div>
          )}
          <div className="mt-6 num text-lg break-all">{url}</div>
          <p className="text-sm text-mute mt-1">Escanea para conectar el dispositivo</p>
          <button className="btn btn-ghost mt-5" onClick={copy}>{copied ? <Check size={16} /> : <Copy size={16} />}{copied ? 'Copiado' : 'Copiar dirección'}</button>
          <div className="w-full mt-6 text-left">
            <label className="text-xs font-semibold text-mute" htmlFor="token-input">Token o enlace de emparejamiento</label>
            <input id="token-input" className="field w-full mt-1" value={tokenInput} placeholder="Token o enlace mitt://pair?…"
              autoComplete="off" autoCorrect="off" autoCapitalize="off" spellCheck={false}
              onChange={(e) => setTokenInput(e.target.value)} onKeyDown={(e) => { if (e.key === 'Enter') saveToken() }} />
            <button className="btn btn-primary w-full mt-2" onClick={saveToken}>Guardar</button>
          </div>
        </div>

        <div className="space-y-4">
          <div className="card p-6">
            <h2 className="font-display font-bold text-xl mb-4">Cómo conectar</h2>
            <ol className="space-y-4">
              {steps.map((t, i) => (
                <li key={i} className="flex gap-4 items-start">
                  <span className="size-8 shrink-0 rounded-full bg-brand-soft text-brand grid place-items-center num text-sm">{i + 1}</span>
                  <span className="pt-1">{t}</span>
                </li>
              ))}
            </ol>
          </div>
          <div className="grid sm:grid-cols-3 gap-4">
            {[
              { i: Server, l: 'Mesas registradas', v: String(tables.length) },
              { i: ScanLine, l: 'Productos publicados', v: String(products.length) },
              { i: Smartphone, l: 'Conexión', v: pill === 'ok' ? 'Activa' : '—' },
            ].map((k) => (
              <div key={k.l} className="card p-5"><k.i size={18} className="text-brand" /><div className="num text-2xl mt-3">{k.v}</div><div className="text-xs text-mute">{k.l}</div></div>
            ))}
          </div>
          <div className="card p-6">
            <div className="flex items-center justify-between mb-4">
              <h2 className="font-display font-bold text-xl">API disponible</h2>
              <button className="btn btn-primary h-9!" onClick={verify} disabled={probing}>{probing ? 'Verificando…' : 'Verificar'}</button>
            </div>
            <ul className="divide-y divide-line">
              {endpoints.map(({ m, p }) => (
                <li key={p} className="py-2.5 flex items-center gap-3 text-sm">
                  <span className={`num text-xs w-14 text-center rounded-md py-0.5 ${m === 'GET' ? 'bg-brand-soft text-brand' : m === 'DELETE' ? 'bg-berry/15 text-berry' : 'bg-sun/25 text-sunink'}`}>{m}</span>
                  <span className="num truncate">{p}</span>
                  {probe[p] && <span className={`num text-xs ml-auto shrink-0 ${probe[p].startsWith('200') ? 'text-brand' : 'text-berry'}`}>{probe[p]}</span>}
                </li>
              ))}
            </ul>
          </div>
        </div>
      </div>
    </>
  )
}
