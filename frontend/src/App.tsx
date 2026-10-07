import { useCallback, useEffect, useRef, useState } from 'react';
import QRCode from 'react-qr-code';
import {
  api,
  baseUrl,
  checkToken,
  money,
  readPairing,
  storedToken,
  type Expense,
  type Pairing,
  type Product,
  type Tab,
} from './api';
import SpotlightCard from './components/SpotlightCard';
import GlareHover from './components/GlareHover';

type PillState = 'empty' | 'checking' | 'ok' | 'bad';

const pillLabel: Record<PillState, string> = {
  empty: 'SIN TOKEN',
  checking: 'VERIFICANDO…',
  ok: 'CONECTADO',
  bad: 'INVÁLIDO',
};

const inputCls =
  'rounded-lg border border-mitt-raised bg-mitt-raised px-3 py-2 min-h-8 text-mitt-text placeholder:text-mitt-muted focus:outline-2 focus:outline-mitt-accent';
const btnPrimary =
  'rounded-lg bg-mitt-accent px-3 py-2 min-h-8 font-semibold text-mitt-on-accent hover:brightness-110 disabled:opacity-50';
const btnGhost =
  'rounded-lg border border-mitt-muted bg-mitt-raised px-3 py-2 min-h-8 text-mitt-text hover:brightness-125 disabled:opacity-50';
const btnDanger =
  'rounded-lg bg-mitt-danger px-3 py-2 min-h-8 font-semibold text-mitt-text hover:brightness-110 disabled:opacity-50';

export default function App() {
  const [pill, setPill] = useState<PillState>(storedToken() ? 'checking' : 'empty');
  const [tokenInput, setTokenInput] = useState(storedToken());
  const [pairing, setPairing] = useState<Pairing | null>(null);
  const [tabs, setTabs] = useState<Tab[]>([]);
  const [catalog, setCatalog] = useState<Product[]>([]);
  const [expenses, setExpenses] = useState<Expense[]>([]);
  const [toast, setToast] = useState('');
  const [tableName, setTableName] = useState('');
  const [productName, setProductName] = useState('');
  const [productPrice, setProductPrice] = useState('');
  const [expenseDesc, setExpenseDesc] = useState('');
  const [expenseQty, setExpenseQty] = useState('');
  const [expenseCost, setExpenseCost] = useState('');
  const [addSelection, setAddSelection] = useState<Record<string, { productId: string; qty: string }>>({});
  const toastTimer = useRef<number | null>(null);

  const showToast = useCallback((msg: string) => {
    setToast(msg);
    if (toastTimer.current !== null) window.clearTimeout(toastTimer.current);
    toastTimer.current = window.setTimeout(() => setToast(''), 4000);
  }, []);

  const fail = useCallback(
    (err: unknown) => {
      showToast(err instanceof Error ? err.message : 'Ocurrió un error. Intente de nuevo.');
    },
    [showToast],
  );

  const paintPill = useCallback(async () => {
    const token = storedToken();
    if (!token) {
      setPill('empty');
      return;
    }
    setPill('checking');
    const probed = token;
    const ok = await checkToken();
    if (storedToken() !== probed) return; // user typed a new token while probing
    setPill(ok ? 'ok' : 'bad');
  }, []);

  const refresh = useCallback(async () => {
    if (!storedToken()) {
      setPill('empty');
      return;
    }
    paintPill();
    try {
      const [tabsRes, prodsRes, expsRes, pairRes] = await Promise.all([
        api<{ tabs: Tab[] }>('GET', '/api/tabs/open'),
        api<{ products: Product[] }>('GET', '/api/products'),
        api<{ expenses: Expense[] }>('GET', '/api/expenses'),
        api<Pairing>('GET', '/api/pairing').catch(() => null),
      ]);
      setTabs(tabsRes.tabs || []);
      setCatalog(prodsRes.products || []);
      setExpenses(expsRes.expenses || []);
      setPairing(pairRes);
    } catch (err) {
      fail(err);
    }
  }, [paintPill, fail]);

  useEffect(() => {
    paintPill();
    if (storedToken()) refresh();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const saveToken = () => {
    const parsed = readPairing(tokenInput);
    const token = (parsed.token || '').trim();
    if (parsed.base) localStorage.setItem('mitt_base', parsed.base);
    localStorage.setItem('mitt_token', token);
    setTokenInput(token || tokenInput);
    paintPill();
    if (token) refresh();
    else showToast('Guarde un token para conectar.');
  };

  const openTable = async () => {
    const label = tableName.trim();
    if (!label) {
      showToast('Escriba el nombre de la mesa.');
      return;
    }
    try {
      await api('POST', '/api/tabs', { table_id: label });
      setTableName('');
      refresh();
    } catch (err) {
      fail(err);
    }
  };

  const addItem = async (tab: Tab) => {
    const sel = addSelection[tab.id] || { productId: '', qty: '1' };
    const qty = parseInt(sel.qty, 10);
    if (!sel.productId || !(qty > 0)) {
      showToast('Elija un producto y una cantidad válida.');
      return;
    }
    try {
      await api('POST', `/api/tabs/${encodeURIComponent(tab.id)}/items`, {
        product_id: sel.productId,
        qty,
      });
      refresh();
    } catch (err) {
      fail(err);
    }
  };

  const closeTab = async (tab: Tab) => {
    if (!confirm(`Cerrar mesa ${tab.table_id} por ${money(tab.total_cents)}?`)) return;
    try {
      const sale = await api<{ table_id: string; total_cents: number }>(
        'POST',
        `/api/tabs/${encodeURIComponent(tab.id)}/close`,
      );
      showToast(`Mesa ${sale.table_id} cobrada: ${money(sale.total_cents)}.`);
      refresh();
    } catch (err) {
      fail(err);
    }
  };

  const toggleProduct = async (p: Product) => {
    try {
      await api('PATCH', `/api/products/${encodeURIComponent(p.id)}`, { available: !p.available });
      refresh();
    } catch (err) {
      fail(err);
    }
  };

  const addProduct = async () => {
    const name = productName.trim();
    const cents = Math.round(parseFloat(productPrice) * 100);
    if (!name || !(cents >= 0)) {
      showToast('Escriba nombre y precio válido.');
      return;
    }
    try {
      await api('POST', '/api/products', { name, price_cents: cents });
      setProductName('');
      setProductPrice('');
      refresh();
    } catch (err) {
      fail(err);
    }
  };

  const addExpense = async () => {
    const desc = expenseDesc.trim();
    const qty = parseFloat(expenseQty);
    const cents = Math.round(parseFloat(expenseCost) * 100);
    if (!desc || !(qty > 0) || !(cents >= 0)) {
      showToast('Complete descripción, cantidad y costo.');
      return;
    }
    try {
      await api('POST', '/api/expenses', { description: desc, qty, cost_cents: cents });
      setExpenseDesc('');
      setExpenseQty('');
      setExpenseCost('');
      refresh();
    } catch (err) {
      fail(err);
    }
  };

  return (
    <div className="min-h-screen">
      <div className="mitt-ambient" aria-hidden="true" />
      <div className="mx-auto max-w-4xl px-4 py-4">
        <header className="flex flex-wrap items-center justify-between gap-3">
          <h1 className="text-[32px] font-bold text-mitt-text">mitt · Panel del bar</h1>
          <button type="button" className={btnGhost} onClick={refresh}>
            REFRESCAR
          </button>
        </header>

        <section>
          <h2 className="mb-2 mt-6 text-[20px] font-semibold text-mitt-text">Conexión</h2>
          <GlareHover className="p-3">
            <div className="flex w-full flex-wrap items-center gap-3 p-1">
              {pairing && (
                <div className="flex items-center gap-3">
                  <div className="rounded-lg bg-white p-2">
                    <QRCode value={pairing.pairing_code} size={112} />
                  </div>
                  <p className="money max-w-55 text-xs break-all text-mitt-muted">{pairing.url}</p>
                </div>
              )}
              <span
                className={`rounded px-3 py-1 text-xs font-semibold ${
                  pill === 'ok' ? 'bg-mitt-success text-mitt-bg' : 'bg-mitt-danger text-mitt-text'
                }`}
              >
                {pillLabel[pill]}
              </span>
              <input
                className={`${inputCls} min-w-60 flex-1`}
                type="text"
                placeholder="Token o enlace mitt://pair?…"
                value={tokenInput}
                autoComplete="off"
                onChange={(e) => setTokenInput(e.target.value)}
              />
              <button type="button" className={btnPrimary} onClick={saveToken}>
                GUARDAR
              </button>
            </div>
          </GlareHover>
          <p className="mt-1 text-xs text-mitt-muted">
            El token se guarda en este navegador. También puede pegar el enlace de emparejamiento
            completo. Base: <span className="money">{baseUrl()}</span>
          </p>
        </section>

        <section>
          <h2 className="mb-2 mt-6 text-[20px] font-semibold text-mitt-text">MESAS</h2>
          <div className="mb-2 flex flex-wrap items-center gap-2 rounded-xl bg-mitt-surface p-3">
            <input
              className={inputCls}
              type="text"
              placeholder="Mesa (ej. T1)"
              size={12}
              value={tableName}
              onChange={(e) => setTableName(e.target.value)}
            />
            <button type="button" className={btnPrimary} onClick={openTable}>
              ABRIR MESA
            </button>
          </div>
          {tabs.length === 0 ? (
            <p className="text-xs text-mitt-muted">Sin mesas abiertas.</p>
          ) : (
            <div className="grid grid-cols-1 gap-2 sm:grid-cols-2">
              {tabs.map((t) => {
                const sel = addSelection[t.id] || { productId: '', qty: '1' };
                return (
                  <SpotlightCard
                    key={t.id}
                    className="border-l-4 border-l-mitt-accent bg-mitt-raised p-3"
                  >
                    <div className="relative">
                      <strong className="text-mitt-text">Mesa {t.table_id}</strong>{' '}
                      <span className="text-xs text-mitt-muted">
                        Ocupada · <span className="money">{money(t.total_cents)}</span>
                      </span>
                      <ul className="my-2 list-none p-0">
                        {(t.items || []).length === 0 && (
                          <li className="text-xs text-mitt-muted">Sin consumos.</li>
                        )}
                        {(t.items || []).map((it) => (
                          <li
                            key={it.product_id}
                            className="border-b border-mitt-raised py-1 text-mitt-text"
                          >
                            <span className="money">
                              {it.qty} × {it.name}
                            </span>{' '}
                            · <span className="money">{money(it.line_total_cents)}</span>
                          </li>
                        ))}
                      </ul>
                      <div className="money text-[32px] font-bold text-mitt-text">
                        {money(t.total_cents)}
                      </div>
                      <div className="mt-2 flex flex-wrap items-center gap-2">
                        <select
                          className={inputCls}
                          value={sel.productId}
                          onChange={(e) =>
                            setAddSelection((s) => ({
                              ...s,
                              [t.id]: { productId: e.target.value, qty: sel.qty },
                            }))
                          }
                        >
                          <option value="">Producto…</option>
                          {catalog.map((p) => (
                            <option key={p.id} value={p.id} disabled={!p.available}>
                              {p.name} · {money(p.price_cents)}
                            </option>
                          ))}
                        </select>
                        <input
                          className={`${inputCls} w-16`}
                          type="number"
                          value={sel.qty}
                          min={1}
                          step={1}
                          onChange={(e) =>
                            setAddSelection((s) => ({
                              ...s,
                              [t.id]: { productId: sel.productId, qty: e.target.value },
                            }))
                          }
                        />
                        <button type="button" className={btnPrimary} onClick={() => addItem(t)}>
                          AGREGAR
                        </button>
                        <button type="button" className={btnDanger} onClick={() => closeTab(t)}>
                          CERRAR
                        </button>
                      </div>
                    </div>
                  </SpotlightCard>
                );
              })}
            </div>
          )}
        </section>

        <section>
          <h2 className="mb-2 mt-6 text-[20px] font-semibold text-mitt-text">CATÁLOGO</h2>
          {catalog.length === 0 ? (
            <p className="text-xs text-mitt-muted">Sin productos.</p>
          ) : (
            <div className="flex flex-col gap-2">
              {catalog.map((p) => (
                <SpotlightCard key={p.id} className="p-3">
                  <div className="relative flex flex-wrap items-center gap-2">
                    <strong className="text-mitt-text">{p.name}</strong>{' '}
                    <span className="money text-mitt-text">{money(p.price_cents)}</span>{' '}
                    {p.available ? (
                      <span className="rounded bg-mitt-success px-2 py-0.5 text-xs font-semibold text-mitt-bg">
                        DISPONIBLE
                      </span>
                    ) : (
                      <span className="rounded bg-mitt-danger px-2 py-0.5 text-xs font-semibold text-mitt-text">
                        SIN STOCK
                      </span>
                    )}
                    <button
                      type="button"
                      className={`${btnGhost} ml-auto`}
                      onClick={() => toggleProduct(p)}
                    >
                      {p.available ? 'MARCAR SIN STOCK' : 'MARCAR DISPONIBLE'}
                    </button>
                  </div>
                </SpotlightCard>
              ))}
            </div>
          )}
          <div className="mt-2 flex flex-wrap items-center gap-2 rounded-xl bg-mitt-surface p-3">
            <input
              className={inputCls}
              type="text"
              placeholder="Nombre"
              size={16}
              value={productName}
              onChange={(e) => setProductName(e.target.value)}
            />
            <input
              className={inputCls}
              type="number"
              placeholder="Precio"
              min={0}
              step="0.01"
              size={8}
              value={productPrice}
              onChange={(e) => setProductPrice(e.target.value)}
            />
            <button type="button" className={btnPrimary} onClick={addProduct}>
              AGREGAR
            </button>
          </div>
        </section>

        <section>
          <h2 className="mb-2 mt-6 text-[20px] font-semibold text-mitt-text">GASTOS</h2>
          {expenses.length === 0 ? (
            <p className="text-xs text-mitt-muted">Sin gastos.</p>
          ) : (
            <SpotlightCard className="p-3">
              <ul className="relative my-2 list-none p-0">
                {expenses.map((e) => (
                  <li
                    key={e.id}
                    className="border-b border-mitt-raised py-1 text-mitt-text"
                  >
                    {e.description} · {e.qty} u ·{' '}
                    <span className="money">{money(e.cost_cents)}</span>
                  </li>
                ))}
              </ul>
            </SpotlightCard>
          )}
          <div className="mt-2 flex flex-wrap items-center gap-2 rounded-xl bg-mitt-surface p-3">
            <input
              className={inputCls}
              type="text"
              placeholder="Descripción"
              size={16}
              value={expenseDesc}
              onChange={(e) => setExpenseDesc(e.target.value)}
            />
            <input
              className={inputCls}
              type="number"
              placeholder="Cantidad"
              min={0}
              step="any"
              size={8}
              value={expenseQty}
              onChange={(e) => setExpenseQty(e.target.value)}
            />
            <input
              className={inputCls}
              type="number"
              placeholder="Costo"
              min={0}
              step="0.01"
              size={8}
              value={expenseCost}
              onChange={(e) => setExpenseCost(e.target.value)}
            />
            <button type="button" className={btnPrimary} onClick={addExpense}>
              AGREGAR
            </button>
          </div>
        </section>
      </div>

      {toast && (
        <div
          role="alert"
          className="fixed bottom-4 left-1/2 max-w-[90vw] -translate-x-1/2 rounded-lg bg-mitt-danger px-5 py-3 text-mitt-text"
        >
          {toast}
        </div>
      )}
    </div>
  );
}
