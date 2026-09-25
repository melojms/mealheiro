// Mirrors docs/API.md. Keep in sync with the backend contract.

export type EntryType = "expense" | "income" | "investment"
export type EntryStatus = "confirmed" | "pending"

export interface Person {
  id: number
  name: string
  kind: "person" | "joint"
}

export interface Category {
  id: number
  parent_id: number | null
  type: EntryType
  name: string
  icon: string
  color: string
  archived: boolean
  usage_count: number
  entry_count: number
}

export interface Entry {
  id: number
  type: EntryType
  date: string
  amount_cents: number
  category_id: number
  category_name: string
  parent_category_id: number | null
  parent_category_name: string | null
  payer_id: number
  payer_name: string
  note: string
  tags: string[]
  status: EntryStatus
  template_id: number | null
  recurring: boolean
  created_at: string
  updated_at: string
}

export interface PendingEntry extends Entry {
  stale: boolean
}

export interface EntryInput {
  type: EntryType
  date: string
  amount_cents: number
  category_id: number
  payer_id: number
  note?: string
  tags?: string[]
}

export interface EntryFilters {
  from?: string
  to?: string
  type?: EntryType
  category_id?: number
  payer_id?: number
  tag?: string
  q?: string
  min_cents?: number
  max_cents?: number
  status?: EntryStatus
  limit?: number
  offset?: number
}

export interface Totals {
  expense_cents: number
  income_cents: number
  investment_cents: number
}

export interface BreakdownItem {
  category_id: number
  name: string
  type: EntryType
  color: string
  icon: string
  amount_cents: number
}

export interface EntryList {
  entries: Entry[]
  total_count: number
  totals: Totals
  breakdown: BreakdownItem[]
}

export interface TagCount {
  name: string
  count: number
}

export interface Template {
  id: number
  type: EntryType
  category_id: number
  category_name: string
  parent_category_name: string | null
  payer_id: number
  payer_name: string
  amount_cents: number
  variable: boolean
  note: string
  start_month: string
  end_month: string | null
  active: boolean
  last_generated_month: string | null
}

export interface TemplateInput {
  type: EntryType
  category_id: number
  payer_id: number
  amount_cents: number
  variable: boolean
  note?: string
  start_month: string
  end_month?: string | null
  active?: boolean
}

export interface BudgetLine {
  category_id: number | null
  name: string
  icon: string
  color: string
  budget_cents: number
  spent_cents: number
  pending_cents: number
  ratio: number
}

export interface Budgets {
  month: string
  overall: BudgetLine | null
  categories: BudgetLine[]
}

export interface BudgetStatus {
  category: BudgetLine | null
  overall: BudgetLine | null
}

export interface Kpis {
  income_cents: number
  expense_cents: number
  investment_cents: number
  leftover_cents: number
  savings_rate: number | null
  pending_cents: number
}

export interface CategoryAmount {
  category_id: number
  name: string
  icon: string
  color: string
  type: EntryType
  amount_cents: number
  pending_cents: number
  prev_month_cents: number
  avg3_cents: number
  subcategories: { category_id: number; name: string; color: string; amount_cents: number }[]
}

export interface MonthReport {
  month: string
  kpis: Kpis
  prev_kpis: Kpis
  expenses: CategoryAmount[]
  income: CategoryAmount[]
  investments: CategoryAmount[]
  recurring: { committed_cents: number; entry_count: number; pending_count: number; pending_cents: number }
}

export interface TrendsReport {
  months: string[]
  series: {
    month: string
    income_cents: number
    expense_cents: number
    investment_cents: number
    savings_cents: number
    by_category: Record<string, number>
  }[]
  categories: { id: number; name: string; color: string; icon: string }[]
}

export interface YearTotals {
  income_cents: number
  expense_cents: number
  investment_cents: number
  leftover_cents: number
  savings_rate: number | null
}

export interface YearReport {
  year: number
  totals: YearTotals
  prev_totals: YearTotals
  categories: {
    category_id: number
    name: string
    icon: string
    color: string
    type: EntryType
    amount_cents: number
    prev_year_cents: number
  }[]
  monthly: { month: string; income_cents: number; expense_cents: number; investment_cents: number }[]
}

export interface Insight {
  kind: "category_up" | "category_down" | "pace" | "top_expense" | "savings_rate" | "new_recurring"
  severity: "info" | "good" | "warning"
  title: string
  detail: string
  category_id: number | null
  entry_id: number | null
  amount_cents: number | null
}

export interface Meta {
  today: string
  month: string
  timezone: string
  version: string
}
