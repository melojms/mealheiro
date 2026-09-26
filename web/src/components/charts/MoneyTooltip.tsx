import { useMemo } from "react"
import { formatCents } from "@/lib/money"
import { cn } from "@/lib/utils"

interface Item {
  name?: string | number
  dataKey?: unknown
  value?: unknown
  color?: string
  payload?: Record<string, unknown> & { fill?: string }
}

/**
 * Tooltip body for cents-valued series (rendered via <ChartTooltip content={<MoneyTooltip … />} />).
 * Recharts injects active/payload/label. Labels resolve through `labels` (dataKey → name).
 */
export function MoneyTooltip({
  active, payload, label, labels, title, sort = false, hideZero = true, total = false,
}: {
  active?: boolean
  payload?: ReadonlyArray<Item>
  label?: unknown
  labels?: Record<string, string>
  title?: (row: Record<string, unknown> | undefined, label: unknown) => string | undefined
  sort?: boolean
  hideZero?: boolean
  total?: boolean
}) {
  const rows = useMemo(() => {
    const items = (payload ?? [])
      .map((p) => ({
        key: String(p.dataKey ?? p.name),
        name: labels?.[String(p.dataKey)] ?? String(p.name ?? ""),
        value: typeof p.value === "number" ? p.value : 0,
        color: p.payload?.fill ?? p.color,
      }))
      .filter((r) => !hideZero || r.value !== 0)
    return sort ? items.sort((a, b) => b.value - a.value) : items
  }, [payload, labels, sort, hideZero])

  if (!active || !payload?.length) return null
  const heading = title ? title(payload[0]?.payload, label) : typeof label === "string" ? label : undefined

  return (
    <div className="border-border/50 bg-background grid min-w-40 gap-1.5 rounded-lg border px-2.5 py-1.5 text-xs shadow-xl">
      {heading && <div className="font-medium">{heading}</div>}
      <div className="grid gap-1">
        {rows.map((r) => (
          <div key={r.key} className="flex items-center gap-2">
            <span className="size-2.5 shrink-0 rounded-[2px]" style={{ backgroundColor: r.color }} />
            <span className="text-muted-foreground flex-1 truncate">{r.name}</span>
            <span className={cn("text-foreground font-mono font-medium tabular-nums", r.value < 0 && "text-rose-600 dark:text-rose-400")}>
              {formatCents(r.value)}
            </span>
          </div>
        ))}
        {rows.length === 0 && <span className="text-muted-foreground">No data</span>}
      </div>
      {total && rows.length > 1 && (
        <div className="flex justify-between border-t pt-1 font-medium">
          <span>Total</span>
          <span className="font-mono tabular-nums">{formatCents(rows.reduce((s, r) => s + r.value, 0))}</span>
        </div>
      )}
    </div>
  )
}
