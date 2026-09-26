import { useState } from "react"
import { Link } from "react-router"
import { useQuery } from "@tanstack/react-query"
import { ReceiptText } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Skeleton } from "@/components/ui/skeleton"
import { EntryEditDialog } from "@/components/entry/EntryEditDialog"
import { EntryRow } from "@/components/entry/EntryRow"
import { useCategories, useToday } from "@/components/entry/hooks"
import { api } from "@/lib/api"
import type { Entry } from "@/lib/types"

export function RecentEntries() {
  const today = useToday()
  const { data, isLoading, isError } = useQuery({ queryKey: ["entries", { limit: 15 }], queryFn: () => api.entries({ limit: 15 }) })
  const { data: categories = [] } = useCategories(true)
  const byId = new Map(categories.map((c) => [c.id, c]))
  const [editing, setEditing] = useState<Entry | null>(null)

  return (
    <section aria-labelledby="recent-title" className="mt-10">
      <div className="mb-2 flex items-center justify-between px-2">
        <h2 id="recent-title" className="text-base font-semibold">Recent entries</h2>
        {!!data?.entries.length && (
          <Button variant="link" size="sm" asChild className="px-0">
            <Link to="/entries">See all</Link>
          </Button>
        )}
      </div>

      {isLoading ? (
        <div className="space-y-1">
          {Array.from({ length: 5 }, (_, i) => (
            <div key={i} className="flex items-center gap-3 px-2 py-2.5">
              <Skeleton className="size-9 rounded-full" />
              <div className="flex-1 space-y-1.5">
                <Skeleton className="h-3.5 w-32" />
                <Skeleton className="h-3 w-48" />
              </div>
              <Skeleton className="h-4 w-16" />
            </div>
          ))}
        </div>
      ) : isError ? (
        <p className="text-muted-foreground px-2 py-6 text-center text-sm">Couldn't load recent entries.</p>
      ) : !data?.entries.length ? (
        <div className="text-muted-foreground flex flex-col items-center gap-2 rounded-xl border border-dashed px-4 py-10 text-center text-sm">
          <ReceiptText className="size-6 opacity-60" />
          No entries yet. Type an amount, tap a category, save.
        </div>
      ) : (
        <div className="divide-border/60 divide-y">
          {data.entries.map((e) => (
            <EntryRow key={e.id} entry={e} category={byId.get(e.category_id)} today={today} onClick={() => setEditing(e)} />
          ))}
        </div>
      )}

      <EntryEditDialog entry={editing} open={editing !== null} onOpenChange={(o) => !o && setEditing(null)} />
    </section>
  )
}
