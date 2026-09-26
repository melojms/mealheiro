import { Repeat } from "lucide-react"
import { CategoryIcon } from "@/components/CategoryIcon"
import { Badge } from "@/components/ui/badge"
import { Skeleton } from "@/components/ui/skeleton"
import { formatCents } from "@/lib/money"
import type { Category, Entry } from "@/lib/types"
import { cn } from "@/lib/utils"
import { shortDate } from "@/components/filters/dates"

const AMOUNT_CLASS: Record<Entry["type"], string> = {
  expense: "",
  income: "text-emerald-600 dark:text-emerald-400",
  investment: "text-sky-600 dark:text-sky-400",
}

/** Tappable entry line. `category` supplies icon/color (Entry itself carries neither). */
export function EntryRow({
  entry, category, onClick, showDate = false,
}: { entry: Entry; category?: Category; onClick: () => void; showDate?: boolean }) {
  const subtitle = [
    showDate ? shortDate(entry.date) : null,
    entry.parent_category_name ? entry.parent_category_name : null,
    entry.payer_name,
    entry.note || null,
  ].filter(Boolean)

  return (
    <button
      type="button"
      onClick={onClick}
      className="hover:bg-muted/60 focus-visible:bg-muted/60 flex w-full items-center gap-3 rounded-lg px-2 py-2.5 text-left transition-colors outline-none"
    >
      <CategoryIcon icon={category?.icon ?? "circle"} color={category?.color ?? "#64748b"} />
      <div className="min-w-0 flex-1">
        <div className="flex items-center gap-1.5">
          <span className="truncate text-sm font-medium">{entry.category_name}</span>
          {entry.recurring && <Repeat className="text-muted-foreground size-3 shrink-0" aria-label="Recurring" />}
          {entry.status === "pending" && (
            <Badge variant="outline" className="h-4 border-amber-500/40 px-1.5 text-[10px] text-amber-700 dark:text-amber-400">
              Pending
            </Badge>
          )}
        </div>
        <div className="text-muted-foreground truncate text-xs">
          {subtitle.join(" · ")}
          {entry.tags.length > 0 && <span className="ml-1">{entry.tags.map((t) => `#${t}`).join(" ")}</span>}
        </div>
      </div>
      <span
        className={cn(
          "shrink-0 text-sm font-semibold tabular-nums",
          AMOUNT_CLASS[entry.type],
          entry.status === "pending" && "opacity-70",
        )}
      >
        {entry.type === "income" ? "+" : ""}
        {formatCents(entry.amount_cents)}
      </span>
    </button>
  )
}

export function EntryRowSkeleton() {
  return (
    <div className="flex items-center gap-3 px-2 py-2.5">
      <Skeleton className="size-9 rounded-full" />
      <div className="flex-1 space-y-1.5">
        <Skeleton className="h-3.5 w-28" />
        <Skeleton className="h-3 w-40" />
      </div>
      <Skeleton className="h-4 w-16" />
    </div>
  )
}
