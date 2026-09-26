import type { LucideIcon } from "lucide-react"
import type { ReactNode } from "react"
import { Card } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { cn } from "@/lib/utils"

/** Compact stat tile: label + icon, big value, delta line and optional footnote (e.g. pending). */
export function KpiCard({
  label, icon: Icon, accent, value, delta, deltaCaption, hint, className,
}: {
  label: string
  icon: LucideIcon
  accent: string // tailwind text color classes for the icon chip
  value: ReactNode
  delta?: ReactNode
  deltaCaption?: string
  hint?: ReactNode
  className?: string
}) {
  return (
    <Card size="sm" role="group" aria-label={label} className={cn("gap-1.5 px-3.5", className)}>
      <div className="text-muted-foreground flex items-center justify-between gap-2 text-xs font-medium">
        <span className="truncate">{label}</span>
        <span className={cn("bg-muted flex size-6 shrink-0 items-center justify-center rounded-full", accent)}>
          <Icon className="size-3.5" aria-hidden />
        </span>
      </div>
      <div className="truncate text-lg font-semibold tracking-tight tabular-nums sm:text-xl">{value}</div>
      {delta && (
        <div className="flex flex-wrap items-center gap-x-1 text-xs">
          {delta}
          {deltaCaption && <span className="text-muted-foreground">{deltaCaption}</span>}
        </div>
      )}
      {hint && <div className="text-muted-foreground text-[11px] leading-tight">{hint}</div>}
    </Card>
  )
}

export function KpiSkeleton() {
  return (
    <Card size="sm" className="gap-2 px-3.5">
      <Skeleton className="h-3 w-16" />
      <Skeleton className="h-6 w-24" />
      <Skeleton className="h-3 w-20" />
    </Card>
  )
}
