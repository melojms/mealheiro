import { Repeat } from "lucide-react"
import { CategoryIcon } from "@/components/CategoryIcon"
import { Money } from "@/components/Money"
import { Badge } from "@/components/ui/badge"
import { formatDay } from "@/components/add/dates"
import type { Category, Entry } from "@/lib/types"
import { cn } from "@/lib/utils"

const AMOUNT_TONE = {
  expense: "",
  income: "text-emerald-600 dark:text-emerald-400",
  investment: "text-sky-600 dark:text-sky-400",
} as const

/** One tappable entry line: icon, category, note · payer · date, amount. */
export function EntryRow({
  entry, category, today, onClick,
}: { entry: Entry; category?: Category; today: string; onClick: () => void }) {
  const meta = [entry.note, entry.payer_name, formatDay(entry.date, today)].filter(Boolean).join(" · ")
  return (
    <button
      type="button"
      onClick={onClick}
      className="hover:bg-muted/60 focus-visible:ring-ring/50 flex w-full items-center gap-3 rounded-xl px-2 py-2.5 text-left transition-colors outline-none focus-visible:ring-3"
    >
      <CategoryIcon icon={category?.icon ?? "circle"} color={category?.color ?? "#64748b"} />
      <div className="min-w-0 flex-1">
        <div className="flex items-center gap-1.5">
          <span className="truncate text-sm font-medium">
            {entry.parent_category_name && <span className="text-muted-foreground font-normal">{entry.parent_category_name} › </span>}
            {entry.category_name}
          </span>
          {entry.recurring && <Repeat className="text-muted-foreground size-3 shrink-0" aria-label="Recurring" />}
          {entry.status === "pending" && (
            <Badge variant="outline" className="border-amber-500/40 text-amber-600 dark:text-amber-400">
              pending
            </Badge>
          )}
          {entry.personal && <Badge variant="outline">personal</Badge>}
        </div>
        <p className="text-muted-foreground truncate text-xs">{meta}</p>
      </div>
      <Money
        cents={entry.amount_cents}
        className={cn("shrink-0 text-sm font-semibold", AMOUNT_TONE[entry.type], entry.status === "pending" && "opacity-70")}
      />
    </button>
  )
}
