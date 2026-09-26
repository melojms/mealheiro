// Keypad amount handling. The amount is kept as the raw typed string with "," as decimal separator
// (e.g. "", "12", "12,", "12,5"), which makes backspace trivial and display exact.
import { parseAmount } from "@/lib/money"

export type KeypadKey = "0" | "1" | "2" | "3" | "4" | "5" | "6" | "7" | "8" | "9" | "," | "backspace" | "clear"

const MAX_INT_DIGITS = 7 // up to 9 999 999,99 €

/** Applies one keypad key to the typed amount string. Invalid keys leave it unchanged. */
export function pressKey(amount: string, key: KeypadKey): string {
  if (key === "clear") return ""
  if (key === "backspace") return amount.slice(0, -1)
  const comma = amount.indexOf(",")
  if (key === ",") {
    if (comma >= 0) return amount
    return amount === "" ? "0," : `${amount},`
  }
  if (comma >= 0) return amount.length - comma - 1 >= 2 ? amount : amount + key
  if (amount === "0") return key
  if (amount.length >= MAX_INT_DIGITS) return amount
  return amount + key
}

/** Maps a hardware keyboard key to a keypad key (null = not handled). */
export function keyFromKeyboard(key: string): KeypadKey | null {
  if (/^[0-9]$/.test(key)) return key as KeypadKey
  if (key === "," || key === ".") return ","
  if (key === "Backspace") return "backspace"
  if (key === "Escape" || key === "Delete") return "clear"
  return null
}

/** Positive cents for the typed amount, or null when empty/zero. */
export function amountCents(amount: string): number | null {
  const cents = parseAmount(amount)
  return cents && cents > 0 ? cents : null
}

const intFormat = new Intl.NumberFormat("pt-PT", { useGrouping: "min2" })

/**
 * Splits the typed amount into the part to show in full and a muted "ghost" suffix that
 * completes it to two decimals: "12" -> {main: "12", ghost: ",00"}; "12,5" -> {main: "12,5", ghost: "0"}.
 */
export function displayParts(amount: string): { main: string; ghost: string } {
  const [int, frac] = amount.split(",")
  const main = (int === "" ? "0" : intFormat.format(Number(int)).replace(/\s/g, " ")) + (frac === undefined ? "" : `,${frac}`)
  if (frac === undefined) return { main, ghost: ",00" }
  return { main, ghost: "0".repeat(2 - frac.length) }
}

/** 1250 -> "12,50", for prefilling a text amount input. */
export function centsToInput(cents: number): string {
  return `${Math.floor(cents / 100)},${String(cents % 100).padStart(2, "0")}`
}
