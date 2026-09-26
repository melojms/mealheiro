import { formatCents } from "@/lib/money"
import type { BudgetLine, BudgetStatus } from "@/lib/types"

export type BudgetLevel = "ok" | "warning" | "over"

export const BUDGET_WARNING_RATIO = 0.8

export function budgetLevel(ratio: number): BudgetLevel {
  if (ratio >= 1) return "over"
  if (ratio >= BUDGET_WARNING_RATIO) return "warning"
  return "ok"
}

/** "Groceries 320,00 € / 400,00 € this month" */
export function budgetLineText(line: BudgetLine): string {
  return `${line.name} ${formatCents(line.spent_cents)} / ${formatCents(line.budget_cents)} this month`
}

/** Lines to show after saving an expense and the worst level among them (null = no budget at all). */
export function budgetSummary(status: BudgetStatus): { lines: string[]; level: BudgetLevel } | null {
  const present = [status.category, status.overall].filter((l): l is BudgetLine => l !== null)
  if (present.length === 0) return null
  const rank: Record<BudgetLevel, number> = { ok: 0, warning: 1, over: 2 }
  const level = present.map((l) => budgetLevel(l.ratio)).reduce((a, b) => (rank[b] > rank[a] ? b : a))
  return { lines: present.map(budgetLineText), level }
}
