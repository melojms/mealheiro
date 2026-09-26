// Local-date helpers on "YYYY-MM-DD" strings. Arithmetic runs in UTC so DST never shifts a day.

const pad = (n: number) => String(n).padStart(2, "0")

/** Local calendar date of a Date as YYYY-MM-DD. */
export function toISODate(d: Date): string {
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`
}

/** YYYY-MM-DD -> local Date at midnight (for calendar widgets). */
export function fromISODate(iso: string): Date {
  const [y, m, d] = iso.split("-").map(Number)
  return new Date(y, m - 1, d)
}

export function addDays(iso: string, days: number): string {
  const [y, m, d] = iso.split("-").map(Number)
  return new Date(Date.UTC(y, m - 1, d + days)).toISOString().slice(0, 10)
}

export type DateChip = "today" | "yesterday" | "other"

export function dateChip(iso: string, today: string): DateChip {
  if (iso === today) return "today"
  if (iso === addDays(today, -1)) return "yesterday"
  return "other"
}

// Fixed names: ICU versions disagree on "Sep" vs "Sept".
const MONTHS = ["Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"]

/** "Today" / "Yesterday" / "12 Mar" / "12 Mar 2025" (year only when different from today's). */
export function formatDay(iso: string, today: string): string {
  const chip = dateChip(iso, today)
  if (chip === "today") return "Today"
  if (chip === "yesterday") return "Yesterday"
  const [y, m, d] = iso.split("-").map(Number)
  const label = `${d} ${MONTHS[m - 1]}`
  return iso.slice(0, 4) === today.slice(0, 4) ? label : `${label} ${y}`
}

/** Shifts a YYYY-MM month by n months. */
export function addMonths(month: string, n: number): string {
  const [y, m] = month.split("-").map(Number)
  const total = y * 12 + (m - 1) + n
  return `${Math.floor(total / 12)}-${pad((total % 12) + 1)}`
}

/** Last day of a YYYY-MM month as YYYY-MM-DD. */
export function monthEnd(month: string): string {
  const [y, m] = month.split("-").map(Number)
  return `${month}-${pad(new Date(Date.UTC(y, m, 0)).getUTCDate())}`
}

/** "2026-03" -> "Mar 2026". */
export function formatMonth(month: string): string {
  const [y, m] = month.split("-").map(Number)
  return `${MONTHS[m - 1]} ${y}`
}
