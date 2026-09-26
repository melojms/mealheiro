import { Link } from "react-router"
import { useQuery } from "@tanstack/react-query"
import { ChevronRight, Inbox } from "lucide-react"
import { api } from "@/lib/api"
import { cn } from "@/lib/utils"

/** Teaser for the "to confirm" inbox; renders nothing when it's empty. */
export function PendingBanner() {
  const { data } = useQuery({ queryKey: ["pending"], queryFn: api.pending })
  if (!data?.length) return null
  const stale = data.some((e) => e.stale)
  return (
    <Link
      to="/month"
      className={cn(
        "mb-4 flex items-center gap-3 rounded-xl border px-3 py-2.5 text-sm transition-colors",
        stale
          ? "border-amber-500/40 bg-amber-500/10 text-amber-800 hover:bg-amber-500/15 dark:text-amber-200"
          : "bg-muted/50 hover:bg-muted",
      )}
    >
      <Inbox className="size-4 shrink-0" />
      <span className="flex-1">
        <span className="font-medium">
          {data.length} {data.length === 1 ? "bill" : "bills"} to confirm
        </span>
        {stale && <span className="opacity-80"> · some from past months</span>}
      </span>
      <ChevronRight className="size-4 opacity-60" />
    </Link>
  )
}
