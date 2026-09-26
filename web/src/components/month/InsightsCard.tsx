import {
  ChevronRight, Gauge, Lightbulb, PiggyBank, Receipt, Repeat, TrendingDown, TrendingUp, type LucideIcon,
} from "lucide-react"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import type { Insight } from "@/lib/types"
import { cn } from "@/lib/utils"

const KIND_ICON: Record<Insight["kind"], LucideIcon> = {
  category_up: TrendingUp,
  category_down: TrendingDown,
  pace: Gauge,
  top_expense: Receipt,
  savings_rate: PiggyBank,
  new_recurring: Repeat,
}

const SEVERITY: Record<Insight["severity"], { chip: string; label: string }> = {
  warning: { chip: "bg-amber-500/15 text-amber-700 dark:text-amber-400", label: "Warning" },
  good: { chip: "bg-emerald-500/15 text-emerald-700 dark:text-emerald-400", label: "Good news" },
  info: { chip: "bg-sky-500/15 text-sky-700 dark:text-sky-400", label: "Info" },
}

export function InsightsCard({
  insights, isPending, onCategory, onEntry,
}: {
  insights: Insight[] | undefined
  isPending: boolean
  onCategory: (id: number) => void
  onEntry: (id: number) => void
}) {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Lightbulb className="text-muted-foreground size-4" /> Insights
        </CardTitle>
      </CardHeader>
      <CardContent className="px-2">
        {isPending ? (
          <div className="space-y-3 px-2">
            {[0, 1, 2].map((i) => (
              <div key={i} className="flex gap-3">
                <Skeleton className="size-8 rounded-full" />
                <div className="flex-1 space-y-1.5">
                  <Skeleton className="h-3.5 w-32" />
                  <Skeleton className="h-3 w-48" />
                </div>
              </div>
            ))}
          </div>
        ) : !insights?.length ? (
          <p className="text-muted-foreground px-2 py-3 text-sm">Nothing notable yet — keep adding entries.</p>
        ) : (
          <ul className="space-y-0.5">
            {insights.map((ins, i) => {
              const Icon = KIND_ICON[ins.kind]
              const sev = SEVERITY[ins.severity]
              const action =
                ins.entry_id !== null
                  ? () => onEntry(ins.entry_id!)
                  : ins.category_id !== null
                    ? () => onCategory(ins.category_id!)
                    : undefined
              const body = (
                <>
                  <span className={cn("flex size-8 shrink-0 items-center justify-center rounded-full", sev.chip)}>
                    <Icon className="size-4" aria-label={sev.label} />
                  </span>
                  <span className="min-w-0 flex-1">
                    <span className="block text-sm font-medium">{ins.title}</span>
                    <span className="text-muted-foreground block text-xs">{ins.detail}</span>
                  </span>
                  {action && <ChevronRight className="text-muted-foreground size-4 shrink-0" />}
                </>
              )
              return (
                <li key={`${ins.kind}-${i}`}>
                  {action ? (
                    <button
                      type="button"
                      onClick={action}
                      className="hover:bg-muted/60 flex w-full items-center gap-3 rounded-lg px-2 py-2 text-left transition-colors"
                    >
                      {body}
                    </button>
                  ) : (
                    <div className="flex items-center gap-3 px-2 py-2">{body}</div>
                  )}
                </li>
              )
            })}
          </ul>
        )}
      </CardContent>
    </Card>
  )
}
