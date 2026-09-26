import { ArrowDownRight, ArrowUpRight, Minus } from "lucide-react"
import { cn } from "@/lib/utils"
import type { Tone } from "./metrics"

const TONE_TEXT: Record<Tone, string> = {
  good: "text-emerald-600 dark:text-emerald-400",
  bad: "text-rose-600 dark:text-rose-400",
  neutral: "text-muted-foreground",
}

/** Small colored "↗ +12%" chip. `direction` is the raw sign; `tone` says whether it's good news. */
export function Delta({
  label, direction, tone, className, title,
}: { label: string | null; direction: number; tone: Tone; className?: string; title?: string }) {
  if (label === null) return <span className={cn("text-muted-foreground text-xs", className)}>—</span>
  const Icon = direction > 0 ? ArrowUpRight : direction < 0 ? ArrowDownRight : Minus
  return (
    <span className={cn("inline-flex items-center gap-0.5 text-xs font-medium tabular-nums", TONE_TEXT[tone], className)} title={title}>
      <Icon className="size-3.5" aria-hidden />
      {label}
    </span>
  )
}
