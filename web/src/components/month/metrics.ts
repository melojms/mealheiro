// Pure helpers for deltas, shares and budget states used by the dashboards.
import { pctChange } from "@/lib/money"
import type { CategoryAmount } from "@/lib/types"

export type Tone = "good" | "bad" | "neutral"

/** Whether a change is good news. For expenses "up" is bad; for income/savings "up" is good. */
export function deltaTone(diff: number, upIsGood: boolean): Tone {
  if (diff === 0) return "neutral"
  return diff > 0 === upIsGood ? "good" : "bad"
}

export interface Delta {
  diff: number
  pct: number | null // null when the baseline is 0
  tone: Tone
}

export function delta(current: number, previous: number, upIsGood: boolean): Delta {
  const diff = current - previous
  return { diff, pct: pctChange(previous, current), tone: deltaTone(diff, upIsGood) }
}

/** Savings-rate change in percentage points, or null when either side is undefined. */
export function ratePoints(current: number | null, previous: number | null): number | null {
  if (current === null || previous === null) return null
  return (current - previous) * 100
}

/** "+12%", "−8%", "0%" (true minus sign), or null for no baseline. */
export function formatSignedPct(pct: number | null): string | null {
  if (pct === null || !Number.isFinite(pct)) return null
  const v = Math.round(pct * 100)
  if (v === 0) return "0%"
  return `${v > 0 ? "+" : "−"}${Math.abs(v)}%`
}

/** "+3.2 pts" style label for savings-rate deltas. */
export function formatPoints(points: number | null): string | null {
  if (points === null || !Number.isFinite(points)) return null
  const v = Math.round(points * 10) / 10
  if (v === 0) return "0 pts"
  return `${v > 0 ? "+" : "−"}${Math.abs(v).toFixed(1)} pts`
}

export function share(part: number, total: number): number {
  return total > 0 ? part / total : 0
}

export type BudgetState = "ok" | "warn" | "over"

export const BUDGET_WARN_RATIO = 0.8

export function budgetState(ratio: number): BudgetState {
  if (ratio >= 1) return "over"
  if (ratio >= BUDGET_WARN_RATIO) return "warn"
  return "ok"
}

/** Sum of pending estimates across a report section (kpis only expose the grand total). */
export function sumPending(items: CategoryAmount[]): number {
  return items.reduce((s, c) => s + c.pending_cents, 0)
}

export function sumAmount(items: { amount_cents: number }[]): number {
  return items.reduce((s, c) => s + c.amount_cents, 0)
}

/** Cents as a plain "84,20" string for editable inputs (no currency, no grouping). */
export function centsToInput(cents: number): string {
  return `${Math.floor(cents / 100)},${String(cents % 100).padStart(2, "0")}`
}
