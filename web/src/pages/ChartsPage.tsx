import { PageHeader } from "@/components/PageHeader"
import { TrendsTab } from "@/components/charts/TrendsTab"
import { YearTab } from "@/components/charts/YearTab"
import { usePayerParam, useUrlParam } from "@/components/filters/hooks"
import { PayerFilter } from "@/components/filters/PayerFilter"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"

type Tab = "trends" | "year"

export default function ChartsPage() {
  const [tabParam, setTab] = useUrlParam("tab")
  const tab: Tab = tabParam === "year" ? "year" : "trends"
  const [payerId, setPayerId] = usePayerParam()

  return (
    <div>
      <PageHeader title="Charts" description="Trends over time and year-over-year" />
      <Tabs value={tab} onValueChange={(v) => setTab(v === "trends" ? null : v)} className="gap-4">
        <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <TabsList className="w-full sm:w-auto">
            <TabsTrigger value="trends" className="px-4">Trends</TabsTrigger>
            <TabsTrigger value="year" className="px-4">Year</TabsTrigger>
          </TabsList>
          <PayerFilter value={payerId} onChange={setPayerId} />
        </div>
        <TabsContent value="trends">
          <TrendsTab payerId={payerId} />
        </TabsContent>
        <TabsContent value="year">
          <YearTab payerId={payerId} />
        </TabsContent>
      </Tabs>
    </div>
  )
}
