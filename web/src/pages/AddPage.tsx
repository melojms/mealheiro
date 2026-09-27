import { useEffect, useEffectEvent, useState } from "react"
import { useMutation, useQueryClient } from "@tanstack/react-query"
import { toast } from "sonner"
import { ChevronDown, Loader2 } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Label } from "@/components/ui/label"
import { Skeleton } from "@/components/ui/skeleton"
import { Textarea } from "@/components/ui/textarea"
import { AmountDisplay } from "@/components/add/AmountDisplay"
import { amountCents, keyFromKeyboard, pressKey, type KeypadKey } from "@/components/add/amount"
import { budgetSummary } from "@/components/add/budget"
import { DevicePayerPrompt } from "@/components/add/DevicePayerPrompt"
import { Keypad } from "@/components/add/Keypad"
import { PendingBanner } from "@/components/add/PendingBanner"
import { RecentEntries } from "@/components/add/RecentEntries"
import { CategoryPicker } from "@/components/entry/CategoryPicker"
import { DateChips } from "@/components/entry/DateChips"
import { useCategories, usePeople, useToday } from "@/components/entry/hooks"
import { PayerChips } from "@/components/entry/PayerChips"
import { PersonalChip } from "@/components/entry/PersonalChip"
import { canBePersonal } from "@/components/entry/personal"
import { TagInput } from "@/components/entry/TagInput"
import { TypeSwitch } from "@/components/entry/TypeSwitch"
import { api } from "@/lib/api"
import { useDefaultPayer } from "@/lib/device"
import { formatCents } from "@/lib/money"
import type { Entry, EntryInput, EntryType } from "@/lib/types"
import { cn } from "@/lib/utils"

export default function AddPage() {
  const queryClient = useQueryClient()
  const today = useToday()
  const defaultPayer = useDefaultPayer()
  const people = usePeople()
  const categories = useCategories()

  const [type, setType] = useState<EntryType>("expense")
  const [amount, setAmount] = useState("")
  const [categoryId, setCategoryId] = useState<number | null>(null)
  const [pickedDate, setPickedDate] = useState<string | null>(null) // null = follow "today"
  const [pickedPayer, setPickedPayer] = useState<number | null>(null) // null = device default
  const [moreOpen, setMoreOpen] = useState(false)
  const [note, setNote] = useState("")
  const [tags, setTags] = useState<string[]>([])
  const [personal, setPersonal] = useState(false) // resets after every save: expenses are shared by default

  const date = pickedDate ?? today
  const peopleList = people.data ?? []
  const knownDefault = peopleList.some((p) => p.id === defaultPayer) ? defaultPayer : null
  const payerId = pickedPayer ?? knownDefault
  const personalAllowed = canBePersonal(type, peopleList.find((p) => p.id === payerId))
  const typeCategories = (categories.data ?? []).filter((c) => c.type === type)
  const cents = amountCents(amount)

  const create = useMutation({
    mutationFn: (input: EntryInput) => api.createEntry(input),
    onSuccess: async (entry) => {
      queryClient.invalidateQueries()
      setAmount("")
      setCategoryId(null)
      setNote("")
      setTags([])
      setPersonal(false)
      setMoreOpen(false)
      await announce(entry)
    },
    onError: (e) => toast.error(`Couldn't save: ${e.message}`),
  })

  const missing = !cents ? "Enter an amount" : categoryId === null ? "Pick a category" : payerId === null ? "Choose who paid" : null

  const save = () => {
    if (missing || create.isPending) return
    create.mutate({ type, date, amount_cents: cents!, category_id: categoryId!, payer_id: payerId!, personal: personalAllowed && personal, note: note.trim(), tags })
  }

  const onKey = (k: KeypadKey) => setAmount((a) => pressKey(a, k))

  // Hardware keyboard: digits/comma/backspace edit the amount, Enter saves.
  const onKeyDown = useEffectEvent((e: KeyboardEvent) => {
    if (e.ctrlKey || e.metaKey || e.altKey) return
    const target = e.target as HTMLElement
    if (target.closest("input, textarea, select, [contenteditable=true], [role=dialog], [role=alertdialog]")) return
    if (e.key === "Enter") {
      if (target === document.body || target.closest("[role=radio], [aria-pressed]")) {
        e.preventDefault()
        save()
      }
      return
    }
    const k = keyFromKeyboard(e.key)
    if (!k) return
    e.preventDefault()
    onKey(k)
  })
  useEffect(() => {
    const handler = (e: KeyboardEvent) => onKeyDown(e)
    window.addEventListener("keydown", handler)
    return () => window.removeEventListener("keydown", handler)
  }, [])

  const changeType = (t: EntryType) => {
    setType(t)
    setCategoryId(null)
  }

  return (
    <div className="mx-auto max-w-5xl">
      <PendingBanner />

      <div className="grid grid-cols-1 gap-5 md:grid-cols-[minmax(0,22rem)_minmax(0,1fr)] md:gap-10">
        <section aria-label="Amount" className="space-y-3 md:sticky md:top-8 md:self-start">
          <TypeSwitch value={type} onChange={changeType} />
          <AmountDisplay amount={amount} type={type} onClear={() => setAmount("")} />
          <Keypad onKey={onKey} />
        </section>

        <section aria-label="Details" className="space-y-5">
          <div className="space-y-2">
            <h2 className="text-muted-foreground text-xs font-medium tracking-wide uppercase">Category</h2>
            <CategoryPicker
              categories={typeCategories}
              value={categoryId}
              onChange={setCategoryId}
              loading={categories.isLoading}
            />
            {categories.isError && <p className="text-destructive text-sm">Couldn't load categories.</p>}
          </div>

          <div className="space-y-2">
            <h2 className="text-muted-foreground text-xs font-medium tracking-wide uppercase">Date</h2>
            <DateChips value={date} today={today} onChange={setPickedDate} />
          </div>

          <div className="space-y-2">
            <h2 className="text-muted-foreground text-xs font-medium tracking-wide uppercase">Paid by</h2>
            {people.isLoading ? (
              <div className="flex gap-2">
                <Skeleton className="h-8 w-20 rounded-full" />
                <Skeleton className="h-8 w-20 rounded-full" />
                <Skeleton className="h-8 w-20 rounded-full" />
              </div>
            ) : knownDefault === null && pickedPayer === null ? (
              <DevicePayerPrompt people={peopleList} />
            ) : (
              <div className="flex flex-wrap items-center gap-2">
                <PayerChips people={peopleList} value={payerId} onChange={setPickedPayer} />
                {personalAllowed && <PersonalChip value={personal} onChange={setPersonal} />}
              </div>
            )}
          </div>

          <div className="rounded-xl border">
            <button
              type="button"
              aria-expanded={moreOpen}
              aria-controls="add-more"
              onClick={() => setMoreOpen((o) => !o)}
              className="hover:bg-muted/50 flex w-full items-center gap-2 rounded-xl px-3 py-2.5 text-sm transition-colors"
            >
              <span className="font-medium">More</span>
              <span className="text-muted-foreground truncate text-xs">
                {[note.trim() && "note", tags.length > 0 && `${tags.length} tag${tags.length > 1 ? "s" : ""}`].filter(Boolean).join(" · ") ||
                  "Note, tags"}
              </span>
              <ChevronDown className={cn("text-muted-foreground ml-auto size-4 transition-transform", moreOpen && "rotate-180")} />
            </button>
            {moreOpen && (
              <div id="add-more" className="animate-in fade-in slide-in-from-top-1 space-y-4 px-3 pt-1 pb-3">
                <div className="space-y-1.5">
                  <Label htmlFor="add-note">Note</Label>
                  <Textarea id="add-note" rows={2} value={note} onChange={(e) => setNote(e.target.value)} placeholder="e.g. Pingo Doce" />
                </div>
                <div className="space-y-1.5">
                  <Label htmlFor="add-tags">Tags</Label>
                  <TagInput id="add-tags" value={tags} onChange={setTags} />
                </div>
              </div>
            )}
          </div>

          {/* Pinned above the phone tab bar while the form is on screen. */}
          <div className="bg-background/85 sticky bottom-[calc(4rem+env(safe-area-inset-bottom))] z-10 -mx-4 px-4 py-2 backdrop-blur md:static md:mx-0 md:bg-transparent md:p-0 md:backdrop-blur-none">
            <Button size="lg" className="h-12 w-full rounded-xl text-base" disabled={!!missing || create.isPending} onClick={save}>
              {create.isPending && <Loader2 className="animate-spin" />}
              {missing ?? `Save ${formatCents(cents!)}`}
            </Button>
          </div>
        </section>
      </div>

      <RecentEntries />
    </div>
  )
}

/** Success toast; for expenses includes budget usage (warning ≥ 80%, error ≥ 100%). */
async function announce(entry: Entry) {
  const title = `Saved ${formatCents(entry.amount_cents)} · ${entry.category_name}`
  let summary = null
  if (entry.type === "expense") {
    try {
      summary = budgetSummary(await api.budgetStatus(entry.category_id))
    } catch {
      // Budget info is a nicety; the entry is saved regardless.
    }
  }
  if (!summary) {
    toast.success(title)
    return
  }
  const description = (
    <div className="space-y-0.5">
      {summary.lines.map((l) => (
        <div key={l} className="tabular-nums">{l}</div>
      ))}
    </div>
  )
  const show = { ok: toast.success, warning: toast.warning, over: toast.error }[summary.level]
  show(title, { description, duration: summary.level === "ok" ? 4000 : 7000 })
}
