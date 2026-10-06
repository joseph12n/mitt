-- mitt schema v1: bar POS catalog, tables, tabs, sales, expenses.
-- Applied once when PRAGMA user_version is 0, then user_version is set to 1.

CREATE TABLE IF NOT EXISTS products(
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  price_cents INTEGER NOT NULL,
  available INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS tables_tbl(
  id TEXT PRIMARY KEY,
  label TEXT NOT NULL,
  status TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS tabs(
  id TEXT PRIMARY KEY,
  table_id TEXT NOT NULL,
  status TEXT NOT NULL,
  opened_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_tabs_table_status ON tabs(table_id, status);

CREATE TABLE IF NOT EXISTS tab_items(
  tab_id TEXT NOT NULL,
  product_id TEXT NOT NULL,
  name TEXT NOT NULL,
  unit_price_cents INTEGER NOT NULL,
  qty INTEGER NOT NULL,
  PRIMARY KEY(tab_id, product_id)
);

CREATE TABLE IF NOT EXISTS sales(
  id TEXT PRIMARY KEY,
  table_id TEXT,
  total_cents INTEGER NOT NULL,
  closed_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS sale_items(
  sale_id TEXT NOT NULL,
  product_id TEXT NOT NULL,
  name TEXT NOT NULL,
  unit_price_cents INTEGER NOT NULL,
  qty INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS expenses(
  id TEXT PRIMARY KEY,
  description TEXT NOT NULL,
  qty REAL NOT NULL,
  cost_cents INTEGER NOT NULL,
  date TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_expenses_date ON expenses(date);
