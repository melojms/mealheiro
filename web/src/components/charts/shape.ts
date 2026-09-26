// Turns report payloads into flat Recharts rows + ChartConfig.
import type { ChartConfig } from "@/components/ui/chart"
import type { TrendsReport, YearReport } from "@/lib/types"
import { shortMonthLabel } from "@/components/filters/dates"

/** Stable ChartConfig/dataKey for a category id (CSS var names can't start with a digit). */
export const catKey = (id: number | string) => `c${id}`

export type TrendRow = { month: string; label: string } & Record<string, number | string>

/** One row per month with a cents column per category (0-filled so stacks line up). */
export function stackedTrendRows(report: TrendsReport): TrendRow[] {
  return report.series.map((s) => {
    const row: TrendRow = { month: s.month, label: shortMonthLabel(s.month) }
    for (const c of report.categories) row[catKey(c.id)] = s.by_category[String(c.id)] ?? 0
    return row
  })
}

export function categoryConfig(categories: { id: number; name: string; color: string }[]): ChartConfig {
  return Object.fromEntries(categories.map((c) => [catKey(c.id), { label: c.name, color: c.color }]))
}

export interface FlowRow {
  month: string
  label: string
  income: number
  expenses: number
  savings: number
}

export function flowRows(report: TrendsReport): FlowRow[] {
  return report.series.map((s) => ({
    month: s.month,
    label: shortMonthLabel(s.month),
    income: s.income_cents,
    expenses: s.expense_cents,
    savings: s.savings_cents,
  }))
}

export interface YearMonthRow {
  month: string
  label: string
  income: number
  expenses: number
  investments: number
}

export function yearMonthlyRows(report: YearReport): YearMonthRow[] {
  return report.monthly.map((m) => ({
    month: m.month,
    // Year view: plain month names, the year is in the header.
    label: shortMonthLabel(m.month).replace(/ '\d\d$/, ""),
    income: m.income_cents,
    expenses: m.expense_cents,
    investments: m.investment_cents,
  }))
}

/** Months in the window that have any data; used to decide between chart and empty state. */
export function hasTrendData(report: TrendsReport): boolean {
  return report.series.some((s) => s.income_cents || s.expense_cents || s.investment_cents)
}

/** Groups year categories by type, preserving the API's amount-desc order within each type. */
export function yearCategoriesByType(report: YearReport) {
  const order = ["expense", "income", "investment"] as const
  return order
    .map((type) => ({ type, items: report.categories.filter((c) => c.type === type) }))
    .filter((g) => g.items.length > 0)
}
