import { CategoryIcon } from "@/components/CategoryIcon"
import { Skeleton } from "@/components/ui/skeleton"
import type { Category } from "@/lib/types"
import { cn } from "@/lib/utils"
import { Chip } from "./Chip"

interface CategoryPickerProps {
  /** Categories of one type, in display order (top-level and children mixed). */
  categories: Category[]
  value: number | null
  onChange: (id: number) => void
  size?: "lg" | "sm"
  loading?: boolean
}

/** Tile grid of top-level categories; a subcategory chip row appears for the selected one if it has children. */
export function CategoryPicker({ categories, value, onChange, size = "lg", loading }: CategoryPickerProps) {
  const selected = categories.find((c) => c.id === value)
  const topId = selected ? (selected.parent_id ?? selected.id) : null
  const tops = categories.filter((c) => c.parent_id === null)
  const children = topId === null ? [] : categories.filter((c) => c.parent_id === topId)
  const top = tops.find((c) => c.id === topId)
  const cols = size === "lg" ? "grid-cols-4 sm:grid-cols-5" : "grid-cols-4"

  if (loading) {
    return (
      <div className={cn("grid gap-2", cols)}>
        {Array.from({ length: 12 }, (_, i) => (
          <Skeleton key={i} className={size === "lg" ? "h-20 rounded-xl" : "h-16 rounded-xl"} />
        ))}
      </div>
    )
  }
  if (tops.length === 0) {
    return <p className="text-muted-foreground rounded-xl border border-dashed p-6 text-center text-sm">No categories yet — add some in Settings.</p>
  }

  return (
    <div className="space-y-3">
      <div role="radiogroup" aria-label="Category" className={cn("grid gap-2", cols)}>
        {tops.map((c) => {
          const active = c.id === topId
          return (
            <button
              key={c.id}
              type="button"
              role="radio"
              aria-checked={active}
              onClick={() => onChange(c.id)}
              className={cn(
                "flex flex-col items-center justify-start gap-1.5 rounded-xl text-center transition-all outline-none select-none",
                "focus-visible:ring-ring/50 hover:bg-muted/60 focus-visible:ring-3 active:scale-95",
                size === "lg" ? "py-2.5" : "py-2",
              )}
              style={active ? { backgroundColor: `${c.color}1f`, boxShadow: `inset 0 0 0 2px ${c.color}` } : undefined}
            >
              <CategoryIcon icon={c.icon} color={c.color} size={size === "lg" ? "lg" : "md"} />
              <span className={cn("line-clamp-2 w-full text-xs leading-tight break-words hyphens-auto", active ? "font-medium" : "text-muted-foreground")}>
                {c.name}
              </span>
            </button>
          )
        })}
      </div>

      {top && children.length > 0 && (
        <div className="animate-in fade-in slide-in-from-top-1 -mx-1 flex gap-2 overflow-x-auto px-1 pb-1" aria-label={`${top.name} subcategory`}>
          <Chip selected={value === top.id} onClick={() => onChange(top.id)}>
            General
          </Chip>
          {children.map((c) => (
            <Chip key={c.id} selected={value === c.id} onClick={() => onChange(value === c.id ? top.id : c.id)}>
              {c.name}
            </Chip>
          ))}
        </div>
      )}
    </div>
  )
}
