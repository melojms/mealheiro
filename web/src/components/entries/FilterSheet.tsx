import { useQuery } from "@tanstack/react-query"
import { Hash } from "lucide-react"
import { useState } from "react"
import { DATE_PRESETS } from "@/components/filters/dates"
import { useCategories, useDebouncedValue, useIsDesktop, usePeople } from "@/components/filters/hooks"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { InputGroup, InputGroupAddon, InputGroupInput } from "@/components/ui/input-group"
import { Label } from "@/components/ui/label"
import { Select, SelectContent, SelectGroup, SelectItem, SelectLabel, SelectTrigger, SelectValue } from "@/components/ui/select"
import {
  Sheet, SheetContent, SheetDescription, SheetFooter, SheetHeader, SheetTitle,
} from "@/components/ui/sheet"
import { ToggleGroup, ToggleGroupItem } from "@/components/ui/toggle-group"
import { api } from "@/lib/api"
import { parseAmount } from "@/lib/money"
import type { Category, EntryType } from "@/lib/types"
import { cn } from "@/lib/utils"
import { datePreset, type UrlFilters } from "./filters"

const TYPE_OPTIONS: { value: EntryType | "all"; label: string }[] = [
  { value: "all", label: "All" },
  { value: "expense", label: "Expense" },
  { value: "income", label: "Income" },
  { value: "investment", label: "Investment" },
]
const TYPE_TITLE: Record<EntryType, string> = { expense: "Expenses", income: "Income", investment: "Investments" }

const pill = "data-[state=on]:bg-primary data-[state=on]:text-primary-foreground px-3"

/** Top-level categories followed by their children, per type. */
function categoryTree(cats: Category[], type?: EntryType) {
  const types: EntryType[] = type ? [type] : ["expense", "income", "investment"]
  return types.map((t) => {
    const tops = cats.filter((c) => c.type === t && c.parent_id === null).sort((a, b) => a.name.localeCompare(b.name))
    return {
      type: t,
      items: tops.flatMap((top) => [
        { cat: top, child: false },
        ...cats
          .filter((c) => c.parent_id === top.id)
          .sort((a, b) => a.name.localeCompare(b.name))
          .map((cat) => ({ cat, child: true })),
      ]),
    }
  })
}

function Field({ label, htmlFor, children }: { label: string; htmlFor?: string; children: React.ReactNode }) {
  return (
    <div className="space-y-2">
      <Label htmlFor={htmlFor} className="text-muted-foreground text-xs font-medium tracking-wide uppercase">
        {label}
      </Label>
      {children}
    </div>
  )
}

/** Edits a draft copy of the filters; nothing hits the URL until "Show results". */
export function FilterSheet({
  open, onOpenChange, value, onApply,
}: { open: boolean; onOpenChange: (o: boolean) => void; value: UrlFilters; onApply: (f: UrlFilters) => void }) {
  const desktop = useIsDesktop()
  const [draft, setDraft] = useState<UrlFilters>(value)
  const { data: cats = [], byId } = useCategories()
  const { data: people = [] } = usePeople()

  // Re-seed the draft each time the sheet opens (state adjustment during render, no effect needed).
  const [wasOpen, setWasOpen] = useState(open)
  if (open !== wasOpen) {
    setWasOpen(open)
    if (open) setDraft(value)
  }

  const set = (patch: UrlFilters) => setDraft((d) => ({ ...d, ...patch }))
  const type = draft.type as EntryType | undefined
  const preset = datePreset(draft)

  const tagQuery = useDebouncedValue(draft.tag ?? "", 200)
  const tags = useQuery({ queryKey: ["tags", tagQuery], queryFn: () => api.tags(tagQuery || undefined), enabled: open })
  const tagSuggestions = (tags.data ?? []).filter((t) => t.name !== draft.tag).slice(0, 8)

  const minBad = !!draft.min && parseAmount(draft.min) === null
  const maxBad = !!draft.max && parseAmount(draft.max) === null

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent
        side={desktop ? "right" : "bottom"}
        className={cn("gap-0 p-0", desktop ? "w-full sm:max-w-md" : "max-h-[90svh] rounded-t-2xl")}
      >
        <SheetHeader className="border-b">
          <SheetTitle>Filters</SheetTitle>
          <SheetDescription>Narrow down the entry list and its totals.</SheetDescription>
        </SheetHeader>

        <div className="flex-1 space-y-6 overflow-y-auto overscroll-contain p-4">
          <Field label="Type">
            <ToggleGroup
              type="single"
              variant="outline"
              size="sm"
              spacing={0}
              value={type ?? "all"}
              onValueChange={(v) => {
                if (!v) return
                const t = v === "all" ? undefined : v
                // A category of another type would yield nothing — drop it.
                const cat = draft.cat ? byId.get(Number(draft.cat)) : undefined
                set({ type: t, cat: t && cat && cat.type !== t ? undefined : draft.cat })
              }}
              className="w-full"
            >
              {TYPE_OPTIONS.map((o) => (
                <ToggleGroupItem key={o.value} value={o.value} className={cn(pill, "flex-1")}>
                  {o.label}
                </ToggleGroupItem>
              ))}
            </ToggleGroup>
          </Field>

          <div className="grid gap-6 sm:grid-cols-2">
            <Field label="Category" htmlFor="f-cat">
              <Select value={draft.cat ?? "all"} onValueChange={(v) => set({ cat: v === "all" ? undefined : v })}>
                <SelectTrigger id="f-cat" className="w-full">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent position="popper" className="max-h-72">
                  <SelectItem value="all">All categories</SelectItem>
                  {categoryTree(cats, type).map((g) =>
                    g.items.length === 0 ? null : (
                      <SelectGroup key={g.type}>
                        <SelectLabel>{TYPE_TITLE[g.type]}</SelectLabel>
                        {g.items.map(({ cat, child }) => (
                          <SelectItem key={cat.id} value={String(cat.id)} className={child ? "pl-6" : undefined}>
                            <span className="size-2 shrink-0 rounded-full" style={{ backgroundColor: cat.color }} />
                            {cat.name}
                            {cat.archived && <span className="text-muted-foreground text-xs">(archived)</span>}
                          </SelectItem>
                        ))}
                      </SelectGroup>
                    ),
                  )}
                </SelectContent>
              </Select>
            </Field>

            <Field label="Payer" htmlFor="f-payer">
              <Select value={draft.payer ?? "all"} onValueChange={(v) => set({ payer: v === "all" ? undefined : v })}>
                <SelectTrigger id="f-payer" className="w-full">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent position="popper">
                  <SelectItem value="all">Everyone</SelectItem>
                  {people.map((p) => (
                    <SelectItem key={p.id} value={String(p.id)}>
                      {p.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </Field>
          </div>

          <Field label="Tag" htmlFor="f-tag">
            <InputGroup className="h-9">
              <InputGroupAddon>
                <Hash />
              </InputGroupAddon>
              <InputGroupInput
                id="f-tag"
                placeholder="Any tag"
                autoComplete="off"
                autoCapitalize="none"
                value={draft.tag ?? ""}
                onChange={(e) => set({ tag: e.target.value.toLowerCase() })}
              />
            </InputGroup>
            {tagSuggestions.length > 0 && (
              <div className="flex flex-wrap gap-1.5" aria-label="Tag suggestions">
                {tagSuggestions.map((t) => (
                  <button
                    key={t.name}
                    type="button"
                    onClick={() => set({ tag: t.name })}
                    className="hover:bg-muted inline-flex items-center gap-1 rounded-full border px-2.5 py-0.5 text-xs transition-colors"
                  >
                    #{t.name}
                    <span className="text-muted-foreground tabular-nums">{t.count}</span>
                  </button>
                ))}
              </div>
            )}
          </Field>

          <Field label="Amount">
            <div className="flex items-center gap-2">
              <InputGroup className="h-9">
                <InputGroupInput
                  aria-label="Minimum amount"
                  placeholder="Min"
                  inputMode="decimal"
                  value={draft.min ?? ""}
                  aria-invalid={minBad}
                  onChange={(e) => set({ min: e.target.value })}
                />
                <InputGroupAddon align="inline-end">€</InputGroupAddon>
              </InputGroup>
              <span className="text-muted-foreground">–</span>
              <InputGroup className="h-9">
                <InputGroupInput
                  aria-label="Maximum amount"
                  placeholder="Max"
                  inputMode="decimal"
                  value={draft.max ?? ""}
                  aria-invalid={maxBad}
                  onChange={(e) => set({ max: e.target.value })}
                />
                <InputGroupAddon align="inline-end">€</InputGroupAddon>
              </InputGroup>
            </div>
          </Field>

          <Field label="Date">
            <div className="flex flex-wrap gap-1.5">
              {[{ value: undefined, label: "Any time" }, ...DATE_PRESETS].map((p) => {
                const active = preset === p.value
                return (
                  <Button
                    key={p.label}
                    type="button"
                    size="sm"
                    variant={active ? "default" : "outline"}
                    aria-pressed={active}
                    onClick={() => set({ range: p.value, from: p.value === "custom" ? draft.from : undefined, to: p.value === "custom" ? draft.to : undefined })}
                  >
                    {p.label}
                  </Button>
                )
              })}
            </div>
            {preset === "custom" && (
              <div className="grid grid-cols-2 gap-2 pt-1">
                <div className="space-y-1">
                  <Label htmlFor="f-from" className="text-xs">From</Label>
                  <Input id="f-from" type="date" value={draft.from ?? ""} max={draft.to} onChange={(e) => set({ from: e.target.value })} />
                </div>
                <div className="space-y-1">
                  <Label htmlFor="f-to" className="text-xs">To</Label>
                  <Input id="f-to" type="date" value={draft.to ?? ""} min={draft.from} onChange={(e) => set({ to: e.target.value })} />
                </div>
              </div>
            )}
          </Field>
        </div>

        <SheetFooter className="flex-row border-t pb-[max(1rem,env(safe-area-inset-bottom))]">
          <Button variant="outline" size="lg" className="flex-1" onClick={() => setDraft({ q: draft.q })}>
            Reset
          </Button>
          <Button
            size="lg"
            className="flex-1"
            disabled={minBad || maxBad}
            onClick={() => {
              onApply(draft)
              onOpenChange(false)
            }}
          >
            Show results
          </Button>
        </SheetFooter>
      </SheetContent>
    </Sheet>
  )
}
