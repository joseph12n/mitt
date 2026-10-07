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
import CountUp from './components/CountUp';
import Sidebar, { type SectionId } from './components/Sidebar';
import Personalizar from './components/Personalizar';
import {
  applyBranding,
  DEFAULT_BRANDING,
  fetchPublicBranding,
  logoUrl,
  type BrandingDTO,
} from './components/branding';
import { btnDanger, btnGhost, btnPrimary, cardCls, inputCls, sectionTitle } from './components/ui';

type PillState = 'empty' | 'checking' | 'ok' | 'bad';
type Theme = 'light' | 'dark';

const pillLabel: Record<PillState, string> = {
  empty: 'SIN TOKEN',
  checking: 'VERIFICANDO…',
  ok: 'CONECTADO',
  bad: 'INVÁLIDO',
};

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

  // Sidebar shell + runtime identity state.
  const [section, setSection] = useState<SectionId>('panel');
  const [theme, setTheme] = useState<Theme>(() =>
    typeof document !== 'undefined' && document.documentElement.dataset.theme === 'light'
      ? 'light'
      : 'dark',
  );
  const [branding, setBranding] = useState<BrandingDTO>(DEFAULT_BRANDING);
  const [logoSrc, setLogoSrc] = useState<string | null>(null);

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
    if (storedToken() !== probed) return; // user saved a new token while probing
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

  // Public identity paints over the theme tokens; a failed fetch keeps the
  // built-in dark defaults working (offline-safe).
  useEffect(() => {
    let live = true;
    fetchPublicBranding(baseUrl()).then((b) => {
      if (!live || !b) return;
      setBranding(b);
      applyBranding(b);
      setLogoSrc(b.has_logo ? logoUrl(baseUrl(), b.updated_at) : null);
    });
    return () => {
      live = false;
    };
  }, []);

  const toggleTheme = useCallback(() => {
    setTheme((t) => {
      const next: Theme = t === 'dark' ? 'light' : 'dark';
      try {
        localStorage.setItem('mitt_theme', next);
      } catch {
        // Private mode: the session keeps the default, nothing breaks.
      }
      document.documentElement.dataset.theme = next;
      return next;
    });
  }, []);

  const handleBrandingSaved = useCallback((next: BrandingDTO) => {
    setBranding(next);
    applyBranding(next);
    setLogoSrc(next.has_logo ? logoUrl(baseUrl(), next.updated_at) : null);
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

  // KPI strip reads only already-fetched state: no extra requests.
  const openCount = tabs.length;
  const inProgressCents = tabs.reduce((sum, t) => sum + t.total_cents, 0);
  const availableCount = catalog.filter((p) => p.available).length;

  return (
    <div className="mitt-shell min-h-screen">
      <div className="mitt-ambient" aria-hidden="true" />
      <Sidebar
        branding={branding}
        logoSrc={logoSrc}
        onLogoError={() => setLogoSrc(null)}
        section={section}
        onSelect={setSection}
        pillText={pillLabel[pill]}
        pillOk={pill === 'ok'}
        theme={theme}
        onToggleTheme={toggleTheme}
        onRefresh={refresh}
      />

      <main className="mitt-main">
        <div className="mitt-content">
          {section === 'panel' && (
            <section aria-label="Panel">
              <div className="mb-2 mt-2 flex flex-wrap items-end justify-between gap-2">
                <h2 className={sectionTitle}>Panel</h2>
                <p className="money text-xs text-mitt-muted">
                  {openCount === 0
                    ? 'Sin mesas abiertas'
                    : `${openCount} abierta${openCount === 1 ? '' : 's'} · ${money(inProgressCents)} en curso`}
                </p>
              </div>
              <div className="grid grid-cols-1 gap-2 sm:grid-cols-3">
                <div className={cardCls}>
                  <p className="text-xs text-mitt-muted">Mesas abiertas</p>
                  <p className="money text-[32px] font-bold text-mitt-text">
                    <CountUp value={openCount} format={(n) => String(n)} durationMs={400} />
                  </p>
                </div>
                <div className={cardCls}>
                  <p className="text-xs text-mitt-muted">En curso</p>
                  <p className="money text-[32px] font-bold text-mitt-text">
                    <CountUp value={inProgressCents} format={money} />
                  </p>
                </div>
                <div className={cardCls}>
                  <p className="text-xs text-mitt-muted">Productos disponibles</p>
                  <p className="money text-[32px] font-bold text-mitt-text">
                    <CountUp
                      value={availableCount}
                      format={(n) => `${n} de ${catalog.length}`}
                      durationMs={400}
                    />
                  </p>
                </div>
              </div>

              <div className="mt-4 grid grid-cols-1 gap-4 lg:grid-cols-2">
                <div className={cardCls}>
                  <h3 className="mb-2 font-semibold text-mitt-text">Emparejar móvil</h3>
                  {pairing ? (
                    <div className="flex items-center gap-4">
                      <div className="rounded-xl bg-white p-3">
                        <QRCode value={pairing.pairing_code} size={168} />
                      </div>
                      <div className="min-w-0">
                        <p className="font-semibold text-mitt-text">Escanee para emparejar</p>
                        <p className="money mt-1 max-w-55 text-xs break-all text-mitt-muted">
                          {pairing.url}
                        </p>
                      </div>
                    </div>
                  ) : (
                    <p className="text-sm text-mitt-muted">
                      Conecte con un token para ver el código QR de emparejamiento.
                    </p>
                  )}
                </div>

                <div className={cardCls}>
                  <div className="mb-2 flex items-center justify-between gap-2">
                    <h3 className="font-semibold text-mitt-text">Mesas abiertas</h3>
                    <span className="money text-xs text-mitt-muted">
                      {openCount === 0 ? 'Barra libre' : money(inProgressCents)}
                    </span>
                  </div>
                  {tabs.length === 0 ? (
                    <p className="text-sm text-mitt-muted">
                      La barra está libre. Abra la primera mesa desde Mesas.
                    </p>
                  ) : (
                    <ul className="list-none p-0">
                      {tabs.map((t) => (
                        <li
                          key={t.id}
                          className="money flex items-baseline justify-between gap-2 border-b border-mitt-raised py-1 text-mitt-text"
                        >
                          <span>
                            Mesa {t.table_id}{' '}
                            <span className="text-xs text-mitt-muted">
                              · {(t.items || []).length} consumo{(t.items || []).length === 1 ? '' : 's'}
                            </span>
                          </span>
                          <span className="font-bold">{money(t.total_cents)}</span>
                        </li>
                      ))}
                    </ul>
                  )}
                  <button
                    type="button"
                    className={`${btnGhost} mt-3`}
                    onClick={() => setSection('mesas')}
                  >
                    IR A MESAS
                  </button>
                </div>
              </div>
            </section>
          )}

          {section === 'mesas' && (
            <section aria-label="Mesas">
              <div className="mb-2 mt-2 flex flex-wrap items-end justify-between gap-2">
                <h2 className={sectionTitle}>Mesas</h2>
                <p className="money text-xs text-mitt-muted">
                  {openCount === 0 ? 'Sin mesas abiertas' : `${openCount} abierta${openCount === 1 ? '' : 's'} · ${money(inProgressCents)} en curso`}
                </p>
              </div>
              <div className="mb-2 flex flex-wrap items-center gap-2 rounded-xl bg-mitt-surface p-3">
                <input
                  className={inputCls}
                  type="text"
                  placeholder="Mesa (ej. T1)"
                  size={12}
                  value={tableName}
                  autoComplete="off"
                  autoCorrect="off"
                  autoCapitalize="off"
                  spellCheck={false}
                  onChange={(e) => setTableName(e.target.value)}
                  onKeyDown={(e) => {
                    if (e.key === 'Enter') openTable();
                  }}
                />
                <button type="button" className={btnPrimary} onClick={openTable}>
                  ABRIR MESA
                </button>
              </div>
              {tabs.length === 0 ? (
                <div className={`${cardCls} text-center`}>
                  <p className="font-semibold text-mitt-text">La barra está libre</p>
                  <p className="mt-1 text-sm text-mitt-muted">
                    Abra la primera mesa con el nombre de arriba para empezar a vender.
                  </p>
                </div>
              ) : (
                <div className="grid grid-cols-1 gap-2 sm:grid-cols-2">
                  {tabs.map((t, i) => {
                    const sel = addSelection[t.id] || { productId: '', qty: '1' };
                    return (
                      <SpotlightCard
                        key={t.id}
                        className="border-l-4 border-l-mitt-accent bg-mitt-raised p-3"
                      >
                        <div
                          className="mitt-enter relative"
                          style={{ animationDelay: `${Math.min(i * 50, 250)}ms` }}
                        >
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
                              aria-label={`Producto para mesa ${t.table_id}`}
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
                              className={`${inputCls} w-20`}
                              type="number"
                              aria-label={`Cantidad para mesa ${t.table_id}`}
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
          )}

          {section === 'catalogo' && (
            <section aria-label="Catálogo">
              <h2 className={`${sectionTitle} mb-2 mt-2`}>Catálogo</h2>
              {catalog.length === 0 ? (
                <div className={`${cardCls} text-center`}>
                  <p className="font-semibold text-mitt-text">Catálogo vacío</p>
                  <p className="mt-1 text-sm text-mitt-muted">
                    Agregue el primer producto con el formulario de abajo.
                  </p>
                </div>
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
                  autoComplete="off"
                  autoCorrect="off"
                  spellCheck={false}
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
          )}

          {section === 'gastos' && (
            <section aria-label="Gastos">
              <h2 className={`${sectionTitle} mb-2 mt-2`}>Gastos</h2>
              {expenses.length === 0 ? (
                <div className={`${cardCls} text-center`}>
                  <p className="font-semibold text-mitt-text">Sin gastos</p>
                  <p className="mt-1 text-sm text-mitt-muted">
                    Registre el primer gasto del día con el formulario de abajo.
                  </p>
                </div>
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
                  autoComplete="off"
                  autoCorrect="off"
                  spellCheck={false}
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
          )}

          {section === 'conexion' && (
            <section aria-label="Conexión">
              <div className="mb-2 mt-2 flex flex-wrap items-end justify-between gap-2">
                <h2 className={sectionTitle}>Conexión</h2>
                <span
                  className={`rounded px-3 py-1 text-xs font-semibold ${
                    pill === 'ok' ? 'bg-mitt-success text-mitt-bg' : 'bg-mitt-danger text-mitt-text'
                  }`}
                >
                  {pillLabel[pill]}
                </span>
              </div>
              <div className={`${cardCls} flex flex-col gap-4`}>
                <details className="mitt-details w-full sm:max-w-md">
                  <summary className={btnGhost}>Conexión manual</summary>
                  <div className="mt-2 flex flex-col gap-2">
                    <label className="text-xs text-mitt-muted" htmlFor="token-input">
                      Token o enlace de emparejamiento
                    </label>
                    <input
                      id="token-input"
                      className={`${inputCls} w-full`}
                      type="text"
                      inputMode="text"
                      enterKeyHint="go"
                      placeholder="Token o enlace mitt://pair?…"
                      value={tokenInput}
                      autoComplete="off"
                      autoCorrect="off"
                      autoCapitalize="off"
                      spellCheck={false}
                      onChange={(e) => setTokenInput(e.target.value)}
                      onKeyDown={(e) => {
                        if (e.key === 'Enter') saveToken();
                      }}
                    />
                    <button type="button" className={btnPrimary} onClick={saveToken}>
                      GUARDAR
                    </button>
                    <p className="text-xs text-mitt-muted">
                      El token se guarda en este navegador. Base:{' '}
                      <span className="money">{baseUrl()}</span>
                    </p>
                  </div>
                </details>
              </div>
            </section>
          )}

          {section === 'personalizar' && (
            <Personalizar
              branding={branding}
              logoSrc={logoSrc}
              hasToken={storedToken() !== ''}
              onSaved={handleBrandingSaved}
              notify={showToast}
              fail={fail}
            />
          )}
        </div>
      </main>

      {toast && (
        <div
          role="alert"
          className="mitt-toast fixed bottom-4 left-1/2 max-w-[90vw] -translate-x-1/2 rounded-lg bg-mitt-danger px-5 py-3 text-mitt-text"
        >
          {toast}
        </div>
      )}
    </div>
  );
}
