import { ChevronDown } from "lucide-react"
import { useState } from "react"
import { CategoryIcon } from "@/components/CategoryIcon"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { formatCents } from "@/lib/money"
import type { EntryList, EntryType } from "@/lib/types"
import { cn } from "@/lib/utils"

const TOTALS: { type: EntryType; key: keyof EntryList["totals"]; label: string; className: string }[] = [
  { type: "expense", key: "expense_cents", label: "Expenses", className: "" },
  { type: "income", key: "income_cents", label: "Income", className: "text-emerald-600 dark:text-emerald-400" },
  { type: "investment", key: "investment_cents", label: "Investments", className: "text-sky-600 dark:text-sky-400" },
]
const COLLAPSED = 5

/** Totals per type + per-category bars for whatever the current filter matches (e.g. a tag). */
export function EntriesSummary({ data, type }: { data: EntryList | undefined; type?: EntryType }) {
  const [expanded, setExpanded] = useState(false)
  if (!data) {
    return (
      <Card size="sm" className="px-3.5">
        <div className="grid grid-cols-3 gap-3">
          {[0, 1, 2].map((i) => (
            <div key={i} className="space-y-1.5">
              <Skeleton className="h-3 w-14" />
              <Skeleton className="h-5 w-20" />
            </div>
          ))}
        </div>
      </Card>
    )
  }

  const totals = TOTALS.filter((t) => (type ? t.type === type : data.totals[t.key] > 0 || t.type === "expense"))
  const breakdown = expanded ? data.breakdown : data.breakdown.slice(0, COLLAPSED)
  // Bars scale per type so a salary doesn't flatten every expense bar.
  const maxByType = new Map<EntryType, number>()
  for (const b of data.breakdown) maxByType.set(b.type, Math.max(maxByType.get(b.type) ?? 0, b.amount_cents))

  return (
    <Card size="sm">
      <CardContent className="space-y-3">
        <div className="flex flex-wrap items-end gap-x-6 gap-y-2">
          {totals.map((t) => (
            <div key={t.key}>
              <div className="text-muted-foreground text-xs">{t.label}</div>
              <div className={cn("text-lg font-semibold tabular-nums", t.className)}>{formatCents(data.totals[t.key])}</div>
            </div>
          ))}
          <div className="text-muted-foreground ml-auto text-xs tabular-nums">
            {data.total_count} {data.total_count === 1 ? "entry" : "entries"}
          </div>
        </div>

        {data.breakdown.length > 1 && (
          <div className="space-y-1.5 border-t pt-3">
            {breakdown.map((b) => (
              <div key={b.category_id} className="flex items-center gap-2.5">
                <CategoryIcon icon={b.icon} color={b.color} size="sm" className="size-6" />
                <span className="w-24 shrink-0 truncate text-xs sm:w-32">{b.name}</span>
                <div className="bg-muted h-1.5 min-w-0 flex-1 overflow-hidden rounded-full">
                  <div
                    className="h-full rounded-full"
                    style={{ width: `${(b.amount_cents / (maxByType.get(b.type) || 1)) * 100}%`, backgroundColor: b.color }}
                  />
                </div>
                <span className="w-20 shrink-0 text-right text-xs font-medium tabular-nums">{formatCents(b.amount_cents)}</span>
              </div>
            ))}
            {data.breakdown.length > COLLAPSED && (
              <Button variant="ghost" size="xs" className="text-muted-foreground" onClick={() => setExpanded((e) => !e)}>
                <ChevronDown className={cn("transition-transform", expanded && "rotate-180")} />
                {expanded ? "Show less" : `Show all ${data.breakdown.length} categories`}
              </Button>
            )}
          </div>
        )}
      </CardContent>
    </Card>
  )
}
