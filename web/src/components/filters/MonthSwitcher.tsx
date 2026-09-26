import { ChevronLeft, ChevronRight } from "lucide-react"
import { Button } from "@/components/ui/button"
import { addMonths, monthLabel } from "./dates"

/** ‹ March 2026 › — `max` (current month) caps forward navigation. */
export function MonthSwitcher({
  month, max, onChange,
}: { month: string; max?: string; onChange: (m: string) => void }) {
  const atMax = max !== undefined && month >= max
  return (
    <div className="bg-card flex items-center gap-1 rounded-xl border p-1 shadow-xs">
      <Button variant="ghost" size="icon" aria-label="Previous month" onClick={() => onChange(addMonths(month, -1))}>
        <ChevronLeft />
      </Button>
      <span className="min-w-32 text-center text-sm font-semibold tabular-nums" aria-live="polite">
        {monthLabel(month)}
      </span>
      <Button
        variant="ghost"
        size="icon"
        aria-label="Next month"
        disabled={atMax}
        onClick={() => onChange(addMonths(month, 1))}
      >
        <ChevronRight />
      </Button>
    </div>
  )
}
