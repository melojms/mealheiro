import { CategoryIcon } from "@/components/CategoryIcon"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { formatCents } from "@/lib/money"
import type { CategoryAmount } from "@/lib/types"
import { Delta } from "./Delta"
import { delta, formatSignedPct, sumAmount } from "./metrics"

function Section({
  title, items, empty, onSelect,
}: { title: string; items: CategoryAmount[]; empty: string; onSelect: (c: CategoryAmount) => void }) {
  return (
    <div>
      <div className="mb-1 flex items-baseline justify-between px-2">
        <h3 className="text-muted-foreground text-xs font-medium tracking-wide uppercase">{title}</h3>
        <span className="text-sm font-semibold tabular-nums">{formatCents(sumAmount(items))}</span>
      </div>
      {items.length === 0 ? (
        <p className="text-muted-foreground px-2 py-2 text-sm">{empty}</p>
      ) : (
        <ul>
          {items.map((c) => {
            const d = delta(c.amount_cents, c.prev_month_cents, true)
            return (
              <li key={c.category_id}>
                <button
                  type="button"
                  onClick={() => onSelect(c)}
                  className="hover:bg-muted/60 flex w-full items-center gap-2.5 rounded-lg px-2 py-1.5 text-left transition-colors"
                >
                  <CategoryIcon icon={c.icon} color={c.color} size="sm" />
                  <span className="min-w-0 flex-1 truncate text-sm">{c.name}</span>
                  <Delta label={formatSignedPct(d.pct)} direction={d.diff} tone={d.tone} className="hidden sm:inline-flex" />
                  <span className="text-sm font-medium tabular-nums">{formatCents(c.amount_cents)}</span>
                  {c.pending_cents > 0 && <span className="size-1.5 rounded-full bg-amber-500" title="Includes pending estimate" />}
                </button>
              </li>
            )
          })}
        </ul>
      )}
    </div>
  )
}

/** Compact income + investments lists. */
export function FlowsCard({
  income, investments, onSelect,
}: { income: CategoryAmount[]; investments: CategoryAmount[]; onSelect: (c: CategoryAmount) => void }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Income & investments</CardTitle>
      </CardHeader>
      <CardContent className="grid gap-4 px-2 sm:grid-cols-2">
        <Section title="Income" items={income} empty="No income yet." onSelect={onSelect} />
        <Section title="Investments" items={investments} empty="Nothing invested yet." onSelect={onSelect} />
      </CardContent>
    </Card>
  )
}
