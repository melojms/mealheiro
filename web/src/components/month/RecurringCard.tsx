import { Repeat } from "lucide-react"
import { Card } from "@/components/ui/card"
import { formatCents } from "@/lib/money"
import type { MonthReport } from "@/lib/types"

export function RecurringCard({ recurring }: { recurring: MonthReport["recurring"] }) {
  return (
    <Card size="sm" className="flex-row items-center gap-3 px-4">
      <span className="bg-primary/10 text-primary flex size-9 shrink-0 items-center justify-center rounded-full">
        <Repeat className="size-4" />
      </span>
      <div className="min-w-0 text-sm">
        <p>
          Recurring this month: <span className="font-semibold tabular-nums">{formatCents(recurring.committed_cents)}</span> committed
          {" · "}
          <span className={recurring.pending_count > 0 ? "font-semibold text-amber-600 dark:text-amber-400" : undefined}>
            {recurring.pending_count} pending
          </span>
        </p>
        <p className="text-muted-foreground text-xs">
          {recurring.entry_count} recurring {recurring.entry_count === 1 ? "expense" : "expenses"}
          {recurring.pending_cents > 0 && ` · ${formatCents(recurring.pending_cents)} still estimated`}
        </p>
      </div>
    </Card>
  )
}
