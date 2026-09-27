import { useState } from "react"
import { useMutation, useQueryClient } from "@tanstack/react-query"
import { toast } from "sonner"
import { CircleCheck, Loader2, Trash2 } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Textarea } from "@/components/ui/textarea"
import { centsToInput } from "@/components/add/amount"
import { api } from "@/lib/api"
import { formatCents, parseAmount } from "@/lib/money"
import type { Entry, EntryInput, EntryType } from "@/lib/types"
import { CategoryPicker } from "./CategoryPicker"
import { ConfirmDialog } from "./ConfirmDialog"
import { DateChips } from "./DateChips"
import { useCategories, usePeople, useToday } from "./hooks"
import { PayerChips } from "./PayerChips"
import { PersonalChip } from "./PersonalChip"
import { canBePersonal } from "./personal"
import { ResponsiveDialog } from "./ResponsiveDialog"
import { TagInput } from "./TagInput"
import { TypeSwitch } from "./TypeSwitch"

export interface EntryEditDialogProps {
  entry: Entry | null
  open: boolean
  onOpenChange: (open: boolean) => void
}

/** Edit / delete an existing entry. On success invalidates all queries. */
export function EntryEditDialog({ entry, open, onOpenChange }: EntryEditDialogProps) {
  // Keep showing the last entry while the close animation runs, and start from fresh form state on every open.
  const [shown, setShown] = useState(entry)
  const [wasOpen, setWasOpen] = useState(open)
  const [session, setSession] = useState(0)
  if (entry && entry !== shown) setShown(entry)
  if (open !== wasOpen) {
    setWasOpen(open)
    if (open) setSession((s) => s + 1)
  }
  if (!shown) return null
  return <EditForm key={`${shown.id}-${session}`} entry={shown} open={open} onOpenChange={onOpenChange} />
}

function EditForm({ entry, open, onOpenChange }: { entry: Entry } & Omit<EntryEditDialogProps, "entry">) {
  const queryClient = useQueryClient()
  const today = useToday()
  const { data: people = [] } = usePeople()
  const { data: allCategories, isLoading: categoriesLoading } = useCategories(true)

  const [type, setType] = useState<EntryType>(entry.type)
  const [amount, setAmount] = useState(centsToInput(entry.amount_cents))
  const [categoryId, setCategoryId] = useState<number | null>(entry.category_id)
  const [date, setDate] = useState(entry.date)
  const [payerId, setPayerId] = useState(entry.payer_id)
  const [personal, setPersonal] = useState(entry.personal)
  const [note, setNote] = useState(entry.note)
  const [tags, setTags] = useState(entry.tags)
  const [confirmDelete, setConfirmDelete] = useState(false)

  // Archived categories are only offered when they're the entry's current one (or its parent).
  const keep = new Set([entry.category_id, entry.parent_category_id])
  const categories = (allCategories ?? []).filter((c) => c.type === type && (!c.archived || keep.has(c.id)))

  const cents = parseAmount(amount)
  const amountInvalid = amount.trim() !== "" && (cents === null || cents <= 0)
  const valid = cents !== null && cents > 0 && categoryId !== null
  const pending = entry.status === "pending"
  const personalAllowed = canBePersonal(type, people.find((p) => p.id === payerId))

  const input = (): EntryInput => ({
    type, date, amount_cents: cents!, category_id: categoryId!, payer_id: payerId,
    personal: personalAllowed && personal, note: note.trim(), tags,
  })

  const done = (message: string) => {
    queryClient.invalidateQueries()
    toast.success(message)
    onOpenChange(false)
  }
  const fail = (e: Error) => toast.error(e.message)

  const save = useMutation({
    mutationFn: async (confirm: boolean) => {
      const updated = await api.updateEntry(entry.id, input())
      return confirm ? api.confirmEntry(entry.id, updated.amount_cents) : updated
    },
    onSuccess: (e, confirm) => done(confirm ? `Confirmed ${formatCents(e.amount_cents)}` : "Entry updated"),
    onError: fail,
  })
  const remove = useMutation({
    mutationFn: () => api.deleteEntry(entry.id),
    onSuccess: () => done("Entry deleted"),
    onError: fail,
  })
  const busy = save.isPending || remove.isPending

  const changeType = (t: EntryType) => {
    setType(t)
    setCategoryId(t === entry.type ? entry.category_id : null)
  }

  return (
    <>
      <ResponsiveDialog
        open={open}
        onOpenChange={onOpenChange}
        title={pending ? "Confirm entry" : "Edit entry"}
        description={pending ? "This is an estimate from a recurring template. Check the amount, then confirm." : undefined}
        footer={
          <>
            <Button variant="destructive" onClick={() => setConfirmDelete(true)} disabled={busy} className="max-sm:order-last">
              <Trash2 /> Delete
            </Button>
            <div className="flex gap-2 max-sm:flex-col">
              {pending && (
                <Button variant="outline" disabled={!valid || busy} onClick={() => save.mutate(false)}>
                  Save as pending
                </Button>
              )}
              <Button disabled={!valid || busy} onClick={() => save.mutate(pending)}>
                {save.isPending ? <Loader2 className="animate-spin" /> : pending && <CircleCheck />}
                {pending ? "Confirm" : "Save changes"}
              </Button>
            </div>
          </>
        }
      >
        <form
          className="space-y-5"
          onSubmit={(e) => {
            e.preventDefault()
            if (valid && !busy) save.mutate(pending)
          }}
        >
          <TypeSwitch value={type} onChange={changeType} />

          <div className="space-y-1.5">
            <Label htmlFor="entry-amount">Amount</Label>
            <div className="relative">
              <Input
                id="entry-amount"
                inputMode="decimal"
                autoComplete="off"
                value={amount}
                aria-invalid={amountInvalid}
                onChange={(e) => setAmount(e.target.value)}
                className="h-12 pr-10 text-2xl font-semibold tabular-nums md:text-2xl"
              />
              <span className="text-muted-foreground pointer-events-none absolute top-1/2 right-3 -translate-y-1/2 text-lg">€</span>
            </div>
            {amountInvalid && <p className="text-destructive text-xs">Enter an amount greater than zero, e.g. 12,50</p>}
          </div>

          <div className="space-y-1.5">
            <Label>Category</Label>
            <CategoryPicker categories={categories} value={categoryId} onChange={setCategoryId} size="sm" loading={categoriesLoading} />
            {categoryId === null && !categoriesLoading && <p className="text-muted-foreground text-xs">Pick a category.</p>}
          </div>

          <div className="space-y-1.5">
            <Label>Date</Label>
            <DateChips value={date} today={today} onChange={setDate} />
          </div>

          <div className="space-y-1.5">
            <Label>Paid by</Label>
            <div className="flex flex-wrap items-center gap-2">
              <PayerChips people={people} value={payerId} onChange={setPayerId} />
              {personalAllowed && <PersonalChip value={personal} onChange={setPersonal} />}
            </div>
          </div>

          <div className="space-y-1.5">
            <Label htmlFor="entry-note">Note</Label>
            <Textarea id="entry-note" rows={2} value={note} onChange={(e) => setNote(e.target.value)} placeholder="Optional" />
          </div>

          <div className="space-y-1.5">
            <Label htmlFor="entry-tags">Tags</Label>
            <TagInput id="entry-tags" value={tags} onChange={setTags} />
          </div>

          {entry.recurring && (
            <p className="text-muted-foreground text-xs">Generated from a recurring template. Editing it here only changes this entry.</p>
          )}
          {/* lets Enter submit from text inputs */}
          <button type="submit" hidden />
        </form>
      </ResponsiveDialog>

      <ConfirmDialog
        open={confirmDelete}
        onOpenChange={setConfirmDelete}
        title="Delete this entry?"
        description={`${entry.category_name} · ${formatCents(entry.amount_cents)} on ${entry.date}. This can't be undone.`}
        onConfirm={() => remove.mutate()}
      />
    </>
  )
}
