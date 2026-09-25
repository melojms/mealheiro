import { formatCents } from "@/lib/money"
import { cn } from "@/lib/utils"

/** Tabular-number money display. */
export function Money({ cents, className }: { cents: number; className?: string }) {
  return <span className={cn("tabular-nums", className)}>{formatCents(cents)}</span>
}
