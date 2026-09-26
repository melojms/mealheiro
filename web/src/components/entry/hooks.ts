import { useSyncExternalStore } from "react"
import { useQuery } from "@tanstack/react-query"
import { api } from "@/lib/api"
import type { EntryType } from "@/lib/types"
import { toISODate } from "@/components/add/dates"

export const ENTRY_TYPES: { value: EntryType; label: string }[] = [
  { value: "expense", label: "Expense" },
  { value: "income", label: "Income" },
  { value: "investment", label: "Investment" },
]

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

export const useIsDesktop = () => useMediaQuery("(min-width: 768px)")

export function useMeta() {
  return useQuery({ queryKey: ["meta"], queryFn: api.meta, staleTime: 5 * 60_000 })
}

/** Today's date in the server's TZ (falls back to the device date while loading). */
export function useToday(): string {
  return useMeta().data?.today ?? toISODate(new Date())
}

export function usePeople() {
  return useQuery({ queryKey: ["people"], queryFn: api.people })
}

/** All categories of every type, ordered by usage (API order). */
export function useCategories(includeArchived = false) {
  return useQuery({
    queryKey: ["categories", { include_archived: includeArchived }],
    queryFn: () => api.categories(includeArchived ? { include_archived: true } : undefined),
  })
}
