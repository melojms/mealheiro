import { useState } from "react"
import { useMutation, useQueryClient } from "@tanstack/react-query"
import { toast } from "sonner"
import { Check, Loader2 } from "lucide-react"
import { CategoryIcon } from "@/components/CategoryIcon"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { ResponsiveDialog } from "@/components/entry/ResponsiveDialog"
import { api } from "@/lib/api"
import { CATEGORY_ICONS } from "@/lib/icons"
import { PALETTE } from "@/lib/palette"
import type { Category, EntryType } from "@/lib/types"
import { cn } from "@/lib/utils"

export type CategoryDialogMode =
  | { kind: "create"; type: EntryType; parent: Category | null }
  | { kind: "edit"; category: Category }

/** Create or edit a category: name, icon, color. */
export function CategoryDialog({
  mode, siblings, open, onOpenChange,
}: {
  mode: CategoryDialogMode
  /** Top-level categories of the same type, used to pick an unused default color. */
  siblings: Category[]
  open: boolean
  onOpenChange: (open: boolean) => void
}) {
  const queryClient = useQueryClient()
  const initial = initialValues(mode, siblings)
  const [name, setName] = useState(initial.name)
  const [icon, setIcon] = useState(initial.icon)
  const [color, setColor] = useState(initial.color)
  const trimmed = name.trim()

  const close = () => onOpenChange(false)
  const save = useMutation({
    mutationFn: () =>
      mode.kind === "create"
        ? api.createCategory({ type: mode.type, name: trimmed, parent_id: mode.parent?.id ?? null, icon, color })
        : api.updateCategory(mode.category.id, { name: trimmed, icon, color }),
    onSuccess: (c) => {
      queryClient.invalidateQueries()
      toast.success(mode.kind === "create" ? `Created ${c.name}` : `Saved ${c.name}`)
      close()
    },
    onError: (e) => toast.error(e.message),
  })

  const title =
    mode.kind === "edit" ? "Edit category" : mode.parent ? `New subcategory of ${mode.parent.name}` : "New category"

  return (
    <ResponsiveDialog
      open={open}
      onOpenChange={onOpenChange}
      title={title}
      footer={
        <>
          <Button variant="outline" onClick={close}>Cancel</Button>
          <Button disabled={!trimmed || save.isPending} onClick={() => save.mutate()}>
            {save.isPending && <Loader2 className="animate-spin" />}
            {mode.kind === "create" ? "Create" : "Save"}
          </Button>
        </>
      }
    >
      <form
        className="space-y-5"
        onSubmit={(e) => {
          e.preventDefault()
          if (trimmed && !save.isPending) save.mutate()
        }}
      >
        <div className="flex items-center gap-3">
          <CategoryIcon icon={icon} color={color} size="lg" />
          <div className="flex-1 space-y-1.5">
            <Label htmlFor="category-name">Name</Label>
            <Input id="category-name" value={name} onChange={(e) => setName(e.target.value)} placeholder="e.g. Pets" autoFocus />
          </div>
        </div>

        <fieldset className="space-y-2">
          <legend className="text-sm font-medium">Color</legend>
          <div className="grid grid-cols-10 gap-1.5">
            {PALETTE.map((c) => (
              <button
                key={c}
                type="button"
                aria-label={`Color ${c}`}
                aria-pressed={color === c}
                onClick={() => setColor(c)}
                className="focus-visible:ring-ring/50 flex aspect-square items-center justify-center rounded-full outline-none focus-visible:ring-3"
                style={{ backgroundColor: c }}
              >
                {color === c && <Check className="size-3.5 text-white drop-shadow" />}
              </button>
            ))}
          </div>
        </fieldset>

        <fieldset className="space-y-2">
          <legend className="text-sm font-medium">Icon</legend>
          <div className="grid grid-cols-7 gap-1.5 sm:grid-cols-8">
            {Object.entries(CATEGORY_ICONS).map(([key, Icon]) => (
              <button
                key={key}
                type="button"
                aria-label={`Icon ${key}`}
                aria-pressed={icon === key}
                onClick={() => setIcon(key)}
                className={cn(
                  "focus-visible:ring-ring/50 hover:bg-muted flex aspect-square items-center justify-center rounded-lg transition-colors outline-none focus-visible:ring-3",
                  icon !== key && "text-muted-foreground",
                )}
                style={icon === key ? { backgroundColor: `${color}22`, color, boxShadow: `inset 0 0 0 2px ${color}` } : undefined}
              >
                <Icon className="size-5" />
              </button>
            ))}
          </div>
        </fieldset>
        <button type="submit" hidden />
      </form>
    </ResponsiveDialog>
  )
}

function initialValues(mode: CategoryDialogMode, siblings: Category[]) {
  if (mode.kind === "edit") return { name: mode.category.name, icon: mode.category.icon, color: mode.category.color }
  if (mode.parent) return { name: "", icon: mode.parent.icon, color: mode.parent.color }
  // Same default as the server: first palette color not used by a top-level category of this type.
  const used = new Set(siblings.map((c) => c.color.toLowerCase()))
  return { name: "", icon: "circle", color: PALETTE.find((c) => !used.has(c)) ?? PALETTE[0] }
}
