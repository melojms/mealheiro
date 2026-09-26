import { useQuery } from "@tanstack/react-query"
import { Plus, Repeat } from "lucide-react"
import { CategoryIcon } from "@/components/CategoryIcon"
import { Money } from "@/components/Money"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Skeleton } from "@/components/ui/skeleton"
import { formatMonth } from "@/components/add/dates"
import { useCategories, useToday } from "@/components/entry/hooks"
import { api } from "@/lib/api"
import type { Category, Template } from "@/lib/types"
import { cn } from "@/lib/utils"
import { TemplateDialog } from "./TemplateDialog"
import { useDialog } from "./useDialog"

export function TemplatesSection() {
  const today = useToday()
  const { data: templates, isLoading, isError } = useQuery({ queryKey: ["templates"], queryFn: api.templates })
  const { data: categories = [] } = useCategories(true)
  const byId = new Map(categories.map((c) => [c.id, c]))
  const dialog = useDialog<Template | null>()

  const active = templates?.filter((t) => t.active) ?? []
  const paused = templates?.filter((t) => !t.active) ?? []

  return (
    <div className="space-y-4">
      <div className="flex justify-end">
        <Button onClick={() => dialog.show(null)}>
          <Plus /> New recurring
        </Button>
      </div>

      {isLoading ? (
        <div className="bg-card divide-y rounded-xl border">
          {Array.from({ length: 4 }, (_, i) => (
            <div key={i} className="flex items-center gap-3 p-3">
              <Skeleton className="size-9 rounded-full" />
              <div className="flex-1 space-y-1.5">
                <Skeleton className="h-3.5 w-32" />
                <Skeleton className="h-3 w-48" />
              </div>
            </div>
          ))}
        </div>
      ) : isError ? (
        <p className="text-muted-foreground text-sm">Couldn't load recurring templates.</p>
      ) : !templates?.length ? (
        <div className="text-muted-foreground flex flex-col items-center gap-2 rounded-xl border border-dashed px-4 py-10 text-center text-sm">
          <Repeat className="size-6 opacity-60" />
          <p>No recurring entries yet.</p>
          <p className="text-xs">Add the mortgage, subscriptions, salary or a monthly ETF contribution once, and they'll show up every month.</p>
        </div>
      ) : (
        <>
          <TemplateList title="Active" templates={active} byId={byId} onOpen={dialog.show} />
          <TemplateList title="Paused" templates={paused} byId={byId} onOpen={dialog.show} />
        </>
      )}

      {dialog.key > 0 ? (
        <TemplateDialog
          key={dialog.key}
          template={dialog.payload}
          currentMonth={today.slice(0, 7)}
          open={dialog.open}
          onOpenChange={dialog.setOpen}
        />
      ) : null}
    </div>
  )
}

function TemplateList({
  title, templates, byId, onOpen,
}: { title: string; templates: Template[]; byId: Map<number, Category>; onOpen: (t: Template) => void }) {
  if (!templates.length) return null
  return (
    <section className="space-y-2">
      <h3 className="text-muted-foreground px-1 text-xs font-medium tracking-wide uppercase">
        {title} · {templates.length}
      </h3>
      <div className="bg-card divide-y rounded-xl border">
        {templates.map((t) => {
          const c = byId.get(t.category_id)
          return (
            <button
              key={t.id}
              type="button"
              onClick={() => onOpen(t)}
              className={cn("hover:bg-muted/50 flex w-full items-center gap-3 p-3 text-left transition-colors first:rounded-t-xl last:rounded-b-xl", !t.active && "opacity-70")}
            >
              <CategoryIcon icon={c?.icon ?? "repeat"} color={c?.color ?? "#64748b"} />
              <div className="min-w-0 flex-1">
                <div className="flex flex-wrap items-center gap-1.5">
                  <span className="truncate text-sm font-medium">
                    {t.parent_category_name && <span className="text-muted-foreground font-normal">{t.parent_category_name} › </span>}
                    {t.category_name}
                  </span>
                  <Badge variant={t.variable ? "outline" : "secondary"}>{t.variable ? "Variable" : "Fixed"}</Badge>
                </div>
                <p className="text-muted-foreground truncate text-xs">
                  {[
                    t.note,
                    t.payer_name,
                    t.end_month ? `${formatMonth(t.start_month)} – ${formatMonth(t.end_month)}` : `Since ${formatMonth(t.start_month)}`,
                  ]
                    .filter(Boolean)
                    .join(" · ")}
                </p>
              </div>
              <div className="shrink-0 text-right">
                <Money cents={t.amount_cents} className="text-sm font-semibold" />
                <p className="text-muted-foreground text-xs">{t.variable ? "est. / month" : "/ month"}</p>
              </div>
            </button>
          )
        })}
      </div>
    </section>
  )
}
