import { ChevronRight, PieChart as PieIcon } from "lucide-react"
import { Pie, PieChart } from "recharts"
import { CategoryIcon } from "@/components/CategoryIcon"
import { MoneyTooltip } from "@/components/charts/MoneyTooltip"
import { catKey, categoryConfig } from "@/components/charts/shape"
import { EmptyState } from "@/components/filters/states"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { ChartContainer, ChartTooltip } from "@/components/ui/chart"
import { Skeleton } from "@/components/ui/skeleton"
import { formatCents, formatCentsCompact, formatPct } from "@/lib/money"
import type { CategoryAmount } from "@/lib/types"
import { Delta } from "./Delta"
import { delta, formatSignedPct, share, sumAmount } from "./metrics"

function Row({ c, total, onSelect }: { c: CategoryAmount; total: number; onSelect: () => void }) {
  const vsPrev = delta(c.amount_cents, c.prev_month_cents, false)
  const vsAvg = delta(c.amount_cents, c.avg3_cents, false)
  const pct = share(c.amount_cents, total)
  return (
    <li>
      <button
        type="button"
        onClick={onSelect}
        className="hover:bg-muted/60 flex w-full items-center gap-3 rounded-lg px-2 py-2 text-left transition-colors"
      >
        <CategoryIcon icon={c.icon} color={c.color} />
        <div className="min-w-0 flex-1">
          <div className="flex items-baseline justify-between gap-2">
            <span className="truncate text-sm font-medium">{c.name}</span>
            <span className="shrink-0 text-sm font-semibold tabular-nums">{formatCents(c.amount_cents)}</span>
          </div>
          <div className="bg-muted mt-1.5 h-1.5 overflow-hidden rounded-full">
            <div className="h-full rounded-full" style={{ width: `${pct * 100}%`, backgroundColor: c.color }} />
          </div>
          <div className="text-muted-foreground mt-1 flex flex-wrap items-center gap-x-3 gap-y-0.5 text-xs">
            <span className="tabular-nums">{formatPct(pct)}</span>
            <span className="inline-flex items-center gap-1">
              <Delta label={formatSignedPct(vsPrev.pct) ?? (c.amount_cents > 0 ? "new" : null)} direction={vsPrev.diff} tone={vsPrev.tone} title={`Last month: ${formatCents(c.prev_month_cents)}`} />
              <span>vs last</span>
            </span>
            <span className="inline-flex items-center gap-1">
              <Delta label={formatSignedPct(vsAvg.pct)} direction={vsAvg.diff} tone={vsAvg.tone} title={`3-month avg: ${formatCents(c.avg3_cents)}`} />
              <span>vs 3-mo avg</span>
            </span>
            {c.pending_cents > 0 && <span>incl. {formatCents(c.pending_cents)} pending</span>}
          </div>
        </div>
        <ChevronRight className="text-muted-foreground size-4 shrink-0" />
      </button>
    </li>
  )
}

/** Donut + ranked list of top-level expense categories for the month. */
export function CategoryBreakdown({
  items, isPending, onSelect,
}: { items: CategoryAmount[] | undefined; isPending: boolean; onSelect: (c: CategoryAmount) => void }) {
  const total = sumAmount(items ?? [])
  const config = categoryConfig((items ?? []).map((c) => ({ id: c.category_id, name: c.name, color: c.color })))
  const data = (items ?? []).map((c) => ({ key: catKey(c.category_id), name: c.name, amount: c.amount_cents, fill: c.color }))
  const labels = Object.fromEntries(data.map((d) => [d.key, d.name]))

  return (
    <Card>
      <CardHeader>
        <CardTitle>Where the money went</CardTitle>
        <CardDescription>Expenses by category · tap one for details</CardDescription>
      </CardHeader>
      <CardContent>
        {isPending ? (
          <div className="grid items-center gap-6 md:grid-cols-[220px_1fr]">
            <Skeleton className="mx-auto size-48 rounded-full" />
            <div className="space-y-3">
              {[0, 1, 2, 3].map((i) => (
                <Skeleton key={i} className="h-10 w-full" />
              ))}
            </div>
          </div>
        ) : !items?.length ? (
          <EmptyState icon={PieIcon} title="No expenses this month" description="Expenses you add will show up here, split by category." />
        ) : (
          <div className="grid items-start gap-4 md:grid-cols-[220px_1fr] md:gap-6">
            <div className="relative mx-auto aspect-square w-full max-w-[220px] md:sticky md:top-4">
              <ChartContainer config={config} className="aspect-square h-full w-full">
                <PieChart>
                  <ChartTooltip cursor={false} content={<MoneyTooltip labels={labels} />} />
                  <Pie
                    data={data}
                    dataKey="amount"
                    nameKey="key"
                    innerRadius="62%"
                    outerRadius="100%"
                    paddingAngle={data.length > 1 ? 1.5 : 0}
                    cornerRadius={4}
                    strokeWidth={0}
                    isAnimationActive={false}
                    onClick={(_, i) => onSelect(items[i])}
                    className="cursor-pointer"
                  />
                </PieChart>
              </ChartContainer>
              <div className="pointer-events-none absolute inset-0 flex flex-col items-center justify-center">
                <span className="text-muted-foreground text-xs">Total</span>
                <span className="text-lg font-semibold tabular-nums">{formatCentsCompact(total)}</span>
              </div>
            </div>
            <ul className="-mx-2 space-y-0.5">
              {items.map((c) => (
                <Row key={c.category_id} c={c} total={total} onSelect={() => onSelect(c)} />
              ))}
            </ul>
          </div>
        )}
      </CardContent>
    </Card>
  )
}
