// STUB — implemented by fe-input. Props are the contract other screens rely on.
import type { Entry } from "@/lib/types"

export interface EntryEditDialogProps {
  entry: Entry | null
  open: boolean
  onOpenChange: (open: boolean) => void
}

/** Edit / delete an existing entry. On success invalidates all queries. */
export function EntryEditDialog(_props: EntryEditDialogProps) {
  return null
}
