import type { Entry } from "@/lib/types"

export interface DayGroup {
  date: string
  entries: Entry[]
  expense_cents: number // day's spend, shown next to the sticky header
}

/** Groups entries (already sorted date desc) into consecutive day buckets, preserving order. */
export function groupByDate(entries: Entry[]): DayGroup[] {
  const groups: DayGroup[] = []
  for (const e of entries) {
    let g = groups.at(-1)
    if (!g || g.date !== e.date) {
      g = { date: e.date, entries: [], expense_cents: 0 }
      groups.push(g)
    }
    g.entries.push(e)
    if (e.type === "expense") g.expense_cents += e.amount_cents
  }
  return groups
}
