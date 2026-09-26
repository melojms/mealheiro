import type { LucideIcon } from "lucide-react"
import { RotateCw, TriangleAlert } from "lucide-react"
import type { ReactNode } from "react"
import { Button } from "@/components/ui/button"
import { cn } from "@/lib/utils"

export function EmptyState({
  icon: Icon, title, description, className, children,
}: { icon: LucideIcon; title: string; description?: ReactNode; className?: string; children?: ReactNode }) {
  return (
    <div className={cn("flex flex-col items-center justify-center gap-2 px-4 py-8 text-center", className)}>
      <span className="bg-muted text-muted-foreground flex size-11 items-center justify-center rounded-full">
        <Icon className="size-5" />
      </span>
      <p className="font-medium">{title}</p>
      {description && <p className="text-muted-foreground max-w-xs text-sm">{description}</p>}
      {children}
    </div>
  )
}

export function ErrorState({ onRetry, className }: { onRetry?: () => void; className?: string }) {
  return (
    <EmptyState icon={TriangleAlert} title="Something went wrong" description="The data couldn't be loaded." className={className}>
      {onRetry && (
        <Button variant="outline" size="sm" onClick={onRetry} className="mt-1">
          <RotateCw /> Try again
        </Button>
      )}
    </EmptyState>
  )
}
