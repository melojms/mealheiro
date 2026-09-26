import { useInfiniteQuery } from "@tanstack/react-query"
import { ListFilter, Loader2, Search, SearchX, X } from "lucide-react"
import { useEffect, useMemo, useState } from "react"
import { useSearchParams } from "react-router"
import { PageHeader } from "@/components/PageHeader"
import { EntryEditDialog } from "@/components/entry/EntryEditDialog"
import { EntriesSummary } from "@/components/entries/EntriesSummary"
import { EntryRow, EntryRowSkeleton } from "@/components/entries/EntryRow"
import { FilterSheet } from "@/components/entries/FilterSheet"
import {
  activeFilterCount, CHIP_KEYS, datePreset, readUrlFilters, toApiFilters, writeUrlFilters, type UrlFilters,
} from "@/components/entries/filters"
import { groupByDate } from "@/components/entries/group"
import { DATE_PRESETS, dayLabel, shortDate } from "@/components/filters/dates"
import { useCategories, useDebouncedValue, useErrorToast, useMeta, usePeople } from "@/components/filters/hooks"
import { EmptyState, ErrorState } from "@/components/filters/states"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { InputGroup, InputGroupAddon, InputGroupButton, InputGroupInput } from "@/components/ui/input-group"
import { api } from "@/lib/api"
import { formatCents } from "@/lib/money"
import type { Entry, EntryType } from "@/lib/types"

const PAGE_SIZE = 50
const TYPE_LABEL: Record<EntryType, string> = { expense: "Expenses", income: "Income", investment: "Investments" }

export default function EntriesPage() {
  const [sp, setSp] = useSearchParams()
  const filters = useMemo(() => readUrlFilters(sp), [sp])
  const setFilters = (f: UrlFilters) => setSp((prev) => writeUrlFilters(prev, f), { replace: true })

  const meta = useMeta()
  const { byId } = useCategories()
  const { data: people } = usePeople()
  const today = meta.data?.today

  // Debounced search box → URL `q`.
  const [search, setSearch] = useState(filters.q ?? "")
  const debounced = useDebouncedValue(search.trim(), 300)
  useEffect(() => {
    if (debounced !== (filters.q ?? "")) setSp((prev) => writeUrlFilters(prev, { ...readUrlFilters(prev), q: debounced }), { replace: true })
    // Only react to the debounced text, not to other URL changes.
  }, [debounced]) // oxlint-disable-line react-hooks/exhaustive-deps

  const apiFilters = useMemo(() => (today ? toApiFilters(filters, today) : undefined), [filters, today])
  const list = useInfiniteQuery({
    queryKey: ["entries", "list", apiFilters],
    queryFn: ({ pageParam }) => api.entries({ ...apiFilters, limit: PAGE_SIZE, offset: pageParam }),
    initialPageParam: 0,
    getNextPageParam: (last, pages) => {
      const loaded = pages.reduce((n, p) => n + p.entries.length, 0)
      return loaded < last.total_count && last.entries.length > 0 ? loaded : undefined
    },
    enabled: !!apiFilters,
  })
  useErrorToast(list.error, "entries")

  const entries = useMemo(() => list.data?.pages.flatMap((p) => p.entries) ?? [], [list.data])
  const groups = useMemo(() => groupByDate(entries), [entries])
  const first = list.data?.pages[0]

  const [sheetOpen, setSheetOpen] = useState(false)
  const [editing, setEditing] = useState<Entry | null>(null)
  const [editOpen, setEditOpen] = useState(false)

  const clear = (keys: readonly string[]) =>
    setFilters(Object.fromEntries(Object.entries(filters).filter(([k]) => !keys.includes(k))) as UrlFilters)

  // Human-readable chips for active filters.
  const chips: { key: keyof typeof CHIP_KEYS; label: string }[] = []
  if (filters.type) chips.push({ key: "type", label: TYPE_LABEL[filters.type as EntryType] ?? filters.type })
  if (filters.cat) chips.push({ key: "cat", label: byId.get(Number(filters.cat))?.name ?? "Category" })
  if (filters.payer) chips.push({ key: "payer", label: people?.find((p) => String(p.id) === filters.payer)?.name ?? "Payer" })
  if (filters.tag) chips.push({ key: "tag", label: `#${filters.tag}` })
  if (filters.min || filters.max)
    chips.push({
      key: "amount",
      label: filters.min && filters.max ? `${filters.min} – ${filters.max} €` : filters.min ? `≥ ${filters.min} €` : `≤ ${filters.max} €`,
    })
  const preset = datePreset(filters)
  if (preset)
    chips.push({
      key: "date",
      label:
        preset === "custom"
          ? `${filters.from ? shortDate(filters.from) : "…"} – ${filters.to ? shortDate(filters.to) : "…"}`
          : DATE_PRESETS.find((p) => p.value === preset)!.label,
    })

  const count = activeFilterCount(filters)
  const filtered = count > 0 || !!filters.q

  return (
    <div className="space-y-4">
      <PageHeader title="Entries" description="Search, filter and review everything you've logged" />

      <div className="flex gap-2">
        <InputGroup className="h-10 flex-1">
          <InputGroupAddon>
            <Search />
          </InputGroupAddon>
          <InputGroupInput
            type="search"
            placeholder="Search notes, categories, tags…"
            aria-label="Search entries"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="[&::-webkit-search-cancel-button]:hidden"
          />
          {search && (
            <InputGroupAddon align="inline-end">
              <InputGroupButton size="icon-xs" aria-label="Clear search" onClick={() => setSearch("")}>
                <X />
              </InputGroupButton>
            </InputGroupAddon>
          )}
        </InputGroup>
        <Button variant="outline" className="h-10 px-3" onClick={() => setSheetOpen(true)}>
          <ListFilter /> <span className="hidden sm:inline">Filters</span>
          {count > 0 && <Badge className="h-5 min-w-5 px-1.5 tabular-nums">{count}</Badge>}
        </Button>
      </div>

      {chips.length > 0 && (
        <div className="flex flex-wrap items-center gap-1.5">
          {chips.map((c) => (
            <Badge key={c.key} variant="secondary" className="h-7 gap-1 pr-1 pl-2.5 text-xs">
              {c.label}
              <button
                type="button"
                onClick={() => clear(CHIP_KEYS[c.key])}
                aria-label={`Remove ${c.label} filter`}
                className="hover:bg-foreground/10 rounded-full p-0.5"
              >
                <X className="size-3" />
              </button>
            </Badge>
          ))}
          <Button variant="ghost" size="xs" className="text-muted-foreground" onClick={() => setFilters({ q: filters.q })}>
            Clear all
          </Button>
        </div>
      )}

      {!list.isError && <EntriesSummary data={first} type={apiFilters?.type} />}

      {list.isError ? (
        <ErrorState onRetry={() => list.refetch()} />
      ) : list.isPending ? (
        <Card className="gap-0 px-2 py-2">
          {Array.from({ length: 8 }, (_, i) => (
            <EntryRowSkeleton key={i} />
          ))}
        </Card>
      ) : groups.length === 0 ? (
        <Card>
          <EmptyState
            icon={SearchX}
            title={filtered ? "No matching entries" : "No entries yet"}
            description={filtered ? "Try a different search or loosen the filters." : "Entries you add will be listed here."}
          >
            {filtered && (
              <Button variant="outline" size="sm" className="mt-1" onClick={() => { setSearch(""); setFilters({}) }}>
                Clear filters
              </Button>
            )}
          </EmptyState>
        </Card>
      ) : (
        // Not a <Card>: its overflow-hidden would break the sticky day headers.
        <div className="bg-card text-card-foreground ring-foreground/10 rounded-xl ring-1">
          {groups.map((g) => (
            <section key={g.date} aria-label={dayLabel(g.date, today ?? "")}>
              <h2 className="bg-card/95 text-muted-foreground sticky top-0 z-10 flex items-center justify-between border-y px-4 py-2 text-xs font-medium backdrop-blur [section:first-child>&]:rounded-t-xl [section:first-child>&]:border-t-0">
                <span>{dayLabel(g.date, today ?? "")}</span>
                {g.expense_cents > 0 && <span className="tabular-nums">{formatCents(g.expense_cents)}</span>}
              </h2>
              <div className="px-2 py-1">
                {g.entries.map((e) => (
                  <EntryRow
                    key={e.id}
                    entry={e}
                    category={byId.get(e.category_id)}
                    onClick={() => {
                      setEditing(e)
                      setEditOpen(true)
                    }}
                  />
                ))}
              </div>
            </section>
          ))}
          {list.hasNextPage && (
            <div className="border-t p-3">
              <Button variant="outline" className="w-full" disabled={list.isFetchingNextPage} onClick={() => list.fetchNextPage()}>
                {list.isFetchingNextPage && <Loader2 className="animate-spin" />}
                Load more
                <span className="text-muted-foreground tabular-nums">
                  ({entries.length} of {first?.total_count})
                </span>
              </Button>
            </div>
          )}
        </div>
      )}

      <FilterSheet open={sheetOpen} onOpenChange={setSheetOpen} value={filters} onApply={(f) => setFilters({ ...f, q: filters.q })} />
      <EntryEditDialog entry={editing} open={editOpen} onOpenChange={setEditOpen} />
    </div>
  )
}
