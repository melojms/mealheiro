import { Delete } from "lucide-react"
import { cn } from "@/lib/utils"
import type { KeypadKey } from "./amount"

const KEYS: KeypadKey[] = ["1", "2", "3", "4", "5", "6", "7", "8", "9", ",", "0", "backspace"]

/** On-screen numeric keypad. Long-press on backspace clears. */
export function Keypad({ onKey, className }: { onKey: (k: KeypadKey) => void; className?: string }) {
  return (
    <div className={cn("grid grid-cols-3 gap-2", className)} aria-label="Amount keypad">
      {KEYS.map((k) => (
        <button
          key={k}
          type="button"
          aria-label={k === "backspace" ? "Delete last digit" : k === "," ? "Decimal comma" : k}
          onClick={() => onKey(k)}
          onContextMenu={(e) => {
            if (k !== "backspace") return
            e.preventDefault()
            onKey("clear")
          }}
          className={cn(
            "bg-card ring-border/70 hover:bg-muted active:bg-accent flex shadow-xs ring-1 dark:bg-muted/60 dark:shadow-none dark:ring-0 h-12 items-center justify-center rounded-xl text-xl font-medium tabular-nums",
            "focus-visible:ring-ring/50 transition-transform outline-none select-none focus-visible:ring-3 active:scale-95 md:h-14",
            k === "backspace" && "text-muted-foreground",
          )}
        >
          {k === "backspace" ? <Delete className="size-5" /> : k}
        </button>
      ))}
    </div>
  )
}
