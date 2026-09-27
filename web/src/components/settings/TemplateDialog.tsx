import { useState } from "react"
import { useMutation, useQueryClient } from "@tanstack/react-query"
import { toast } from "sonner"
import { Loader2, Trash2, X } from "lucide-react"
import { CategoryIcon } from "@/components/CategoryIcon"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Select, SelectContent, SelectGroup, SelectItem, SelectLabel, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Switch } from "@/components/ui/switch"
import { Textarea } from "@/components/ui/textarea"
import { centsToInput } from "@/components/add/amount"
import { ConfirmDialog } from "@/components/entry/ConfirmDialog"
import { useCategories, usePeople } from "@/components/entry/hooks"
import { PayerChips } from "@/components/entry/PayerChips"
import { canBePersonal } from "@/components/entry/personal"
import { ResponsiveDialog } from "@/components/entry/ResponsiveDialog"
import { TypeSwitch } from "@/components/entry/TypeSwitch"
import { api } from "@/lib/api"
import { useDefaultPayer } from "@/lib/device"
import { parseAmount } from "@/lib/money"
import type { EntryType, Template, TemplateInput } from "@/lib/types"

const MONTH_RE = /^\d{4}-(0[1-9]|1[0-2])$/

/** Create (template = null) or edit a recurring template. */
export function TemplateDialog({
  template, currentMonth, open, onOpenChange,
}: { template: Template | null; currentMonth: string; open: boolean; onOpenChange: (o: boolean) => void }) {
  const queryClient = useQueryClient()
  const { data: people = [] } = usePeople()
  const { data: allCategories = [] } = useCategories(true)
  const defaultPayer = useDefaultPayer()

  const [type, setType] = useState<EntryType>(template?.type ?? "expense")
  const [categoryId, setCategoryId] = useState<number | null>(template?.category_id ?? null)
  const [payerId, setPayerId] = useState<number | null>(template?.payer_id ?? defaultPayer)
  const [amount, setAmount] = useState(template ? centsToInput(template.amount_cents) : "")
  const [variable, setVariable] = useState(template?.variable ?? false)
  const [personal, setPersonal] = useState(template?.personal ?? false)
  const [note, setNote] = useState(template?.note ?? "")
  const [startMonth, setStartMonth] = useState(template?.start_month ?? currentMonth)
  const [endMonth, setEndMonth] = useState(template?.end_month ?? "")
  const [active, setActive] = useState(template?.active ?? true)
  const [confirmDelete, setConfirmDelete] = useState(false)

  const categories = allCategories.filter((c) => c.type === type && (!c.archived || c.id === template?.category_id))
  const tops = categories.filter((c) => c.parent_id === null).sort((a, b) => a.name.localeCompare(b.name))

  const personalAllowed = canBePersonal(type, people.find((p) => p.id === payerId))
  const cents = parseAmount(amount)
  const endInvalid = endMonth !== "" && (!MONTH_RE.test(endMonth) || endMonth < startMonth)
  const errors = {
    amount: amount.trim() !== "" && (cents === null || cents <= 0),
    start: !MONTH_RE.test(startMonth),
    end: endInvalid,
  }
  const valid = cents !== null && cents > 0 && categoryId !== null && payerId !== null && !errors.start && !errors.end

  const done = (msg: string) => {
    queryClient.invalidateQueries()
    toast.success(msg)
    onOpenChange(false)
  }
  const save = useMutation({
    mutationFn: () => {
      const input: TemplateInput = {
        type, category_id: categoryId!, payer_id: payerId!, amount_cents: cents!, variable,
        personal: personalAllowed && personal, note: note.trim(),
        start_month: startMonth, end_month: endMonth || null, active,
      }
      return template ? api.updateTemplate(template.id, input) : api.createTemplate(input)
    },
    onSuccess: () => done(template ? "Template saved" : "Template created"),
    onError: (e) => toast.error(e.message),
  })
  const remove = useMutation({
    mutationFn: () => api.deleteTemplate(template!.id),
    onSuccess: () => done("Template deleted"),
    onError: (e) => toast.error(e.message),
  })
  const busy = save.isPending || remove.isPending

  return (
    <>
      <ResponsiveDialog
        open={open}
        onOpenChange={onOpenChange}
        title={template ? "Edit recurring" : "New recurring"}
        description="Creates an entry on the 1st of every month. Changes apply to future months only."
        footer={
          <>
            {template ? (
              <Button variant="destructive" onClick={() => setConfirmDelete(true)} disabled={busy} className="max-sm:order-last">
                <Trash2 /> Delete
              </Button>
            ) : (
              <span className="max-sm:hidden" />
            )}
            <Button disabled={!valid || busy} onClick={() => save.mutate()}>
              {save.isPending && <Loader2 className="animate-spin" />}
              {template ? "Save" : "Create"}
            </Button>
          </>
        }
      >
        <form
          className="space-y-5"
          onSubmit={(e) => {
            e.preventDefault()
            if (valid && !busy) save.mutate()
          }}
        >
          <TypeSwitch
            value={type}
            onChange={(t) => {
              setType(t)
              setCategoryId(t === template?.type ? template.category_id : null)
            }}
          />

          <div className="space-y-1.5">
            <Label htmlFor="tpl-category">Category</Label>
            <Select value={categoryId === null ? "" : String(categoryId)} onValueChange={(v) => setCategoryId(Number(v))}>
              <SelectTrigger id="tpl-category" className="h-9 w-full">
                <SelectValue placeholder="Choose a category" />
              </SelectTrigger>
              <SelectContent position="popper" className="max-h-72">
                {tops.map((top) => (
                  <SelectGroup key={top.id}>
                    {categories.some((c) => c.parent_id === top.id) && <SelectLabel className="sr-only">{top.name}</SelectLabel>}
                    <SelectItem value={String(top.id)}>
                      <CategoryIcon icon={top.icon} color={top.color} size="sm" className="size-6" />
                      {top.name}
                    </SelectItem>
                    {categories
                      .filter((c) => c.parent_id === top.id)
                      .sort((a, b) => a.name.localeCompare(b.name))
                      .map((c) => (
                        <SelectItem key={c.id} value={String(c.id)} className="pl-10">
                          {c.name}
                        </SelectItem>
                      ))}
                  </SelectGroup>
                ))}
              </SelectContent>
            </Select>
          </div>

          <div className="grid gap-5 sm:grid-cols-2">
            <div className="space-y-1.5">
              <Label htmlFor="tpl-amount">{variable ? "Estimated amount" : "Amount"}</Label>
              <div className="relative">
                <Input
                  id="tpl-amount"
                  inputMode="decimal"
                  autoComplete="off"
                  placeholder="0,00"
                  value={amount}
                  aria-invalid={errors.amount}
                  onChange={(e) => setAmount(e.target.value)}
                  className="h-9 pr-8 tabular-nums"
                />
                <span className="text-muted-foreground pointer-events-none absolute top-1/2 right-3 -translate-y-1/2 text-sm">€</span>
              </div>
            </div>
            <div className="space-y-1.5">
              <Label>Paid by</Label>
              <PayerChips people={people} value={payerId} onChange={setPayerId} />
            </div>
          </div>

          <div className="bg-muted/40 flex items-start gap-3 rounded-xl border p-3">
            <Switch id="tpl-variable" checked={variable} onCheckedChange={setVariable} className="mt-0.5" />
            <div className="space-y-0.5">
              <Label htmlFor="tpl-variable">Variable amount</Label>
              <p className="text-muted-foreground text-xs">
                {variable
                  ? "Each month's entry is created as pending, prefilled with the last amount, and waits for you to confirm it (e.g. electricity, salary)."
                  : "Each month's entry is created confirmed with exactly this amount (e.g. mortgage, subscriptions)."}
              </p>
            </div>
          </div>

          {personalAllowed && (
            <div className="bg-muted/40 flex items-start gap-3 rounded-xl border p-3">
              <Switch id="tpl-personal" checked={personal} onCheckedChange={setPersonal} className="mt-0.5" />
              <div className="space-y-0.5">
                <Label htmlFor="tpl-personal">Personal</Label>
                <p className="text-muted-foreground text-xs">
                  {personal
                    ? "Generated entries are personal and left out of the shared split (e.g. a gym membership)."
                    : "Generated entries count as shared expenses (e.g. mortgage, electricity)."}
                </p>
              </div>
            </div>
          )}

          <div className="grid grid-cols-2 gap-3">
            <div className="space-y-1.5">
              <Label htmlFor="tpl-start">Start month</Label>
              <Input
                id="tpl-start"
                type="month"
                value={startMonth}
                aria-invalid={errors.start}
                onChange={(e) => setStartMonth(e.target.value)}
                className="h-9"
              />
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="tpl-end">End month</Label>
              <div className="relative">
                <Input
                  id="tpl-end"
                  type="month"
                  value={endMonth}
                  min={startMonth}
                  aria-invalid={errors.end}
                  onChange={(e) => setEndMonth(e.target.value)}
                  className="h-9"
                />
                {endMonth && (
                  <button
                    type="button"
                    aria-label="Clear end month"
                    onClick={() => setEndMonth("")}
                    className="text-muted-foreground hover:bg-muted absolute top-1/2 right-1.5 -translate-y-1/2 rounded p-1"
                  >
                    <X className="size-3.5" />
                  </button>
                )}
              </div>
            </div>
            <p className="text-muted-foreground col-span-2 -mt-1 text-xs">
              {errors.end ? <span className="text-destructive">End month must be after the start month.</span> : "Leave the end month empty to repeat indefinitely."}
            </p>
          </div>

          <div className="space-y-1.5">
            <Label htmlFor="tpl-note">Note</Label>
            <Textarea id="tpl-note" rows={2} value={note} onChange={(e) => setNote(e.target.value)} placeholder="e.g. Netflix" />
          </div>

          <div className="flex items-center justify-between gap-3">
            <div>
              <Label htmlFor="tpl-active">Active</Label>
              <p className="text-muted-foreground text-xs">Paused templates stop generating; past entries are kept.</p>
            </div>
            <Switch id="tpl-active" checked={active} onCheckedChange={setActive} />
          </div>
          <button type="submit" hidden />
        </form>
      </ResponsiveDialog>

      <ConfirmDialog
        open={confirmDelete}
        onOpenChange={setConfirmDelete}
        title="Delete this recurring template?"
        description="Entries it already generated are kept. No new ones will be created."
        onConfirm={() => remove.mutate()}
      />
    </>
  )
}
