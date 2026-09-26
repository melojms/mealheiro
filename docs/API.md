# Mealheiro — HTTP API contract

Source of truth for backend and frontend. Frontend types live in `web/src/lib/types.ts` and must match this file.

## Conventions

- JSON, `snake_case` keys. Money is **integer cents** (`*_cents`), always positive for entries; the entry `type` gives meaning.
- Dates `YYYY-MM-DD` (local, `TZ`), months `YYYY-MM`. Timestamps RFC3339 UTC strings.
- Errors: `{"error": "message"}` with `400` (malformed/invalid input), `404` (not found), `409` (conflict, e.g. delete of referenced category).
- Empty lists are `[]`, never `null`. Optional values are `null` when absent.
- `payer_id` query param on reports filters entries by payer (omit = everyone).
- "Top-level category" = category with `parent_id = null`. Rollups aggregate a subcategory into its top-level parent.
- Pending entries are included in all totals unless stated otherwise; `pending_cents` fields break them out.

## Shared types

```ts
type EntryType = "expense" | "income" | "investment"
type EntryStatus = "confirmed" | "pending"

interface Person { id: number; name: string; kind: "person" | "joint" }

interface Category {
  id: number; parent_id: number | null; type: EntryType
  name: string; icon: string /* lucide kebab-case name */; color: string /* #rrggbb */
  archived: boolean
  usage_count: number // entries in the last 90 days; for top-level = self + children
  entry_count: number // all-time entries referencing this category (self only)
}

interface Entry {
  id: number; type: EntryType; date: string; amount_cents: number
  category_id: number; category_name: string
  parent_category_id: number | null; parent_category_name: string | null
  payer_id: number; payer_name: string
  note: string; tags: string[]            // sorted asc
  status: EntryStatus
  template_id: number | null
  recurring: boolean                       // generated from a template (template_month != null)
  created_at: string; updated_at: string
}

interface EntryInput {
  type: EntryType; date: string; amount_cents: number
  category_id: number; payer_id: number
  note?: string; tags?: string[]           // tags lowercased + trimmed + deduped server-side
}
```

## Meta

| Method & path | Response |
|---|---|
| `GET /healthz` | `{"status":"ok"}` |
| `GET /api/meta` | `{today, month, timezone, version}` |

## People  *(owner: be-core)*

| Method & path | Body | Response |
|---|---|---|
| `GET /api/people` | | `Person[]` (sort_order) |
| `PATCH /api/people/{id}` | `{name}` (non-empty, ≤ 40 chars) | `Person` |

## Categories  *(owner: be-core)*

| Method & path | Body | Response |
|---|---|---|
| `GET /api/categories?type=&include_archived=true` | | `Category[]`: ordered by `usage_count` desc, then name. Archived excluded unless `include_archived=true`. |
| `POST /api/categories` | `{type, name, parent_id?, icon?, color?}` | `201 Category`. Parent must be top-level, non-archived, same type. Default icon: parent's icon or `circle`; default color: parent's color or the first palette color unused by top-level categories of that type (palette in `internal/api/palette.go`). Duplicate name at same level → 409. |
| `PATCH /api/categories/{id}` | `{name?, icon?, color?, archived?}` | `Category`. Archiving a parent archives its children; unarchiving a child unarchives its parent. |
| `DELETE /api/categories/{id}` | | `204`. `409` if referenced by any entry, template, budget or child category (message suggests archiving). |

## Entries  *(owner: be-core)*

| Method & path | Body | Response |
|---|---|---|
| `GET /api/entries` | | `EntryList` (below) |
| `GET /api/entries/{id}` | | `Entry` |
| `POST /api/entries` | `EntryInput` | `201 Entry` (status `confirmed`). Validation: amount > 0; valid date; category exists, not archived, `category.type == type`; payer exists. |
| `PUT /api/entries/{id}` | `EntryInput` | `Entry`. Archived category allowed if unchanged. Status unchanged. |
| `POST /api/entries/{id}/confirm` | `{amount_cents?}` | `Entry` with status `confirmed` (optionally updating the amount). 400 if not pending. |
| `DELETE /api/entries/{id}` | | `204` |

`GET /api/entries` query params (all optional): `from`, `to` (dates, inclusive), `type`, `category_id` (matches the category **or its children**), `payer_id`, `tag`, `q` (case-insensitive substring over note, category name, parent category name, tags), `min_cents`, `max_cents`, `status`, `limit` (default 50, max 500), `offset`.

```ts
interface EntryList {
  entries: Entry[]                         // date desc, id desc
  total_count: number                      // matching entries, ignoring limit/offset
  totals: { expense_cents: number; income_cents: number; investment_cents: number }
  breakdown: { category_id: number; name: string; type: EntryType; color: string; icon: string; amount_cents: number }[]
                                           // matching entries grouped by top-level category, amount desc
}
```

## Tags  *(owner: be-core)*

| Method & path | Response |
|---|---|
| `GET /api/tags?q=` | `{name: string; count: number}[]`: tags used by ≥1 entry, prefix match on `q`, count desc then name, max 20 |

## Recurring templates  *(owner: be-ops)*

```ts
interface Template {
  id: number; type: EntryType
  category_id: number; category_name: string; parent_category_name: string | null
  payer_id: number; payer_name: string
  amount_cents: number; variable: boolean; note: string
  start_month: string; end_month: string | null; active: boolean
  last_generated_month: string | null
}
interface TemplateInput {
  type: EntryType; category_id: number; payer_id: number; amount_cents: number
  variable: boolean; note?: string; start_month: string; end_month?: string | null; active?: boolean // default true
}
```

| Method & path | Body | Response |
|---|---|---|
| `GET /api/templates` | | `Template[]` (active first, then type, category name) |
| `POST /api/templates` | `TemplateInput` | `201 Template`; then runs generation immediately. |
| `PUT /api/templates/{id}` | `TemplateInput` | `Template`; affects future generations only; runs generation. |
| `DELETE /api/templates/{id}` | | `204`; generated entries are kept (`template_id` → null, still `recurring`). |
| `POST /api/recurring/run` | | `{created: number}` |
| `GET /api/pending` | | `PendingEntry[]` = `Entry & {stale: boolean}`: all pending entries, date asc. `stale` = entry month < current month. |

Generation rules: for each active template, for each month in `[start_month, min(end_month, current month)]` without a `template_runs` row: insert entry dated `YYYY-MM-01` with the template's type/category/payer/note, `template_id`, `template_month`; status `pending` if `variable` else `confirmed`; amount = variable ? (most recent entry of this template by date, else template amount) : template amount; insert `template_runs` row. All in one transaction per template.

## Budgets  *(owner: be-ops)*

```ts
interface BudgetLine {
  category_id: number | null               // null = overall cap
  name: string                             // "Overall" for the cap
  icon: string; color: string              // "wallet" / "#64748b" for the cap
  budget_cents: number
  spent_cents: number                      // expenses in month incl. pending (top-level rollup; cap = all expenses)
  pending_cents: number
  ratio: number                            // spent/budget (float, can exceed 1)
}
```

| Method & path | Body | Response |
|---|---|---|
| `GET /api/budgets?month=` (default current) | | `{month, overall: BudgetLine \| null, categories: BudgetLine[]}`: lines whose effective budget for that month is non-null; categories sorted by ratio desc. |
| `PUT /api/budgets` | `{category_id: number \| null, amount_cents: number \| null}` | `204`. Upserts row `effective_from = current month`. `amount_cents: null` removes the budget from the current month on. Category must be a top-level expense category. |
| `GET /api/budgets/status?category_id=&month=` | | `{category: BudgetLine \| null, overall: BudgetLine \| null}`: for the add-time toast. `category_id` may be a subcategory (resolved to its parent). |

Effective budget for month M = row with greatest `effective_from <= M` for that category (null amount = none).

## Export & backup  *(owner: be-ops)*

| Method & path | Response |
|---|---|
| `GET /api/export.csv?from=&to=&types=expense,income,investment` | `text/csv; charset=utf-8`, `Content-Disposition: attachment; filename="mealheiro_<from>_<to>.csv"`. UTF-8 BOM, `;` separator, CRLF, header `date;type;category;subcategory;amount;payer;note;tags;recurring;status`. `category` = top-level name, `subcategory` = leaf name or empty. `amount` decimal comma `12,50`. `tags` joined with `,`. `recurring` `yes`/`no`. Date asc. Fields quoted per RFC 4180 when needed. Missing from/to = unbounded. |
| `GET /api/backup` | `application/octet-stream`, `Content-Disposition: attachment; filename="mealheiro-<YYYYMMDD-HHMMSS>.db"`: consistent `VACUUM INTO` snapshot streamed then deleted. |

Nightly: `backup.Nightly` writes `<DATA_DIR>/backups/mealheiro-YYYYMMDD.db` once per day and keeps the newest 30.

## Reports  *(owner: be-reports)*

All accept optional `payer_id`.

### `GET /api/reports/month?month=` (default current)

```ts
interface CategoryAmount {
  category_id: number; name: string; icon: string; color: string; type: EntryType
  amount_cents: number; pending_cents: number
  prev_month_cents: number                 // same category, previous month
  avg3_cents: number                       // average of the 3 months before `month` (integer, rounded)
  subcategories: { category_id: number; name: string; color: string; amount_cents: number }[]
                                           // children with >0; entries booked on the parent itself appear as a child named "(general)" with the parent's id
}
interface MonthReport {
  month: string
  kpis: {
    income_cents: number; expense_cents: number; investment_cents: number
    leftover_cents: number                 // income − expense − investment
    savings_rate: number | null            // (income − expense) / income; null if income = 0
    pending_cents: number                  // all pending entries in month (any type)
  }
  prev_kpis: MonthReport["kpis"]           // previous month, same shape
  expenses: CategoryAmount[]               // top-level expense categories with amount > 0, amount desc
  income: CategoryAmount[]
  investments: CategoryAmount[]
  recurring: { committed_cents: number; entry_count: number; pending_count: number; pending_cents: number }
                                           // expense entries in month with recurring = true
}
```

### `GET /api/reports/trends?end=&months=12&category_id=`

```ts
interface TrendsReport {
  months: string[]                         // ascending, `months` items ending at `end` (default current month)
  series: {
    month: string; income_cents: number; expense_cents: number; investment_cents: number
    savings_cents: number                  // income − expense
    by_category: Record<string, number>    // top-level expense category id → cents (only non-zero)
  }[]
  categories: { id: number; name: string; color: string; icon: string }[]
                                           // top-level expense categories present in the window, total desc
}
```
`category_id` (top-level expense) restricts `expense_cents` and `by_category` to that category.

### `GET /api/reports/year?year=` (default current)

```ts
interface YearReport {
  year: number
  totals: { income_cents: number; expense_cents: number; investment_cents: number; leftover_cents: number; savings_rate: number | null }
  prev_totals: YearReport["totals"]
  categories: { category_id: number; name: string; icon: string; color: string; type: EntryType; amount_cents: number; prev_year_cents: number }[]
                                           // all types, top-level, non-zero in either year, type then amount desc
  monthly: { month: string; income_cents: number; expense_cents: number; investment_cents: number }[] // 12 items
}
```

### `GET /api/reports/years`

`number[]`: years that have entries, desc; always includes the current year.

## Insights  *(owner: be-reports)*

### `GET /api/insights?month=` (default current)

```ts
interface Insight {
  kind: "category_up" | "category_down" | "pace" | "top_expense" | "savings_rate" | "new_recurring"
  severity: "info" | "good" | "warning"
  title: string                            // short, e.g. "Groceries up 32%"
  detail: string                           // e.g. "412,30 € vs 312,00 € 3-month average"
  category_id: number | null
  entry_id: number | null
  amount_cents: number | null
}
```
Rules (thresholds are named constants in `internal/reports`):
1. `category_up`/`category_down` (top-level expense): |month − avg3| > 20% of avg3 **and** > 2000 cents; requires avg3 > 0. up = warning, down = good. For the current month, only `category_up` is emitted (month incomplete).
2. `pace` (current month only, day ≥ 5): non-recurring expenses so far / day × days in month. info; warning if projection > previous month's non-recurring expenses by > 10%.
3. `top_expense`: the 3 largest expense entries in the month (info), each with `entry_id`.
4. `savings_rate`: this month vs previous month (only when both have income). good if up, warning if down by > 5 points, else info.
5. `new_recurring`: templates whose `start_month` = month (info).

Order: warnings first, then good, then info.
