import { categoryIcon } from "@/lib/icons"
import { cn } from "@/lib/utils"

/** Round colored badge with the category's Lucide icon. */
export function CategoryIcon({
  icon, color, className, size = "md",
}: { icon: string; color: string; className?: string; size?: "sm" | "md" | "lg" }) {
  const Icon = categoryIcon(icon)
  const box = { sm: "size-7", md: "size-9", lg: "size-12" }[size]
  const glyph = { sm: "size-3.5", md: "size-4.5", lg: "size-6" }[size]
  return (
    <span
      className={cn("inline-flex shrink-0 items-center justify-center rounded-full", box, className)}
      style={{ backgroundColor: `${color}22`, color }}
    >
      <Icon className={glyph} />
    </span>
  )
}
