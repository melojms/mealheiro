import { useQuery, useQueryClient } from "@tanstack/react-query"
import { useState } from "react"
import { toast } from "sonner"
import { PageHeader } from "@/components/PageHeader"
import { EntryEditDialog } from "@/components/entry/EntryEditDialog"
import { isMonth } from "@/components/filters/dates"
import { useCategories, useErrorToast, useMeta, usePayerParam, useUrlParam } from "@/components/filters/hooks"
import { MonthSwitcher } from "@/components/filters/MonthSwitcher"
import { PayerFilter } from "@/components/filters/PayerFilter"
import { ErrorState } from "@/components/filters/states"
import { BudgetsCard } from "@/components/month/BudgetsCard"
import { CategoryBreakdown } from "@/components/month/CategoryBreakdown"
import { CategoryDrillDown, type DrillTarget } from "@/components/month/CategoryDrillDown"
import { FlowsCard } from "@/components/month/FlowsCard"
import { InsightsCard } from "@/components/month/InsightsCard"
import { KpiGrid, KpiGridSkeleton } from "@/components/month/KpiGrid"
import { PendingInbox } from "@/components/month/PendingInbox"
import { RecurringCard } from "@/components/month/RecurringCard"
import { Skeleton } from "@/components/ui/skeleton"
import { api } from "@/lib/api"
import type { Entry } from "@/lib/types"

export default function MonthPage() {
  const qc = useQueryClient()
  const meta = useMeta()
  const [mParam, setMonth] = useUrlParam("m")
  const [payerId, setPayerId] = usePayerParam()
  const { byId } = useCategories()

  const current = meta.data?.month
  const month = isMonth(mParam) ? mParam : current

  const report = useQuery({
    queryKey: ["reports", "month", month, payerId],
    queryFn: () => api.monthReport(month, payerId),
    enabled: !!month,
  })
  const insights = useQuery({
    queryKey: ["insights", month, payerId],
    queryFn: () => api.insights(month, payerId),
    enabled: !!month,
  })
  // Budgets are household-wide (no payer filter in the API).
  const budgets = useQuery({ queryKey: ["budgets", month], queryFn: () => api.budgets(month), enabled: !!month })
  const pending = useQuery({ queryKey: ["pending"], queryFn: api.pending })

  useErrorToast(meta.error, "settings")
  useErrorToast(report.error, "month report")
  useErrorToast(insights.error, "insights")
  useErrorToast(budgets.error, "budgets")
  useErrorToast(pending.error, "pending entries")

  const [drill, setDrill] = useState<DrillTarget | null>(null)
  const [drillOpen, setDrillOpen] = useState(false)
  const [editing, setEditing] = useState<Entry | null>(null)
  const [editOpen, setEditOpen] = useState(false)

  const openDrill = (t: DrillTarget) => {
    setDrill(t)
    setDrillOpen(true)
  }
  const openCategory = (id: number) => {
    const r = report.data
    const fromReport = [...(r?.expenses ?? []), ...(r?.income ?? []), ...(r?.investments ?? [])].find((c) => c.category_id === id)
    const cat = byId.get(id)
    if (fromReport) openDrill(fromReport)
    else if (cat) openDrill({ category_id: id, name: cat.name, icon: cat.icon, color: cat.color })
  }
  const openEntry = (e: Entry) => {
    setEditing(e)
    setEditOpen(true)
  }
  const openEntryById = async (id: number) => {
    try {
      openEntry(await qc.fetchQuery({ queryKey: ["entry", id], queryFn: () => api.entry(id) }))
    } catch (err) {
      toast.error("Couldn't open entry", { description: (err as Error).message })
    }
  }

  // Pending inbox is global; with a payer filter only show theirs.
  const pendingEntries = pending.data?.filter((e) => payerId === undefined || e.payer_id === payerId)

  return (
    <div className="space-y-4">
      <PageHeader title="Month" description="How this month is going" />

      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        {month ? (
          <MonthSwitcher month={month} max={current} onChange={(m) => setMonth(m === current ? null : m)} />
        ) : (
          <Skeleton className="h-10 w-56 rounded-xl" />
        )}
        <PayerFilter value={payerId} onChange={setPayerId} />
      </div>

      {report.isError ? (
        <ErrorState onRetry={() => report.refetch()} />
      ) : report.data ? (
        <KpiGrid report={report.data} />
      ) : (
        <KpiGridSkeleton />
      )}

      <div className="flex flex-col gap-4 lg:grid lg:grid-cols-[minmax(0,1fr)_minmax(0,380px)] lg:items-start">
        <div className="flex flex-col gap-4 lg:order-2">
          <InsightsCard
            insights={insights.data}
            isPending={insights.isPending}
            onCategory={openCategory}
            onEntry={openEntryById}
          />
          <PendingInbox entries={pendingEntries} isPending={pending.isPending} byId={byId} onOpen={openEntry} />
          <BudgetsCard budgets={budgets.data} isPending={budgets.isPending} onCategory={openCategory} />
        </div>

        <div className="flex flex-col gap-4 lg:order-1">
          <CategoryBreakdown items={report.data?.expenses} isPending={report.isPending} onSelect={openDrill} />
          {report.data && <RecurringCard recurring={report.data.recurring} />}
          {report.data && (
            <FlowsCard income={report.data.income} investments={report.data.investments} onSelect={openDrill} />
          )}
        </div>
      </div>

      {month && (
        <CategoryDrillDown
          target={drill}
          open={drillOpen}
          onOpenChange={setDrillOpen}
          month={month}
          payerId={payerId}
          onEntry={openEntry}
        />
      )}
      <EntryEditDialog entry={editing} open={editOpen} onOpenChange={setEditOpen} />
    </div>
  )
}
