/** Colors of the people in the shared split bar, by position. */
export const PERSON_COLORS = ["#0ea5e9", "#f97316", "#8b5cf6", "#10b981"] as const

/**
 * Whole-number share of each amount, summing to exactly 100 (largest remainder).
 * null when there is nothing to split.
 */
export function splitPercents(amounts: number[]): number[] | null {
  const total = amounts.reduce((s, a) => s + a, 0)
  if (total <= 0) return null
  const exact = amounts.map((a) => (a / total) * 100)
  const out = exact.map(Math.floor)
  const byRemainder = exact.map((e, i) => [e - out[i], i] as const).sort((a, b) => b[0] - a[0] || a[1] - b[1])
  for (let k = 0, left = 100 - out.reduce((s, p) => s + p, 0); k < left; k++) out[byRemainder[k][1]]++
  return out
}
