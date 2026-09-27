# Mealheiro — Specification

Household budget web app for two people (+ a joint account). Goal: see where money goes, optimize month by month, save as much as possible. Runs self-hosted via Docker on a home server; reached over LAN/VPN. No authentication (for now).

## 1. Principles

- **Input must be frictionless.** Phone-first. Adding an expense = type amount → tap category → Save (3 taps + digits).
- **Numbers must be trustworthy.** Money stored as integer cents; estimates (pending recurring entries) are visibly flagged.
- **Lightweight.** Single Go binary with the SPA embedded; ~20MB distroless image; SQLite file on a volume.

## 2. Stack

| Layer | Choice |
|---|---|
| Backend | Go 1.27, stdlib `net/http` (method+path routing), `log/slog` |
| DB | SQLite via `modernc.org/sqlite` (pure Go, no CGO), WAL mode |
| Queries | `sqlc` (one generated file per query file, no interface emission) |
| Migrations | `goose`, SQL files embedded in the binary, run on startup |
| Frontend | React + Vite + TypeScript + Tailwind + shadcn/ui + shadcn charts (Recharts), Lucide icons |
| Packaging | Multi-stage Dockerfile (node → go → distroless, starts as root only to chown `/data` and drop to `PUID:PGID`); SPA embedded via `embed` |
| Tests | Go tests on real temp SQLite; Vitest for amount parsing/formatting; one Playwright smoke test |

## 3. Domain

### 3.1 Entry types
- **expense** — money spent.
- **income** — salary, bonus, refund, side income.
- **investment** — money moved to investments/savings. Not an expense; not spendable.

Monthly: `leftover = income − expenses − investments`. `savings rate = (income − expenses) / income` (investments count as saved).

### 3.2 People / payer
- One `people` list: two persons + **Joint** (joint account, e.g. mortgage direct debit). Names editable.
- Every entry (all types) has a payer. Finances are pooled, **no settle-up math**; the payer only drives the shared-expenses split (§3.9).
- Each device stores its default payer (localStorage); quick-add pre-selects it.

### 3.3 Categories
- Two levels: top-level category → optional subcategories. Each category belongs to one entry type.
- Each category has a Lucide **icon** and a **color** (from a fixed palette; auto-assigned, editable).
- Editable in Settings. Delete = **archive** (hidden from quick-add, history kept). Hard delete only when zero entries/templates reference it. Rename updates history.
- An entry references a leaf or top-level category; rollups aggregate subcategories into their parent.
- Quick-add tile order: **auto by usage** (count of entries in the last 90 days, desc; ties by name).

Seed:
- **Expense**: Water · Electricity · Gas · Internet & Phone · House (Rent, Mortgage, Management, Insurance, Maintenance) · Groceries · Subscriptions · Car · Transport (Public, Taxi/Uber) · Clothes · Technology · Eating out · Health · Leisure · Travel · Gifts · Personal care · Education · Other
- **Income**: Salary · Bonus · Refund · Side income
- **Investment**: ETFs/Stocks · Savings account · BTC

### 3.4 Entries
Fields: type, date (local date, no time), amount (cents, always > 0; sign from type), category, payer, personal (bool, default false), note (optional), tags (optional, many), status (`confirmed` | `pending`), template (if generated), created_at/updated_at.
- **Personal**: expenses are shared by default; the "Personal" toggle keeps one out of the shared split (§3.9). Only expenses paid by a person can be personal — the API clears the flag for income/investments and for Joint. The Add screen resets it to off after every save.
- Full edit/delete. "Recent entries" list under quick-add.
- Currency: EUR only. Input accepts `12,50` and `12.50`. Display `12,50 €` style (pt-PT number format), English UI (strings centralized for future PT).
- Dates/“today”/recurring generation use `TZ` (default `Europe/Lisbon`).

### 3.5 Tags
Free-form, many per entry, lowercase-normalized, autocomplete from existing tags. Entry list filterable by tag with total and per-category breakdown. Included in CSV. No tag charts in MVP.

### 3.6 Recurring templates
- Apply to all three types (e.g. mortgage, subscriptions, salary, monthly ETF contribution).
- Fields: type, category, payer, amount, `variable` flag, `personal` flag (same rules as entries; copied onto generated entries), note, `start_month`, optional `end_month`, `active`.
- Frequency: **monthly only**. Generated entries are dated the **1st of the month** (no due-day field).
- **Fixed** templates → entry created `confirmed`. **Variable** templates (e.g. electricity, salary) → entry created `pending`, amount prefilled with the most recent entry from that template (or template amount), shown in the **"to confirm" inbox**. Confirming allows editing the amount.
- Pending entries count in totals and budgets as estimates, visually flagged. Pending entries from past months get a warning badge; never auto-confirmed.
- Editing a template affects **future** generations only. Pausing/deleting keeps already generated entries.
- Generation runs on startup and on a daily tick; catches up missed months; idempotent via unique `(template_id, month)`.

### 3.7 Budgets
- Per **top-level expense category**, covering its subcategories. Plus an optional **overall monthly spend cap** (category = null).
- One monthly amount; changes take effect from the current month onward (history via `effective_from`). Categories without a budget are fine.
- Usage includes pending estimates. Progress bar: amber ≥ 80%, red ≥ 100%.
- After saving an expense: toast `Groceries 320,00 € / 400,00 € this month`.

### 3.8 Insights (month view)
Thresholds are named constants.
1. Category > 20% **and** > €20 above its 3-month average (and notable drops, same thresholds).
2. **Pace**: projection of non-recurring expense for the current month (`spent_so_far / day_of_month × days_in_month`), excluding template-generated entries.
3. Top 3 largest expenses of the month.
4. Savings rate vs last month.
5. New recurring templates whose first generated entry is in this month.

### 3.9 Shared expenses
- Answers "are we both paying our half?". Target split is a fixed **50/50**; the app only shows totals — no "owes" wording, no settle-up.
- Counts **expenses** that are not personal and were paid by a **person**. Joint-paid expenses are left out (the joint account is already shared). Pending estimates count and are flagged.
- Per person: amount paid and share of the total (whole percentages summing to 100). Covers a month or its calendar year.
- Household-wide: ignores the payer filter.

## 4. Screens

1. **Add** (landing): amount keypad focused → category tile grid (subcategory row appears if any) → Save. Visible, prefilled chips: date (Today · Yesterday · pick), payer (device default), "Personal" toggle (expenses paid by a person only), type switch (Expense/Income/Investment). Collapsed "More": note, tags. Below: recent entries (edit/delete).
2. **Month**: KPIs (income · expenses · investments · leftover · savings %), budgets with progress bars, insights, category breakdown (donut + ranked list with Δ vs last month and vs 3-month avg; tap → subcategory split + entries), "Recurring this month: €X committed · N pending", pending inbox. Payer filter. Right after the KPIs, a **Shared expenses** card (§3.9) with a Month | Year toggle: total, split bar with percentages, amount per person.
3. **Trends**: stacked bars of expenses per category per month (last 12 months), line of income vs expenses vs savings, single-category toggle. Payer filter.
4. **Year**: totals per category per year, year-over-year comparison. Payer filter.
5. **Entries**: search & filter (text, type, category, payer, shared/personal, tag, amount range, date range) with totals. Personal entries carry a badge. "Shared" lists exactly what the Shared expenses card counts.
6. **Settings**: people, categories (icon, color, archive), recurring templates, budgets, device default payer, theme (light/dark/system), CSV export, backup download.

Responsive: phone-first; dashboards tuned for desktop too. PWA manifest (installable to home screen). Dark mode.

## 5. CSV export

Date-range dialog + type filter. All entry types with a `type` column. Columns: `date;type;category;subcategory;amount;payer;note;tags;recurring;status;personal`. Separator `;`, decimal comma (`12,50`), UTF-8 with BOM.

## 6. Operations

- Data dir `/data` (SQLite `mealheiro.db`).
- Nightly backup: `VACUUM INTO /data/backups/mealheiro-YYYYMMDD.db`, keep last 30. Settings has "Download backup" (consistent snapshot `.db`).
- `/healthz` endpoint (DB ping).
- `compose.yaml`: publishes `${HOST_PORT:-7447}:${PORT:-7447}` (server listens on `PORT`, default 7447), bind-mounts `${DATA_LOCATION:-./data}:/data`, `restart: unless-stopped`, `container_name: mealheiro`, runs as `PUID:PGID` (default 1000) after fixing `/data` ownership itself, `read_only: true`, `cap_drop: [ALL]` + `cap_add: [CHOWN, SETUID, SETGID, DAC_READ_SEARCH]`, `security_opt: [no-new-privileges:true]`, healthcheck.
- `compose.proxy.yaml` override: joins external `proxy` network, removes the published port.
- `.env.example`, Makefile (`up`, `down`, `logs`, `build`, `test`, `dev`, `generate`). Image `melojms/mealheiro:${VERSION}` built locally.

## 7. Out of scope (MVP)

Auth · multi-currency · bank import · savings goals (later) · receipt photos · settle-up (shared expenses show totals only) · custom split ratios · payment method / merchant fields · offline queue.
