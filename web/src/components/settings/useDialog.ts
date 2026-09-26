import { useState } from "react"

/**
 * Open/close state for a dialog that edits some payload. The payload survives closing (so the exit
 * animation still has content) and `key` changes on every show so the form remounts with fresh state.
 */
export function useDialog<T>() {
  const [state, setState] = useState<{ payload: T | null; open: boolean; key: number }>({ payload: null, open: false, key: 0 })
  return {
    ...state,
    show: (payload: T) => setState((s) => ({ payload, open: true, key: s.key + 1 })),
    setOpen: (open: boolean) => setState((s) => ({ ...s, open })),
  }
}
