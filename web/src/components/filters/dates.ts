// Pure date/month helpers. Months are "YYYY-MM", dates "YYYY-MM-DD" (local, no time).
// All Date objects are built in UTC so the browser's timezone never shifts a day.

function parseMonth(m: string): [number, number] {
  const [y, mo] = m.split("-").map(Number)
  return [y, mo]
}

const pad = (n: number) => String(n).padStart(2, "0")

export function isMonth(s: string | null | undefined): s is string {
  return !!s && /^\d{4}-(0[1-9]|1[0-2])$/.test(s)
}

export function isDate(s: string | null | undefined): s is string {
  return !!s && /^\d{4}-\d{2}-\d{2}$/.test(s) && !Number.isNaN(Date.parse(`${s}T00:00:00Z`))
}

/** addMonths("2026-01", -1) -> "2025-12" */
export function addMonths(m: string, n: number): string {
  const [y, mo] = parseMonth(m)
  const idx = y * 12 + (mo - 1) + n
  return `${Math.floor(idx / 12)}-${pad((idx % 12) + 1)}`
}

export function daysInMonth(m: string): number {
  const [y, mo] = parseMonth(m)
  return new Date(Date.UTC(y, mo, 0)).getUTCDate()
}

/** Inclusive date range covering the month. */
export function monthRange(m: string): { from: string; to: string } {
  return { from: `${m}-01`, to: `${m}-${pad(daysInMonth(m))}` }
}

// Fixed English names: ICU builds disagree on "Sep" vs "Sept" and on weekday punctuation.
const MONTHS = ["Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"]
const WEEKDAYS = ["Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"]

function utc(m: string, day = 1): Date {
  const [y, mo] = parseMonth(m)
  return new Date(Date.UTC(y, mo - 1, day))
}

/** "2026-03" -> "March 2026" */
export function monthLabel(m: string): string {
  return utc(m).toLocaleDateString("en-GB", { month: "long", year: "numeric", timeZone: "UTC" })
}

/** "2026-03" -> "Mar"; January also carries the year ("Jan '26") so chart axes stay unambiguous. */
export function shortMonthLabel(m: string): string {
  const name = MONTHS[parseMonth(m)[1] - 1]
  return m.endsWith("-01") ? `${name} '${m.slice(2, 4)}` : name
}

/** Tooltip-friendly "Mar 2026". */
export function mediumMonthLabel(m: string): string {
  const [y, mo] = parseMonth(m)
  return `${MONTHS[mo - 1]} ${y}`
}

/** Heading for a day in a list: "Today", "Yesterday", "Sat, 14 Mar" (+ year when not the current one). */
export function dayLabel(date: string, today: string): string {
  if (date === today) return "Today"
  const d = new Date(`${date}T00:00:00Z`)
  const t = new Date(`${today}T00:00:00Z`)
  if (t.getTime() - d.getTime() === 86_400_000) return "Yesterday"
  const weekday = WEEKDAYS[d.getUTCDay()]
  const month = MONTHS[d.getUTCMonth()]
  const year = date.slice(0, 4) === today.slice(0, 4) ? "" : ` ${date.slice(0, 4)}`
  return `${weekday}, ${d.getUTCDate()} ${month}${year}`
}

/** "2026-03-01" -> "1 Mar 2026" */
export function shortDate(date: string): string {
  const [y, m, d] = date.split("-").map(Number)
  return `${d} ${MONTHS[m - 1]} ${y}`
}

export const DATE_PRESETS = [
  { value: "this-month", label: "This month" },
  { value: "last-month", label: "Last month" },
  { value: "last-3-months", label: "Last 3 months" },
  { value: "this-year", label: "This year" },
  { value: "custom", label: "Custom" },
] as const

export type DatePreset = (typeof DATE_PRESETS)[number]["value"]

export function isPreset(s: string | null): s is DatePreset {
  return DATE_PRESETS.some((p) => p.value === s)
}

/** Resolves a preset relative to `today`. "custom" returns the given from/to untouched. */
export function presetRange(
  preset: DatePreset,
  today: string,
  custom: { from?: string; to?: string } = {},
): { from?: string; to?: string } {
  const month = today.slice(0, 7)
  switch (preset) {
    case "this-month":
      return monthRange(month)
    case "last-month":
      return monthRange(addMonths(month, -1))
    case "last-3-months":
      return { from: monthRange(addMonths(month, -2)).from, to: monthRange(month).to }
    case "this-year":
      return { from: `${today.slice(0, 4)}-01-01`, to: `${today.slice(0, 4)}-12-31` }
    case "custom":
      return { from: custom.from, to: custom.to }
  }
}
