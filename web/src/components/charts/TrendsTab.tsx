import { useQuery } from "@tanstack/react-query"
import { ChartColumnStacked, X } from "lucide-react"
import { Area, AreaChart, Bar, BarChart, CartesianGrid, ReferenceLine, XAxis, YAxis } from "recharts"
import { CategoryIcon } from "@/components/CategoryIcon"
import { mediumMonthLabel } from "@/components/filters/dates"
import { useCategories, useErrorToast, useUrlParam } from "@/components/filters/hooks"
import { EmptyState, ErrorState } from "@/components/filters/states"
import { Button } from "@/components/ui/button"
import { Card, CardAction, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { ChartContainer, ChartLegend, ChartLegendContent, ChartTooltip, type ChartConfig } from "@/components/ui/chart"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Skeleton } from "@/components/ui/skeleton"
import { api } from "@/lib/api"
import { formatCents, formatCentsCompact } from "@/lib/money"
import { cn } from "@/lib/utils"
import { MoneyTooltip } from "./MoneyTooltip"
import { catKey, categoryConfig, flowRows, hasTrendData, stackedTrendRows } from "./shape"

const FLOW_CONFIG = {
  income: { label: "Income", color: "#10b981" },
  expenses: { label: "Expenses", color: "#f43f5e" },
  savings: { label: "Savings", color: "#6366f1" },
} satisfies ChartConfig

const FLOW_LABELS = { income: "Income", expenses: "Expenses", savings: "Savings" }

const monthTitle = (row: Record<string, unknown> | undefined) =>
  typeof row?.month === "string" ? mediumMonthLabel(row.month) : undefined

const axisProps = {
  tickLine: false,
  axisLine: false,
} as const

export function TrendsTab({ payerId }: { payerId: number | undefined }) {
  const [catParam, setCat] = useUrlParam("cat")
  const categoryId = catParam && Number(catParam) > 0 ? Number(catParam) : undefined
  const { data: allCats, byId } = useCategories()
  const expenseTops = (allCats ?? []).filter((c) => c.type === "expense" && c.parent_id === null)

  const trends = useQuery({
    queryKey: ["reports", "trends", categoryId, payerId],
    queryFn: () => api.trends({ months: 12, category_id: categoryId, payer_id: payerId }),
  })
  useErrorToast(trends.error, "trends")

  const selected = categoryId !== undefined ? byId.get(categoryId) : undefined
  const data = trends.data
  const rows = data ? stackedTrendRows(data) : []
  const config = data ? categoryConfig(data.categories) : {}
  const labels = Object.fromEntries((data?.categories ?? []).map((c) => [catKey(c.id), c.name]))
  const avg = data && data.series.length ? data.series.reduce((s, x) => s + x.expense_cents, 0) / data.series.length : 0

  return (
    <div className="space-y-4">
      <Card>
        <CardHeader>
          <CardTitle>{selected ? selected.name : "Spending by category"}</CardTitle>
          <CardDescription>
            Last 12 months{selected && avg > 0 ? ` · avg ${formatCents(Math.round(avg))}/month` : ""}
          </CardDescription>
          <CardAction>
            <Select value={categoryId ? String(categoryId) : "all"} onValueChange={(v) => setCat(v === "all" ? null : v)}>
              <SelectTrigger size="sm" className="max-w-40" aria-label="Category">
                <SelectValue />
              </SelectTrigger>
              <SelectContent position="popper" align="end">
                <SelectItem value="all">All categories</SelectItem>
                {expenseTops.map((c) => (
                  <SelectItem key={c.id} value={String(c.id)}>
                    <CategoryIcon icon={c.icon} color={c.color} size="sm" className="size-5" />
                    {c.name}
                    {c.archived && <span className="text-muted-foreground text-xs">(archived)</span>}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </CardAction>
        </CardHeader>
        <CardContent>
          {trends.isPending ? (
            <Skeleton className="h-64 w-full" />
          ) : trends.isError ? (
            <ErrorState onRetry={() => trends.refetch()} />
          ) : !hasTrendData(data!) ? (
            <EmptyState icon={ChartColumnStacked} title="No data yet" description="Charts fill in as you add entries." />
          ) : (
            <>
              <ChartContainer config={config} className="aspect-auto h-64 w-full min-w-0 [&_.recharts-responsive-container]:min-w-0">
                <BarChart data={rows} margin={{ left: 4, right: 4 }} accessibilityLayer>
                  <CartesianGrid vertical={false} />
                  <XAxis dataKey="label" {...axisProps} tickMargin={8} interval="preserveStartEnd" minTickGap={4} />
                  <YAxis {...axisProps} width={64} tickFormatter={(v: number) => formatCentsCompact(v)} />
                  <ChartTooltip
                    cursor={{ fillOpacity: 0.5 }}
                    content={<MoneyTooltip labels={labels} title={monthTitle} sort total />}
                  />
                  {data!.categories.map((c, i) => (
                    <Bar
                      key={c.id}
                      dataKey={catKey(c.id)}
                      stackId="spend"
                      fill={`var(--color-${catKey(c.id)})`}
                      radius={i === data!.categories.length - 1 ? [4, 4, 0, 0] : 0}
                      maxBarSize={40}
                    />
                  ))}
                  {selected && avg > 0 && (
                    <ReferenceLine y={avg} stroke="currentColor" strokeDasharray="4 4" className="text-muted-foreground" />
                  )}
                </BarChart>
              </ChartContainer>
              {/* Legend doubles as a category picker. */}
              <div className="mt-3 flex flex-wrap gap-1.5">
                {data!.categories.map((c) => (
                  <button
                    key={c.id}
                    type="button"
                    onClick={() => setCat(categoryId === c.id ? null : String(c.id))}
                    className={cn(
                      "hover:bg-muted inline-flex items-center gap-1.5 rounded-full border px-2 py-0.5 text-xs transition-colors",
                      categoryId === c.id && "bg-muted",
                    )}
                  >
                    <span className="size-2 rounded-full" style={{ backgroundColor: c.color }} />
                    {c.name}
                  </button>
                ))}
                {selected && (
                  <Button variant="ghost" size="xs" onClick={() => setCat(null)}>
                    <X /> Show all
                  </Button>
                )}
              </div>
            </>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Income vs expenses</CardTitle>
          <CardDescription>
            Savings = income − expenses{selected ? ` · expenses limited to ${selected.name}` : ""}
          </CardDescription>
        </CardHeader>
        <CardContent>
          {trends.isPending ? (
            <Skeleton className="h-64 w-full" />
          ) : trends.isError || !hasTrendData(data!) ? (
            <p className="text-muted-foreground py-6 text-center text-sm">No data to show.</p>
          ) : (
            <ChartContainer config={FLOW_CONFIG} className="aspect-auto h-64 w-full min-w-0 [&_.recharts-responsive-container]:min-w-0">
              <AreaChart data={flowRows(data!)} margin={{ left: 4, right: 8 }} accessibilityLayer>
                <defs>
                  {Object.keys(FLOW_CONFIG).map((k) => (
                    <linearGradient key={k} id={`flow-${k}`} x1="0" y1="0" x2="0" y2="1">
                      <stop offset="5%" stopColor={`var(--color-${k})`} stopOpacity={0.3} />
                      <stop offset="95%" stopColor={`var(--color-${k})`} stopOpacity={0.02} />
                    </linearGradient>
                  ))}
                </defs>
                <CartesianGrid vertical={false} />
                <XAxis dataKey="label" {...axisProps} tickMargin={8} interval="preserveStartEnd" minTickGap={4} />
                <YAxis {...axisProps} width={64} tickFormatter={(v: number) => formatCentsCompact(v)} />
                <ReferenceLine y={0} stroke="currentColor" className="text-border" />
                <ChartTooltip content={<MoneyTooltip labels={FLOW_LABELS} title={monthTitle} hideZero={false} />} />
                <ChartLegend content={<ChartLegendContent />} />
                {(["income", "expenses", "savings"] as const).map((k) => (
                  <Area
                    key={k}
                    dataKey={k}
                    type="monotone"
                    stroke={`var(--color-${k})`}
                    strokeWidth={2}
                    fill={`url(#flow-${k})`}
                    dot={false}
                    activeDot={{ r: 4 }}
                  />
                ))}
              </AreaChart>
            </ChartContainer>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
