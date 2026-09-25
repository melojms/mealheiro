const eur = new Intl.NumberFormat("pt-PT", { style: "currency", currency: "EUR", minimumFractionDigits: 2 })
const eurCompact = new Intl.NumberFormat("pt-PT", { style: "currency", currency: "EUR", maximumFractionDigits: 0 })

/** 123456 -> "1234,56 €" (pt-PT). Uses a normal space before €. */
export function formatCents(cents: number): string {
  return eur.format(cents / 100).replace(/ /g, " ")
}

/** Rounded to whole euros, for chart axes/labels: 123456 -> "1235 €". */
export function formatCentsCompact(cents: number): string {
  return eurCompact.format(Math.round(cents / 100)).replace(/ /g, " ")
}

/**
 * Parses user input into positive cents. Accepts "12,50", "12.5", "1 234,56", "1.234,56", "1,234.56", "7".
 * The last '.' or ',' followed by 1-2 digits is the decimal separator. Returns null when invalid.
 */
export function parseAmount(input: string): number | null {
  let s = input.trim().replace(/€/g, "").replace(/\s/g, "")
  if (s === "" || /[^0-9.,]/.test(s)) return null
  let intPart = s
  let frac = ""
  const i = Math.max(s.lastIndexOf("."), s.lastIndexOf(","))
  if (i >= 0 && s.length - i - 1 <= 2) {
    intPart = s.slice(0, i)
    frac = s.slice(i + 1)
  }
  intPart = intPart.replace(/[.,]/g, "")
  if (intPart === "") intPart = "0"
  s = intPart + frac.padEnd(2, "0")
  if (!/^\d+$/.test(s)) return null
  const cents = Number(s)
  return Number.isSafeInteger(cents) ? cents : null
}

/** Signed percentage change a→b as a fraction, or null when a is 0. */
export function pctChange(from: number, to: number): number | null {
  if (from === 0) return null
  return (to - from) / from
}

export function formatPct(fraction: number | null, digits = 0): string {
  if (fraction === null || !Number.isFinite(fraction)) return "—"
  return `${(fraction * 100).toFixed(digits)}%`
}
