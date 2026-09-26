// Entries screen filters: URL search params <-> API query. Pure so it can be unit-tested.
import { isDate, isPreset, presetRange, type DatePreset } from "@/components/filters/dates"
import { parseAmount } from "@/lib/money"
import type { EntryFilters, EntryType } from "@/lib/types"

export const FILTER_KEYS = ["q", "type", "cat", "payer", "tag", "min", "max", "range", "from", "to"] as const
export type FilterKey = (typeof FILTER_KEYS)[number]

/** Raw, user-facing filter values as stored in the URL (amounts stay as typed, e.g. "12,50"). */
export type UrlFilters = Partial<Record<FilterKey, string>>

const TYPES: EntryType[] = ["expense", "income", "investment"]

export function readUrlFilters(sp: URLSearchParams): UrlFilters {
  const f: UrlFilters = {}
  for (const k of FILTER_KEYS) {
    const v = sp.get(k)?.trim()
    if (v) f[k] = v
  }
  return f
}

/** Replaces all filter keys in `sp` with `f`, keeping unrelated params. */
export function writeUrlFilters(sp: URLSearchParams, f: UrlFilters): URLSearchParams {
  const next = new URLSearchParams(sp)
  for (const k of FILTER_KEYS) {
    const v = f[k]?.trim()
    if (v) next.set(k, v)
    else next.delete(k)
  }
  return next
}

const posInt = (s?: string) => {
  const n = Number(s)
  return s && Number.isInteger(n) && n > 0 ? n : undefined
}

export function datePreset(f: UrlFilters): DatePreset | undefined {
  if (isPreset(f.range ?? null)) return f.range as DatePreset
  // Bare from/to (e.g. a shared link) behave as a custom range.
  return f.from || f.to ? "custom" : undefined
}

/** Maps URL filters to API params; invalid values are dropped rather than sent. */
export function toApiFilters(f: UrlFilters, today: string): EntryFilters {
  const out: EntryFilters = {}
  if (f.q) out.q = f.q
  if (f.type && TYPES.includes(f.type as EntryType)) out.type = f.type as EntryType
  const cat = posInt(f.cat)
  if (cat) out.category_id = cat
  const payer = posInt(f.payer)
  if (payer) out.payer_id = payer
  if (f.tag) out.tag = f.tag.toLowerCase()
  const min = f.min ? parseAmount(f.min) : null
  if (min !== null) out.min_cents = min
  const max = f.max ? parseAmount(f.max) : null
  if (max !== null) out.max_cents = max

  const preset = datePreset(f)
  if (preset) {
    const r = presetRange(preset, today, { from: isDate(f.from) ? f.from : undefined, to: isDate(f.to) ? f.to : undefined })
    if (r.from) out.from = r.from
    if (r.to) out.to = r.to
  }
  return out
}

/** Number of active filters, excluding the free-text search (shown on the Filters button). */
export function activeFilterCount(f: UrlFilters): number {
  let n = 0
  if (f.type) n++
  if (f.cat) n++
  if (f.payer) n++
  if (f.tag) n++
  if (f.min || f.max) n++
  if (datePreset(f)) n++
  return n
}

/** Keys to clear when removing a single chip. */
export const CHIP_KEYS = {
  type: ["type", "cat"],
  cat: ["cat"],
  payer: ["payer"],
  tag: ["tag"],
  amount: ["min", "max"],
  date: ["range", "from", "to"],
} satisfies Record<string, FilterKey[]>
