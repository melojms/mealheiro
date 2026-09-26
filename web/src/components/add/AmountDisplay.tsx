import { X } from "lucide-react"
import type { EntryType } from "@/lib/types"
import { cn } from "@/lib/utils"
import { displayParts } from "./amount"

const TONE: Record<EntryType, string> = {
  expense: "text-foreground",
  income: "text-emerald-600 dark:text-emerald-400",
  investment: "text-sky-600 dark:text-sky-400",
}

export function AmountDisplay({ amount, type, onClear }: { amount: string; type: EntryType; onClear: () => void }) {
  const { main, ghost } = displayParts(amount)
  return (
    <div className="relative flex items-center justify-center py-3 md:py-5">
      <output
        aria-live="polite"
        aria-label="Amount"
        className={cn("text-5xl font-semibold tracking-tight tabular-nums transition-colors md:text-6xl", amount === "" && "text-muted-foreground/60", amount !== "" && TONE[type])}
      >
        {main}
        <span className="text-muted-foreground/40">{ghost}</span>
        <span className="text-muted-foreground ml-1.5 text-3xl font-normal md:text-4xl">€</span>
      </output>
      {amount !== "" && (
        <button
          type="button"
          onClick={onClear}
          aria-label="Clear amount"
          className="text-muted-foreground hover:bg-muted absolute right-0 rounded-full p-2 transition-colors"
        >
          <X className="size-4" />
        </button>
      )}
    </div>
  )
}
