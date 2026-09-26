import type { EntryType } from "@/lib/types"
import { cn } from "@/lib/utils"
import { ENTRY_TYPES } from "./hooks"

/** Segmented Expense / Income / Investment control. */
export function TypeSwitch({
  value, onChange, className,
}: { value: EntryType; onChange: (t: EntryType) => void; className?: string }) {
  return (
    <div role="radiogroup" aria-label="Entry type" className={cn("bg-muted grid grid-cols-3 rounded-xl p-1", className)}>
      {ENTRY_TYPES.map((t) => (
        <button
          key={t.value}
          type="button"
          role="radio"
          aria-checked={value === t.value}
          onClick={() => onChange(t.value)}
          className={cn(
            "focus-visible:ring-ring/50 h-8 rounded-lg text-sm font-medium transition-all outline-none focus-visible:ring-3",
            value === t.value
              ? "bg-background text-foreground dark:bg-input/50 shadow-sm"
              : "text-muted-foreground hover:text-foreground",
          )}
        >
          {t.label}
        </button>
      ))}
    </div>
  )
}
