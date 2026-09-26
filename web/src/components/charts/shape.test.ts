import { catKey, categoryConfig, flowRows, hasTrendData, stackedTrendRows, yearCategoriesByType, yearMonthlyRows } from "./shape"
import type { TrendsReport, YearReport } from "@/lib/types"

const trends: TrendsReport = {
  months: ["2025-12", "2026-01"],
  series: [
    { month: "2025-12", income_cents: 300000, expense_cents: 120000, investment_cents: 0, savings_cents: 180000, by_category: { "7": 100000, "9": 20000 } },
    { month: "2026-01", income_cents: 0, expense_cents: 5000, investment_cents: 0, savings_cents: -5000, by_category: { "9": 5000 } },
  ],
  categories: [
    { id: 7, name: "House", color: "#0ea5e9", icon: "house" },
    { id: 9, name: "Groceries", color: "#10b981", icon: "shopping-cart" },
  ],
}

describe("trends shaping", () => {
  it("builds 0-filled stacked rows", () => {
    expect(stackedTrendRows(trends)).toEqual([
      { month: "2025-12", label: "Dec", c7: 100000, c9: 20000 },
      { month: "2026-01", label: "Jan '26", c7: 0, c9: 5000 },
    ])
  })

  it("builds a config keyed by category", () => {
    expect(categoryConfig(trends.categories)).toEqual({
      c7: { label: "House", color: "#0ea5e9" },
      c9: { label: "Groceries", color: "#10b981" },
    })
    expect(catKey(3)).toBe("c3")
  })

  it("maps flows incl. negative savings", () => {
    expect(flowRows(trends)[1]).toEqual({ month: "2026-01", label: "Jan '26", income: 0, expenses: 5000, savings: -5000 })
  })

  it("detects empty windows", () => {
    expect(hasTrendData(trends)).toBe(true)
    expect(hasTrendData({ ...trends, series: trends.series.map((s) => ({ ...s, income_cents: 0, expense_cents: 0 })) })).toBe(false)
    expect(hasTrendData({ months: [], series: [], categories: [] })).toBe(false)
  })
})

describe("year shaping", () => {
  const zero = { income_cents: 0, expense_cents: 0, investment_cents: 0, leftover_cents: 0, savings_rate: null }
  const year: YearReport = {
    year: 2026,
    totals: zero,
    prev_totals: zero,
    categories: [
      { category_id: 1, name: "House", icon: "house", color: "#000000", type: "expense", amount_cents: 5, prev_year_cents: 0 },
      { category_id: 2, name: "Food", icon: "utensils", color: "#000000", type: "expense", amount_cents: 3, prev_year_cents: 9 },
      { category_id: 3, name: "ETF", icon: "trending-up", color: "#000000", type: "investment", amount_cents: 1, prev_year_cents: 1 },
    ],
    monthly: [{ month: "2026-01", income_cents: 1, expense_cents: 2, investment_cents: 3 }],
  }

  it("labels months without year", () => {
    expect(yearMonthlyRows(year)).toEqual([{ month: "2026-01", label: "Jan", income: 1, expenses: 2, investments: 3 }])
  })

  it("groups categories by type in fixed order, skipping empty types", () => {
    expect(yearCategoriesByType(year).map((g) => [g.type, g.items.map((c) => c.category_id)])).toEqual([
      ["expense", [1, 2]],
      ["investment", [3]],
    ])
  })
})
