import { addMonths, monthEnd } from "@/components/add/dates"

export type ExportPreset = "this_month" | "last_month" | "this_year" | "last_year" | "all"

export const EXPORT_PRESETS: { value: ExportPreset; label: string }[] = [
  { value: "this_month", label: "This month" },
  { value: "last_month", label: "Last month" },
  { value: "this_year", label: "This year" },
  { value: "last_year", label: "Last year" },
  { value: "all", label: "All" },
]

/** Inclusive date range for a preset relative to today (YYYY-MM-DD). Empty strings = unbounded. */
export function presetRange(preset: ExportPreset, today: string): { from: string; to: string } {
  const month = today.slice(0, 7)
  const year = Number(today.slice(0, 4))
  switch (preset) {
    case "this_month":
      return { from: `${month}-01`, to: monthEnd(month) }
    case "last_month": {
      const prev = addMonths(month, -1)
      return { from: `${prev}-01`, to: monthEnd(prev) }
    }
    case "this_year":
      return { from: `${year}-01-01`, to: `${year}-12-31` }
    case "last_year":
      return { from: `${year - 1}-01-01`, to: `${year - 1}-12-31` }
    case "all":
      return { from: "", to: "" }
  }
}

/** Which preset (if any) a range corresponds to. */
export function matchPreset(range: { from: string; to: string }, today: string): ExportPreset | null {
  return EXPORT_PRESETS.find(({ value }) => {
    const r = presetRange(value, today)
    return r.from === range.from && r.to === range.to
  })?.value ?? null
}
