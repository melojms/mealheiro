import { ArrowDownCircle, ArrowUpCircle, PiggyBank, TrendingUp, Wallet } from "lucide-react"
import { formatCents, formatPct } from "@/lib/money"
import type { MonthReport } from "@/lib/types"
import { Delta } from "./Delta"
import { KpiCard, KpiSkeleton } from "./KpiCard"
import { delta, deltaTone, formatPoints, formatSignedPct, ratePoints, sumPending } from "./metrics"

const pendingHint = (cents: number) => (cents > 0 ? `incl. ${formatCents(cents)} pending` : undefined)

/** Income · expenses · investments · leftover · savings rate, each vs the previous month. */
export function KpiGrid({ report }: { report: MonthReport }) {
  const { kpis, prev_kpis: prev } = report
  const money = [
    { label: "Income", icon: ArrowDownCircle, accent: "text-emerald-600 dark:text-emerald-400", cur: kpis.income_cents, prev: prev.income_cents, upIsGood: true, pending: sumPending(report.income) },
    { label: "Expenses", icon: ArrowUpCircle, accent: "text-rose-600 dark:text-rose-400", cur: kpis.expense_cents, prev: prev.expense_cents, upIsGood: false, pending: sumPending(report.expenses) },
    { label: "Investments", icon: TrendingUp, accent: "text-sky-600 dark:text-sky-400", cur: kpis.investment_cents, prev: prev.investment_cents, upIsGood: true, pending: sumPending(report.investments) },
  ]
  const left = delta(kpis.leftover_cents, prev.leftover_cents, true)
  const pts = ratePoints(kpis.savings_rate, prev.savings_rate)

  return (
    <div className="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-5">
      {money.map((m) => {
        const d = delta(m.cur, m.prev, m.upIsGood)
        return (
          <KpiCard
            key={m.label}
            label={m.label}
            icon={m.icon}
            accent={m.accent}
            value={formatCents(m.cur)}
            delta={<Delta label={formatSignedPct(d.pct)} direction={d.diff} tone={d.tone} title={`Last month: ${formatCents(m.prev)}`} />}
            deltaCaption="vs last month"
            hint={pendingHint(m.pending)}
          />
        )
      })}
      <KpiCard
        label="Leftover"
        icon={Wallet}
        accent={kpis.leftover_cents < 0 ? "text-rose-600 dark:text-rose-400" : "text-foreground"}
        value={<span className={kpis.leftover_cents < 0 ? "text-rose-600 dark:text-rose-400" : undefined}>{formatCents(kpis.leftover_cents)}</span>}
        delta={
          <Delta
            label={`${left.diff >= 0 ? "+" : "−"}${formatCents(Math.abs(left.diff))}`}
            direction={left.diff}
            tone={left.tone}
          />
        }
        deltaCaption="vs last month"
        hint={kpis.pending_cents > 0 ? "incl. pending estimates" : undefined}
      />
      <KpiCard
        label="Savings rate"
        icon={PiggyBank}
        accent="text-violet-600 dark:text-violet-400"
        className="col-span-2 sm:col-span-1"
        value={formatPct(kpis.savings_rate)}
        delta={pts === null ? undefined : <Delta label={formatPoints(pts)} direction={pts} tone={deltaTone(Math.round(pts * 10), true)} />}
        deltaCaption={pts === null ? undefined : "vs last month"}
        hint={kpis.savings_rate === null ? "No income this month" : undefined}
      />
    </div>
  )
}

export function KpiGridSkeleton() {
  return (
    <div className="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-5">
      {Array.from({ length: 5 }, (_, i) => (
        <KpiSkeleton key={i} />
      ))}
    </div>
  )
}
