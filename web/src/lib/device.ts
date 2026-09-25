// Per-device preferences stored in localStorage (never shared between devices).
import { useSyncExternalStore } from "react"

const PAYER_KEY = "mm-budget.default-payer"
const listeners = new Set<() => void>()

function read(): number | null {
  try {
    const v = localStorage.getItem(PAYER_KEY)
    return v ? Number(v) : null
  } catch {
    return null
  }
}

export function setDefaultPayer(id: number | null) {
  try {
    if (id === null) localStorage.removeItem(PAYER_KEY)
    else localStorage.setItem(PAYER_KEY, String(id))
  } catch {
    /* storage unavailable */
  }
  listeners.forEach((l) => l())
}

/** Device default payer id (null = not set yet). */
export function useDefaultPayer(): number | null {
  return useSyncExternalStore(
    (cb) => {
      listeners.add(cb)
      return () => listeners.delete(cb)
    },
    read,
    () => null,
  )
}
