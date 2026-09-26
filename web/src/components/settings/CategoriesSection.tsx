import { useState } from "react"
import { useMutation, useQueryClient } from "@tanstack/react-query"
import { toast } from "sonner"
import { Archive, ArchiveRestore, MoreHorizontal, Pencil, Plus, Trash2 } from "lucide-react"
import { CategoryIcon } from "@/components/CategoryIcon"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import {
  DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuSeparator, DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { Label } from "@/components/ui/label"
import { Skeleton } from "@/components/ui/skeleton"
import { Switch } from "@/components/ui/switch"
import { ConfirmDialog } from "@/components/entry/ConfirmDialog"
import { useCategories } from "@/components/entry/hooks"
import { TypeSwitch } from "@/components/entry/TypeSwitch"
import { api } from "@/lib/api"
import type { Category, EntryType } from "@/lib/types"
import { cn } from "@/lib/utils"
import { CategoryDialog, type CategoryDialogMode } from "./CategoryDialog"
import { useDialog } from "./useDialog"

const byName = (a: Category, b: Category) => a.name.localeCompare(b.name)

export function CategoriesSection() {
  const queryClient = useQueryClient()
  const [type, setType] = useState<EntryType>("expense")
  const [showArchived, setShowArchived] = useState(false)
  const { data, isLoading } = useCategories(true)
  const dialog = useDialog<CategoryDialogMode>()
  const [deleting, setDeleting] = useState<Category | null>(null)

  const ofType = (data ?? []).filter((c) => c.type === type)
  const visible = ofType.filter((c) => showArchived || !c.archived)
  const tops = visible.filter((c) => c.parent_id === null).sort(byName)
  const archivedCount = ofType.filter((c) => c.archived).length

  const setArchived = useMutation({
    mutationFn: ({ c, archived }: { c: Category; archived: boolean }) => api.updateCategory(c.id, { archived }),
    onSuccess: (c, { archived }) => {
      queryClient.invalidateQueries()
      toast.success(archived ? `Archived ${c.name}` : `Restored ${c.name}`, {
        description: archived ? "Hidden from quick-add; history is kept." : undefined,
        action: { label: "Undo", onClick: () => setArchived.mutate({ c, archived: !archived }) },
      })
    },
    onError: (e) => toast.error(e.message),
  })

  const remove = useMutation({
    mutationFn: (c: Category) => api.deleteCategory(c.id),
    onSuccess: (_, c) => {
      queryClient.invalidateQueries()
      toast.success(`Deleted ${c.name}`)
    },
    // 409: still referenced; the server message suggests archiving, so offer it right there.
    onError: (e, c) =>
      toast.error(e.message, {
        duration: 8000,
        action: c.archived ? undefined : { label: "Archive", onClick: () => setArchived.mutate({ c, archived: true }) },
      }),
  })

  const actions = {
    edit: (c: Category) => dialog.show({ kind: "edit", category: c }),
    addChild: (c: Category) => dialog.show({ kind: "create", type, parent: c }),
    archive: (c: Category) => setArchived.mutate({ c, archived: !c.archived }),
    remove: (c: Category) => setDeleting(c),
  }

  return (
    <div className="space-y-4">
      <TypeSwitch value={type} onChange={setType} className="max-w-md" />

      <div className="flex flex-wrap items-center justify-between gap-3">
        <div className="flex items-center gap-2">
          <Switch id="show-archived" checked={showArchived} onCheckedChange={setShowArchived} />
          <Label htmlFor="show-archived" className="font-normal">
            Show archived {archivedCount > 0 && <span className="text-muted-foreground">({archivedCount})</span>}
          </Label>
        </div>
        <Button onClick={() => dialog.show({ kind: "create", type, parent: null })}>
          <Plus /> New category
        </Button>
      </div>

      <div className="bg-card divide-y rounded-xl border">
        {isLoading ? (
          Array.from({ length: 6 }, (_, i) => (
            <div key={i} className="flex items-center gap-3 p-3">
              <Skeleton className="size-9 rounded-full" />
              <Skeleton className="h-4 w-40" />
            </div>
          ))
        ) : tops.length === 0 ? (
          <p className="text-muted-foreground p-8 text-center text-sm">No categories of this type yet.</p>
        ) : (
          tops.map((top) => (
            <div key={top.id} className="py-1">
              <CategoryRow
                category={top}
                actions={actions}
                // Parents show their subtree's usage; entry_count alone is self-only.
                count={ofType.reduce((n, c) => (c.parent_id === top.id ? n + c.entry_count : n), top.entry_count)}
              />
              {visible
                .filter((c) => c.parent_id === top.id)
                .sort(byName)
                .map((child) => (
                  <CategoryRow key={child.id} category={child} actions={actions} child />
                ))}
            </div>
          ))
        )}
      </div>

      {dialog.payload && (
        <CategoryDialog
          key={dialog.key}
          mode={dialog.payload}
          siblings={ofType.filter((c) => c.parent_id === null)}
          open={dialog.open}
          onOpenChange={dialog.setOpen}
        />
      )}

      <ConfirmDialog
        open={deleting !== null}
        onOpenChange={(o) => !o && setDeleting(null)}
        title={`Delete ${deleting?.name ?? "category"}?`}
        description="Only categories that were never used can be deleted. Otherwise, archive it to hide it from quick-add while keeping history."
        onConfirm={() => deleting && remove.mutate(deleting)}
      />
    </div>
  )
}

interface RowActions {
  edit: (c: Category) => void
  addChild: (c: Category) => void
  archive: (c: Category) => void
  remove: (c: Category) => void
}

function CategoryRow({
  category: c, actions, child, count = c.entry_count,
}: { category: Category; actions: RowActions; child?: boolean; count?: number }) {
  return (
    <div className={cn("flex items-center gap-3 py-1.5 pr-2", child ? "pl-8 sm:pl-12" : "pl-3", c.archived && "opacity-60")}>
      {child && <span className="bg-border -ml-4 h-px w-3 shrink-0" aria-hidden />}
      <CategoryIcon icon={c.icon} color={c.color} size={child ? "sm" : "md"} />
      <button type="button" onClick={() => actions.edit(c)} className="min-w-0 flex-1 text-left">
        <span className={cn("block truncate", child ? "text-sm" : "text-sm font-medium")}>{c.name}</span>
        <span className="text-muted-foreground block text-xs">
          {count} {count === 1 ? "entry" : "entries"}
        </span>
      </button>
      {c.archived && <Badge variant="secondary">Archived</Badge>}
      <DropdownMenu modal={false}>
        <DropdownMenuTrigger asChild>
          <Button variant="ghost" size="icon" aria-label={`Actions for ${c.name}`}>
            <MoreHorizontal />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end">
          <DropdownMenuItem onSelect={() => actions.edit(c)}>
            <Pencil /> Edit
          </DropdownMenuItem>
          {!child && !c.archived && (
            <DropdownMenuItem onSelect={() => actions.addChild(c)}>
              <Plus /> Add subcategory
            </DropdownMenuItem>
          )}
          <DropdownMenuItem onSelect={() => actions.archive(c)}>
            {c.archived ? <ArchiveRestore /> : <Archive />}
            {c.archived ? "Unarchive" : "Archive"}
          </DropdownMenuItem>
          <DropdownMenuSeparator />
          <DropdownMenuItem variant="destructive" onSelect={() => actions.remove(c)}>
            <Trash2 /> Delete
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  )
}
