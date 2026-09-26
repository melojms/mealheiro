import type { ComponentProps } from "react"
import { cn } from "@/lib/utils"

/** Pill-shaped toggle button used for quick choices (date, payer, subcategory). */
export function Chip({ selected, className, ...props }: ComponentProps<"button"> & { selected?: boolean }) {
  return (
    <button
      type="button"
      aria-pressed={selected}
      className={cn(
        "inline-flex h-8 shrink-0 items-center gap-1.5 rounded-full border px-3 text-sm whitespace-nowrap transition-colors outline-none select-none",
        "focus-visible:ring-ring/50 focus-visible:ring-3 active:scale-[0.97] [&_svg]:size-3.5",
        selected
          ? "border-primary bg-primary text-primary-foreground"
          : "bg-background text-foreground hover:bg-muted dark:bg-input/30 dark:hover:bg-input/50",
        className,
      )}
      {...props}
    />
  )
}
