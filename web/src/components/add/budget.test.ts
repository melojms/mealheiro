import type { BudgetLine } from "@/lib/types"
import { budgetLevel, budgetLineText, budgetSummary } from "./budget"

const line = (name: string, spent: number, budget: number, category_id: number | null = 1): BudgetLine => ({
  category_id, name, icon: "wallet", color: "#64748b", budget_cents: budget, spent_cents: spent, pending_cents: 0,
  ratio: spent / budget,
})

describe("budgetLevel", () => {
  it.each([[0, "ok"], [0.79, "ok"], [0.8, "warning"], [0.99, "warning"], [1, "over"], [1.5, "over"]])(
    "%f -> %s",
    (r, want) => expect(budgetLevel(r)).toBe(want),
  )
})

describe("budgetSummary", () => {
  it("returns null without budgets", () => expect(budgetSummary({ category: null, overall: null })).toBeNull())

  it("formats the category line", () => {
    expect(budgetLineText(line("Groceries", 32000, 40000))).toBe("Groceries 320,00 € / 400,00 € this month")
  })

  it("combines category and overall with the worst level", () => {
    const s = budgetSummary({ category: line("Groceries", 10000, 40000), overall: line("Overall", 190000, 200000, null) })
    expect(s).toEqual({
      lines: ["Groceries 100,00 € / 400,00 € this month", "Overall 1900,00 € / 2000,00 € this month"],
      level: "warning",
    })
  })

  it("overall only", () => {
    expect(budgetSummary({ category: null, overall: line("Overall", 250000, 200000, null) })?.level).toBe("over")
  })
})
