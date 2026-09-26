// Shared data + URL-state hooks for the Month, Charts and Entries screens.
import { useCallback, useEffect, useMemo, useState, useSyncExternalStore } from "react"
import { useQuery } from "@tanstack/react-query"
import { useSearchParams } from "react-router"
import { toast } from "sonner"
import { api } from "@/lib/api"
import type { Category } from "@/lib/types"

export function useMeta() {
  return useQuery({ queryKey: ["meta"], queryFn: api.meta, staleTime: 5 * 60_000 })
}

export function usePeople() {
  return useQuery({ queryKey: ["people"], queryFn: api.people })
}

/** All categories incl. archived (history screens must resolve old ids), plus an id → category map. */
export function useCategories() {
  const q = useQuery({ queryKey: ["categories", "all"], queryFn: () => api.categories({ include_archived: true }) })
  const byId = useMemo(() => new Map<number, Category>((q.data ?? []).map((c) => [c.id, c])), [q.data])
  return { ...q, byId }
}

/**
 * A single search param as state. Setting null/"" removes it. Uses `replace` so filter
 * tweaks don't flood the history stack.
 */
export function useUrlParam(key: string): [string | null, (v: string | null) => void] {
  const [sp, setSp] = useSearchParams()
  const set = useCallback(
    (v: string | null) =>
      setSp(
        (prev) => {
          const next = new URLSearchParams(prev)
          if (v === null || v === "") next.delete(key)
          else next.set(key, v)
          return next
        },
        { replace: true },
      ),
    [key, setSp],
  )
  return [sp.get(key), set]
}

/** Payer filter from `?payer=<id>`; undefined = everyone. */
export function usePayerParam(): [number | undefined, (id: number | undefined) => void] {
  const [raw, set] = useUrlParam("payer")
  const n = Number(raw)
  const id = raw && Number.isInteger(n) && n > 0 ? n : undefined
  return [id, useCallback((v: number | undefined) => set(v === undefined ? null : String(v)), [set])]
}

/** Shows one toast per distinct error (deduped by id so refetch loops don't spam). */
export function useErrorToast(error: unknown, what: string) {
  useEffect(() => {
    if (error) toast.error(`Couldn't load ${what}`, { id: `load-${what}`, description: (error as Error).message })
  }, [error, what])
}

export function useDebouncedValue<T>(value: T, ms = 300): T {
  const [v, setV] = useState(value)
  useEffect(() => {
    const t = setTimeout(() => setV(value), ms)
    return () => clearTimeout(t)
  }, [value, ms])
  return v
}

export function useMediaQuery(query: string): boolean {
  return useSyncExternalStore(
    (cb) => {
      const mql = window.matchMedia(query)
      mql.addEventListener("change", cb)
      return () => mql.removeEventListener("change", cb)
    },
    () => window.matchMedia(query).matches,
    () => false,
  )
}

/** Tailwind `md` breakpoint. */
export const useIsDesktop = () => useMediaQuery("(min-width: 768px)")
