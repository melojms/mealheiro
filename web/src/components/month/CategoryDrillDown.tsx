import { useQuery } from "@tanstack/react-query"
import { Inbox } from "lucide-react"
import { CategoryIcon } from "@/components/CategoryIcon"
import { EntryRow, EntryRowSkeleton } from "@/components/entries/EntryRow"
import { monthLabel, monthRange } from "@/components/filters/dates"
import { useCategories, useErrorToast, useIsDesktop } from "@/components/filters/hooks"
import { EmptyState, ErrorState } from "@/components/filters/states"
import { Sheet, SheetContent, SheetDescription, SheetHeader, SheetTitle } from "@/components/ui/sheet"
import { api } from "@/lib/api"
import { formatCents, formatPct } from "@/lib/money"
import type { CategoryAmount, Entry } from "@/lib/types"
import { cn } from "@/lib/utils"
import { share, sumAmount } from "./metrics"

export type DrillTarget = Pick<CategoryAmount, "category_id" | "name" | "icon" | "color"> & Partial<CategoryAmount>

/** Category detail for one month: subcategory split + the month's entries. */
export function CategoryDrillDown({
  target, open, onOpenChange, month, payerId, onEntry,
}: {
  target: DrillTarget | null
  open: boolean
  onOpenChange: (open: boolean) => void
  month: string
  payerId: number | undefined
  onEntry: (e: Entry) => void
}) {
  const desktop = useIsDesktop()
  const { byId } = useCategories()
  const range = monthRange(month)
  const entries = useQuery({
    queryKey: ["entries", "drill", target?.category_id, month, payerId],
    queryFn: () => api.entries({ category_id: target!.category_id, ...range, payer_id: payerId, limit: 500 }),
    enabled: open && !!target,
  })
  useErrorToast(entries.error, "category entries")

  const subs = target?.subcategories ?? []
  const subTotal = sumAmount(subs)
  const maxSub = Math.max(0, ...subs.map((s) => s.amount_cents))
  // Fall back to the entries' own total when the target came from an insight/budget without report data.
  const total = target?.amount_cents ?? entries.data?.totals.expense_cents

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent
        side={desktop ? "right" : "bottom"}
        className={cn("gap-0 p-0", desktop ? "w-full sm:max-w-md" : "max-h-[88svh] rounded-t-2xl")}
      >
        {target && (
          <>
            <SheetHeader className="flex-row items-center gap-3 border-b pr-12">
              <CategoryIcon icon={target.icon} color={target.color} size="lg" />
              <div className="min-w-0">
                <SheetTitle className="truncate text-lg">{target.name}</SheetTitle>
                <SheetDescription>
                  {monthLabel(month)}
                  {total !== undefined && (
                    <>
                      {" · "}
                      <span className="text-foreground font-semibold tabular-nums">{formatCents(total)}</span>
                    </>
                  )}
                </SheetDescription>
              </div>
            </SheetHeader>

            <div className="flex-1 overflow-y-auto overscroll-contain pb-[env(safe-area-inset-bottom)]">
              {subs.length > 1 && (
                <section className="border-b p-4">
                  <h3 className="text-muted-foreground mb-3 text-xs font-medium tracking-wide uppercase">Subcategories</h3>
                  <ul className="space-y-2.5">
                    {subs.map((s) => (
                      <li key={`${s.category_id}-${s.name}`}>
                        <div className="flex justify-between gap-2 text-sm">
                          <span className="truncate">{s.name}</span>
                          <span className="shrink-0 tabular-nums">
                            <span className="font-medium">{formatCents(s.amount_cents)}</span>
                            <span className="text-muted-foreground ml-2 text-xs">{formatPct(share(s.amount_cents, subTotal))}</span>
                          </span>
                        </div>
                        <div className="bg-muted mt-1 h-2 overflow-hidden rounded-full">
                          <div
                            className="h-full rounded-full"
                            style={{ width: `${share(s.amount_cents, maxSub) * 100}%`, backgroundColor: s.color || target.color }}
                          />
                        </div>
                      </li>
                    ))}
                  </ul>
                </section>
              )}

              <section className="p-2">
                <h3 className="text-muted-foreground px-2 pt-2 pb-1 text-xs font-medium tracking-wide uppercase">
                  Entries{entries.data ? ` · ${entries.data.total_count}` : ""}
                </h3>
                {entries.isPending ? (
                  [0, 1, 2, 3].map((i) => <EntryRowSkeleton key={i} />)
                ) : entries.isError ? (
                  <ErrorState onRetry={() => entries.refetch()} />
                ) : !entries.data.entries.length ? (
                  <EmptyState icon={Inbox} title="No entries" description="Nothing booked in this category this month." />
                ) : (
                  entries.data.entries.map((e) => (
                    <EntryRow key={e.id} entry={e} category={byId.get(e.category_id)} onClick={() => onEntry(e)} showDate />
                  ))
                )}
              </section>
            </div>
          </>
        )}
      </SheetContent>
    </Sheet>
  )
}
