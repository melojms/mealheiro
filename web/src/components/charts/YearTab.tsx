import { useQuery } from "@tanstack/react-query"
import { ArrowDownCircle, ArrowUpCircle, CalendarRange, PiggyBank, TrendingUp, Wallet } from "lucide-react"
import { Bar, BarChart, CartesianGrid, XAxis, YAxis } from "recharts"
import { CategoryIcon } from "@/components/CategoryIcon"
import { useErrorToast, useMeta, useUrlParam } from "@/components/filters/hooks"
import { EmptyState, ErrorState } from "@/components/filters/states"
import { Delta } from "@/components/month/Delta"
import { KpiCard } from "@/components/month/KpiCard"
import { KpiGridSkeleton } from "@/components/month/KpiGrid"
import { delta, deltaTone, formatPoints, formatSignedPct, ratePoints } from "@/components/month/metrics"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { ChartContainer, ChartLegend, ChartLegendContent, ChartTooltip, type ChartConfig } from "@/components/ui/chart"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Skeleton } from "@/components/ui/skeleton"
import { api } from "@/lib/api"
import { formatCents, formatCentsCompact, formatPct } from "@/lib/money"
import type { EntryType, YearReport } from "@/lib/types"
import { MoneyTooltip } from "./MoneyTooltip"
import { yearCategoriesByType, yearMonthlyRows } from "./shape"

const MONTHLY_CONFIG = {
  income: { label: "Income", color: "#10b981" },
  expenses: { label: "Expenses", color: "#f43f5e" },
  investments: { label: "Investments", color: "#0ea5e9" },
} satisfies ChartConfig
const MONTHLY_LABELS = { income: "Income", expenses: "Expenses", investments: "Investments" }

const TYPE_TITLE: Record<EntryType, string> = { expense: "Expenses", income: "Income", investment: "Investments" }

function Totals({ report }: { report: YearReport }) {
  const { totals: t, prev_totals: p } = report
  const cards = [
    { label: "Income", icon: ArrowDownCircle, accent: "text-emerald-600 dark:text-emerald-400", cur: t.income_cents, prev: p.income_cents, up: true },
    { label: "Expenses", icon: ArrowUpCircle, accent: "text-rose-600 dark:text-rose-400", cur: t.expense_cents, prev: p.expense_cents, up: false },
    { label: "Investments", icon: TrendingUp, accent: "text-sky-600 dark:text-sky-400", cur: t.investment_cents, prev: p.investment_cents, up: true },
    { label: "Leftover", icon: Wallet, accent: "text-foreground", cur: t.leftover_cents, prev: p.leftover_cents, up: true },
  ]
  const pts = ratePoints(t.savings_rate, p.savings_rate)
  const caption = `vs ${report.year - 1}`
  return (
    <div className="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-5">
      {cards.map((c) => {
        const d = delta(c.cur, c.prev, c.up)
        return (
          <KpiCard
            key={c.label}
            label={c.label}
            icon={c.icon}
            accent={c.accent}
            value={formatCents(c.cur)}
            delta={<Delta label={formatSignedPct(d.pct)} direction={d.diff} tone={d.tone} title={`${report.year - 1}: ${formatCents(c.prev)}`} />}
            deltaCaption={caption}
          />
        )
      })}
      <KpiCard
        label="Savings rate"
        icon={PiggyBank}
        accent="text-violet-600 dark:text-violet-400"
        className="col-span-2 sm:col-span-1"
        value={formatPct(t.savings_rate)}
        delta={pts === null ? undefined : <Delta label={formatPoints(pts)} direction={pts} tone={deltaTone(Math.round(pts * 10), true)} />}
        deltaCaption={pts === null ? undefined : caption}
      />
    </div>
  )
}

function CategoryTable({ report }: { report: YearReport }) {
  const groups = yearCategoriesByType(report)
  if (!groups.length) return <EmptyState icon={CalendarRange} title="No categories" description="Nothing booked this year or last." />
  return (
    <div className="space-y-5">
      {groups.map((g) => (
        <div key={g.type}>
          <div className="text-muted-foreground mb-1 grid grid-cols-[1fr_auto_auto] gap-3 px-2 text-xs font-medium tracking-wide uppercase sm:grid-cols-[1fr_7rem_7rem_4rem]">
            <span>{TYPE_TITLE[g.type]}</span>
            <span className="w-24 text-right sm:w-auto">{report.year}</span>
            <span className="hidden text-right sm:block">{report.year - 1}</span>
            <span className="w-14 text-right sm:w-auto">Δ</span>
          </div>
          <ul>
            {g.items.map((c) => {
              const d = delta(c.amount_cents, c.prev_year_cents, g.type !== "expense")
              return (
                <li
                  key={c.category_id}
                  className="hover:bg-muted/40 grid grid-cols-[1fr_auto_auto] items-center gap-3 rounded-lg px-2 py-1.5 sm:grid-cols-[1fr_7rem_7rem_4rem]"
                >
                  <div className="flex min-w-0 items-center gap-2.5">
                    <CategoryIcon icon={c.icon} color={c.color} size="sm" />
                    <div className="min-w-0">
                      <div className="truncate text-sm">{c.name}</div>
                      <div className="text-muted-foreground truncate text-xs tabular-nums sm:hidden">
                        {report.year - 1}: {formatCents(c.prev_year_cents)}
                      </div>
                    </div>
                  </div>
                  <span className="w-24 text-right text-sm font-medium tabular-nums sm:w-auto">{formatCents(c.amount_cents)}</span>
                  <span className="text-muted-foreground hidden text-right text-sm tabular-nums sm:block">
                    {formatCents(c.prev_year_cents)}
                  </span>
                  <span className="w-14 text-right sm:w-auto">
                    <Delta
                      label={formatSignedPct(d.pct) ?? (c.amount_cents > 0 ? "new" : null)}
                      direction={d.diff}
                      tone={d.tone}
                    />
                  </span>
                </li>
              )
            })}
          </ul>
        </div>
      ))}
    </div>
  )
}

export function YearTab({ payerId }: { payerId: number | undefined }) {
  const meta = useMeta()
  const [yearParam, setYear] = useUrlParam("year")
  const years = useQuery({ queryKey: ["reports", "years"], queryFn: api.years })
  const currentYear = meta.data ? Number(meta.data.today.slice(0, 4)) : undefined
  const year = yearParam && /^\d{4}$/.test(yearParam) ? Number(yearParam) : currentYear

  const report = useQuery({
    queryKey: ["reports", "year", year, payerId],
    queryFn: () => api.yearReport(year, payerId),
    enabled: year !== undefined,
  })
  useErrorToast(years.error, "years")
  useErrorToast(report.error, "year report")

  const yearOptions = [...new Set([...(years.data ?? []), ...(year ? [year] : [])])].sort((a, b) => b - a)

  return (
    <div className="space-y-4">
      <div className="flex items-center gap-2">
        <span className="text-muted-foreground text-sm">Year</span>
        {year === undefined ? (
          <Skeleton className="h-8 w-24" />
        ) : (
          <Select value={String(year)} onValueChange={(v) => setYear(Number(v) === currentYear ? null : v)}>
            <SelectTrigger aria-label="Year" className="w-28 font-semibold">
              <SelectValue />
            </SelectTrigger>
            <SelectContent position="popper">
              {yearOptions.map((y) => (
                <SelectItem key={y} value={String(y)}>
                  {y}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        )}
      </div>

      {report.isError ? (
        <ErrorState onRetry={() => report.refetch()} />
      ) : !report.data ? (
        <>
          <KpiGridSkeleton />
          <Skeleton className="h-72 w-full rounded-xl" />
        </>
      ) : (
        <>
          <Totals report={report.data} />

          <Card>
            <CardHeader>
              <CardTitle>Month by month</CardTitle>
              <CardDescription>{report.data.year}</CardDescription>
            </CardHeader>
            <CardContent>
              <ChartContainer config={MONTHLY_CONFIG} className="aspect-auto h-64 w-full min-w-0 [&_.recharts-responsive-container]:min-w-0">
                <BarChart data={yearMonthlyRows(report.data)} margin={{ left: 4, right: 4 }} barGap={2} accessibilityLayer>
                  <CartesianGrid vertical={false} />
                  <XAxis dataKey="label" tickLine={false} axisLine={false} tickMargin={8} interval="preserveStartEnd" minTickGap={4} />
                  <YAxis tickLine={false} axisLine={false} width={64} tickFormatter={(v: number) => formatCentsCompact(v)} />
                  <ChartTooltip cursor={{ fillOpacity: 0.5 }} content={<MoneyTooltip labels={MONTHLY_LABELS} hideZero={false} />} />
                  <ChartLegend content={<ChartLegendContent />} />
                  {(["income", "expenses", "investments"] as const).map((k) => (
                    <Bar key={k} dataKey={k} fill={`var(--color-${k})`} radius={[3, 3, 0, 0]} maxBarSize={14} />
                  ))}
                </BarChart>
              </ChartContainer>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Categories</CardTitle>
              <CardDescription>
                {report.data.year} vs {report.data.year - 1}
              </CardDescription>
            </CardHeader>
            <CardContent className="px-2">
              <CategoryTable report={report.data} />
            </CardContent>
          </Card>
        </>
      )}
    </div>
  )
}
