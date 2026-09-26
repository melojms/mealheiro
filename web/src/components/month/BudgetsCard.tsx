import { Target } from "lucide-react"
import { Link } from "react-router"
import { CategoryIcon } from "@/components/CategoryIcon"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { formatCents } from "@/lib/money"
import type { BudgetLine, Budgets } from "@/lib/types"
import { cn } from "@/lib/utils"
import { budgetState, type BudgetState } from "./metrics"

const STATE_BAR: Record<BudgetState, string | null> = { ok: null, warn: "#f59e0b", over: "#ef4444" }
const STATE_TEXT: Record<BudgetState, string> = {
  ok: "text-muted-foreground",
  warn: "text-amber-600 dark:text-amber-400",
  over: "text-rose-600 dark:text-rose-400",
}

function BudgetRow({ line, onSelect }: { line: BudgetLine; onSelect?: () => void }) {
  const state = budgetState(line.ratio)
  const color = STATE_BAR[state] ?? line.color
  const cap = Math.max(line.budget_cents, line.spent_cents, 1)
  const confirmedPct = ((line.spent_cents - line.pending_cents) / cap) * 100
  const pendingPct = (line.pending_cents / cap) * 100
  const left = line.budget_cents - line.spent_cents

  const Comp = onSelect ? "button" : "div"
  return (
    <Comp
      type={onSelect ? "button" : undefined}
      onClick={onSelect}
      className={cn("block w-full rounded-lg px-2 py-2 text-left", onSelect && "hover:bg-muted/60 transition-colors")}
    >
      <div className="flex items-center gap-2.5">
        <CategoryIcon icon={line.icon} color={line.color} size="sm" />
        <span className="min-w-0 flex-1 truncate text-sm font-medium">{line.name}</span>
        <span className={cn("text-xs font-semibold tabular-nums", STATE_TEXT[state])}>{Math.round(line.ratio * 100)}%</span>
      </div>
      <div
        className="bg-muted relative mt-2 flex h-2 overflow-hidden rounded-full"
        role="meter"
        aria-label={`${line.name} budget used`}
        aria-valuemin={0}
        aria-valuemax={line.budget_cents}
        aria-valuenow={line.spent_cents}
      >
        <div className="h-full transition-all" style={{ width: `${confirmedPct}%`, backgroundColor: color }} />
        {/* Pending estimates: same hue, hatched, so they read as "probably". */}
        <div
          className="h-full opacity-50"
          style={{
            width: `${pendingPct}%`,
            backgroundImage: `repeating-linear-gradient(135deg, ${color} 0 3px, transparent 3px 6px)`,
          }}
        />
        {line.spent_cents > line.budget_cents && (
          <div className="bg-background absolute inset-y-0 w-0.5" style={{ left: `${(line.budget_cents / cap) * 100}%` }} />
        )}
      </div>
      <div className="text-muted-foreground mt-1.5 flex justify-between gap-2 text-xs tabular-nums">
        <span>
          <span className="text-foreground font-medium">{formatCents(line.spent_cents)}</span> / {formatCents(line.budget_cents)}
          {line.pending_cents > 0 && <span> · {formatCents(line.pending_cents)} pending</span>}
        </span>
        <span className={cn(left < 0 && STATE_TEXT.over)}>
          {left >= 0 ? `${formatCents(left)} left` : `${formatCents(-left)} over`}
        </span>
      </div>
    </Comp>
  )
}

export function BudgetsCard({
  budgets, isPending, onCategory,
}: { budgets: Budgets | undefined; isPending: boolean; onCategory: (id: number) => void }) {
  const empty = !budgets?.overall && !budgets?.categories.length
  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Target className="text-muted-foreground size-4" /> Budgets
        </CardTitle>
      </CardHeader>
      <CardContent className="px-2">
        {isPending ? (
          <div className="space-y-4 px-2">
            {[0, 1, 2].map((i) => (
              <div key={i} className="space-y-2">
                <Skeleton className="h-4 w-32" />
                <Skeleton className="h-2 w-full" />
              </div>
            ))}
          </div>
        ) : empty ? (
          <div className="flex items-center justify-between gap-3 px-2 py-1">
            <p className="text-muted-foreground text-sm">No budgets set for this month.</p>
            <Button variant="outline" size="sm" asChild>
              <Link to="/settings">Set budgets</Link>
            </Button>
          </div>
        ) : (
          <div className="space-y-1">
            {budgets!.overall && <BudgetRow line={budgets!.overall} />}
            {budgets!.overall && budgets!.categories.length > 0 && <div className="bg-border mx-2 my-1 h-px" />}
            {budgets!.categories.map((l) => (
              <BudgetRow key={l.category_id} line={l} onSelect={() => l.category_id !== null && onCategory(l.category_id)} />
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  )
}
