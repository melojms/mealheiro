import { useState } from "react"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { toast } from "sonner"
import { Loader2 } from "lucide-react"
import { CategoryIcon } from "@/components/CategoryIcon"
import { Money } from "@/components/Money"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Progress } from "@/components/ui/progress"
import { Skeleton } from "@/components/ui/skeleton"
import { centsToInput } from "@/components/add/amount"
import { budgetLevel } from "@/components/add/budget"
import { formatMonth } from "@/components/add/dates"
import { useCategories } from "@/components/entry/hooks"
import { api } from "@/lib/api"
import { parseAmount } from "@/lib/money"
import type { BudgetLine } from "@/lib/types"
import { cn } from "@/lib/utils"

const OVERALL = "overall"

interface Row {
  key: string
  categoryId: number | null
  name: string
  icon: string
  color: string
  line: BudgetLine | undefined
}

/** Draft value -> cents to save (null = remove) or "invalid". */
function draftCents(draft: string): number | null | "invalid" {
  if (draft.trim() === "") return null
  const c = parseAmount(draft)
  return c && c > 0 ? c : "invalid"
}

export function BudgetsSection() {
  const queryClient = useQueryClient()
  const budgets = useQuery({ queryKey: ["budgets"], queryFn: () => api.budgets() })
  const categories = useCategories()
  const [drafts, setDrafts] = useState<Record<string, string>>({})

  const lines = new Map((budgets.data?.categories ?? []).map((l) => [l.category_id, l]))
  const rows: Row[] = [
    { key: OVERALL, categoryId: null, name: "Overall monthly cap", icon: "wallet", color: "#64748b", line: budgets.data?.overall ?? undefined },
    ...(categories.data ?? [])
      .filter((c) => c.type === "expense" && c.parent_id === null)
      .sort((a, b) => a.name.localeCompare(b.name))
      .map((c) => ({ key: String(c.id), categoryId: c.id, name: c.name, icon: c.icon, color: c.color, line: lines.get(c.id) })),
  ]

  const valueOf = (r: Row) => drafts[r.key] ?? (r.line ? centsToInput(r.line.budget_cents) : "")
  const changes = rows.flatMap((r) => {
    if (!(r.key in drafts)) return []
    const cents = draftCents(drafts[r.key])
    if (cents === "invalid" || cents === (r.line?.budget_cents ?? null)) return []
    return [{ categoryId: r.categoryId, cents }]
  })
  const hasInvalid = rows.some((r) => r.key in drafts && draftCents(drafts[r.key]) === "invalid")

  const save = useMutation({
    mutationFn: () => Promise.all(changes.map((c) => api.setBudget(c.categoryId, c.cents))),
    onSuccess: () => {
      queryClient.invalidateQueries()
      setDrafts({})
      toast.success(changes.length === 1 ? "Budget saved" : `${changes.length} budgets saved`)
    },
    onError: (e) => {
      queryClient.invalidateQueries()
      toast.error(e.message)
    },
  })

  const loading = budgets.isLoading || categories.isLoading

  return (
    <div className="space-y-4">
      <p className="text-muted-foreground text-sm">
        One monthly amount per top-level expense category (it covers its subcategories). Leave empty for no budget. Changes apply
        from {budgets.data ? formatMonth(budgets.data.month) : "this month"} onward.
      </p>

      <div className="bg-card divide-y rounded-xl border">
        {loading
          ? Array.from({ length: 6 }, (_, i) => (
              <div key={i} className="flex items-center gap-3 p-3">
                <Skeleton className="size-9 rounded-full" />
                <Skeleton className="h-4 flex-1" />
                <Skeleton className="h-9 w-28" />
              </div>
            ))
          : rows.map((r) => {
              const invalid = r.key in drafts && draftCents(drafts[r.key]) === "invalid"
              return (
                <div key={r.key} className={cn("flex items-center gap-3 p-3", r.key === OVERALL && "bg-muted/30 rounded-t-xl")}>
                  <CategoryIcon icon={r.icon} color={r.color} />
                  <div className="min-w-0 flex-1 space-y-1">
                    <label htmlFor={`budget-${r.key}`} className="block truncate text-sm font-medium">
                      {r.name}
                    </label>
                    {r.line ? <Usage line={r.line} /> : <p className="text-muted-foreground text-xs">No budget</p>}
                  </div>
                  <div className="relative w-28 shrink-0 sm:w-32">
                    <Input
                      id={`budget-${r.key}`}
                      inputMode="decimal"
                      autoComplete="off"
                      placeholder="—"
                      value={valueOf(r)}
                      aria-invalid={invalid}
                      onChange={(e) => setDrafts((d) => ({ ...d, [r.key]: e.target.value }))}
                      className="h-9 pr-7 text-right tabular-nums"
                    />
                    <span className="text-muted-foreground pointer-events-none absolute top-1/2 right-2.5 -translate-y-1/2 text-sm">€</span>
                  </div>
                </div>
              )
            })}
      </div>

      <div className="bg-background/85 sticky bottom-[calc(4rem+env(safe-area-inset-bottom))] z-10 -mx-4 flex items-center justify-end gap-2 px-4 py-2 backdrop-blur md:bottom-4 md:mx-0 md:px-0">
        {hasInvalid && <p className="text-destructive mr-auto text-xs">Some amounts are invalid.</p>}
        {Object.keys(drafts).length > 0 && (
          <Button variant="ghost" onClick={() => setDrafts({})} disabled={save.isPending}>
            Reset
          </Button>
        )}
        <Button onClick={() => save.mutate()} disabled={changes.length === 0 || hasInvalid || save.isPending}>
          {save.isPending && <Loader2 className="animate-spin" />}
          {changes.length > 1 ? `Save ${changes.length} changes` : "Save"}
        </Button>
      </div>
    </div>
  )
}

const INDICATOR = {
  ok: "",
  warning: "[&_[data-slot=progress-indicator]]:bg-amber-500",
  over: "[&_[data-slot=progress-indicator]]:bg-destructive",
}
const TEXT = { ok: "text-muted-foreground", warning: "text-amber-600 dark:text-amber-400", over: "text-destructive" }

function Usage({ line }: { line: BudgetLine }) {
  const level = budgetLevel(line.ratio)
  return (
    <div className="space-y-1">
      <Progress value={Math.min(line.ratio, 1) * 100} className={cn("h-1.5", INDICATOR[level])} aria-label={`${line.name} usage`} />
      <p className={cn("text-xs tabular-nums", TEXT[level])}>
        <Money cents={line.spent_cents} /> of <Money cents={line.budget_cents} /> this month · {Math.round(line.ratio * 100)}%
        {line.pending_cents > 0 && <span className="text-muted-foreground"> · incl. estimates</span>}
      </p>
    </div>
  )
}
