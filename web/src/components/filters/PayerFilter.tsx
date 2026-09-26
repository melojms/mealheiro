import { Users } from "lucide-react"
import { ToggleGroup, ToggleGroupItem } from "@/components/ui/toggle-group"
import { Skeleton } from "@/components/ui/skeleton"
import { cn } from "@/lib/utils"
import { usePeople } from "./hooks"

/** "All · Ana · Rui · Joint" segmented filter; scrolls horizontally when names are long. */
export function PayerFilter({
  value, onChange, className,
}: { value: number | undefined; onChange: (id: number | undefined) => void; className?: string }) {
  const { data: people, isPending } = usePeople()
  if (isPending) return <Skeleton className={cn("h-8 w-52 rounded-lg", className)} />
  if (!people?.length) return null

  return (
    <div className={cn("-mx-1 max-w-full overflow-x-auto px-1 [scrollbar-width:none]", className)}>
      <ToggleGroup
        type="single"
        variant="outline"
        size="sm"
        spacing={0}
        value={value === undefined ? "all" : String(value)}
        // Radix emits "" when the active item is clicked again; treat that as "keep".
        onValueChange={(v) => v && onChange(v === "all" ? undefined : Number(v))}
        aria-label="Filter by payer"
      >
        <ToggleGroupItem value="all" className="data-[state=on]:bg-primary data-[state=on]:text-primary-foreground px-3">
          <Users className="size-3.5" /> All
        </ToggleGroupItem>
        {people.map((p) => (
          <ToggleGroupItem
            key={p.id}
            value={String(p.id)}
            className="data-[state=on]:bg-primary data-[state=on]:text-primary-foreground px-3"
          >
            {p.name}
          </ToggleGroupItem>
        ))}
      </ToggleGroup>
    </div>
  )
}
