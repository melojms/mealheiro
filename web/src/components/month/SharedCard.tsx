import { useQuery } from "@tanstack/react-query"
import { Scale } from "lucide-react"
import { useState } from "react"
import { Money } from "@/components/Money"
import { monthLabel } from "@/components/filters/dates"
import { useErrorToast } from "@/components/filters/hooks"
import { ErrorState } from "@/components/filters/states"
import { Card, CardAction, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { ToggleGroup, ToggleGroupItem } from "@/components/ui/toggle-group"
import { api } from "@/lib/api"
import { formatCents } from "@/lib/money"
import type { SharedReport } from "@/lib/types"
import { PERSON_COLORS, splitPercents } from "./shared"

type Scope = "month" | "year"

/**
 * How much each person paid towards shared (non-personal) expenses in the month
 * or its calendar year. Household-wide: ignores the page's payer filter.
 */
export function SharedCard({ month }: { month: string }) {
  const [scope, setScope] = useState<Scope>("month")
  const year = Number(month.slice(0, 4))
  const report = useQuery({
    queryKey: ["reports", "shared", scope, scope === "month" ? month : year],
    queryFn: () => api.sharedReport(scope === "month" ? { month } : { year }),
  })
  useErrorToast(report.error, "shared expenses")
  const period = scope === "month" ? monthLabel(month) : String(year)

  return (
    <Card size="sm" role="region" aria-label="Shared expenses">
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Scale className="text-muted-foreground size-4" /> Shared expenses
        </CardTitle>
        <CardAction>
          <ToggleGroup
            type="single"
            variant="outline"
            size="sm"
            spacing={0}
            value={scope}
            // Radix emits "" when the active item is clicked again; treat that as "keep".
            onValueChange={(v) => v && setScope(v as Scope)}
            aria-label="Shared expenses period"
          >
            <ToggleGroupItem value="month" className="data-[state=on]:bg-primary data-[state=on]:text-primary-foreground px-3">
              Month
            </ToggleGroupItem>
            <ToggleGroupItem value="year" className="data-[state=on]:bg-primary data-[state=on]:text-primary-foreground px-3">
              Year
            </ToggleGroupItem>
          </ToggleGroup>
        </CardAction>
      </CardHeader>
      <CardContent>
        {report.isError ? (
          <ErrorState onRetry={() => report.refetch()} />
        ) : report.data ? (
          <SharedSplit report={report.data} period={period} />
        ) : (
          <div className="space-y-3">
            <Skeleton className="h-6 w-32" />
            <Skeleton className="h-2.5 w-full rounded-full" />
            <Skeleton className="h-4 w-48" />
          </div>
        )}
      </CardContent>
    </Card>
  )
}

function SharedSplit({ report, period }: { report: SharedReport; period: string }) {
  const pcts = splitPercents(report.people.map((p) => p.amount_cents))
  if (!pcts) return <p className="text-muted-foreground text-sm">No shared expenses in {period}.</p>

  const people = report.people.map((p, i) => ({ ...p, pct: pcts[i], color: PERSON_COLORS[i % PERSON_COLORS.length] }))
  return (
    <div className="space-y-3">
      <p className="text-sm">
        <Money cents={report.total_cents} className="text-xl font-semibold" />
        <span className="text-muted-foreground"> shared in {period}</span>
        {report.pending_cents > 0 && (
          <span className="block text-xs text-amber-600 dark:text-amber-400">incl. {formatCents(report.pending_cents)} estimated</span>
        )}
      </p>

      <div
        className="bg-muted flex h-2.5 overflow-hidden rounded-full"
        role="img"
        aria-label={people.map((p) => `${p.name} ${p.pct}%`).join(", ")}
      >
        {people.map((p) => (
          <div key={p.person_id} className="h-full transition-all" style={{ width: `${p.pct}%`, backgroundColor: p.color }} />
        ))}
      </div>

      <ul className="grid grid-cols-2 gap-3">
        {people.map((p) => (
          <li key={p.person_id} className="min-w-0">
            <div className="flex items-center gap-1.5 text-sm">
              <span className="size-2.5 shrink-0 rounded-full" style={{ backgroundColor: p.color }} />
              <span className="truncate font-medium">{p.name}</span>
              <span className="ml-auto font-semibold tabular-nums">{p.pct}%</span>
            </div>
            <p className="text-muted-foreground pl-4 text-xs tabular-nums">
              {formatCents(p.amount_cents)}
              {p.pending_cents > 0 && <span className="text-amber-600 dark:text-amber-400"> · {formatCents(p.pending_cents)} est.</span>}
            </p>
          </li>
        ))}
      </ul>
    </div>
  )
}
